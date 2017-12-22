package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/djavorszky/ddn/common/logger"
	"github.com/djavorszky/ddn/common/model"
	"github.com/djavorszky/ddn/server/database/data"
	"github.com/djavorszky/ddn/server/database/dbutil"
	"github.com/djavorszky/sutils"
	webpush "github.com/sherclockholmes/webpush-go"

	// DB
	_ "github.com/mattn/go-sqlite3"
)

// DB implements the BackendConnection
type DB struct {
	DBLocation string

	conn *sql.DB
}

// ConnectAndPrepare establishes a database connection and initializes the tables, if needed
func (lite *DB) ConnectAndPrepare() error {
	conn, err := sql.Open("sqlite3", lite.DBLocation)
	if err != nil {
		return fmt.Errorf("could not open connection: %v", err)
	}

	err = conn.Ping()
	if err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}
	lite.conn = conn

	err = lite.initTables()
	if err != nil {
		return fmt.Errorf("failed updating tables: %v", err)
	}

	return nil
}

// Close closes the database connection
func (lite *DB) Close() error {
	return lite.conn.Close()
}

// FetchByID returns the entry associated with that ID, or
// an error if it does not exist
func (lite *DB) FetchByID(ID int) (data.Row, error) {
	if err := lite.alive(); err != nil {
		return data.Row{}, fmt.Errorf("database down: %s", err.Error())
	}

	row := lite.conn.QueryRow("SELECT * FROM databases WHERE id = ?", ID)
	res, err := dbutil.ReadRow(row)
	if err != nil {
		return data.Row{}, fmt.Errorf("failed reading result: %v", err)
	}

	return res, nil
}

// FetchByDBNameAgent returns the entry for the database with the given name, from the given agent,
// or an error if it does not exist
func (lite *DB) FetchByDBNameAgent(dbname, agent string) (data.Row, error) {
	if err := lite.alive(); err != nil {
		return data.Row{}, fmt.Errorf("database down: %s", err.Error())
	}

	row := lite.conn.QueryRow("SELECT * FROM `databases` WHERE dbname = ? AND agentName = ?", dbname, agent)
	res, err := dbutil.ReadRow(row)
	if err != nil {
		return data.Row{}, fmt.Errorf("failed reading result: %v", err)
	}

	return res, nil
}

// FetchByCreator returns public entries that were created by the
// specified user, an empty list if it's not the user does
// not have any entries, or an error if something went
// wrong
func (lite *DB) FetchByCreator(creator string) ([]data.Row, error) {
	if err := lite.alive(); err != nil {
		return nil, fmt.Errorf("database down: %s", err.Error())
	}

	var entries []data.Row

	rows, err := lite.conn.Query("SELECT * FROM databases WHERE creator = ? AND visibility = 0 ORDER BY id DESC", creator)
	if err != nil {
		return nil, fmt.Errorf("couldn't execute query: %s", err.Error())
	}

	for rows.Next() {
		row, err := dbutil.ReadRows(rows)
		if err != nil {
			return nil, fmt.Errorf("error reading result from query: %s", err.Error())
		}

		entries = append(entries, row)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error reading result from query: %s", err.Error())
	}

	return entries, nil
}

// FetchPublic returns all entries that have "Public" set to true
func (lite *DB) FetchPublic() ([]data.Row, error) {
	if err := lite.alive(); err != nil {
		return nil, fmt.Errorf("database down: %s", err.Error())
	}

	var entries []data.Row

	rows, err := lite.conn.Query("SELECT * FROM `databases` WHERE visibility = 1 ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("failed running query: %v", err)
	}

	for rows.Next() {
		row, err := dbutil.ReadRows(rows)
		if err != nil {
			return nil, fmt.Errorf("error reading result from query: %s", err.Error())
		}

		entries = append(entries, row)
	}

	return entries, nil
}

// FetchAll returns all entries.
func (lite *DB) FetchAll() ([]data.Row, error) {
	if err := lite.alive(); err != nil {
		return nil, fmt.Errorf("database down: %s", err.Error())
	}

	var entries []data.Row

	rows, err := lite.conn.Query("SELECT * FROM `databases` ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("failed running query: %v", err)
	}

	for rows.Next() {
		row, err := dbutil.ReadRows(rows)
		if err != nil {
			return nil, fmt.Errorf("error reading result from query: %s", err.Error())
		}

		entries = append(entries, row)
	}

	return entries, nil
}

// FetchUserPushSubscriptions fetches the subscriptions for the specified user
func (lite *DB) FetchUserPushSubscriptions(subscriber string) ([]webpush.Subscription, error) {
	if err := lite.alive(); err != nil {
		return nil, fmt.Errorf("database down: %s", err.Error())
	}

	if !sutils.Present(subscriber) {
		return nil, fmt.Errorf("missing subscriber")
	}

	var entries []webpush.Subscription

	rows, err := lite.conn.Query("SELECT endpoint, p256dh_key, auth_key FROM `push_subscriptions` WHERE subscriber = ?", subscriber)
	if err != nil {
		return nil, fmt.Errorf("couldn't execute query: %s", err.Error())
	}
	// FetchAll returns all entries.

	for rows.Next() {
		row, err := dbutil.ReadSubscriptionRows(rows)
		if err != nil {
			return nil, fmt.Errorf("error reading result from query: %s", err.Error())
		}

		entries = append(entries, row)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error reading result from query: %s", err.Error())
	}

	return entries, nil
}

// Insert adds an entry to the database, returning its ID
func (lite *DB) Insert(row *data.Row) error {
	if err := lite.alive(); err != nil {
		return fmt.Errorf("database down: %s", err.Error())
	}

	query := "INSERT INTO `databases` (`dbname`, `dbuser`, `dbpass`, `dbsid`, `dumpfile`, `createDate`, `expiryDate`, `creator`, `agentName`, `dbAddress`, `dbPort`, `dbvendor`, `status`, `message`, `visibility`, `comment`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	res, err := lite.conn.Exec(query,
		row.DBName,
		row.DBUser,
		row.DBPass,
		row.DBSID,
		row.Dumpfile,
		row.CreateDate,
		row.ExpiryDate,
		row.Creator,
		row.AgentName,
		row.DBAddress,
		row.DBPort,
		row.DBVendor,
		row.Status,
		row.Message,
		row.Public,
		row.Comment,
	)
	if err != nil {
		return fmt.Errorf("insert failed: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed getting new ID: %v", err)
	}

	row.ID = int(id)

	return nil
}

// InsertPushSubscription adds a record to the push_subscriptions table
func (lite *DB) InsertPushSubscription(subscription *model.PushSubscription, subscriber string) error {
	if err := lite.alive(); err != nil {
		return fmt.Errorf("database down: %s", err.Error())
	}

	if !sutils.Present(subscriber) {
		return fmt.Errorf("missing subscriber")
	}

	if !sutils.Present(subscription.Endpoint) {
		return fmt.Errorf("missing endpoint")
	}

	query := "INSERT INTO `push_subscriptions` (`subscriber`, `endpoint`, `p256dh_key`, `auth_key`) VALUES (?, ?, ?, ?)"

	_, err := lite.conn.Exec(query,
		subscriber,
		subscription.Endpoint,
		subscription.Keys.P256dh,
		subscription.Keys.Auth,
	)
	if err != nil {
		return fmt.Errorf("saving push subscription to the database failed: %v", err)
	}

	return nil
}

// Update updates an already existing entry
func (lite *DB) Update(entry *data.Row) error {
	if err := lite.alive(); err != nil {
		return fmt.Errorf("database down: %s", err.Error())
	}

	var count int

	err := lite.conn.QueryRow("SELECT count(*) FROM `databases` WHERE id = ?", entry.ID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed existence check: %v", err)
	}

	if count == 0 {
		return lite.Insert(entry)
	}

	query := "UPDATE `databases` SET `dbname`= ?, `dbuser`= ?, `dbpass`= ?, `dbsid`= ?, `dumpfile`= ?, `createDate`= ?, `expiryDate`= ?, `creator`= ?, `agentName`= ?, `dbAddress`= ?, `dbPort`= ?, `dbvendor`= ?, `status`= ?, `message`= ?, `visibility`= ?, `comment` = ? WHERE id = ?"

	_, err = lite.conn.Exec(query,
		entry.DBName,
		entry.DBUser,
		entry.DBPass,
		entry.DBSID,
		entry.Dumpfile,
		entry.CreateDate,
		entry.ExpiryDate,
		entry.Creator,
		entry.AgentName,
		entry.DBAddress,
		entry.DBPort,
		entry.DBVendor,
		entry.Status,
		entry.Message,
		entry.Public,
		entry.Comment,
		entry.ID,
	)
	if err != nil {
		return fmt.Errorf("failed update: %v", err)
	}

	return nil
}

// Delete removes the entry from the database
func (lite *DB) Delete(entry data.Row) error {
	if err := lite.alive(); err != nil {
		return fmt.Errorf("database down: %s", err.Error())
	}

	_, err := lite.conn.Exec("DELETE FROM `databases` WHERE id = ?", entry.ID)

	return err
}

// DeletePushSubscription deletes a record from the push_subscriptions table
func (lite *DB) DeletePushSubscription(subscription *model.PushSubscription, subscriber string) error {
	if err := lite.alive(); err != nil {
		return fmt.Errorf("database down: %s", err.Error())
	}

	if !sutils.Present(subscriber) {
		return fmt.Errorf("missing subscriber")
	}

	if !sutils.Present(subscription.Endpoint) {
		return fmt.Errorf("missing endpoint")
	}

	_, err := lite.conn.Exec("DELETE FROM `push_subscriptions` WHERE subscriber = ? AND endpoint = ?", subscriber, subscription.Endpoint)

	return err
}

type dbUpdate struct {
	Query   string
	Comment string
}

var queries = []dbUpdate{
	{
		Query:   "CREATE TABLE version (queryId INTEGER PRIMARY KEY, query TEXT NULL, comment TEXT NULL, date DATETIME NULL);",
		Comment: "Create the version table",
	},
	{
		Query:   "CREATE TABLE databases (id INTEGER PRIMARY KEY, dbname VARCHAR(255) NULL, dbuser VARCHAR(255) NULL, dbpass VARCHAR(255) NULL, dbsid VARCHAR(45) NULL, dumpfile TEXT NULL, createDate DATETIME NULL, expiryDate DATETIME NULL, creator VARCHAR(255) NULL, connectorName VARCHAR(255) NULL, dbAddress VARCHAR(255) NULL, dbPort VARCHAR(45) NULL, dbvendor VARCHAR(255) NULL, status INTEGER, message TEXT, visibility INTEGER DEFAULT 0);",
		Comment: "Create the databases table",
	},
	{
		Query:   "UPDATE databases SET message = '' WHERE message IS NULL;",
		Comment: "Update 'message' columns to empty where null",
	},
	{
		Query:   "ALTER TABLE databases RENAME TO databases_tmp",
		Comment: "Update column in 'databases': Create temp table",
	},
	{
		Query:   "CREATE TABLE databases (id INTEGER PRIMARY KEY AUTOINCREMENT, dbname VARCHAR(255) NULL, dbuser VARCHAR(255) NULL, dbpass VARCHAR(255) NULL, dbsid VARCHAR(45) NULL, dumpfile TEXT NULL, createDate DATETIME NULL, expiryDate DATETIME NULL, creator VARCHAR(255) NULL, agentName VARCHAR(255) NULL, dbAddress VARCHAR(255) NULL, dbPort VARCHAR(45) NULL, dbvendor VARCHAR(255) NULL, status INTEGER, message TEXT, visibility INTEGER DEFAULT 0);",
		Comment: "Update column in 'databases': Create updated table",
	},
	{
		Query:   "INSERT INTO databases SELECT id, dbname, dbuser, dbpass, dbsid, dumpfile, createDate, expiryDate, creator, connectorName AS agentName, dbAddress, dbPort, dbvendor, status, message, visibility FROM databases_tmp;",
		Comment: "Update column in 'databases': Insert data to updated table",
	},
	{
		Query:   "DROP TABLE databases_tmp;",
		Comment: "Update column in 'databases': Drop temp table",
	},
	{
		Query:   "CREATE TABLE `push_subscriptions` ( `subscriber` VARCHAR(255) NOT NULL, `endpoint` VARCHAR(255) NOT NULL, `p256dh_key` VARCHAR(255) NOT NULL, `auth_key` VARCHAR(255) NOT NULL);",
		Comment: "Create the push_subscriptions table",
	},
	{
		Query:   "CREATE UNIQUE INDEX IF NOT EXISTS `push_subscription` ON `push_subscriptions` (`subscriber`, `endpoint`);",
		Comment: "Create unique index on columns (subscriber, endpoint) for table push_subscriptions",
	},
	{
		Query:   "ALTER TABLE `databases` ADD COLUMN `comment`;",
		Comment: "Add 'comment' column",
	},
	{
		Query:   "UPDATE databases SET comment = '' WHERE comment IS NULL;",
		Comment: "Update 'comment' columns to empty where null",
	},
	{
		Query:   "CREATE UNIQUE INDEX IF NOT EXISTS `agent_db_idx` ON `databases` (`dbname`, `agentName`);",
		Comment: "Create unique index on columns (dbname, agentName) for table databases",
	},
}

func (lite *DB) initTables() error {
	var (
		err      error
		startLoc int
	)

	lite.conn.QueryRow("SELECT count(*) FROM version").Scan(&startLoc)

	for _, q := range queries[startLoc:] {
		logger.Info("Updating database %q", q.Comment)
		_, err = lite.conn.Exec(q.Query)
		if err != nil {
			return fmt.Errorf("executing query %q (%q) failed: %s", q.Comment, q.Query, sutils.TrimNL(err.Error()))
		}

		_, err = lite.conn.Exec("INSERT INTO version (query, comment, date) VALUES (?, ?, ?)", q.Query, q.Comment, time.Now())
		if err != nil {
			return fmt.Errorf("updating version table with query %q (%q) failed: %s", q.Comment, q.Query, sutils.TrimNL(err.Error()))
		}
	}

	return nil
}

// Alive checks whether the connection is alive. Returns error if not.
func (lite *DB) alive() error {
	defer func() {
		if p := recover(); p != nil {
			logger.Error("Panic Attack! Database seems to be down.")
		}
	}()

	_, err := lite.conn.Exec("select * from databases WHERE 1 = 0")
	if err != nil {
		return fmt.Errorf("executing stayalive query failed: %s", sutils.TrimNL(err.Error()))
	}

	return nil
}
