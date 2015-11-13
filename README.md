# maxscale_string_tester
Test script for testing that the MaxScale binlog router responds correctly to certain commands

Sample output from a version I was testing:

```
$ MYSQL_DSN="user:password@tcp(maxscale.example.com:3306)/" go run main.go
Using dsn defined in environment variable MYSQL_DSN
Connected to database
show variables like 'maxscale%'
OK: MAXSCALE_VERSION ('1.2.1.18') has no nulls
show slave status:
OK: Slave_IO_Running ('Yes') has no nulls
OK: Slave_SQL_Running ('Yes') has no nulls
OK: Master_Log_File ('binlog.067620') has no nulls
OK: Relay_Master_Log_File ('binlog.067620') has no nulls
OK: Relay_Log_File ('binlog.067620') has no nulls
OK: Executed_Gtid_Set ('') has no nulls
OK: UsingMariaDBGTID ('') has no nulls
OK: Master_Host ('master.example.com') has no nulls
other commands:
OK: VERSION() ('5.6.99-log') has no nulls
OK: @@hostname ('master.example.com') has no nulls
WARNING: @@report_host gave an error: 'Error 1064: You have an error in your SQL syntax; Check the syntax the MaxScale binlog router accepts.'
```
