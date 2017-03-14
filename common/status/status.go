// Package status contains status codes not unline the http
// statuses, but tailored toward the ddn ecosystem
package status

// Labels contains the labels of the statuses.
var Labels map[int]string

func init() {
	Labels = make(map[int]string)

	// Info
	Labels[Started] = "Started"
	Labels[InProgress] = "In Progress"
	Labels[DownloadInProgress] = "Downloading"
	Labels[ExtractingArchive] = "Extracting Archive"
	Labels[ValidatingDump] = "Validating Dump"
	Labels[ImportInProgress] = "Importing"

	// Success
	Labels[Success] = "Completed"
	Labels[Created] = "Created"
	Labels[Accepted] = "Accepted"
	Labels[Update] = "Update"

	// Client Error
	Labels[ClientError] = "Client Error"
	Labels[NotFound] = "Not found"

	// Server Error
	Labels[ServerError] = "Server Error"

	// Warnings
	Labels[RemovalScheduled] = "Removal scheduled"
}

// Info statuses are used to convey that something has happened
// but has not finished yet. It is not a success, nor a failure.
//
// They can range from 1 to 99
const (
	Started    int = 1 // status.Started
	InProgress int = 2 // status.InProgress

	DownloadInProgress int = 3 // status.DownloadInProgress
	ExtractingArchive  int = 4 // status.ExtractingArchive
	ValidatingDump     int = 5 // status.ValidatingDump
	ImportInProgress   int = 6 // status.ImportInProgress
)

// Success statuses are used to convey a successful result.
const (
	Success  int = 100 // status.Success
	Created  int = 101 // status.Created
	Accepted int = 102 // status.Accepted
	Update   int = 103 // status.Update
)

// Client errors are used to convey that something was
// wrong with a client request.
const (
	ClientError int = 200 // status.ClientError
	NotFound    int = 201 // status.NotFound
)

// Server errors are used to convey that something went wrong
// on the server.
const (
	ServerError int = 300 // status.ServerError
)

// Warnings are for issuing warnings.
const (
	RemovalScheduled int = 400 // status.RemovalScheduled
)
