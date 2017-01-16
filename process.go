package main

import (
	"log"
	"os"
)

func startImport(dbreq DBRequest) {
	log.Println("Starting download")

	filepath, err := downloadFile(dbreq.DumpLocation)
	if err != nil {
		log.Println("Downloading file failed:", err.Error())
	}
	defer os.Remove(filepath)

	dbreq.DumpLocation = filepath

	log.Println("Download finished, starting import")

	if err = db.ImportDatabase(dbreq); err != nil {
		log.Println("Importing dump failed:", err.Error())
		return
	}

	log.Println("Import finished.")
}
