WHENEVER OSERROR EXIT FAILURE
WHENEVER SQLERROR EXIT SQL.SQLCODE
SET VERIFY OFF

create user &1 identified by &2 default tablespace &3 quota unlimited on &3;

grant connect, resource to &1;

exit;