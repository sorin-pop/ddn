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
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(location)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return filepath, nil
}

func fileExists(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	}

	return false
}
