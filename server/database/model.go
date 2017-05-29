package database

import (
	"time"

	"github.com/djavorszky/ddn/common/status"
)

// Entry represents a row in the database
type Entry struct {
	ID            int       `json:"id"`
	DBVendor      string    `json:"vendor"`
	DBName        string    `json:"dbname"`
	DBUser        string    `json:"dbuser"`
	DBPass        string    `json:"dbpass"`
	DBSID         string    `json:"sid"`
	Dumpfile      string    `json:"dumplocation"`
	CreateDate    time.Time `json:"createdate"`
	ExpiryDate    time.Time `json:"expirydate"`
	Creator       string    `json:"creator"`
	ConnectorName string    `json:"connector"`
	DBAddress     string    `json:"dbaddress"`
	DBPort        string    `json:"dbport"`
	Status        int       `json:"status"`
	Message       string    `json:"message"`
	Public        int       `json:"public"`
}

// InProgress returns true if the DBEntry's status denotes that something's in progress.
func (entry Entry) InProgress() bool {
	return entry.Status < 100
}

// IsStatusOk returns true if the DBEntry's status is OK.
func (entry Entry) IsStatusOk() bool {
	return entry.Status > 99 && entry.Status < 200
}

// IsClientErr returns true if something went wrong with the client request.
func (entry Entry) IsClientErr() bool {
	return entry.Status > 199 && entry.Status < 300
}

// IsServerErr returns true if something went wrong on the server.
func (entry Entry) IsServerErr() bool {
	return entry.Status > 299 && entry.Status < 400
}

// IsErr returns true if something went wrong either on the server or with the client request.
func (entry Entry) IsErr() bool {
	return entry.IsServerErr() || entry.IsClientErr()
}

// IsWarn returns true if something went wrong either on the server or with the client request.
func (entry Entry) IsWarn() bool {
	return entry.Status > 399
}

// StatusLabel returns the string representation of the status
func (entry Entry) StatusLabel() string {
	label, ok := status.Labels[entry.Status]
	if !ok {
		return "Unknown"
	}

	return label
}

// Progress returns the progress as 0 <= progress <= 100 of its current import.
// If error, returns 0; If success, returns 100;
func (entry Entry) Progress() int {
	if entry.IsClientErr() || entry.IsServerErr() {
		return 0
	}

	if entry.IsStatusOk() {
		return 100
	}

	switch entry.Status {
	case status.DownloadInProgress:
		return 0
	case status.ExtractingArchive:
		return 25
	case status.ValidatingDump:
		return 50
	case status.ImportInProgress:
		return 75
	default:
		return 0
	}
}
