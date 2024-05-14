package engine

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/failsafe-go/failsafe-go/failsafehttp"
)

// Part piece of the file to download when HTTP Range Requests are supported
type Part struct {
	Start int64
	End   int64
	Size  int64
}

// Download downloads file from the given URL and stores the result in the given output location.
// Will utilize multiple concurrent connections to increase transfer speed. The latter is only
// possible when the remote server supports HTTP Range Requests, otherwise it falls back
// to single connection download.
func Download(url url.URL, outputFilepath string, workers int, tlsSkipVerify bool,
	retryDelay time.Duration, retryMaxDelay time.Duration, maxRetries int) (*time.Duration, error) {

	client := createHTTPClient(tlsSkipVerify, retryDelay, retryMaxDelay, maxRetries)

	outputFile, _ := os.OpenFile(outputFilepath, os.O_CREATE|os.O_RDWR, 0644)
	defer outputFile.Close()

	start := time.Now()
	supportRanges, contentLength := checkRemoteFile(url, client)
	if supportRanges {
		downloadWithMultipleConnections(url, outputFile, contentLength, workers, client)
	} else {
		downloadWithSingleConnection(url, outputFile, client)
	}
	timeSpent := time.Since(start)
	return &timeSpent, nil
}

func checkRemoteFile(url url.URL, client *http.Client) (supportRanges bool, contentLength int64) {
	res, err := client.Head(url.String())
	if err != nil {
		log.Fatal(fmt.Sprintf("Error on get url %s: %q\n", url.String(), err))
	}
	defer res.Body.Close()

	contentLength = res.ContentLength
	supportRanges = res.Header.Get(HeaderAcceptRanges) == "bytes" && contentLength != 0
	return
}

func downloadWithSingleConnection(url url.URL, outputFile *os.File, client *http.Client) {
	res, err := client.Get(url.String())
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	buf := make([]byte, 3*1024*1024)
	io.CopyBuffer(outputFile, res.Body, buf)
}

func downloadWithMultipleConnections(url url.URL, outputFile *os.File, contentLength int64, workers int, client *http.Client) {
	var wg sync.WaitGroup
	parts := make([]Part, workers)
	partSize := contentLength / int64(workers)
	remainder := contentLength % int64(workers)

	for i, part := range parts {
		start := int64(i) * partSize
		end := start + partSize
		if remainder != 0 && i == len(parts)-1 {
			end += remainder
		}
		part = Part{start, end, partSize}
		wg.Add(1)
		go downloadRange(client, url, outputFile.Name(), part, &wg)
	}
	wg.Wait()
}

func downloadRange(client *http.Client, url url.URL, outputFilepath string, part Part, wg *sync.WaitGroup) {
	defer wg.Done()
	outputFile, _ := os.OpenFile(outputFilepath, os.O_RDWR, 0664)
	defer outputFile.Close()
	outputFile.Seek(part.Start, 0)

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set(HeaderRange, fmt.Sprintf("bytes=%d-%d", part.Start, part.End-1))
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusPartialContent {
		panic("Remote doesn't responded with the expected code: 206")
	}

	buf := make([]byte, 3*1024*1024)
	io.CopyBuffer(outputFile, res.Body, buf)
}

func createHTTPClient(tlsSkipVerify bool, retryDelay time.Duration,
	retryMaxDelay time.Duration, maxRetries int) *http.Client {

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: tlsSkipVerify,
		},
	}
	retryPolicy := failsafehttp.RetryPolicyBuilder().
		WithBackoff(retryDelay, retryMaxDelay).
		WithMaxRetries(maxRetries).
		Build()
	return &http.Client{
		Transport: failsafehttp.NewRoundTripper(transport, retryPolicy),
	}
}
