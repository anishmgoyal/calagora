package cache

import "database/sql"

// Base contains essential information for cache operations
var Base struct {
	// Db is the handle to the database connection
	Db *sql.DB
}

// BaseInitialization sets up any caching that is to be performed
func BaseInitialization(db *sql.DB) {
	Base.Db = db
	initPlaceCache()
}
