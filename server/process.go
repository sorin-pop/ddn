package main

import (
	"log"
	"time"

	"github.com/djavorszky/ddn/common/status"
)

// maintain runs each day and checks the databases about when they will expire.
//
// If they expire within 7 days, an email is sent. If they expire the next day,
// another email is sent.
//
// If they are expired, then they are dropped.
//Second
// Maintain should always be ran in a goroutine.
func maintain() {
	ticker := time.NewTicker(24 * time.Hour)

	for range ticker.C {
		dbs, err := db.list()
		if err != nil {
			log.Printf("Failed listing databases: %s", err.Error())
		}

		for _, dbe := range dbs {
			now := time.Now()

			// if expired
			if dbe.ExpiryDate.Before(now) {
				conn, ok := registry[dbe.ConnectorName]
				if !ok {
					log.Printf("Wanted to drop database %q but its connector %q is offline", dbe.DBName, dbe.ConnectorName)
					continue
				}

				conn.DropDatabase(getID(), dbe.DBName, dbe.DBUser)
				db.delete(int64(dbe.ID))
			}

			// if expires within a week:
			weekPlus := now.AddDate(0, 0, 7)
			if dbe.ExpiryDate.Before(weekPlus) {
				db.updateColumn(dbe.ID, "status", status.RemovalScheduled)
			}
		}
	}

}
