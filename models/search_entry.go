package models

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/anishmgoyal/calagora/utils"
)

const (
	maxQueryTermsToConsider = 10
	searchPageSize          = 50
)

// SearchEntry encapsulates a search entry index for a word in a listing
type SearchEntry struct {
	ID       int       `json:"-"`
	Word     string    `json:"word"`
	Count    int       `json:"count"`
	Listing  Listing   `json:"listing"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// DoRebuildSearchIndex deletes any search index entries for a listing,
// then rebuilds them
func (l *Listing) DoRebuildSearchIndex(db *sql.DB) (bool, error) {
	deleteStatement := "DELETE FROM search_entries WHERE listing_id = $1"
	_, err := db.Exec(deleteStatement, l.ID)
	if err != nil {
		return false, err
	}

	if l.Published {
		success := true
		var lastError error

		typeName, _ := ListingTypes[l.Type]
		fullString := l.Name + " " + l.Name + " " + typeName + " " + l.Description

		images, err := l.GetImages(db)
		if err != nil {
			return false, err
		}

		if len(images) == 0 {
			images = append(images, Image{URL: ImageNotFound})
		}

		minOrdinal := 0
		for i := 0; i < len(images); i++ {
			if images[i].Ordinal < minOrdinal {
				images[0] = images[i]
				minOrdinal = images[i].Ordinal
			}
		}

		termMap := utils.GetSearchTermsForString(fullString, true)
		for word, count := range termMap {
			insertStatement := "INSERT INTO search_entries (word, count, " +
				"listing_id, listing_name, listing_price, listing_image, " +
				"place_id, listing_type) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
			_, err := db.Exec(insertStatement, word, count, l.ID, l.Name, l.Price,
				images[0].URL, l.User.PlaceID, l.Type)
			if err != nil {
				success = false
				lastError = err
			}
		}
		return success, lastError
	}

	return true, nil
}

// GetPageCountForTerms gets the number of pages available for
// a given set of search terms
func GetPageCountForTerms(db *sql.DB, terms []string, placeID int) int {
	wordList := ""
	args := make([]interface{}, len(terms))
	for i := 0; i < len(terms); i++ {
		if i > 0 {
			wordList = wordList + ", "
		}
		wordList = wordList + "$" + strconv.Itoa(i+1)
		args[i] = terms[i]
	}

	query := "SELECT COUNT(1) FROM search_entries WHERE word IN (" +
		wordList + ")"
	if placeID > -1 {
		query += " AND place_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, placeID)
	}
	query += " GROUP BY listing_id"

	res := db.QueryRow(query, args...)
	var numRecords int

	err := res.Scan(&numRecords)
	if err != nil {
		return 0
	}

	pageCount := numRecords / 50
	if numRecords%50 > 0 {
		pageCount++
	}
	return pageCount
}

// DoSearchForTerms attempts to find listings matching a list of query
// terms
func DoSearchForTerms(db *sql.DB, terms []string, page int, placeID int) (
	[]Listing, error) {

	args := make([]interface{}, 0, maxQueryTermsToConsider+2)
	argCount := 0

	listings := make([]Listing, 0, 50)
	numFound := 0

	queryTermCount := len(terms)
	if queryTermCount > maxQueryTermsToConsider {
		queryTermCount = maxQueryTermsToConsider
	}

	termList := ""
	for i := 0; i < queryTermCount; i++ {
		args = append(args, terms[i])
		argCount++
		if i > 0 {
			termList = termList + ", "
		}
		termList = termList + "$" + strconv.Itoa(argCount)
	}

	query := "SELECT listing_id, min(listing_name), min(listing_price), " +
		"min(listing_image) FROM search_entries WHERE word IN (" + termList + ")"

	if placeID > -1 {
		args = append(args, placeID)
		argCount++
		query = query + " AND place_id = $" + strconv.Itoa(argCount)
	}

	query = query + " GROUP BY listing_id ORDER BY sum(count) DESC LIMIT $" +
		strconv.Itoa(argCount+1) + " OFFSET $" + strconv.Itoa(argCount+2)

	args = append(args, searchPageSize, searchPageSize*page)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var listing Listing
		err := rows.Scan(&listing.ID, &listing.Name, &listing.Price,
			&listing.ImageURL)
		if err == nil {
			listing.PriceClient = utils.PriceServerToClient(listing.Price)
			if listing.ImageURL == nil {
				listing.ImageURL = &ImageNotFound
			}
			listings = append(listings, listing)
			numFound++
		}
	}

	return listings[:numFound], nil
}
