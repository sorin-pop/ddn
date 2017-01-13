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

#### Example request:
Simply navigate to `http://ip-address:port/list-databases`

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

### Drop database
**API endpoint:** POST `/drop-database`
#### Explanation
Used to drop the database. Dropping the database also drops the tablespace, its contents and datafiles (if any) and the user specified in the request.

#### Post request details
|Key|Value|
|---|---|
|database_name|Valid name for a database. Must be non-empty|
|tablespace_name|Valid name for a tablespace. Can be empty (in case db used does not use tablespaces)|
|username|Valid name for user. Must be non-empty|

#### Example Curl request
```
curl -X POST -d '{"database_name":"exampleDatabase", "tablespace_name":"", "username":"exampleUser"}' localhost:7001/drop-database
```
### Heartbeat
**API endpoint:** GET `/heartbeat`
#### Explanation
Used to ping the connector to see if its up. Returns immediately with information on whether the database is up or not.
#### Example request:
Simply navigate to `http://ip-address:port/heartbeat`

### Whoami
**API endpoint:** GET `/whoami`
#### Explanation
Used to query information about the connector, and its capabilities. Returns immediately with the following information:
1. Configured database server vendor
2. Configured database server version



#### Example request:
Simply navigate to `http://ip-address:port/heartbeat`
