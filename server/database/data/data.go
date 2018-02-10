package data

import (
	"time"

	"github.com/djavorszky/ddn/common/status"
)

// Row represents a row in the database
type Row struct {
	ID         int       `json:"id"`
	DBVendor   string    `json:"vendor"`
	DBName     string    `json:"dbname"`
	DBUser     string    `json:"dbuser"`
	DBPass     string    `json:"dbpass"`
	DBSID      string    `json:"sid"`
	Dumpfile   string    `json:"dumplocation"`
	CreateDate time.Time `json:"createdate"`
	ExpiryDate time.Time `json:"expirydate"`
	Creator    string    `json:"creator"`
	AgentName  string    `json:"agent"`
	DBAddress  string    `json:"dbaddress"`
	DBPort     string    `json:"dbport"`
	Status     int       `json:"status"`
	Label      string    `json:"status_label"`
	Comment    string    `json:"comment"`
	Message    string    `json:"message"`
	Public     int       `json:"public"`
}

// InProgress returns true if the DBEntry's status denotes that something's in progress.
func (row Row) InProgress() bool {
	return row.Status < 100
}

// IsStatusOk returns true if the DBEntry's status is OK.
func (row Row) IsStatusOk() bool {
	return row.Status > 99 && row.Status < 200
}

// IsClientErr returns true if something went wrong with the client request.
func (row Row) IsClientErr() bool {
	return row.Status > 199 && row.Status < 300
}

// IsServerErr returns true if something went wrong on the server.
func (row Row) IsServerErr() bool {
	return row.Status > 299 && row.Status < 400
}

// IsErr returns true if something went wrong either on the server or with the client request.
func (row Row) IsErr() bool {
	return row.IsServerErr() || row.IsClientErr()
}

// IsWarn returns true if something went wrong either on the server or with the client request.
func (row Row) IsWarn() bool {
	return row.Status > 399
}

// StatusLabel returns the string representation of the status
func (row Row) StatusLabel() string {
	label, ok := status.Labels[row.Status]
	if !ok {
		return "Unknown"
	}

	return label
}

// Progress returns the progress as 0 <= progress <= 100 of its current import.
// If error, returns 0; If success, returns 100;
func (row Row) Progress() int {
	if row.IsClientErr() || row.IsServerErr() {
		return 0
	}

	if row.IsStatusOk() {
		return 100
	}

	switch row.Status {
	case status.DownloadInProgress, status.CopyInProgress:
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
