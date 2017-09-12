package dbutil

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/djavorszky/ddn/server/database/data"
)

// CompareRows compares the two entries to see if they are the same or not.
func CompareRows(first, second data.Row) error {
	if first.ID != second.ID {
		return fmt.Errorf("ID mismatch. First: '%d' vs Second: '%d'", first.ID, second.ID)
	}

	if first.DBVendor != second.DBVendor {
		return fmt.Errorf("DBVendor mismatch. First: %q vs Second: %q", first.DBVendor, second.DBVendor)
	}

	if first.DBName != second.DBName {
		return fmt.Errorf("DBName mismatch. First: %q vs Second: %q", first.DBName, second.DBName)
	}

	if first.DBUser != second.DBUser {
		return fmt.Errorf("DBUser mismatch. First: %q vs Second: %q", first.DBUser, second.DBUser)
	}

	if first.DBPass != second.DBPass {
		return fmt.Errorf("DBPass mismatch. First: %q vs Second: %q", first.DBPass, second.DBPass)
	}

	if first.DBSID != second.DBSID {
		return fmt.Errorf("DBSID mismatch. First: %q vs Second: %q", first.DBSID, second.DBSID)
	}

	if first.Dumpfile != second.Dumpfile {
		return fmt.Errorf("Dumpfile mismatch. First: %q vs Second: %q", first.Dumpfile, second.Dumpfile)
	}

	delta := first.CreateDate.Sub(second.CreateDate)
	if delta < -1*time.Second || delta > 1*time.Second {
		return fmt.Errorf("CreateDate mismatch. First: %q vs Second: %q", first.CreateDate.Round(time.Second).Format(time.ANSIC), second.CreateDate.Round(time.Second).Format(time.ANSIC))
	}

	delta = first.ExpiryDate.Sub(second.ExpiryDate)
	if delta < -1*time.Second || delta > 1*time.Second {
		return fmt.Errorf("ExpiryDate mismatch. First: %q vs Second: %q", first.ExpiryDate.Round(time.Second).Format(time.ANSIC), second.ExpiryDate.Round(time.Second).Format(time.ANSIC))
	}

	if first.Creator != second.Creator {
		return fmt.Errorf("Creator mismatch. First: %q vs Second: %q", first.Creator, second.Creator)
	}

	if first.ConnectorName != second.ConnectorName {
		return fmt.Errorf("ConnectorName mismatch. First: %q vs Second: %q", first.ConnectorName, second.ConnectorName)
	}

	if first.DBAddress != second.DBAddress {
		return fmt.Errorf("DBAddress mismatch. First: %q vs Second: %q", first.DBAddress, second.DBAddress)
	}

	if first.DBPort != second.DBPort {
		return fmt.Errorf("DBPort mismatch. First: %q vs Second: %q", first.DBPort, second.DBPort)
	}

	if first.Status != second.Status {
		return fmt.Errorf("Status mismatch. First: %q vs Second: %q", first.Status, second.Status)

	}

	if first.Public != second.Public {
		return fmt.Errorf("Public mismatch. First: %q vs Second: %q", first.Public, second.Public)
	}

	return nil
}

// ReadRow reads an sql.Row into a data.Row
func ReadRow(result *sql.Row) (data.Row, error) {
	var row data.Row

	err := result.Scan(
		&row.ID,
		&row.DBName,
		&row.DBUser,
		&row.DBPass,
		&row.DBSID,
		&row.Dumpfile,
		&row.CreateDate,
		&row.ExpiryDate,
		&row.Creator,
		&row.ConnectorName,
		&row.DBAddress,
		&row.DBPort,
		&row.DBVendor,
		&row.Status,
		&row.Message,
		&row.Public)
	if err != nil {
		return data.Row{}, fmt.Errorf("failed reading row: %v", err)
	}

	return row, nil
}

// ReadRows reads an sql.Rows into a data.Row
func ReadRows(rows *sql.Rows) (data.Row, error) {
	var row data.Row

	err := rows.Scan(
		&row.ID,
		&row.DBName,
		&row.DBUser,
		&row.DBPass,
		&row.DBSID,
		&row.Dumpfile,
		&row.CreateDate,
		&row.ExpiryDate,
		&row.Creator,
		&row.ConnectorName,
		&row.DBAddress,
		&row.DBPort,
		&row.DBVendor,
		&row.Status,
		&row.Message,
		&row.Public)
	if err != nil {
		return data.Row{}, fmt.Errorf("failed reading row: %v", err)
	}

	return row, nil
}
