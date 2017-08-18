package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/djavorszky/ddn/server/database/data"
	"github.com/djavorszky/ddn/server/database/dbutil"
)

const (
	testDBFile = "./test.db"
)

var (
	testConn *sql.DB
	lite     DB

	gmt, _ = time.LoadLocation("GMT")

	testEntry = data.Row{
		ID:            1,
		DBName:        "testDB",
		DBUser:        "testUser",
		DBPass:        "testPass",
		DBSID:         "testsid",
		Dumpfile:      "testloc",
		CreateDate:    time.Now().In(gmt),
		ExpiryDate:    time.Now().In(gmt).AddDate(0, 0, 30),
		Creator:       "test@gmail.com",
		ConnectorName: "mysql-55",
		DBAddress:     "localhost",
		DBPort:        "3306",
		DBVendor:      "mysql",
		Message:       "",
		Status:        100,
	}
)

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		fmt.Printf("Failed setup: %s", err.Error())
		os.Exit(-1)
	}

	res := m.Run()

	teardown()

	os.Exit(res)
}

func setup() error {
	var err error

	os.Remove(testDBFile)

	testConn, err = sql.Open("sqlite3", testDBFile)
	if err != nil {
		return fmt.Errorf("could not open connection to database: %v", err)
	}
	lite.conn = testConn

	return nil
}

func teardown() {
	testConn.Close()
	lite.Close()
	os.Remove(testDBFile)
}

func TestInitTables(t *testing.T) {
	var err error

	_, err = testConn.Exec("SELECT 1 FROM version LIMIT 1;")
	if err == nil {
		t.Errorf("Version table already exists before test even ran.")
	}

	_, err = testConn.Exec("SELECT 1 FROM `databases` LIMIT 1;")
	if err == nil {
		t.Errorf("Databases table already exists before test even ran.")
	}

	err = lite.initTables()
	if err != nil {
		t.Errorf("Failed initializing tables: %s", err.Error())
	}

	_, err = testConn.Exec("SELECT 1 FROM version LIMIT 1;")
	if err != nil {
		t.Errorf("Version table has not been created.")
	}

	_, err = testConn.Exec("SELECT 1 FROM `databases` LIMIT 1;")
	if err != nil {
		t.Errorf("Databases table has not been created.")
	}

	type versiontest struct {
		queryID int
		query   string
		comment string
		date    time.Time
	}

	rows, _ := testConn.Query("SELECT * FROM version ORDER BY queryId DESC")

	for rows.Next() {
		var row versiontest

		err = rows.Scan(&row.queryID, &row.query, &row.comment, &row.date)
		if err != nil {
			t.Errorf("failed reading row: %v", err)
		}

		dbu := queries[row.queryID-1]

		if row.query != dbu.Query {
			t.Errorf("Query mismatch. Expected: %q, got: %q", dbu.Query, row.query)
		}

		if row.comment != dbu.Comment {
			t.Errorf("Comment mismatch: Expected %q, got: %q", dbu.Comment, row.comment)
		}
	}
	err = rows.Err()
	if err != nil {
		t.Errorf("error reading result from query: %s", err.Error())
	}
	rows.Close()
}

func TestInsert(t *testing.T) {
	err := lite.Insert(&testEntry)
	if err != nil {
		t.Errorf("lite.Insert(testEntry) failed with error: %v", err)
	}

	if testEntry.ID == 0 {
		t.Errorf("lite.Insert(testEntry) resulted in  id of 0")
	}

	result, err := lite.FetchByID(testEntry.ID)
	if err != nil {
		t.Errorf("FetchById(%d) resulted in error: %v", testEntry.ID, err)
	}

	if err = dbutil.CompareRows(testEntry, result); err != nil {
		t.Errorf("Persisted and read results not the same: %v", err)
	}
}

func TestFetchByID(t *testing.T) {
	lite.Insert(&testEntry)

	res, err := lite.FetchByID(testEntry.ID)
	if err != nil {
		t.Errorf("FetchById(%d) failed with error: %v", testEntry.ID, err)
	}

	if err := dbutil.CompareRows(res, testEntry); err != nil {
		t.Errorf("Fetched result not the same as queried: %v", err)
	}
}

func TestFetchByCreator(t *testing.T) {
	creator := "someone@somewhere.com"

	testEntry.Creator = creator

	lite.Insert(&testEntry)
	lite.Insert(&testEntry)

	results, err := lite.FetchByCreator(creator)
	if err != nil {
		t.Errorf("failed to fetch by creator: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected resultset to have 2 results, %d instead", len(results))
	}

	for _, res := range results {
		if res.Creator != creator {
			t.Errorf("Creator mismatch: Got %q, expected %q", res.Creator, creator)
		}
	}
}

func TestFetchPublic(t *testing.T) {
	res, err := lite.FetchPublic()
	if err != nil {
		t.Errorf("FetchPublic() error: %v", err)
	}

	if len(res) != 0 {
		t.Errorf("FetchPublic() returned with entries, shouldn't have")
	}

	testEntry.Public = 1

	lite.Insert(&testEntry)

	res, err = lite.FetchPublic()
	if err != nil {
		t.Errorf("FetchPublic() error: %v", err)
	}

	if len(res) != 1 {
		t.Errorf("FetchPublic() expected 1 result, got %d instead", len(res))
	}

	if err := dbutil.CompareRows(res[0], testEntry); err != nil {
		t.Errorf("Read and persisted mismatch: %v", err)
	}
}

func TestFetchAll(t *testing.T) {
	var count int

	lite.conn.QueryRow("SELECT count(*) FROM `databases`").Scan(&count)

	entries, err := lite.FetchAll()
	if err != nil {
		t.Errorf("FetchAll() encountered error: %v", err)
	}

	if len(entries) != count {
		t.Errorf("Expected size %d, got %d instead", count, len(entries))
	}
}

func TestUpdate(t *testing.T) {
	lite.Insert(&testEntry)

	// We're updating by ID - this should updated the row for "testEntry"
	updatedEntry := data.Row{
		ID:            testEntry.ID,
		DBName:        "updatedtestDB",
		DBUser:        "updatedtestUser",
		DBPass:        "updatedtestPass",
		DBSID:         "updatedtestsid",
		Dumpfile:      "updatedtestloc",
		CreateDate:    time.Now().In(gmt),
		ExpiryDate:    time.Now().In(gmt).AddDate(0, 0, 30),
		Creator:       "updatedtest@gmail.com",
		ConnectorName: "updatedysql-55",
		DBAddress:     "updatedlocalhost",
		DBPort:        "updated3306",
		DBVendor:      "updatedsqlite",
		Message:       "updated",
		Status:        200,
	}

	err := lite.Update(&updatedEntry)
	if err != nil {
		t.Errorf("Update(updatedEntry) failed: %v", err)
	}

	readEntry, _ := lite.FetchByID(testEntry.ID)

	if err := dbutil.CompareRows(updatedEntry, readEntry); err != nil {
		t.Errorf("Updated and read entries not the same: %v", err)
	}
}

func TestDelete(t *testing.T) {
	lite.Insert(&testEntry)

	err := lite.Delete(testEntry)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	_, err = lite.FetchByID(testEntry.ID)
	if err == nil {
		t.Errorf("Row not deleted")
	}
}

func TestReadRow(t *testing.T) {
	err := lite.Insert(&testEntry)
	if err != nil {
		t.Errorf("Failed adding a entry: %s", err.Error())
	}

	rows, err := testConn.Query("SELECT * FROM `databases` WHERE id = ?", testEntry.ID)
	if err != nil {
		t.Errorf("Failed querying for entries: %s", err.Error())
	}

	for rows.Next() {
		row, err := dbutil.ReadRows(rows)
		if err != nil {
			t.Errorf("Failed reading row from rows: %s", err.Error())
		}

		if err = dbutil.CompareRows(testEntry, row); err != nil {
			t.Errorf("Persisted and read DBEntry not the same: %s", err.Error())
		}
	}

	// cleanup
	_, err = testConn.Exec("DELETE FROM `databases` WHERE ID = ?", testEntry.ID)
	if err != nil {
		t.Errorf("Could not delete created entry")
	}

	testEntry.ID++
}
