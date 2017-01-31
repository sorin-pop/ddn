package main

import (
	"log"
	"os"

	"net/http"

	"github.com/djavorszky/notif"
)

func startImport(dbreq DBRequest) {
	var err error

	ch := notif.New(dbreq.ID, conf.MasterAddress)

	ch <- notif.Y{StatusCode: http.StatusOK, Msg: "Starting download"}
	if err != nil {
		log.Println(err)
	}

	filepath, err := downloadFile(dbreq.DumpLocation)
	if err != nil {
		ch <- notif.Y{StatusCode: http.StatusInternalServerError, Msg: "Downlading file failed: " + err.Error()}
		return
	}
	defer os.Remove(filepath)

	dbreq.DumpLocation = filepath

	ch <- notif.Y{StatusCode: http.StatusOK, Msg: "Starting import"}

	// TODO: Connector dies if import fails, e.g. if dumpfile is of wrong version.

	if err = db.ImportDatabase(dbreq); err != nil {
		ch <- notif.Y{StatusCode: http.StatusInternalServerError, Msg: "Importing dump failed: " + err.Error()}

		return
	}
	ch <- notif.Y{StatusCode: http.StatusOK, Msg: "Import finished successfully."}
	close(ch)
}
