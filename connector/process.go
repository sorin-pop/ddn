package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/djavorszky/notif"

	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/common/status"
)

func startImport(dbreq model.DBRequest) {
	upd8Path := fmt.Sprintf("%s/%s", conf.MasterAddress, "echo")

	ch := notif.New(dbreq.ID, upd8Path)
	defer close(ch)

	ch <- notif.Y{StatusCode: status.Started, Msg: "Starting download"}

	path, err := inet.DownloadFile(usr.HomeDir, dbreq.DumpLocation)
	if err != nil {
		db.DropDatabase(dbreq)
		log.Printf("could not download file: %s", err.Error())

		ch <- notif.Y{StatusCode: status.Update, Msg: "Downloading file failed: " + err.Error()}
		return
	}
	defer os.Remove(path)

	if isArchive(path) {
		ch <- notif.Y{StatusCode: status.Update, Msg: "Extracting archive"}

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
			db.DropDatabase(dbreq)
			log.Println("import process stopped; encountered unsupported archive")

			ch <- notif.Y{StatusCode: status.ClientError, Msg: "Unsupported archive"}
			return
		}
		for _, f := range files {
			defer os.Remove(f)
		}

		if err != nil {
			db.DropDatabase(dbreq)
			log.Printf("could not extract archive: %s", err.Error())

			ch <- notif.Y{StatusCode: status.ServerError, Msg: "Extracting file failed: " + err.Error()}
			return
		}

		if len(files) > 1 {
			db.DropDatabase(dbreq)
			log.Println("import process stopped; more than one file found in archive")

			ch <- notif.Y{StatusCode: status.ClientError, Msg: "Archive contains more than one file, import stopped"}
			return
		}

		path = files[0]
	}

	if mdb, ok := db.(*mysql); ok {
		ch <- notif.Y{StatusCode: status.Update, Msg: "Validating MySQL dump"}
		path, err = mdb.validateDump(path)

		if err != nil {
			db.DropDatabase(dbreq)
			log.Printf("database validation failed: %s", err.Error())

			ch <- notif.Y{StatusCode: status.ServerError, Msg: "Validating dump failed: " + err.Error()}
			return
		}
	}

	dbreq.DumpLocation = path

	ch <- notif.Y{StatusCode: status.Update, Msg: "Starting import"}

	err = db.ImportDatabase(dbreq)
	if err != nil {
		log.Printf("could not import database: %s", err.Error())

		ch <- notif.Y{StatusCode: status.ServerError, Msg: "Importing dump failed: " + err.Error()}
		return
	}

	ch <- notif.Y{StatusCode: status.Success, Msg: "Import finished successfully"}
}
