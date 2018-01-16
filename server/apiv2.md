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



## GET /api/agents
Example `curl -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/agents`

### Payload
none

### Returns
List of agents objects, each one containing all known information

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

## GET /api/agents/${agentName}
Example `curl -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/agents/mariadb-10`

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

## GET /api/databases
Example `curl -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/databases`

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

## GET /api/databases/${id}
Example `curl -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/databases/15`

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

## GET /api/databases/${agent}/${dbname}
Example `curl -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/databases/mariadb-10/gel_component`

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


## DELETE /api/databases/${id}
Example `curl -X DELETE -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/databases/15`

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

## DELETE /api/databases/${agent}/${dbname}
Example `curl -X DELETE -H "Authorization:daniel.javorszky@liferay.com" http://localhost:7010/api/databases/mariadb-10/gel_component`

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

## POST /api/databases
Example `curl -X POST  -H "Authorization:daniel.javorszky@liferay.com" -H "Content-Type: application/json" -d '{"agent_identifier":"mariadb-10"}' http://localhost:7010/api/databases`

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

## PUT /api/databases/${id}
Example `curl -X PUT -H 'Authorization:daniel.javorszky@liferay.com'  http://localhost:7010/api/databases/16`

### Payload
`${id}` - the id of the metadata itself.

### Returns
Recreates the database with the given ID. Basically drops the database and creates a new one with the same information

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

## GET /api/browse/${loc}
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

## PUT /api/databases/${id}/visibility/${vis}
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

