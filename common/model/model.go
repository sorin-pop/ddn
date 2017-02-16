package model

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
