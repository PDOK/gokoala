package engine

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/failsafe-go/failsafe-go/failsafehttp"
	"golang.org/x/sync/errgroup"
)

const bufferSize = 1 * 1024 * 1024 // 1MiB

// Part piece of the file to download when HTTP Range Requests are supported
type Part struct {
	Start int64
	End   int64
	Size  int64
}

// Download downloads file from the given URL and stores the result in the given output location.
// Will utilize multiple concurrent connections to increase transfer speed. The latter is only
// possible when the remote server supports HTTP Range Requests, otherwise it falls back
// to a regular/single connection download. Additionally, failed requests will be retried according
// to the given settings.
func Download(url url.URL, outputFilepath string, parallelism int, tlsSkipVerify bool, timeout time.Duration,
	retryDelay time.Duration, retryMaxDelay time.Duration, maxRetries int) (*time.Duration, error) {

	client := createHTTPClient(tlsSkipVerify, timeout, retryDelay, retryMaxDelay, maxRetries)
	outputFile, err := os.OpenFile(outputFilepath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer outputFile.Close()

	start := time.Now()

	supportRanges, contentLength, err := checkRemoteFile(url, client)
	if err != nil {
		return nil, err
	}
	if supportRanges && parallelism > 1 {
		err = downloadWithMultipleConnections(url, outputFile, contentLength, int64(parallelism), client)
	} else {
		err = downloadWithSingleConnection(url, outputFile, client)
	}
	if err != nil {
		return nil, err
	}
	err = assertFileValid(outputFile, contentLength)
	if err != nil {
		return nil, err
	}

	timeSpent := time.Since(start)
	return &timeSpent, err
}

func checkRemoteFile(url url.URL, client *http.Client) (supportRanges bool, contentLength int64, err error) {
	res, err := client.Head(url.String())
	if err != nil {
		return
	}
	defer res.Body.Close()

	contentLength = res.ContentLength
	supportRanges = res.Header.Get(HeaderAcceptRanges) == "bytes" && contentLength != 0
	return
}

func downloadWithSingleConnection(url url.URL, outputFile *os.File, client *http.Client) error {
	res, err := client.Get(url.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()

	buf := make([]byte, bufferSize)
	_, err = io.CopyBuffer(outputFile, res.Body, buf)
	return err
}

func downloadWithMultipleConnections(url url.URL, outputFile *os.File, contentLength int64, parallelism int64, client *http.Client) error {
	parts := make([]Part, parallelism)
	partSize := contentLength / parallelism
	remainder := contentLength % parallelism

	wg, _ := errgroup.WithContext(context.Background())
	for i, part := range parts {
		start := int64(i) * partSize
		end := start + partSize
		if remainder != 0 && i == len(parts)-1 {
			end += remainder
		}
		part = Part{start, end, partSize}
		wg.Go(func() error {
			return downloadPart(client, url, outputFile.Name(), part)
		})
	}
	return wg.Wait()
}

func downloadPart(client *http.Client, url url.URL, outputFilepath string, part Part) error {
	outputFile, err := os.OpenFile(outputFilepath, os.O_RDWR, 0664)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	_, err = outputFile.Seek(part.Start, 0)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set(HeaderRange, fmt.Sprintf("bytes=%d-%d", part.Start, part.End-1))
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("server advertises HTTP Range Request support "+
			"but doesn't return status %d", http.StatusPartialContent)
	}

	buf := make([]byte, bufferSize)
	_, err = io.CopyBuffer(outputFile, res.Body, buf)
	return err
}

func assertFileValid(outputFile *os.File, contentLength int64) error {
	fi, err := outputFile.Stat()
	if err != nil {
		return err
	}
	if fi.Size() != contentLength {
		return fmt.Errorf("invalid file, content-length %d and file size %d mismatch", contentLength, fi.Size())
	}
	return nil
}

func createHTTPClient(tlsSkipVerify bool, timeout time.Duration, retryDelay time.Duration,
	retryMaxDelay time.Duration, maxRetries int) *http.Client {

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: tlsSkipVerify, //nolint:gosec // on purpose, default is false
		},
	}
	//nolint:bodyclose // false positive
	retryPolicy := failsafehttp.NewRetryPolicyBuilder().
		WithBackoff(retryDelay, retryMaxDelay). //nolint:bodyclose // false positive
		WithMaxRetries(maxRetries).             //nolint:bodyclose // false positive
		Build()                                 //nolint:bodyclose // false positive
	return &http.Client{
		Timeout:   timeout,
		Transport: failsafehttp.NewRoundTripper(transport, retryPolicy),
	}
}
