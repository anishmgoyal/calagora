package utils

import "database/sql"

// SetTimezoneUtc sets the timezone to UTC in the database
// to ensure consistency of dates
func SetTimezoneUtc(db *sql.DB) error {
	_, err := db.Exec("SET TIME ZONE 'UTC'")
	return err
}

// SetTimezone sets the timezone to the one specified in the database
// to ensure consistency of dates
func setTimezone(db *sql.DB, timezone string) error {
	_, err := db.Exec("SET TIME ZONE $1", timezone)
	return err
}
