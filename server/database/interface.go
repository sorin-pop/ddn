package database

import "github.com/djavorszky/ddn/server/database/data"

// BackendConnection defines all required methods that is needed
// in order to fulfill the requirements of backend communication
//
// Any initial setup should be done with struct fields
type BackendConnection interface {
	ConnectAndPrepare() error
	Close() error

	FetchByID(ID int) (data.Row, error)
	FetchByCreator(creator string) ([]data.Row, error)
	FetchPublic() ([]data.Row, error)
	FetchAll() ([]data.Row, error)

	Insert(row *data.Row) error
	Update(row *data.Row) error
	Delete(row data.Row) error
}
