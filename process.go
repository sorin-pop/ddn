package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func startImport(dbreq DBRequest) {

	log.Println("Starting download")

	filepath, err := downloadFile(dbreq.DumpLocation)
	if err != nil {
		log.Println(err)
	}
	defer os.Remove(filepath)

	log.Println("Download finished, starting import")

	userArg, pwArg, dbnameArg, pathArg := fmt.Sprintf("-u %s", conf.User), fmt.Sprintf("-p%s", conf.Password), dbreq.DatabaseName, fmt.Sprintf("< %s", filepath)

	log.Println(conf.Exec, userArg, pwArg, dbnameArg, pathArg)

	cmd := exec.Command(conf.Exec, userArg, pwArg, dbnameArg, pathArg)

	err = cmd.Run()
	if err != nil {
		log.Println("This fails all the time due to some reason.. The exact same command works from terminal.")
		log.Fatal(err)
	}

	log.Println("Import finished.")
}
