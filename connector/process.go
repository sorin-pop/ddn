package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/djavorszky/notif"

	"strings"

	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/common/status"
)

func startImport(dbreq model.DBRequest) {
	upd8Path := fmt.Sprintf("%s/%s", conf.MasterAddress, "upd8")

	ch := notif.New(dbreq.ID, upd8Path)
	defer close(ch)

	ch <- notif.Y{StatusCode: status.DownloadInProgress, Msg: "Downloading dump"}

	path, err := inet.DownloadFile("dumps", dbreq.DumpLocation)
	if err != nil {
		db.DropDatabase(dbreq)
		log.Printf("could not download file: %s", err.Error())

		ch <- notif.Y{StatusCode: status.DownloadFailed, Msg: "Downloading file failed: " + err.Error()}
		return
	}
	defer os.Remove(path)

	if isArchive(path) {
		ch <- notif.Y{StatusCode: status.ExtractingArchive, Msg: "Extracting archive"}

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

			ch <- notif.Y{StatusCode: status.ArchiveNotSupported, Msg: "archive not supported"}
			return
		}
		for _, f := range files {
			defer os.Remove(f)
		}

		if err != nil {
			db.DropDatabase(dbreq)
			log.Printf("could not extract archive: %s", err.Error())

			ch <- notif.Y{StatusCode: status.ExtractingArchiveFailed, Msg: "Extracting file failed: " + err.Error()}
			return
		}

		if len(files) > 1 {
			db.DropDatabase(dbreq)
			log.Println("import process stopped; more than one file found in archive")

			ch <- notif.Y{StatusCode: status.MultipleFilesInArchive, Msg: "Archive contains more than one file, import stopped"}
			return
		}

		path = files[0]
	}

	if mdb, ok := db.(*mysql); ok {
		ch <- notif.Y{StatusCode: status.ValidatingDump, Msg: "Validating dump"}
		path, err = mdb.validateDump(path)

		if err != nil {
			db.DropDatabase(dbreq)
			log.Printf("database validation failed: %s", err.Error())

			ch <- notif.Y{StatusCode: status.ValidationFailed, Msg: "Validating dump failed: " + err.Error()}
			return
		}
	}

	if !strings.Contains(path, "dumps") {
		oldPath := path
		path = "dumps" + string(os.PathSeparator) + path

		os.Rename(oldPath, path)
	}

	path, _ = filepath.Abs(path)
	defer os.Remove(path)

	dbreq.DumpLocation = path

	ch <- notif.Y{StatusCode: status.ImportInProgress, Msg: "Importing"}

	err = db.ImportDatabase(dbreq)
	if err != nil {
		log.Printf("could not import database: %s", err.Error())

		ch <- notif.Y{StatusCode: status.ImportFailed, Msg: "Importing dump failed: " + err.Error()}
		return
	}

	ch <- notif.Y{StatusCode: status.Success, Msg: "Completed"}
}

// This method should always be called asynchronously
func keepAlive() {
	endpoint := fmt.Sprintf("%s/%s", conf.MasterAddress, "alive")

	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		if !inet.AddrExists(endpoint) {
			if registered {
				log.Println("Lost connection to master server, will attempt to reconnect once it's back.")

				registered = false
			}

			continue
		}

		if !registered {
			log.Println("Master server back online.")

			err := registerConnector()
			if err != nil {
				log.Printf("couldn't register with master: %s", err.Error())
			}
		}
	}
}
