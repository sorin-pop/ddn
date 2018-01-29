package main

import (
	"fmt"
	"time"

	"github.com/djavorszky/ddn/common/inet"
	"github.com/djavorszky/ddn/common/logger"
	"github.com/djavorszky/ddn/common/status"
	"github.com/djavorszky/ddn/server/mail"
	"github.com/djavorszky/ddn/server/registry"
)

// maintain runs each day and checks the databases about when they will expire.
//
// If they expire within 7 days, an email is sent. If they expire the next day,
// another email is sent.
//
// If they are expired, then they are dropped.
//
// Maintain should always be ran in a goroutine.
func maintain() {
	ticker := time.NewTicker(24 * time.Hour)

	for range ticker.C {
		dbs, err := db.FetchAll()
		if err != nil {
			logger.Error("Failed listing databases: %s", err.Error())
		}

		for _, dbe := range dbs {
			now := time.Now()

			// if expired
			if (dbe.ExpiryDate.Year() == now.Year()) && (dbe.ExpiryDate.YearDay() == now.YearDay()) {
				agent, ok := registry.Get(dbe.AgentName)
				if !ok {
					logger.Error("drop database %q - agent %q offline", dbe.DBName, dbe.AgentName)
					continue
				}

				_, err = agent.DropDatabase(registry.ID(), dbe.DBName, dbe.DBUser)
				if err != nil {
					dbe.Status = status.DropDatabaseFailed
					dbe.Message = err.Error()
					db.Update(&dbe)

					logger.Error("failed dropping database: %v", err)
					continue
				}
				db.Delete(dbe)

				mail.Send(dbe.Creator, fmt.Sprintf("[Cloud DB] Database %q dropped", dbe.DBName), fmt.Sprintf(`
<h3>Database dropped</h3>
				
<p>This is to inform you that the database %q has been dropped.</p>
<p>Thank you for using <a href="http://cloud-db.liferay.int">Cloud DB</a>.</p>`, dbe.DBName))

				err = sendUserNotifications(dbe.Creator, fmt.Sprintf("Database %s has been dropped.", dbe.DBName))
				if err != nil {
					logger.Error("failed notifying user: %v", err)
				}

				continue
			}

			// if expires within a day:
			// Note, not adding a check to see if an email has been sent
			// already, as these are only checked once per day, meaning,
			// on the next check the expiry date will be in the past.
			dayPlus := now.AddDate(0, 0, 1)
			if dbe.ExpiryDate.Before(dayPlus) {
				mail.Send(dbe.Creator, fmt.Sprintf("[Cloud DB] Database %q to be removed in 1 day", dbe.DBName), fmt.Sprintf(`
<h3>Database removal imminent</h3>
				
<p>This is to inform you that the database %q will be removed in one day.</p>
<p>If you'd like to extend it, please visit <a href="http://cloud-db.liferay.int">Cloud DB</a>.</p>
<p>Cheers</p>`, dbe.DBName))

				continue
			}

			if dbe.Status == status.RemovalScheduled || dbe.Status == status.ImportFailed {
				continue
			}

			// if expires within a week:
			weekPlus := now.AddDate(0, 0, 7)
			if dbe.ExpiryDate.Before(weekPlus) {
				dbe.Status = status.RemovalScheduled

				db.Update(&dbe)

				mail.Send(dbe.Creator, fmt.Sprintf("[Cloud DB] Database %q to be removed in one week", dbe.DBName), fmt.Sprintf(`
<h3>Database removal scheduled</h3>
				
<p>This is to inform you that the database %q will be removed in 7 days.</p>
<p>If you'd like to extend it, please visit <a href="http://cloud-db.liferay.int">Cloud DB</a>.</p>
<p>Cheers</p>`, dbe.DBName))

				err = sendUserNotifications(dbe.Creator, fmt.Sprintf("Database %s to be removed in one week.", dbe.DBName))
				if err != nil {
					logger.Error("failed notifying user: %v", err)
				}
			}
		}
	}
}

// checkAgents checks whether the registered agents are alive or not.
// If they are not alive, it'll update their status.
func checkAgents() {
	ticker := time.NewTicker(30 * time.Second)

	for range ticker.C {
		for _, agent := range registry.List() {
			addr := fmt.Sprintf("%s:%s/heartbeat", agent.Address, agent.AgentPort)

			if !inet.AddrExists(addr) && agent.Up {
				agent.Up = false

				registry.Store(agent)

				for _, addr := range config.AdminEmail {
					mail.Send(addr, "[Cloud DB] Agent disappeared without trace",
						fmt.Sprintf("Agent %q at %q no longer exists.", agent.ShortName, agent.Address))
				}

				continue
			}

			if !agent.Up && inet.AddrExists(addr) {
				agent.Up = true

				registry.Store(agent)
			}
		}
	}
}
