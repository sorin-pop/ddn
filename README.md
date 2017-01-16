# Distributed Database Network Connector

Distributed Database Network Connector, or ddnc for short, is a minimal JSON REST API server to run on a virtual or physical machine which has one or more database servers installed. The purpose of the DDN Connector is to act as a unified interface between the outside world and the database server to handle request to create a database / schema along with a connecting user and tablespace (whichever makes sense for the configured database), as well as to list databases / schemas, drop them, and finally, to create a database from a previously provided dump.

## Supported Database Vendors
1. MySQL

### Tested versions:
1. MySQL "5.5.53"

## Configuration
On first run, the connector will generate a `ddnc.properties` file with dummy values inside that provides some needed configuration for the connector. The default configuration can be found below:
```
vendor="mysql"
version="5.5.53"
executable="/usr/bin/mysql"
dbport="3306"
connectorPort="7000"
username="root"
password="root"
masterAddress="127.0.0.1"
```
Currently, only `vendor`, `username`, `password` and `connectorPort` are used for configuration. `vendor` and `version` are also returned when the api `/whoami` is called.


## API
### List databases
**API endpoint:** GET `/list-databases`
#### Explanation
Used to list the available databases in a JSON format. Each entry in the list should contain the name of the database. Database server specific system databases should be omitted from the list.

#### Example request
Simply navigate to `http://ip-address:port/list-databases`

#### Example response
```
{"Status":200,"Message":["61x","61xsamlidp","61xsamlsp","62x","62xcluster"]}
```
### Create database
**API endpoint:** POST `/create-database`
#### Explanation
Used to create a database with a unique database name and tablespace name (if required), as well as a user (with name and password), all provided in the POST request.

Returns an error if:
1. DB Name is not unique (already exists), or invalid
2. Tablespace Name is not unique (already exists), or invalid
3. Username is not unique (already exists), or invalid

Creating a database is a logical step; it encompasses the creation of a database and a connecting user, plus a tablespace, if needed. Privileges are also granted to the created user.

Returns immediately with a success message if creating the database, tablespace and user have all completed successfully.

#### Post request details
|Key|Value|
|---|---|
|database_name|Valid name for a database. Must be non-empty|
|tablespace_name|Valid name for a tablespace. Can be empty (in case db used does not use tablespaces)|
|username|Valid name for user. Must be non-empty|
|password|Valid string for password. Must be non-empty|

#### Example Curl request
```
curl -X POST -d '{"database_name":"exampleDatabase", "tablespace_name":"", "username":"exampleUser", "password":"liferay"}' localhost:7000/create-database
```

#### Example responses
**Success:**
```
{"Status":200,"Message":"Successfully created the database and user!"}
```
**Failures:**
```
{"Status":500,"Message":"Database 'exampleDatabase' already exists"}
```
```
{"Status":500,"Message":"User 'exampleUser' already exists"}
```

### Import database
**API endpoint:** POST `/import-database`
#### Explanation
Starts an import process to import a dumpfile. The dumpfile should be reachable via a non-authenticated HTTP call. The file will be fetched by the Connector, imported, then discarded. A new database (with tablespace and user) will be created, privileges will be granted to the created user. (This is done by calling the `createDatabase` API)

Returns immediately with a Failure if the JSON is malformed, there are missing fields (see below), creating the database and user fails, or the file does not exist. Returns with true immediately if all of the above complete without an issue and starts an import process in the background.

#### Post request details
|Key|Value|
|---|---|
|database_name|Valid name for a database. Must be non-empty|
|tablespace_name|Valid name for a tablespace. Can be empty (in case db used does not use tablespaces)|
|dumpfile_location|Valid location of dumpfile. Must be non-empty.|
|username|Valid name for user. Must be non-empty|
|password|Valid string for password. Must be non-empty|

#### Example Curl request
```
curl -X POST -d '{"database_name":"exampleDatabase", "tablespace_name":"", "username":"exampleUser", "password":"liferay", "dumpfile_location":"http://r2d2.liferay.int/share/route/to/valid/dumpfile.sql"}' localhost:7000/import-database
```

#### Example responses
**Success:**
```
{"Status":200,"Message":"Understood request, starting import process."}
```
**Failures:**
The failures from `/create-database`, plus:
```
{"Status":404,"Message":"Specified file doesn't exist or is not reachable."}
```

### Drop database
**API endpoint:** POST `/drop-database`
#### Explanation
Used to drop the database. Dropping the database also drops the tablespace, its contents and datafiles (if any) and the user specified in the request. It responds with a success even if the database and user do not actually exist.

#### Post request details
|Key|Value|
|---|---|
|database_name|Valid name for a database. Must be non-empty|
|tablespace_name|Valid name for a tablespace. Can be empty (in case db used does not use tablespaces)|
|username|Valid name for user. Must be non-empty|

#### Example Curl request
```
curl -X POST -d '{"database_name":"exampleDatabase", "tablespace_name":"", "username":"exampleUser"}' localhost:7000/drop-database
```
#### Example response
```
{"Status":200,"Message":"Successfully dropped the database and user!"}
```

### Heartbeat
**API endpoint:** GET `/heartbeat`
#### Explanation
Used to ping the connector to see if its up. Returns immediately with information on whether the database is up or not.
#### Example request
Simply navigate to `http://ip-address:port/heartbeat`

#### Example responses
**Success:**
```
{"Status":200,"Message":"Still alive"}
```
**Database down:**
```
{"Status":503,"Message":"The server is unable to process requests as the underlying database is down."}
```

### Whoami
**API endpoint:** GET `/whoami`
#### Explanation
Used to query information about the connector, and its capabilities. Returns immediately with the following information:
1. Configured database server vendor
2. Configured database server version

#### Example request:
Simply navigate to `http://ip-address:port/whoami`

#### Example responses:
```
{"Status":200,"Message":{"vendor":"mysql","version":"5.5.53"}}
```

### Additional information
If there are required fields and one or more of them are missing from the request, the following is sent back:
```
{"Status":400,"Message":"One or more required fields are missing from the call"}
```

If the a malformed request is sent (not JSON or invalid JSON):
```
{"Status":400,"Message":"Invalid JSON request, received error: unexpected EOF"}
```
```
{"Status":400,"Message":"Invalid JSON request, received error: invalid character 's' looking for beginning of value"}
```