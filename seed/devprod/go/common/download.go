package common

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadFile(url string, filePath string) error {
	log.Printf("Downloading %v to %v", url, filePath)
	response, err := http.Get(url)
	if err != nil {
		return WrapTrace(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return WrapTrace(fmt.Errorf("failed to download %v: %v", url, response.Status))
	}
	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return WrapTrace(err)
	}
	file, err := os.Create(filePath)
	if err != nil {
		return WrapTrace(err)
	}
	defer file.Close()
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return WrapTrace(err)
	}
	err = file.Close()
	if err != nil {
		return WrapTrace(err)
	}
	return nil
}
