package main

import (
	"log"
	"os"

	"net/http"

	"github.com/djavorszky/notif"
)

func startImport(dbreq DBRequest) {
	var err error

	notif.Dest = conf.MasterAddress

	msg := notif.Msg{ID: dbreq.ID, StatusID: http.StatusOK, Message: "Starting download"}

	err = notif.Snd(msg)
	if err != nil {
		log.Println(err)
	}

	filepath, err := downloadFile(dbreq.DumpLocation)
	if err != nil {
		msg.StatusID = http.StatusInternalServerError
		msg.Message = "Downlading file failed: " + err.Error()

		err = notif.Snd(msg)
		if err != nil {
			log.Println(err)
		}

		return
	}
	defer os.Remove(filepath)

	dbreq.DumpLocation = filepath

	msg.Message = "Download finished, starting import"

	err = notif.Snd(msg)
	if err != nil {
		log.Println(err)
	}

	// TODO: Connector dies if import fails, e.g. if dumpfile is of wrong version.

	if err = db.ImportDatabase(dbreq); err != nil {
		msg.StatusID = http.StatusInternalServerError
		msg.Message = "Importing dump failed: " + err.Error()

		err = notif.Snd(msg)
		if err != nil {
			log.Println(err)
		}

		return
	}

	msg.Message = "Import finished successfully."

	err = notif.Snd(msg)
	if err != nil {
		log.Println(err)
	}
}
