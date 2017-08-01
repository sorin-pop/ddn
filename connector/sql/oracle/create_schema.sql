WHENEVER OSERROR EXIT FAILURE
WHENEVER SQLERROR EXIT SQL.SQLCODE
SET VERIFY OFF

var  datafile1 VARCHAR2(2000);
var  datafile2 VARCHAR2(2000);

EXECUTE :datafile1 := '&3' || '&1' || '_01.dbf';
EXECUTE :datafile2 := '&3' || '&1' || '_02.dbf';

BEGIN
    EXECUTE IMMEDIATE 'CREATE SMALLFILE TABLESPACE ' || '&1' || ' DATAFILE ' || '''' || :datafile1 || '''' || ' SIZE 32M AUTOEXTEND ON MAXSIZE UNLIMITED';
    EXECUTE IMMEDIATE 'ALTER TABLESPACE ' || '&1' || ' ADD DATAFILE ' || '''' || :datafile2 || '''' || ' SIZE 1M AUTOEXTEND ON MAXSIZE UNLIMITED';
END;
/

CREATE USER &1 IDENTIFIED BY &2 DEFAULT TABLESPACE &1 QUOTA UNLIMITED ON &1;

GRANT CONNECT, RESOURCE TO &1;

EXIT;