## CloudDB API responses

### Required header
All API calls requires an "Authorization" header to be set. Currently, the value of the header should be the email address of your user. If testing with `curl`, the following should be added to the command (as done in the example calls):

`-H "Authorization:your.email@example.com"`

If the Authorization header is not specified, an error will be returned:
```
{
    "success":false,
    "error":["ERR_ACCESS_DENIED"]
}
```

### Response patterns

#### Success
```
{
   "success":true,
   "data": // string, array or object
}
```

#### Fail
```
{
    "success":false,
    "error":["ERR_MSG", "optional params"] // array of string messages
}
```

## List all agents

### GET /api/agents
Example

`curl -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/agents`

### Payload
none

### Returns
List of agents objects, each one containing all known information. Also returns agents that are not up.

Example success return:
```
{
   "success":true,
   "data":[
      {
         "id":1,
         "vendor":"mariadb",
         "dbport":"3309",
         "dbaddress":"172.17.0.2",
         "sid":"",
         "agent":"mariadb-10",
         "agent_long":"mariadb 10.2.11",
         "agent_identifier":"myhostname-mariadb-10",
         "agent_port":"7005",
         "agent_version":"3",
         "agent_address":"http://172.16.20.230",
         "agent_token":"",
         "agent_up":true
      }
   ]
}
```

Failed return:
```
{
    "success":false,
    "error":["ERR_NO_AGENTS_AVAILABLE"]
}
```
## List all active agents

### GET /api/agents/active
Example

`curl -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/agents/active`

### Payload
none

### Returns
List of agents objects, each one containing all known information. Only returns active agents.

Example success return:
```
{
   "success":true,
   "data":[
      {
         "id":1,
         "vendor":"mariadb",
         "dbport":"3309",
         "dbaddress":"172.17.0.2",
         "sid":"",
         "agent":"mariadb-10",
         "agent_long":"mariadb 10.2.11",
         "agent_identifier":"myhostname-mariadb-10",
         "agent_port":"7005",
         "agent_version":"3",
         "agent_address":"http://172.16.20.230",
         "agent_token":"",
         "agent_up":true
      }
   ]
}
```

Failed return:
```
{
    "success":false,
    "error":["ERR_NO_AGENTS_AVAILABLE"]
}
```

## Get all information on a specific agent

### GET /api/agents/${agentName}
Example

`curl -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/agents/mariadb-10`

### Payload
`${agentName}` - the shortname of the agent (`agent` field in response)

### Returns
All known information about the specified agent

Example success return:
```
{
   "success":true,
   "data":{
      "id":1,
      "vendor":"mariadb",
      "dbport":"3309",
      "dbaddress":"172.17.0.2",
      "sid":"",
      "agent":"mariadb-10",
      "agent_long":"mariadb 10.2.11",
      "agent_identifier":"myhostname-mariadb-10",
      "agent_port":"7005",
      "agent_version":"3",
      "agent_address":"http://172.16.20.230",
      "agent_token":"",
      "agent_up":true
   }
}
```

Failed returns:
```
{
    "success":false,
    "error":["ERR_AGENT_NOT_FOUND"]
}
```

## List databases
### GET /api/databases
Example

`curl -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/databases`

### Payload
none

### Returns
All metadata about the public databases and the ones created by the requester.

Example success return:
```
{  
   "success":true,
   "data":[  
      {  
         "id":16,
         "vendor":"mariadb",
         "dbname":"electric_adapter",
         "dbuser":"electric_adapter",
         "dbpass":"tag_tuner",
         "sid":"",
         "dumplocation":"",
         "createdate":"2018-01-07T13:25:46.148399484Z",
         "expirydate":"2018-02-07T13:25:46.148399558Z",
         "creator":"daniel.javorszky@liferay.com",
         "agent":"mariadb-10",
         "dbaddress":"172.17.0.2",
         "dbport":"3309",
         "status":100,
         "comment":"",
         "message":"",
         "public":0
      },
      // .. more
}
```
## Get metadata of specific database by id

### GET /api/databases/${id}
Example

`curl -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/databases/15`

### Payload
`${id}` - the id of the metadata itself.

### Returns
All metadata about the database that has the id `${id}`

Example success return:
```
{
   "success":true,
   "data":{
      "id":15,
      "vendor":"mariadb",
      "dbname":"gel_component",
      "dbuser":"performance_air",
      "dbpass":"gel_gel",
      "sid":"",
      "dumplocation":"",
      "createdate":"2017-12-11T15:14:27.03707071Z",
      "expirydate":"2018-01-11T15:14:27.037070856Z",
      "creator":"daniel.javorszky@liferay.com",
      "agent":"mariadb-10",
      "dbaddress":"172.17.0.2",
      "dbport":"3309",
      "status":100,
      "comment":"",
      "message":"",
      "public":0
   }
}
```

Example failed return:
```
{
    "success":false,
    "error":["ERR_DATABASE_NO_RESULT"]
}
```
## Get metadata of specific database by agent and database name

### GET /api/databases/${agent}/${dbname}
Example

`curl -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/databases/mariadb-10/gel_component`

### Payload
`${agent}` - Shortname of the agent

`${dbname}` - Database name (or in some cases like Oracle, the name of the user)

### Returns
All metadata about the database that has been created (/imported) by the `${agent}` agent and has the name of `${dbname}`.

Example success return:
```
{
   "success":true,
   "data":{
      "id":15,
      "vendor":"mariadb",
      "dbname":"gel_component",
      "dbuser":"performance_air",
      "dbpass":"gel_gel",
      "sid":"",
      "dumplocation":"",
      "createdate":"2017-12-11T15:14:27.03707071Z",
      "expirydate":"2018-01-11T15:14:27.037070856Z",
      "creator":"daniel.javorszky@liferay.com",
      "agent":"mariadb-10",
      "dbaddress":"172.17.0.2",
      "dbport":"3309",
      "status":100,
      "comment":"",
      "message":"",
      "public":0
   }
}
```

Example failed return:
```
{
    "success":false,
    "error":["ERR_DATABASE_NO_RESULT"]
}
```

## Drop database by its id

### DELETE /api/databases/${id}
Example

`curl -X DELETE -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/databases/15`

### Payload
`${id}` - the id of the metadata itself.

### Returns
Drops the database with id `${id}`

Example success return:
```
{
   "success":true,
   "data":"Delete successful"
}
```

Example failed return:
```
{
    "success":false,
    "error":["ERR_DATABASE_NO_RESULT"]
}
```

## Drop database by agent and database name

### DELETE /api/databases/${agent}/${dbname}
Example

`curl -X DELETE -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/databases/mariadb-10/gel_component`

### Payload
`${agent}` - Shortname of the agent

`${dbname}` - Database name (or in some cases like Oracle, the name of the user)

### Returns
Drops the database with the `${dbname}` database name managed by the `${agent}` agent.

Example success return:
```
{
   "success":true,
   "data":"Delete successful"
}
```

Example failed return:
```
{
    "success":false,
    "error":["ERR_DATABASE_NO_RESULT"]
}
```

## Create an empty database

### POST /api/databases/create
Example

`curl -X POST  -H "Authorization:daniel.javorszky@liferay.com" -H "Content-Type: application/json" -d '{"agent_identifier":"mariadb-10"}' http://localhost:7010/api/databases/create`

### Payload
#### Required
`agent_identifier` - Shortname of the agent

#### Optional
`database_name` - Name of the database to be created.

`username` - Name of the user to be created.

`password` - Password to set for the created user

### Returns
All data about the created database.


Example success return:
```
{
   "success":true,
   "data":{
      "id":34,
      "vendor":"mariadb",
      "dbname":"gps_video",
      "dbuser":"gps_video",
      "dbpass":"air_viewer",
      "sid":"",
      "dumplocation":"",
      "createdate":"2018-01-16T01:14:33.41554638Z",
      "expirydate":"2018-02-16T01:14:33.415546478Z",
      "creator":"daniel.javorszky@liferay.com",
      "agent":"mariadb-10",
      "dbaddress":"172.17.0.2",
      "dbport":"3309",
      "status":100,
      "comment":"",
      "message":"",
      "public":0
   }
}
```

Example failed returns:
```
{
    "success":false,
    "error":["ERR_MISSING_PARAMETERS","agent_identifier"]
}

// or

{
    "success":false,
    "error":["ERR_AGENT_NOT_FOUND","nonexistent_agent"]
}
```

## Import a database

### POST /api/databases/import
Example

`curl -X POST  -H "Authorization:daniel.javorszky@liferay.com" -H "Content-Type: application/json" -d '{"agent_identifier":"mariadb-10", "dumpfile_location":"/folder/file.sql"}' http://localhost:7010/api/databases/import`

`curl -X POST  -H "Authorization:daniel.javorszky@liferay.com" -H "Content-Type: application/json" -d '{"agent_identifier":"mariadb-10", "dumpfile_location":"http://localhost/somedumpfile.sql"}' http://localhost:7010/api/databases/import`

### Payload
#### Required
`agent_identifier` - Shortname of the agent

`dumpfile_location` - Location of the dumpfile. Can be absolute path  (if folder is mounted) or http link to download.

#### Optional
`database_name` - Name of the database to be created.

`username` - Name of the user to be created.

`password` - Password to set for the created user

### Returns
All data about the imported database.


Example success return:
```
{
   "success":true,
   "data":{
      "id":34,
      "vendor":"mariadb",
      "dbname":"gps_video",
      "dbuser":"gps_video",
      "dbpass":"air_viewer",
      "sid":"",
      "dumplocation":"http://localhost/somedumpfile.sql",
      "createdate":"2018-01-16T01:14:33.41554638Z",
      "expirydate":"2018-02-16T01:14:33.415546478Z",
      "creator":"daniel.javorszky@liferay.com",
      "agent":"mariadb-10",
      "dbaddress":"172.17.0.2",
      "dbport":"3309",
      "status":100,
      "comment":"",
      "message":"",
      "public":0
   }
}
```

Example failed returns:
```
{
    "success":false,
    "error":["ERR_MISSING_PARAMETERS","agent_identifier"]
}

// or

{
    "success":false,
    "error":["ERR_MISSING_PARAMETERS","dumpfile_location"]
}

// or

{
    "success":false,
    "error":["ERR_AGENT_NOT_FOUND","nonexistent_agent"]
}
```
## Recreate a database

Recreates the database with the given ID. Basically drops the database and creates a new one with the same information

### PUT /api/databases/${id}/recreate
Example

`curl -X PUT -H 'Authorization:daniel.javorszky@liferay.com'  http://localhost:7010/api/databases/16/recreate`

### Payload
`${id}` - the id of the metadata itself.

### Returns
Returns all information on the recreated database.

Example success return:
```
{
   "success":true,
   "data":{
      "id":15,
      "vendor":"mariadb",
      "dbname":"gel_component",
      "dbuser":"performance_air",
      "dbpass":"gel_gel",
      "sid":"",
      "dumplocation":"",
      "createdate":"2017-12-11T15:14:27.03707071Z",
      "expirydate":"2018-01-11T15:14:27.037070856Z",
      "creator":"daniel.javorszky@liferay.com",
      "agent":"mariadb-10",
      "dbaddress":"172.17.0.2",
      "dbport":"3309",
      "status":100,
      "comment":"",
      "message":"",
      "public":0
   }
}
```

Example failed return:
```
{
    "success":false,
    "error":["ERR_DATABASE_NO_RESULT"]
}
```
## List files in mounted folder

### GET /api/browse/${loc}
List the files and folders in the mounted folder

Examples:

`curl -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/browse`

`curl -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/browse/somefolder`

### Payload
`${loc}` - Relative path on the server. Can be empty (e.g. `api/browse`) or a valid path (`api/browse/folder`)

### Returns
Returns a directory listing, containing all files and folders, as well as some information about the folder that was queried.

Example success return:
```
{
   "success":true,
   "data":{
      "OnRoot":true,
      "Path":"/",
      "Parent":"",
      "Entries":[
         {
            "Name":"a",
            "Path":"/a",
            "Size":12345,
            "Folder":false
         },
         {
            "Name":"hello",
            "Path":"/hello",
            "Size":4096,
            "Folder":true
         }
      ]
   }
}
```

Example failed returns:
```
{
   "success":false,
   "error":[
      "ERR_DIR_LIST_FAILED",
      "failed reading dir: open /ddn/ftp/asd: no such file or directory"
   ]
   // or
   "error":[
       "ERR_DIR_LIST_FAILED",
       "failed reading dir: readdirent: not a directory"
   ]
}
```
## Update database visibility

### PUT /api/databases/${id}/visibility/${vis}
Change the visibility of database `${id}` to private or public
Examples:

`curl -X PUT -H 'Authorization:daniel.javorszky@liferay.com'  http://localhost:7010/api/databases/16/visibility/public`

`curl -X PUT -H 'Authorization:daniel.javorszky@liferay.com'  http://localhost:7010/api/databases/16/visibility/private`

### Payload
`${id}` - the id of the metadata itself.

`${vis}` - either public or private.

### Returns
Returns a success message if successful, or an error if not. If no change needed to take effect (e.g. public->public), it is still considered to be a success.

Example success return:
```
{
   "success":true,
   "data":"Visibility updated successfully"
   // or
   "data":"Visibility already set to public"
}
```

Example failed returns:
```
{
    "success":false,
    "error":["ERR_DATABASE_NO_RESULT"]
}
```
## Extend database expiry
### PUT /api/databases/${id}/expiry/extend/${amount}/${unit}
Extend the expiry of database `${id}` by `${amount}` `${unit}`
Examples:

`curl -X PUT -H 'Authorization:daniel.javorszky@liferay.com'  http://localhost:7010/api/databases/16/expiry/extend/13/days`

`curl -X PUT -H 'Authorization:daniel.javorszky@liferay.com'  http://localhost:7010/api/databases/16/expiry/extend/4/months`

`curl -X PUT -H 'Authorization:daniel.javorszky@liferay.com'  http://localhost:7010/api/databases/16/expiry/extend/1/years`


### Payload
`${id}` - the id of the metadata itself.

`${amount}` - integer

`${unit}` - can be `days`, `months` or `years`

### Returns
Returns the new expiry date if successful, or error message if something went wrong.

Example success return:
```
{
    "success":true,
    "data":"2018-03-16T13:25:46.148399558Z"
}
```

Example failed returns:
```
{
    "success":false,
    "error":["ERR_DATABASE_NO_RESULT"]
}
```
## Fetch access information of a database by id
### GET /api/databases/${id}/accessinfo
Get accesss info for the database denoted by meta id `${id}`
Examples:

`curl -H 'Authorization:daniel.javorszky@liferay.com'  http://localhost:7010/api/databases/16/accessinfo`


### Payload
`${id}` - the id of the metadata itself.


### Returns
Returns the database access information. May contain a jdbc_url_6210 key-value as well in case it's different from the dxp one. Also contains "database" in case vendor is not oracle.

Example success return:
```
{  
   "success":true,
   "data":{  
      database: "gel_controller"
      jdbc_driver: "jdbc.default.driverClassName=com.mysql.jdbc.Driver"
      jdbc_url: "jdbc.default.url=jdbc:mysql://127.0.0.1:3306/gel_controller?characterEncoding=UTF-8&dontTrackOpenResources=true&holdResultsOpenOverStatementClose=true&useFastDateParsing=false&useUnicode=true&useSSL=false"
      jdbc_url_6210: "jdbc.default.url=jdbc:mysql://127.0.0.1:3306/gel_controller?useUnicode=true&characterEncoding=UTF-8&useFastDateParsing=false&useSSL=false"
      password: "auto_air"
      url: "127.0.0.1:3306"
      user: "gel_controller"
   }
}
```

Example failed returns:
```
{
    "success":false,
    "error":["ERR_DATABASE_NO_RESULT"]
}
```

## Fetch access information of a database by agent and database name

## GET /api/databases/${agent}/${dbname}/accessinfo
Get accesss info for the database `${agent}` and `${dbname}`
Examples:

`curl -H 'Authorization:daniel.javorszky@liferay.com'  http://localhost:7010/api/databases/mariadb-10/electric_adapter/accessinfo`


### Payload
`${agent}` - Shortname of the agent

`${dbname}` - Database name (or in some cases like Oracle, the name of the user)

### Returns
Returns the database access information. May contain a jdbc_url_6210 key-value as well in case it's different from the dxp one. Also contains "database" in case vendor is not oracle.

Example success return:
```
{  
   "success":true,
   "data":{  
      database: "gel_controller"
      jdbc_driver: "jdbc.default.driverClassName=com.mysql.jdbc.Driver"
      jdbc_url: "jdbc.default.url=jdbc:mysql://127.0.0.1:3306/gel_controller?characterEncoding=UTF-8&dontTrackOpenResources=true&holdResultsOpenOverStatementClose=true&useFastDateParsing=false&useUnicode=true&useSSL=false"
      jdbc_url_6210: "jdbc.default.url=jdbc:mysql://127.0.0.1:3306/gel_controller?useUnicode=true&characterEncoding=UTF-8&useFastDateParsing=false&useSSL=false"
      password: "auto_air"
      url: "127.0.0.1:3306"
      user: "gel_controller"
   }
}
```

Example failed returns:
```
{
    "success":false,
    "error":["ERR_DATABASE_NO_RESULT"]
}
```
## Change the loglevel of the server
### PUT /api/loglevel/${level}
Updates the loglevel of the server.

Example

`curl -X PUT -H 'Authorization:daniel.javorszky@liferay.com'  http://localhost:7010/api/loglevel/debug`

### Payload
`${level}` - loglevel. Can be either `fatal`, `error`, `warn`, `info` or `debug`

### Returns
Example success return:
```
{
   "success":true,
   "data":"Loglevel changed from info to debug"
}
```

Example failed return:
```
{
    "success":false,
    "error":["ERR_UNKNOWN_PARAMETER","debugz"]
}
```
