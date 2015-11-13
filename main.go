// program to test MaxScale does not return with responses which are invalid
package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"

	"github.com/outbrain/golib/sqlutils"
)

/*
 *  Early versions of MaxScale may be returning invalid string results.
 *  This can confuse orchestrator. This test program allows you to run
 *  some queries against MaxScale to see if it's behaving as expected.
 *
 *  Usage:  MYSQL_DSN="user:password@tcp(host-to-check:port)/" ./maxscale_string_tester
 */

const (
	defaultDSN = "user:host@tcp(127.0.0.1:3306)/mydb"
)

var (
	verbose bool
)

// Clean the hostname if it doesn't have proper characters.
// - valid characters are
//   0-9 a-z A-Z -_.
// - any other characters will be removed from the string
func cleanHostname(dirty string) string {
	var clean string

	for i := range dirty {
		c := dirty[i]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c == '-') || (c == '_') || (c == '.') {
			clean += string(c)
		}
	}

	return clean
}

func checkForNulls(dirty string) int {
	count := 0

	for i := range dirty {
		if dirty[i] == '\x00' {
			count++
		}
	}
	return count
}

func namedCheckForNulls(name, dirty string) {
	if count := checkForNulls(dirty); count > 0 {
		fmt.Printf("WARNING: %+v ('%+v') has %d nulls (% x)\n", name, dirty, count, dirty)
	} else {
		fmt.Printf("OK: %+v ('%+v') has no nulls\n", name, dirty)
	}
}

func singleRowQuery(db *sql.DB, query string) (string, error) {
	var value string

	err := db.QueryRow(query).Scan(&value)
	switch {
	case err == sql.ErrNoRows:
		return "", err
	case err != nil:
		return "", err
	}

	return value, nil
}

func namedSingleRowQuery(db *sql.DB, name, query string) {
	if value, err := singleRowQuery(db, query); err != nil {
		fmt.Printf("WARNING: %+v gave an error: '%s'\n", name, err.Error())
	} else {
		namedCheckForNulls(name, value)
	}
}

func main() {
	flag.BoolVar(&verbose, "verbose", false, "Verbose logging")
	flag.Parse()

	dsn := os.Getenv("MYSQL_DSN") // better name ?
	if dsn == "" {
		fmt.Println("Using default dsn", defaultDSN)
		dsn = defaultDSN
	} else {
		fmt.Println("Using dsn defined in environment variable MYSQL_DSN")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// really connect
	if err = db.Ping(); err != nil {
		log.Fatal("Ping failure:", err)
	}
	fmt.Println("Connected to database")

	fmt.Println("show variables like 'maxscale%'")
	rows, err := db.Query("show variables like 'maxscale%'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var variable, value string
		if err := rows.Scan(&variable, &value); err != nil {
			log.Fatal(err)
		}
		namedCheckForNulls(variable, value)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// show slave status
	fmt.Println("show slave status:")
	err = sqlutils.QueryRowsMap(db, "show slave status", func(m sqlutils.RowMap) error {
		namedCheckForNulls("Slave_IO_Running", m.GetString("Slave_IO_Running"))
		namedCheckForNulls("Slave_SQL_Running", m.GetString("Slave_SQL_Running"))
		namedCheckForNulls("Master_Log_File", m.GetString("Master_Log_File"))
		namedCheckForNulls("Relay_Master_Log_File", m.GetString("Relay_Master_Log_File"))
		namedCheckForNulls("Relay_Log_File", m.GetString("Relay_Log_File"))
		namedCheckForNulls("Executed_Gtid_Set", m.GetString("Executed_Gtid_Set"))
		namedCheckForNulls("UsingMariaDBGTID", m.GetString("Using_Gtid"))
		namedCheckForNulls("Master_Host", m.GetString("Master_Host"))

		return nil
	})

	fmt.Println("other commands:")
	namedSingleRowQuery(db, "VERSION()", "SELECT VERSION()")
	namedSingleRowQuery(db, "@@hostname", "SELECT @@hostname")
	namedSingleRowQuery(db, "@@report_host", "SELECT @@report_host")
}
