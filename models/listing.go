package models

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/anishmgoyal/calagora/utils"
)

// Listing contains fields that represent a listing
// posted by a user for sale
type Listing struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Price       int       `json:"price_server"`
	PriceClient string    `json:"price"`
	Type        string    `json:"type"`
	Condition   string    `json:"condition"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	ImageURL    *string   `json:"image_url"`
	Published   bool      `json:"published"`
	User        User      `json:"user"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

// ListingError contains fields that can be used to return
// a server side validation error
type ListingError struct {
	Name        string `json:"name,omitempty"`
	Price       string `json:"price,omitempty"`
	Type        string `json:"type,omitempty"`
	Condition   string `json:"conition,omitempty"`
	Status      string `json:"status,omitempty"`
	Description string `json:"description,omitempty"`
	Global      string `json:"global,omitempty"`
}

// ListingQueryOpts allows a user to specify how
// a query for listings should be constructed
type ListingQueryOpts struct {
	PlaceID         int
	RestrictByPlace bool
	SkipByPlace     bool

	UserID         int
	RestrictByUser bool
	SkipByUser     bool

	Status           string
	RestrictByStatus bool
	SkipByStatus     bool

	Type           string
	RestrictByType bool
	SkipByType     bool

	HideDraft     bool
	HidePublished bool

	PageSize  int
	PageNum   int
	UsePaging bool
}

const (
	// ListingMisc refers to miscellaneous listings
	ListingMisc = "misc"
	// ListingTextbook refers to textbook listings
	ListingTextbook = "textbook"
	// ListingHomegoods refers to home goods and furniture listings
	ListingHomegoods = "homegoods"
	// ListingHousing refers to listings of rentals and homes
	ListingHousing = "housing"
	// ListingAutomotive refers to car listings
	ListingAutomotive = "automotive"
	// ListingElectronics refers to listings of miscellaneous electronic devices
	ListingElectronics = "electronics"
	// ListingClothing refers to clothing listings
	ListingClothing = "clothing"
	// ListingAthletics refers to athletic equipment listings
	ListingAthletics = "athletics"
)

// ListingTypeNames is an array of type names for listings
var ListingTypeNames = []string{
	ListingMisc,
	ListingTextbook,
	ListingHomegoods,
	ListingAutomotive,
	ListingElectronics,
	ListingClothing,
	ListingAthletics,
}

// ListingTypes is a map of type names and type descriptions
var ListingTypes = map[string]string{
	ListingMisc:        "Miscellaneous",
	ListingTextbook:    "Textbooks",
	ListingHomegoods:   "Home Goods & Furniture",
	ListingHousing:     "Housing",
	ListingAutomotive:  "Automotive",
	ListingElectronics: "Electronics",
	ListingClothing:    "Clothing",
	ListingAthletics:   "Athletic Equipment",
}

const (
	// ListingListed is a listing with no active offers
	ListingListed = "listed"
	// ListingSold is a listing that has been sold
	ListingSold = "sold"
)

var listingStatuses = map[string]bool{
	ListingListed: true,
	ListingSold:   true,
}

const (
	// ListingCondNA refers to listings that don't have a condition
	ListingCondNA = "na"
	// ListingCondNew refers to listings that are like new
	ListingCondNew = "new"
	// ListingCondExcellent refers to listings in excellent condition
	ListingCondExcellent = "excellent"
	// ListingCondGood refers to listings in good condition
	ListingCondGood = "good"
	// ListingCondFair refers to listings in fair condition
	ListingCondFair = "fair"
	// ListingCondPoor refers to listings in poor condition
	ListingCondPoor = "poor"
	// ListingCondForParts refers to listings that can no longer be used
	ListingCondForParts = "forparts"
)

// ListingConditionNames is an array of listing conditions
var ListingConditionNames = []string{
	ListingCondNA,
	ListingCondNew,
	ListingCondExcellent,
	ListingCondGood,
	ListingCondFair,
	ListingCondPoor,
	ListingCondForParts,
}

// ListingConditions is a map of listing conditions and descriptions
var ListingConditions = map[string]string{
	ListingCondNA:        "Not Applicable",
	ListingCondNew:       "New",
	ListingCondExcellent: "Excellent",
	ListingCondGood:      "Good",
	ListingCondFair:      "Fair",
	ListingCondPoor:      "Poor",
	ListingCondForParts:  "For Parts",
}

// Validate checks if the fields of a given listing confirm
// to certain constraints
func (listing *Listing) Validate() (bool, ListingError) {

	var valid = true
	var err ListingError

	if len(listing.Name) < 3 || len(listing.Name) > 40 {
		valid = false
		err.Name = "Listing names may be between 3 and 40 characters long."
	}

	if _, ok := ListingTypes[listing.Type]; !ok {
		valid = false
		err.Type = "The product type " + listing.Type + " is invalid."
	}

	price, convErr := utils.PriceClientToServer(listing.PriceClient)
	if convErr != nil {
		valid = false
		err.Price = "That price is invalid."
		listing.Price = 0
	} else {
		listing.Price = price
	}

	if listing.Price < 0 {
		valid = false
		err.Price = "Prices can't be negative."
	}

	if _, ok := listingStatuses[listing.Status]; !ok {
		valid = false
		err.Status = "The status " + listing.Status + " is invalid."
	}

	if _, ok := ListingConditions[listing.Condition]; !ok {
		valid = false
		err.Condition = "The condition " + listing.Condition + " is invalid."
	}

	if len(listing.Description) >= 2500 {
		valid = false
		err.Description = "The description cannot exceed 2500 characters."
	}

	return valid, err
}

// Normalize removes artifacts like extra spaces
func (listing *Listing) Normalize() {
	if len(listing.Name) < 1000 && len(listing.Description) < 10000 {
		listing.Name = strings.TrimSpace(listing.Name)
		listing.Description = strings.TrimSpace(listing.Description)
	}
}

// Create creates a new listing, or returns validation errors
func (listing *Listing) Create(db *sql.DB) (bool, *ListingError) {

	listing.Normalize()
	valid, validationError := listing.Validate()
	if !valid {
		return false, &validationError
	}

	rows, err := db.Query("INSERT INTO listings (name, price, type, condition, "+
		"status, description, published, place_id, user_id) VALUES ($1, $2, $3, "+
		"$4, $5, $6, $7, $8, $9) RETURNING id", listing.Name, listing.Price,
		listing.Type, listing.Condition, listing.Status, listing.Description,
		listing.Published, listing.User.PlaceID, listing.User.ID)
	if err != nil {
		fmt.Println("ERROR!")
		fmt.Println(err.Error())
		return false, &ListingError{
			Global: "An unexpected error occurred.",
		}
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&listing.ID)
	} else {
		// The listing was created... but we don't have its ID
		return true, &ListingError{
			Global: "An unexpected error occurred.",
		}
	}

	go listing.DoRebuildSearchIndex(db)
	return true, nil
}

// Save updates a listing in the database with new changes
func (listing *Listing) Save(db *sql.DB) (bool, *ListingError) {
	listing.Normalize()
	valid, validationError := listing.Validate()
	if !valid {
		return false, &validationError
	}

	res, err := db.Exec("UPDATE listings SET name = $1, price = $2, "+
		"type = $3, condition = $4, status = $5, description = $6, "+
		"published = $7, modified = now() WHERE id = $8", listing.Name,
		listing.Price, listing.Type, listing.Condition, listing.Status,
		listing.Description, listing.Published, listing.ID)
	if err != nil {
		fmt.Println("ERROR!")
		fmt.Println(err.Error())
		return false, &ListingError{
			Global: "An unexpected error occurred.",
		}
	}

	go listing.DoRebuildSearchIndex(db)

	numAffected, _ := res.RowsAffected()
	if numAffected == 1 {
		return true, nil
	}
	return true, &ListingError{
		Global: "Wrong number of rows updated: " +
			strconv.FormatInt(numAffected, 10),
	}
}

// Delete attempts to delete a listing, and returns whether or not
// the operation was successful, as well as any related error messages
func (listing *Listing) Delete(db *sql.DB) (bool, error) {
	images, err := listing.GetImages(db)
	if err != nil {
		return false, errors.New("Failed to get images for deletion")
	}
	for _, image := range images {
		ok, _ := image.Delete(db)
		if !ok {
			return false, errors.New("Failed to delete all images")
		}
	}

	res, err := db.Exec("DELETE FROM listings WHERE id = $1", listing.ID)
	if err != nil {
		return false, err
	}
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if num != 1 {
		return false, errors.New("Deleted incorrect number of rows (" +
			strconv.FormatInt(num, 10) + ")")
	}
	return true, nil
}

// MarkSold attempts to mark the listing sold in the database
func (listing *Listing) MarkSold(db *sql.DB) (bool, error) {
	res, err := db.Exec("UPDATE listings SET status = '"+ListingSold+
		"' WHERE id = $1", listing.ID)
	if err != nil {
		return false, err
	}
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if num != 1 {
		return false, errors.New("Updated incorrect number of rows (" +
			strconv.FormatInt(num, 10) + ")")
	}
	return true, nil
}

// GetListingByID attempts to find a listing by its ID, returns nil if
// it couldn't be found
func GetListingByID(db *sql.DB, id int) (*Listing, error) {
	rows, err := db.Query("SELECT l.id, l.name, l.price, l.type, l.condition, "+
		"l.status, l.description, l.place_id, l.published, u.id, u.username, "+
		"u.display_name, u.email_address, l.created, l.modified FROM listings l, "+
		"users u WHERE l.user_id = u.id AND l.id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		var listing Listing
		rows.Scan(&listing.ID, &listing.Name, &listing.Price, &listing.Type,
			&listing.Condition, &listing.Status, &listing.Description,
			&listing.User.PlaceID, &listing.Published, &listing.User.ID,
			&listing.User.Username, &listing.User.DisplayName,
			&listing.User.EmailAddress, &listing.Created, &listing.Modified)
		listing.PriceClient = utils.PriceServerToClient(listing.Price)
		return &listing, nil
	}
	return nil, nil
}

// GetListingList gets listings that match certain criteria
// For example: you can hide listings by a specific user, show only
// drafts or only published listings, etc.
func GetListingList(db *sql.DB, options ListingQueryOpts) []Listing {

	var buffer bytes.Buffer
	buffer.WriteString("SELECT l.id, l.name, l.price, l.type, l.condition, " +
		"l.status, l.description, l.published, l.place_id, u.id, u.username, " +
		"u.display_name, u.email_address, u.place_id, i.URL FROM listings l JOIN " +
		"users u ON l.user_id = u.id LEFT JOIN images i ON i.media_id = l.id " +
		"WHERE (i.id = (SELECT id FROM images WHERE media='" + MediaListing +
		"' AND media_id = l.id ORDER BY ordinal ASC LIMIT 1) OR i.id IS NULL)")

	var args = make([]interface{}, 0, 5)

	var argCount = 1

	if options.RestrictByPlace {
		buffer.WriteString(" AND (l.place_id = $" + strconv.Itoa(argCount) + ")")
		args = append(args, options.PlaceID)
		argCount++
	} else if options.SkipByPlace {
		buffer.WriteString(" AND (l.place_id = $" + strconv.Itoa(argCount) + ")")
		args = append(args, options.PlaceID)
		argCount++
	}

	if options.RestrictByUser {
		buffer.WriteString(" AND (l.user_id = $" + strconv.Itoa(argCount) + ")")
		args = append(args, options.UserID)
		argCount++
	} else if options.SkipByUser {
		buffer.WriteString(" AND (NOT l.user_id = $" + strconv.Itoa(argCount) + ")")
		args = append(args, options.UserID)
		argCount++
	}

	if options.RestrictByStatus {
		buffer.WriteString(" AND (l.status = $" + strconv.Itoa(argCount) + ")")
		args = append(args, options.Status)
		argCount++
	} else if options.SkipByStatus {
		buffer.WriteString(" AND (l.status <> $" + strconv.Itoa(argCount) + ")")
		args = append(args, options.Status)
		argCount++
	}

	if options.RestrictByType {
		buffer.WriteString(" AND (l.type = $" + strconv.Itoa(argCount) + ")")
		args = append(args, options.Type)
		argCount++
	} else if options.SkipByType {
		buffer.WriteString(" AND (l.type <> $" + strconv.Itoa(argCount) + ")")
		args = append(args, options.Type)
		argCount++
	}

	if options.HideDraft {
		buffer.WriteString(" AND l.published")
	} else if options.HidePublished {
		buffer.WriteString(" AND NOT l.published")
	}

	buffer.WriteString(" ORDER BY l.id DESC")

	if options.UsePaging {
		buffer.WriteString(" LIMIT $" + strconv.Itoa(argCount))
		args = append(args, options.PageSize)
		argCount++
		buffer.WriteString(" OFFSET $" + strconv.Itoa(argCount))
		args = append(args, options.PageNum*options.PageSize)
		argCount++
	}

	rows, err := db.Query(buffer.String(), args[:argCount-1]...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var listings = make([]Listing, 0, 10)
	var found = 0
	for rows.Next() {
		var l Listing
		err = rows.Scan(&l.ID, &l.Name, &l.Price, &l.Type, &l.Condition,
			&l.Status, &l.Description, &l.Published, &l.User.PlaceID, &l.User.ID,
			&l.User.Username, &l.User.DisplayName, &l.User.EmailAddress,
			&l.User.PlaceID, &l.ImageURL)
		if err == nil {
			if l.ImageURL == nil {
				l.ImageURL = &ImageNotFound
			}
			l.PriceClient = utils.PriceServerToClient(l.Price)
			listings = append(listings, l)
			found++
		}
	}

	return listings[:found]
}
