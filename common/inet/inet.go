// Package inet contains convenience features for operations on the internet.
package inet

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// WriteHeader updates the header's Content-Type to application/json and charset to
// UTF-8. Additionally, it also adds the http status to it.
func WriteHeader(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
}

// DownloadFile downloads the file from the url and places it into the
// `dest` folder
func DownloadFile(dest, url string) (string, error) {
	i, j := strings.LastIndex(url, "/"), len(url)
	filename := url[i+1 : j]

	filepath := fmt.Sprintf("%s/%s", dest, filename)

	out, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("could not create file: %s", err.Error())
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("couldn't get url '%s': %s", url, err.Error())
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("downloading file failed: %s", err.Error())
	}

	return filepath, nil
}

// AddrExists checks the URL to see if it's valid, downloadable file or not.
func AddrExists(url string) bool {
	defer func() {
		if p := recover(); p != nil {
			// panic happens, no need to log anything. It's usually a refusal.
			// log.Printf("Remote end %q refused the connection", url)
		}
	}()

	resp, err := http.Get(url)
	if err != nil {
		resp.Body.Close()
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	}

	return false
}

// SendResponse composes the message, writes the header, then writes the bytes
// to the ResponseWriter
func SendResponse(w http.ResponseWriter, msg JSONMessage) {
	b, status := msg.Compose()

	WriteHeader(w, status)

	w.Write(b)
}
