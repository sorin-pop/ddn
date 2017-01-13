package main

import (
	"log"
	"os"
)

func startImport(dbreq DBRequest) {

	log.Println("Starting download")

	filepath, err := downloadFile(dbreq.DumpLocation)
	if err != nil {
		log.Println(err)
	}
	defer os.Remove(filepath)

	log.Println("Download finished, starting import")

	// TODO actually start an import..
}
