package main

import (
	"fmt"

	"github.com/djavorszky/ddn/common/model"
)

func jdbcClassName(vendor string) string {
	var class string

	switch vendor {
	case "oracle":
		class = "oracle.jdbc.driver.OracleDriver"
	case "mysql":
		class = "com.mysql.jdbc.Driver"
	case "postgres":
		class = "org.postgresql.Driver"
	}

	return class
}

func ojdbcURL(sid, address, port string) string {
	return fmt.Sprintf("jdbc:oracle:thin:@%s:%s:%s", address, port, sid)
}

func mjdbcURL(dbname, version, address, port string) string {
	var url string
	switch version {
	case "6210", "6.2.10", "6.2", "6.2 EE", "62":
		url = fmt.Sprintf("jdbc:mysql://%s:%s/%s?useUnicode=true&characterEncoding=UTF-8&useFastDateParsing=false", address, port, dbname)
	case "7010", "7", "DXP", "DE":
		url = fmt.Sprintf("jdbc:mysql://%s:%s/%s?characterEncoding=UTF-8&dontTrackOpenResources=true&holdResultsOpenOverStatementClose=true&useFastDateParsing=false&useUnicode=true", address, port, dbname)
	}

	return url
}

func pjdbcURL(dbname, address, port string) string {
	return fmt.Sprintf("jdbc:postgresql://%s:%s/%s", address, port, dbname)
}

func portalExt(dbentry model.DBEntry, dxp bool) model.PortalExt {
	var ext model.PortalExt

	version := "6210"
	if dxp {
		version = "7"
	}

	var url string
	switch dbentry.DBVendor {
	case "mysql":
		url = mjdbcURL(dbentry.DBName, version, dbentry.DBAddress, dbentry.DBPort)
	case "oracle":
		url = ojdbcURL(dbentry.DBSID, dbentry.DBAddress, dbentry.DBPort)
	case "postgres":
		url = pjdbcURL(dbentry.DBName, dbentry.DBAddress, dbentry.DBPort)
	}

	ext.URL = fmt.Sprintf("jdbc.default.url=%s", url)

	ext.Driver = fmt.Sprintf("jdbc.default.driverClassName=%s", jdbcClassName(dbentry.DBVendor))

	ext.User = fmt.Sprintf("jdbc.default.username=%s", dbentry.DBUser)
	ext.Password = fmt.Sprintf("jdbc.default.password=%s", dbentry.DBPass)

	return ext
}
