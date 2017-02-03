package main

import (
	"log"
	"os"
	"path/filepath"

	"net/http"

	"github.com/djavorszky/ddn/inet"
	"github.com/djavorszky/ddn/model"
	"github.com/djavorszky/notif"
)

func startImport(dbreq model.DBRequest) {
	ch := notif.New(dbreq.ID, conf.MasterAddress)
	defer close(ch)

	ch <- notif.Y{StatusCode: http.StatusOK, Msg: "Starting download"}

	path, err := inet.DownloadFile(usr.HomeDir, dbreq.DumpLocation)
	if err != nil {
		log.Printf("could not download file: %s", err.Error())

		ch <- notif.Y{StatusCode: http.StatusInternalServerError, Msg: "Downlading file failed: " + err.Error()}
		return
	}
	defer os.Remove(path)

	if isArchive(path) {
		ch <- notif.Y{StatusCode: http.StatusOK, Msg: "Extracting archive"}

		var (
			files []string
			err   error
		)

		switch filepath.Ext(path) {
		case ".zip":
			files, err = unzip(path)
		case ".gz":
			files, err = ungzip(path)
		case ".tar":
			files, err = untar(path)
		default:
			log.Println("import process stopped; encountered unsupported archive")

			ch <- notif.Y{StatusCode: http.StatusBadRequest, Msg: "Unsupported archive"}
			return
		}
		for _, f := range files {
			defer os.Remove(f)
		}

		if err != nil {
			log.Printf("could not extract archive: %s", err.Error())

			ch <- notif.Y{StatusCode: http.StatusInternalServerError, Msg: "Extracting file failed: " + err.Error()}
			return
		}

		if len(files) > 1 {
			log.Println("import process stopped; more than one file found in archive")

			ch <- notif.Y{StatusCode: http.StatusBadRequest, Msg: "Archive contains more than one file, import stopped"}
			return
		}

		path = files[0]
	}

	dbreq.DumpLocation = path

	ch <- notif.Y{StatusCode: http.StatusOK, Msg: "Starting import"}

	// TODO: Connector dies if import fails, e.g. if dumpfile is of wrong version.

	if err = db.ImportDatabase(dbreq); err != nil {
		log.Printf("could not import database: %s", err.Error())

		ch <- notif.Y{StatusCode: http.StatusInternalServerError, Msg: "Importing dump failed: " + err.Error()}
		return
	}
	ch <- notif.Y{StatusCode: http.StatusOK, Msg: "Import finished successfully"}
}
