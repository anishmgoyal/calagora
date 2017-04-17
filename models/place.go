package models

import (
	"database/sql"
	"strings"
)

// Place contains fields describing a subsection of Calagora's market
type Place struct {
	ID           int    `json:"id"`
	Abbreviation string `json:"abbreviation"`
	Name         string `json:"name"`
	EmailDomain  string `json:"email_domain"`
}

var unknownPlace = Place{
	ID:           0,
	Abbreviation: "UNKN",
	Name:         "an unknown school",
	EmailDomain:  ".edu",
}

// GetPlaceByID gets information about a place by its ID
func GetPlaceByID(db *sql.DB, id int) (*Place, error) {
	rows, err := db.Query("SELECT abbr, name, email_domain FROM places WHERE "+
		"id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		var place Place
		err = rows.Scan(&place.Abbreviation, &place.Name, &place.EmailDomain)
		if err != nil {
			return nil, err
		}
		return &place, nil
	}
	return &unknownPlace, nil
}

// FindPlaceID gets a place id from an email address if available
func FindPlaceID(db *sql.DB, emailAddress string) (int, error) {
	parts := strings.Split(emailAddress, "@")
	if len(parts) != 2 {
		return -1, nil
	}
	domain := parts[1]
	placeID := -1
	for {
		rows, err := db.Query("SELECT id FROM places WHERE email_domain = $1",
			domain)
		if err != nil {
			return -1, err
		}
		if rows.Next() {
			rows.Scan(&placeID)
			break
		}
		rows.Close()

		index := strings.Index(domain, ".")
		if index == -1 {
			break
		}
		domain = domain[index+1:]
	}
	return placeID, nil
}
