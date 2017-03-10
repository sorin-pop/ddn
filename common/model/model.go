package model

import (
	"fmt"

	"encoding/json"

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
	Requester           string `json:"requester_email"`
	PortalVersion       string `json:"portal_version"`
	DBRequest
}

// PortalExt contains the portal-ext.properties for the created database
type PortalExt struct {
	Driver   string
	URL      string
	User     string
	Password string
}

// RegisterRequest is used to represent a JSON call between the connector and the server.
// ID can be null if it's the initial registration, but must correspond to the connector's
// ID when unregistering
type RegisterRequest struct {
	ConnectorName string `json:"connector_name"`
	DBVendor      string `json:"dbvendor"`
	DBPort        string `json:"dbport"`
	DBSID         string `json:"dbsid"`
	ShortName     string `json:"short_name"`
	LongName      string `json:"long_name"`
	Version       string `json:"version"`
	Port          string `json:"port"`
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
	ID            int
	DBVendor      string
	DBName        string
	DBUser        string
	DBPass        string
	DBSID         string
	Dumpfile      string
	CreateDate    string
	ExpiryDate    string
	Creator       string
	ConnectorName string
	DBAddress     string
	DBPort        string
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

func (c Connector) executeAction(dbreq DBRequest, endpoint string) (string, error) {
	dest := fmt.Sprintf("http://%s:%s/%s", c.Address, c.ConnectorPort, endpoint)

	resp, err := notif.SndLoc(dbreq, dest)
	if err != nil {
		return "", fmt.Errorf("sending json message failed: %s", err.Error())
	}

	var respMsg inet.Message

	json.Unmarshal([]byte(resp), &respMsg)

	switch respMsg.Status {
	case status.Success, status.Accepted, status.Started, status.Created:
		return respMsg.Message, nil
	default:
		return "", fmt.Errorf("executing action on endpoint %q failed: %s", endpoint, respMsg.Message)
	}
}
