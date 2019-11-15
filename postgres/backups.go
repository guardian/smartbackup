package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

/**
tells Postgres to put itself into a consistent state for backup
returns a string of the WAL segment of the consistent state if successful (this can be ignored) or an error
*/
func StartBackup(config *DatabaseConfig, backupName string) (string, error) {
	log.Printf("DEBUG: connection string is %s", config.GetConnectionString())
	db, err := sql.Open("postgres", config.GetConnectionString())
	if err != nil {
		log.Printf("ERROR: Can't connect to postgres: %s", err)
		return "", err
	}

	var fastStartString string
	if config.FastStart {
		fastStartString = "true"
	} else {
		fastStartString = "false"
	}

	queryStr := fmt.Sprintf("select pg_start_backup('%s', %s);", backupName, fastStartString)
	log.Printf("DEBUG - query string is %s", queryStr)
	log.Printf("Putting database into consistent recovery state...")
	result, queryErr := db.Query(queryStr)

	if queryErr != nil {
		log.Printf("ERROR: Could not start postgres backup: %s", queryErr)
		return "", queryErr
	}

	var snapshotLocation string
	result.Next()

	scanErr := result.Scan(&snapshotLocation)
	if scanErr != nil {
		log.Printf("ERROR: Postgres backup apparently started but could not understand server response: %s", scanErr)
		return "", scanErr
	}
	log.Printf("Done, consistent recovery state reached at %s", snapshotLocation)
	return snapshotLocation, nil
}

/**
tells Postgres that the backup operation is complete and it can reset its state
*/
func StopBackup(config *DatabaseConfig) (string, error) {
	db, err := sql.Open("postgres", config.GetConnectionString())
	if err != nil {
		log.Printf("ERROR: Can't connect to postgres: %s", err)
		return "", err
	}

	result, queryErr := db.Query("select pg_stop_backup();")
	if queryErr != nil {
		log.Printf("ERROR: Could not stop backup: %s", queryErr)
		return "", queryErr
	}

	var snapshotLocation string
	result.Next()

	scanErr := result.Scan(&snapshotLocation)
	if scanErr != nil {
		log.Printf("ERROR: Postgres backup apparently stopped but could not understand server response: %s", scanErr)
		return "", scanErr
	}
	return snapshotLocation, nil
}
