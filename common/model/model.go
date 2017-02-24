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

// CreateDatabase sends a request to the connector to create a database.
func (c Connector) CreateDatabase(id int, dbname, dbuser, dbpass string) (string, error) {

	if ok := sutils.Present(dbname, dbpass, dbuser); !ok {
		return "", fmt.Errorf("asked to persist database with missing values: dbname: %q, dbuser: %q, dbpass: %q", dbname, dbpass, dbuser)
	}

	dbreq := DBRequest{
		ID:           id,
		DatabaseName: dbname,
		Username:     dbuser,
		Password:     dbpass,
	}

	dest := fmt.Sprintf("http://%s:%s/create-database", c.Address, c.ConnectorPort)

	resp, err := notif.SndLoc(dbreq, dest)
	if err != nil {
		return "", fmt.Errorf("sending json message failed: %s", err.Error())
	}

	var respMsg inet.Message

	json.Unmarshal([]byte(resp), &respMsg)

	if respMsg.Status != status.Success {
		return "", fmt.Errorf("creating database failed: %s", respMsg.Message)
	}

	return respMsg.Message, nil
}
