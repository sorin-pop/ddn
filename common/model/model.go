package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/status"
	"github.com/djavorszky/notif"
	"github.com/djavorszky/sutils"
)

// DBRequest is used to represent JSON call about creating, dropping or importing databases
type DBRequest struct {
	ID           int    `json:"id"`
	DatabaseName string `json:"database_name"`
	DumpLocation string `json:"dumpfile_location"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

// ClientRequest is used to represent a JSON call between a client and the server
type ClientRequest struct {
	ConnectorIdentifier string `json:"connector_identifier"`
	RequesterEmail      string `json:"requester_email"`
	DBRequest
}

// RegisterRequest is used to represent a JSON call between the connector and the server.
// ID can be null if it's the initial registration, but must correspond to the connector's
// ID when unregistering
type RegisterRequest struct {
	ConnectorName string `json:"connector_name"`
	DBVendor      string `json:"dbvendor"`
	DBPort        string `json:"dbport"`
	DBAddr        string `json:"dbaddr"`
	DBSID         string `json:"dbsid"`
	ShortName     string `json:"short_name"`
	LongName      string `json:"long_name"`
	Version       string `json:"version"`
	Port          string `json:"port"`
	Addr          string `json:"address"`
}

// RegisterResponse is used as the response to the RegisterRequest
type RegisterResponse struct {
	ID      int    `json:"id"`
	Address string `json:"address"`
	Token   string `json:"token"`
}

// Connector is used to represent a DDN Connector.
type Connector struct {
	ID            int
	DBVendor      string
	DBPort        string
	DBAddr        string
	DBSID         string
	ShortName     string
	LongName      string
	Identifier    string
	ConnectorPort string
	Version       string
	Address       string
	Token         string
	Up            bool
}

// DBEntry represents a row in the "databases" table.
type DBEntry struct {
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
func (dbe DBEntry) InProgress() bool {
	return dbe.Status < 100
}

// IsStatusOk returns true if the DBEntry's status is OK.
func (dbe DBEntry) IsStatusOk() bool {
	return dbe.Status > 99 && dbe.Status < 200
}

// IsClientErr returns true if something went wrong with the client request.
func (dbe DBEntry) IsClientErr() bool {
	return dbe.Status > 199 && dbe.Status < 300
}

// IsServerErr returns true if something went wrong on the server.
func (dbe DBEntry) IsServerErr() bool {
	return dbe.Status > 299 && dbe.Status < 400
}

// IsErr returns true if something went wrong either on the server or with the client request.
func (dbe DBEntry) IsErr() bool {
	return dbe.IsServerErr() || dbe.IsClientErr()
}

// IsWarn returns true if something went wrong either on the server or with the client request.
func (dbe DBEntry) IsWarn() bool {
	return dbe.Status > 399
}

// StatusLabel returns the string representation of the status
func (dbe DBEntry) StatusLabel() string {
	label, ok := status.Labels[dbe.Status]
	if !ok {
		return "Unknown"
	}

	return label
}

// Progress returns the progress as 0 <= progress <= 100 of its current import.
// If error, returns 0; If success, returns 100;
func (dbe DBEntry) Progress() int {
	if dbe.IsClientErr() || dbe.IsServerErr() {
		return 0
	}

	if dbe.IsStatusOk() {
		return 100
	}

	switch dbe.Status {
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

// CreateDatabase sends a request to the connector to create a database.
func (c Connector) CreateDatabase(id int, dbname, dbuser, dbpass string) (string, error) {

	if ok := sutils.Present(dbname, dbuser, dbpass); !ok {
		return "", fmt.Errorf("asked to create database with missing values: dbname: %q, dbuser: %q, dbpass: %q", dbname, dbuser, dbpass)
	}

	dbreq := DBRequest{
		ID:           id,
		DatabaseName: dbname,
		Username:     dbuser,
		Password:     dbpass,
	}

	return c.executeAction(dbreq, "create-database")
}

// ImportDatabase starts the import on the connector.
func (c Connector) ImportDatabase(id int, dbname, dbuser, dbpass, dumploc string) (string, error) {
	if ok := sutils.Present(dbname, dbuser, dbpass, dumploc); !ok {
		return "", fmt.Errorf("asked to import database with missing values: dbname: %q, dbuser: %q, dbpass: %q, dumploc: %q", dbname, dbuser, dbpass, dumploc)
	}

	dbreq := DBRequest{
		ID:           id,
		DatabaseName: dbname,
		Username:     dbuser,
		Password:     dbpass,
		DumpLocation: dumploc,
	}

	return c.executeAction(dbreq, "import-database")
}

// DropDatabase sends a request to the connector to create a database.
func (c Connector) DropDatabase(id int, dbname, dbuser string) (string, error) {

	if ok := sutils.Present(dbname, dbuser); !ok {
		return "", fmt.Errorf("asked to create database with missing values: dbname: %q, dbuser: %q", dbname, dbuser)
	}

	dbreq := DBRequest{
		ID:           id,
		DatabaseName: dbname,
		Username:     dbuser,
	}

	return c.executeAction(dbreq, "drop-database")
}

func (c Connector) executeAction(dbreq DBRequest, endpoint string) (string, error) {
	dest := fmt.Sprintf("%s:%s/%s", c.Address, c.ConnectorPort, endpoint)

	if !strings.HasPrefix(dest, "http://") && !strings.HasPrefix(dest, "https://") {
		dest = fmt.Sprintf("http://%s", dest)
	}

	resp, err := notif.SndLoc(dbreq, dest)
	if err != nil && resp == "" {
		return "", fmt.Errorf("sending json message failed: %s", err.Error())
	}

	var respMsg inet.Message

	json.Unmarshal([]byte(resp), &respMsg)

	switch respMsg.Status {
	case status.Success, status.Accepted, status.Started, status.Created:
		return respMsg.Message, nil
	case status.MissingParameters:
		return "", fmt.Errorf("missing parameters from the request")
	case status.InvalidJSON:
		return "", fmt.Errorf("invalid JSON request")
	case status.CreateDatabaseFailed, status.ListDatabaseFailed, status.DropDatabaseFailed:
		return "", fmt.Errorf("connector issue: %s", respMsg.Message)
	default:
		return "", fmt.Errorf("executing action on endpoint %q failed: %s", endpoint, respMsg.Message)
	}
}
