package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func writeHeader(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
}

func downloadFile(location string) (string, error) {
	i, j := strings.LastIndex(location, "/"), len(location)
	filename := location[i+1 : j]

	filepath := fmt.Sprintf("%s/%s", usr.HomeDir, filename)

	out, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("could not create file: %s", err.Error())
	}
	defer out.Close()

	resp, err := http.Get(location)
	if err != nil {
		return "", fmt.Errorf("couldn't get url '%s': %s", location, err.Error())
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("downloading file failed: %s", err.Error())
	}

	return filepath, nil
}

func fileExists(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("could not get url '%s': %s", url, err)

		resp.Body.Close()
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	}

	return false
}
