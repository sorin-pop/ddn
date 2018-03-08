package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/server/database/data"
	"github.com/djavorszky/ddn/server/database/dbutil"
	_ "github.com/mattn/go-sqlite3"
	webpush "github.com/sherclockholmes/webpush-go"
)

const (
	testDBFile = "./test.db"
)

var (
	testConn *sql.DB
	lite     DB

	gmt, _ = time.LoadLocation("GMT")
)

func getTestEntry(agentName, databaseName string) data.Row {
	entry := data.Row{
		ID:         1,
		DBName:     databaseName,
		DBUser:     "testUser",
		DBPass:     "testPass",
		DBSID:      "testsid",
		Dumpfile:   "testloc",
		CreateDate: time.Now().In(gmt),
		ExpiryDate: time.Now().In(gmt).AddDate(0, 0, 30),
		Creator:    "test@gmail.com",
		AgentName:  agentName,
		DBAddress:  "localhost",
		DBPort:     "3306",
		DBVendor:   "mysql",
		Message:    "",
		Status:     100,
		Comment:    "Just some random comment",
	}

	return entry
}

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
	insertTest := getTestEntry("testInsert", "firstDB")
	err := lite.Insert(&insertTest)
	if err != nil {
		t.Errorf("lite.Insert(insertTest) failed with error: %v", err)
		return
	}

	if insertTest.ID == 0 {
		t.Errorf("lite.Insert(insertTest) resulted in id of 0")
		return
	}

	result, err := lite.FetchByID(insertTest.ID)
	if err != nil {
		t.Errorf("FetchById(%d) resulted in error: %v", insertTest.ID, err)
		return
	}

	if err = dbutil.CompareRows(insertTest, result); err != nil {
		t.Errorf("Persisted and read results not the same: %v", err)
	}
}

func TestFetchByID(t *testing.T) {
	fetchTest := getTestEntry("testFetch", "secondDB")

	err := lite.Insert(&fetchTest)
	if err != nil {
		t.Errorf("Insert failed: %v", err)
		return
	}

	res, err := lite.FetchByID(fetchTest.ID)
	if err != nil {
		t.Errorf("FetchById(%d) failed with error: %v", fetchTest.ID, err)
		return
	}

	if err := dbutil.CompareRows(res, fetchTest); err != nil {
		t.Errorf("Fetched result not the same as queried: %v", err)
	}
}

func TestDupInsert(t *testing.T) {
	insertTest := getTestEntry("testDupInsert", "quack")
	err := lite.Insert(&insertTest)
	if err != nil {
		t.Errorf("lite.Insert(insertTest) failed with error: %v", err)
		return
	}

	err = lite.Insert(&insertTest)
	if err == nil {
		t.Errorf("Second lite.Insert(insertTest) should have failed.")
		return
	}
}

func TestFetchByDBNameAgent(t *testing.T) {
	fetchTest2 := getTestEntry("testFetch2", "thirdDB")
	err := lite.Insert(&fetchTest2)
	if err != nil {
		t.Errorf("Insert failed: %v", err)
		return
	}

	res, err := lite.FetchByDBNameAgent(fetchTest2.DBName, fetchTest2.AgentName)
	if err != nil {
		t.Errorf("FetchByDBNameAgent(%s, %s) failed with error: %v", fetchTest2.DBName, fetchTest2.AgentName, err)
		return
	}

	if err := dbutil.CompareRows(res, fetchTest2); err != nil {
		t.Errorf("Fetched result not the same as queried: %v", err)
	}
}

func TestFetchByCreator(t *testing.T) {
	creator := "someone@somewhere.com"

	creatorFetch := getTestEntry("creatorFetch", "uniqueDB")
	creatorFetch.Creator = creator
	lite.Insert(&creatorFetch)

	creatorFetch.DBName = "fetchByCreator_2"
	lite.Insert(&creatorFetch)

	results, err := lite.FetchByCreator(creator)
	if err != nil {
		t.Errorf("failed to fetch by creator: %v", err)
		return
	}

	if len(results) != 2 {
		t.Errorf("Expected resultset to have 2 results, %d instead", len(results))
		return
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
		return
	}

	if len(res) != 0 {
		t.Errorf("FetchPublic() returned with entries, shouldn't have")
		return
	}

	fetchPublic := getTestEntry("fetchByPublic", "someDBName")

	fetchPublic.Public = 1
	lite.Insert(&fetchPublic)

	res, err = lite.FetchPublic()
	if err != nil {
		t.Errorf("FetchPublic() error: %v", err)
		return
	}

	if len(res) != 1 {
		t.Errorf("FetchPublic() expected 1 result, got %d instead", len(res))
		return
	}

	if err := dbutil.CompareRows(res[0], fetchPublic); err != nil {
		t.Errorf("Read and persisted mismatch: %v", err)
		return
	}
}

func TestFetchAll(t *testing.T) {
	var count int

	lite.conn.QueryRow("SELECT count(*) FROM `databases`").Scan(&count)

	entries, err := lite.FetchAll()
	if err != nil {
		t.Errorf("FetchAll() encountered error: %v", err)
		return
	}

	if len(entries) != count {
		t.Errorf("Expected size %d, got %d instead", count, len(entries))
	}
}

func TestUpdate(t *testing.T) {
	testUpdate := getTestEntry("update", "ununun")
	err := lite.Insert(&testUpdate)
	if err != nil {
		t.Errorf("Insert failed: %v", err)
		return
	}

	// We're updating by ID - this should updated the row for "testUpdate"
	updatedEntry := data.Row{
		ID:         testUpdate.ID,
		DBName:     "updatedtestDB",
		DBUser:     "updatedtestUser",
		DBPass:     "updatedtestPass",
		DBSID:      "updatedtestsid",
		Dumpfile:   "updatedtestloc",
		CreateDate: time.Now().In(gmt),
		ExpiryDate: time.Now().In(gmt).AddDate(0, 0, 30),
		Creator:    "updatedtest@gmail.com",
		AgentName:  "updatedysql-55",
		DBAddress:  "updatedlocalhost",
		DBPort:     "updated3306",
		DBVendor:   "updatedsqlite",
		Message:    "updated",
		Status:     200,
		Comment:    "Something else I suppose",
	}

	err = lite.Update(&updatedEntry)
	if err != nil {
		t.Errorf("Update(updatedEntry) failed: %v", err)
		return
	}

	readEntry, _ := lite.FetchByID(testUpdate.ID)

	if err := dbutil.CompareRows(updatedEntry, readEntry); err != nil {
		t.Errorf("Updated and read entries not the same: %v", err)
	}
}

func TestDelete(t *testing.T) {
	testDelete := getTestEntry("delete", "testdatabase")
	err := lite.Insert(&testDelete)
	if err != nil {
		t.Errorf("Insert failed: %v", err)
		return
	}

	err = lite.Delete(testDelete)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
		return
	}

	row, _ := lite.FetchByID(testDelete.ID)
	if row.ID == testDelete.ID {
		t.Errorf("Row was not deleted, managed to fetch it back")
	}
}

func TestReadRow(t *testing.T) {
	readTest := getTestEntry("readtest", "somethingoranother")
	err := lite.Insert(&readTest)
	if err != nil {
		t.Errorf("Failed adding an entry: %s", err.Error())
		return
	}

	rows, err := testConn.Query("SELECT * FROM `databases` WHERE id = ?", readTest.ID)
	if err != nil {
		t.Errorf("Failed querying for entries: %s", err.Error())
		return
	}

	for rows.Next() {
		row, err := dbutil.ReadRows(rows)
		if err != nil {
			t.Errorf("Failed reading row from rows: %s", err.Error())
			return
		}

		if err = dbutil.CompareRows(readTest, row); err != nil {
			t.Errorf("Persisted and read DBEntry not the same: %s", err.Error())
		}
	}
}

func TestInsertPushSubscription(t *testing.T) {
	type args struct {
		subscription *model.PushSubscription
		subscriber   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Success", args{
			subscriber: "test@example.com",
			subscription: &model.PushSubscription{
				Endpoint:       "testEndpoint",
				ExpirationTime: "testExpirationTime",
				Keys: webpush.Keys{
					P256dh: "randomTestKey",
					Auth:   "randomTestAuth",
				},
			},
		}, false},
		{"Missing Subscriber", args{
			subscription: &model.PushSubscription{
				Endpoint:       "testEndpoint",
				ExpirationTime: "testExpirationTime",
				Keys: webpush.Keys{
					P256dh: "randomTestKey",
					Auth:   "randomTestAuth",
				},
			},
		}, true},
		{"Missing Endpoint", args{
			subscriber: "test@example.com",
			subscription: &model.PushSubscription{
				ExpirationTime: "testExpirationTime",
				Keys: webpush.Keys{
					P256dh: "randomTestKey",
					Auth:   "randomTestAuth",
				},
			},
		}, true},
		{"Missing ExpirationTime", args{
			subscriber: "test@example.com",
			subscription: &model.PushSubscription{
				Endpoint: "testEndpoint",
				Keys: webpush.Keys{
					P256dh: "randomTestKey",
					Auth:   "randomTestAuth",
				},
			},
		}, true},
		{"Missing Keys", args{
			subscriber: "test@example.com",
			subscription: &model.PushSubscription{
				Endpoint:       "testEndpoint",
				ExpirationTime: "testExpirationTime",
			},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lite := &DB{
				DBLocation: testDBFile,
				conn:       testConn,
			}
			if err := lite.InsertPushSubscription(tt.args.subscription, tt.args.subscriber); (err != nil) != tt.wantErr {
				t.Errorf("DB.InsertPushSubscription() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			read, _ := lite.FetchUserPushSubscriptions(tt.args.subscriber)
			if len(read) == 0 {
				t.Errorf("Did not find inserted data after DB.InsertPushSubscription")
				return
			}

			s := read[0]
			if s.Endpoint != tt.args.subscription.Endpoint {
				t.Errorf("endpoint mismatch; expected %v, got %v", tt.args.subscription.Endpoint, s.Endpoint)
			}

			if s.Keys.Auth != tt.args.subscription.Keys.Auth {
				t.Errorf("auth mismatch; expected %v, got %v", tt.args.subscription.Keys.Auth, s.Keys.Auth)
			}

			if s.Keys.P256dh != tt.args.subscription.Keys.P256dh {
				t.Errorf("P256Dh mismatch; expected %v, got %v", tt.args.subscription.Keys.P256dh, s.Keys.P256dh)
			}
		})
	}
}

func TestFetchUserPushSubscriptions(t *testing.T) {
	testUser := "test@example.com"
	testSubscription := &model.PushSubscription{
		Endpoint:       "testEndpoint",
		ExpirationTime: "testExpirationTime",
		Keys: webpush.Keys{
			P256dh: "randomTestKey",
			Auth:   "randomTestAuth",
		},
	}

	tests := []struct {
		name          string
		subscriber    string
		expectedCount int
		wantErr       bool
	}{
		{"Success", testUser, 1, false},
		{"No subscription for user", "random@user.com", 0, false},
		{"No user specified", "", 0, true},
	}

	lite.InsertPushSubscription(testSubscription, testUser)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				read []webpush.Subscription
				err  error
			)

			lite := &DB{
				DBLocation: testDBFile,
				conn:       testConn,
			}

			if read, err = lite.FetchUserPushSubscriptions(tt.subscriber); (err != nil) != tt.wantErr {
				t.Errorf("DB.FetchUserPushSubscriptions() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr || tt.expectedCount == 0 {
				return
			}

			if len(read) != tt.expectedCount {
				t.Errorf("Wrong number of results returned. Expected %v, got %v", tt.expectedCount, len(read))
				return
			}

			s := read[0]
			if s.Endpoint != testSubscription.Endpoint {
				t.Errorf("endpoint mismatch; expected %v, got %v", testSubscription.Endpoint, s.Endpoint)
			}

			if s.Keys.Auth != testSubscription.Keys.Auth {
				t.Errorf("auth mismatch; expected %v, got %v", testSubscription.Keys.Auth, s.Keys.Auth)
			}

			if s.Keys.P256dh != testSubscription.Keys.P256dh {
				t.Errorf("P256Dh mismatch; expected %v, got %v", testSubscription.Keys.P256dh, s.Keys.P256dh)
			}
		})
	}
}

func TestDeleteUserPushNotification(t *testing.T) {
	testUser := "test@example.com"
	testSubscription := &model.PushSubscription{
		Endpoint:       "testEndpoint",
		ExpirationTime: "testExpirationTime",
		Keys: webpush.Keys{
			P256dh: "randomTestKey",
			Auth:   "randomTestAuth",
		},
	}

	tests := []struct {
		name       string
		subscriber string
		endpoint   string
		wantErr    bool
	}{
		{"Success", testUser, testSubscription.Endpoint, false},
		{"No Subscription for user", "random@user.com", testSubscription.Endpoint, false},
		{"No User specified", "", testSubscription.Endpoint, true},
		{"No Endpoint specified", testUser, "", true},
	}

	lite.InsertPushSubscription(testSubscription, testUser)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lite := &DB{
				DBLocation: testDBFile,
				conn:       testConn,
			}

			tmpSub := &model.PushSubscription{Endpoint: tt.endpoint}

			if err := lite.DeletePushSubscription(tmpSub, tt.subscriber); (err != nil) != tt.wantErr {
				t.Errorf("DB.FetchUserPushSubscriptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
