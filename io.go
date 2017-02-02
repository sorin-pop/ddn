package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

func generateProps(filename string) error {
	filename, conf := generateInteractive(filename)

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("couldn't create file: %s", err.Error())
	}
	defer file.Close()

	prop := `vendor="{{.Vendor}}"
version="{{.Version}}"
executable="{{.Exec}}"
dbport="{{.DBPort}}"
dbAddress="{{.DBAddress}}"
connectorPort="{{.ConnectorPort}}"
username="{{.User}}"
password="{{.Password}}"
masterAddress="{{.MasterAddress}}"
`

	if conf.SID != "" {
		prop += "oracle-sid=\"{{.SID}}\"\n"
	}

	if conf.DefaultTablespace != "" {
		prop += "default-tablespace=\"{{.DefaultTablespace}}\"\n"
	}

	tmpl, err := template.New("props").Parse(prop)
	if err != nil {
		return fmt.Errorf("couldn't parse template: %s", err.Error())
	}

	err = tmpl.Execute(file, conf)
	if err != nil {
		return fmt.Errorf("couldn't execute template: %s", err.Error())
	}

	file.Sync()

	return nil
}

func unzip(filepath string) ([]string, error) {
	r, err := zip.OpenReader(filepath)
	if err != nil {
		return nil, fmt.Errorf("creating zip reader failed: %s", err.Error())
	}
	defer r.Close()

	var files []string
	for _, f := range r.File {
		name, err := unzipFile(f)
		if err != nil {
			return nil, fmt.Errorf("extracting zip file failed: %s", err.Error())
		}

		files = append(files, name)
	}

	return files, nil
}

func unzipFile(f *zip.File) (string, error) {
	src, err := f.Open()
	if err != nil {
		return "", fmt.Errorf("opening zipfile failed: %s", err.Error())
	}
	defer src.Close()

	dst, err := os.OpenFile(f.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return "", fmt.Errorf("opening destination file failed: %s", err.Error())
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", fmt.Errorf("copying from archive failed: %s", err.Error())
	}

	return f.Name, nil
}

func ungzip(path string) ([]string, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening gzipfile failed: %s", err.Error())
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("creating gzip reader failed: %s", err.Error())
	}
	defer archive.Close()

	ext := filepath.Ext(path)
	name := archive.Header.Name
	if name == "" {
		dstName := filepath.Base(path)

		name = dstName[:len(dstName)-len(ext)]
	}

	dst, err := os.Create(name)
	if err != nil {
		return nil, fmt.Errorf("could not create output file: %s", err.Error())
	}
	defer dst.Close()

	_, err = io.Copy(dst, archive)
	if err != nil {
		return nil, fmt.Errorf("uncompressing gzip failed: %s", err.Error())
	}

	if filepath.Ext(dst.Name()) == ".tar" {
		return untar(fmt.Sprintf("%s/%s", filepath.Dir(dst.Name()), dst.Name()))
	}

	return []string{dst.Name()}, nil
}

func untar(path string) ([]string, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("opening tarball failed: %s", err.Error())
	}
	defer file.Close()

	tarBallReader := tar.NewReader(file)

	for {
		header, err := tarBallReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, fmt.Errorf("encountered error while reading tarball: %s", err.Error())
		}

		filename := header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			if err != nil {
				return nil, fmt.Errorf("tarball contains folder, stopping")
			}
		case tar.TypeReg:
			writer, err := os.Create(filename)
			if err != nil {
				return nil, fmt.Errorf("could not create output file: %s", err.Error())
			}
			defer writer.Close()

			_, err = io.Copy(writer, tarBallReader)
			if err != nil {
				return nil, fmt.Errorf("uncompressing tarball failed: %s", err.Error())
			}

			return []string{writer.Name()}, nil
		default:
			fmt.Printf("Unable to untar type : %c in file %s", header.Typeflag, filename)
		}
	}

	return []string{path}, nil
}

func isArchive(path string) bool {
	switch filepath.Ext(path) {
	case ".zip", ".tar", ".gz":
		return true
	}

	return false
}
