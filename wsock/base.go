package wsock

import "database/sql"

// Base contains all info needed by wsock code
var Base struct {
	// Db is the connection to the database
	Db *sql.DB
}

// BaseInitialization sets up the wsock module
func BaseInitialization(db *sql.DB) {
	Base.Db = db
}
