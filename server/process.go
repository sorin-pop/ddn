package main

import (
	"fmt"
	"log"
	"time"

	gomail "gopkg.in/gomail.v2"

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
	ticker := time.NewTicker(1000 * time.Millisecond)

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

				sendMail(dbe.Creator, fmt.Sprintf("[Cloud DB] Database %q dropped", dbe.DBName), fmt.Sprintf(`
<h3>Database removal scheduled</h3>
				
<p>This is to inform you that the database %q has been dropped.</p>
<p>Thank you for using <a href="http://cloud-db.liferay.int">Cloud DB</a>.</p>`, dbe.DBName))

				continue
			}

			// if expires within a day:
			// Note, not adding a check to see if an email has been sent
			// already, as these are only checked once per day, meaning,
			// on the next check the expiry date will be in the past.
			dayPlus := now.AddDate(0, 0, 1)
			if dbe.ExpiryDate.Before(dayPlus) {
				sendMail(dbe.Creator, fmt.Sprintf("[Cloud DB] Database %q to be removed in 1 day", dbe.DBName), fmt.Sprintf(`
<h3>Database removal imminent</h3>
				
<p>This is to inform you that the database %q will be removed in one day.</p>
<p>If you'd like to extend it, please visit <a href="http://cloud-db.liferay.int">Cloud DB</a>.</p>
<p>Cheers</p>`, dbe.DBName))

				continue
			}

			if dbe.Status == status.RemovalScheduled {
				continue
			}

			// if expires within a week:
			weekPlus := now.AddDate(0, 0, 7)
			if dbe.ExpiryDate.Before(weekPlus) {
				db.updateColumn(dbe.ID, "status", status.RemovalScheduled)

				sendMail(dbe.Creator, fmt.Sprintf("[Cloud DB] Database %q to be removed in one week", dbe.DBName), fmt.Sprintf(`
<h3>Database removal scheduled</h3>
				
<p>This is to inform you that the database %q will be removed in 7 days.</p>
<p>If you'd like to extend it, please visit <a href="http://cloud-db.liferay.int">Cloud DB</a>.</p>
<p>Cheers</p>`, dbe.DBName))
			}
		}
	}
}

// sendMail sends an email to "to" with subject "subj" and body "body".
// It only returns with an error if something went wrong in this process.
//
// If the server is not configured to send an email (e.g. address, port or EmailSender
// is empty, it silently returns)
func sendMail(to, subj, body string) error {
	if config.SMTPAddr == "" || config.SMTPPort == 0 || config.EmailSender == "" {
		log.Println("Returning because not configured to send email.")
		return nil
	}

	m := gomail.NewMessage()

	m.SetHeader("From", config.EmailSender)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subj)

	m.SetBody("text/html", body)

	dialer := gomail.Dialer{
		Host:     config.SMTPAddr,
		Port:     config.SMTPPort,
		Username: config.SMTPUser,
		Password: config.SMTPPass,
	}

	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %s", err.Error())
	}

	return nil
}
