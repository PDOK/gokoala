package util

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
)

// ReadFile read a plain or gzipped file and return contents as string.
func ReadFile(filePath string) string {
	gzipFile := filePath + ".gz"
	var fileContents string
	if _, err := os.Stat(gzipFile); !errors.Is(err, fs.ErrNotExist) {
		fileContents, err = readGzipContents(gzipFile)
		if err != nil {
			log.Fatalf("unable to decompress gzip file %s", gzipFile)
		}
	} else {
		fileContents, err = readPlainContents(filePath)
		if err != nil {
			log.Fatalf("unable to read file %s", filePath)
		}
	}

	return fileContents
}

// decompress gzip files, return contents as string.
func readGzipContents(filePath string) (string, error) {
	gzipFile, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(gzipFile *os.File) {
		err := gzipFile.Close()
		if err != nil {
			log.Println("failed to close gzip file")
		}
	}(gzipFile)
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return "", err
	}
	defer func(gzipReader *gzip.Reader) {
		err := gzipReader.Close()
		if err != nil {
			log.Println("failed to close gzip reader")
		}
	}(gzipReader)
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, gzipReader) //nolint:gosec
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// read file, return contents as string.
func readPlainContents(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println("failed to close file")
		}
	}(file)
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, file)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
