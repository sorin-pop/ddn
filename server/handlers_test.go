package main

import "testing"

func Test_ensureValues(t *testing.T) {
	type args struct {
		dbname string
		dbuser string
		dbpass string
		vendor string
	}
	tests := []struct {
		name string
		args args
	}{
		{"allpresent", args{"dbname", "dbuser", "dbpass", "mysql"}},
		{"mssql_allempty", args{"", "", "", "mssql"}},
		{"allempty", args{"", "", "", ""}},
		{"dbname_empty", args{"", "dbuser", "dbpass", "mssql"}},
		{"dbuser_empty", args{"dbname", "", "dbpass", "mssql"}},
		{"dbpass_empty", args{"dbname", "dbname", "", "mssql"}},
		{"dbname_dbuser_empty", args{"", "", "dbpass", "mssql"}},
		{"dbuser_dbpass_empty", args{"dbname", "", "", "mssql"}},
		{"dbname_dbpass_empty", args{"", "dbname", "", "mssql"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ensureValues(&tt.args.dbname, &tt.args.dbuser, &tt.args.dbpass, tt.args.vendor)
		})

		if tt.args.dbname == "" {
			t.Errorf("%s: dbname remained empty", tt.name)
		}

		if tt.args.dbuser == "" {
			t.Errorf("%s: dbuser remained empty", tt.name)
		}

		if tt.args.dbpass == "" {
			t.Errorf("%s: dbpass remained empty", tt.name)
		}
	}
}
