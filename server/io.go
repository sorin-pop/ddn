package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

func createProps(filename string, conf Config) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("couldn't create file: %s", err.Error())
	}
	defer file.Close()

	toml.NewEncoder(file).Encode(conf)

	return nil
}
