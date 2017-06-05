# Distributed Database Network Server

Description
-----------

The Distributed Database Network server, or `ddns` for short, is the central server for the whole ddn network. It is an API server with JSON API capabilities with a web UI planned once the API part is released.

The main task of the server is to keep track of all the registered `connectors` (see [Distributed Database Network Connector](https://github.com/djavorszky/ddnc)), as well as to provide an endpoint for all end users to call.

Keeping track of the `connectors` will be done in a local MySQL database. Once a `connector` comes online, it will register itself with the server and provide periodic updates that it is still alive. If the connector goes down, the updates will cease - In this case, the `server` marks the connector as down and removes it from the registry.

Installation
------------

The required frontend dependencies can be installed with running the following command inside the "web" folder:

    npm install

If you want to update these dependencies, just run `npm update` inside the "web" folder

Documentation
-------------

For first release, the following API endpoints are defined:
## List connectors
**API endpoint:** GET `/list-connectors`
### Explanation
Returns a JSON encoded string containing the databases, and their versions, on the available connectors, ordered by their shortname.

### Example request
Simply navigate to `http://ip-address:port/list-connectors`

### Example response
```
{"Status":200,"Message":["mysql-55":"MySQL 5.5.53","mysql-56":"MySQL 5.6.43","mysql-57":"MySQL 5.7.15","postgresql-94":"PostgreSQL 9.4.9","oracle-11g":"Oracle 11.0.2.0"]}
```

## Create database
**API endpoint:** POST `/create-database`
### Explanation
Used to create a database through on of its connectors. Passes the values forward, then gets the result back. Provides a human readable success or failure message.

Returns an error if whenever the connector errors out, namely:

1. DB Name is not unique (already exists), or invalid
2. Username is not unique (already exists), or invalid

In addition to that, it also returns an error if the specified database name does not exist in its registry

Returns immediately with a success message if creating the database and user have all completed successfully, or failure if something went wrong.

### Post request details
|Key|Value|
|---|---|
|database_identifier|Must correspond to a valid connector identifier. See `/list-connectors` for what those are|
|database_name|Valid name for a database / schema. Must be non-empty|
|username|Valid name for user. Must be non-empty|
|password|Valid string for password. Must be non-empty|

### Example Curl request
```
curl -X POST -d '{"database_identifier":"mysql-55", "database_name":"exampleDatabase", "username":"exampleUser", "password":"liferay"}' localhost:6000/create-database
```

### Example responses
**Success:**

```
{"Status":200,"Message":"Successfully created the database and user!"}
```

**Failures:**

```
{"Status":500,"Message":"Unknown identifier 'somename'"}
```

```
{"Status":500,"Message":"Database 'exampleDatabase' already exists"}
```

```
{"Status":500,"Message":"User 'exampleUser' already exists"}
```

# API used by the connectors only
The below APIs are used by the connectors only and should not be used manually.

## Alive
**API endpoint:** GET `/alive`
### Explanation
Used by the connectors to check if the server is online or not. Simply returns http status code 200 and no response body when called.

## register
**API endpoint:** POST `/register`
### Explanation
Used by the connector to register itself with the server. The `model.RegisterRequest` struct should be used for requesting access, for which a `model.RegisterResponse` should be the response

## unregister
**API endpoint:** POST `/unregister`
### Explanation
Used by the connector to unregister itself from the server. The `model.Connector` struct should be used for unregistering. There is no response to this request.
