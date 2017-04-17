package models

import (
	"database/sql"
	"time"

	"github.com/anishmgoyal/calagora/utils"
)

const (
	// OfferOffered is for an offer not yet accepted
	OfferOffered = "offered"
	// OfferAccepted is for an offer accepted by a seller
	OfferAccepted = "accepted"
	// OfferCompleted is for an offer accepted and completed by a seller
	OfferCompleted = "completed"
)

var offerStatuses = map[string]bool{
	OfferOffered:   true,
	OfferAccepted:  true,
	OfferCompleted: true,
}

// Offer is a type for price offers a person may give to a seller
type Offer struct {
	ID            int       `json:"id"`
	Price         int       `json:"price_server"`
	PriceClient   string    `json:"price"`
	Counter       int       `json:"counter_server"`
	CounterClient string    `json:"counter"`
	IsCountered   bool      `json:"is_countered"`
	BuyerComment  string    `json:"buyer_comment"`
	SellerComment string    `json:"seller_comment"`
	Status        string    `json:"status"`
	UnreadCount   int       `json:"unread_count"`
	Listing       Listing   `json:"listing"`
	Buyer         User      `json:"buyer"`
	Seller        User      `json:"seller"`
	Created       time.Time `json:"created"`
	Modified      time.Time `json:"modified"`
}

// OfferError contains descriptions of validation errors that may exist in
// an offer object
type OfferError struct {
	Price         string `json:"price"`
	Counter       string `json:"counter"`
	BuyerComment  string `json:"buyer_comment"`
	SellerComment string `json:"seller_comment"`
	Status        string `json:"status"`
	Global        string `json:"global"`
}

// Validate checks if the fields of an offer object are valid
func (o *Offer) Validate() (bool, OfferError) {
	err := OfferError{}
	valid := true
	if o.Price < 0 {
		err.Price = "An offer can't be negative."
		valid = false
	}
	if o.Counter < 0 {
		err.Counter = "A counter offer can't be negative."
		valid = false
	}
	if len(o.BuyerComment) > 140 {
		err.BuyerComment = "A comment on an offer can't be longer than 140 " +
			"characters."
		valid = false
	}
	if len(o.SellerComment) > 140 {
		err.SellerComment = "A comment on an offer can't be longer than 140 " +
			"characters."
		valid = false
	}
	if _, ok := offerStatuses[o.Status]; !ok {
		err.Status = "The offer status is invalid"
		valid = false
	}
	return valid, err
}

// Create inserts an offer into the database
func (o *Offer) Create(db *sql.DB) (bool, *OfferError) {

	o.Status = OfferOffered

	valid, validationError := o.Validate()
	if !valid {
		return valid, &validationError
	}

	rows, err := db.Query("INSERT INTO offers (price, counter, buyer_comment, "+
		"seller_comment, status, listing_id, buyer_id, seller_id) VALUES "+
		"($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id", o.Price, o.Counter,
		o.BuyerComment, o.SellerComment, o.Status, o.Listing.ID, o.Buyer.ID,
		o.Seller.ID)
	if err != nil {
		return false, &OfferError{Global: "An unexpected error occurred."}
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&o.ID)
	} else {
		// The offer was created... but we don't have its ID
		return false, &OfferError{Global: "An unexpected error occurred."}
	}

	return true, nil
}

// Save saves changes made to an offer
func (o *Offer) Save(db *sql.DB) (bool, *OfferError) {

	valid, validationError := o.Validate()
	if !valid {
		return valid, &validationError
	}

	res, err := db.Exec("UPDATE offers SET price = $1, counter = $2, "+
		"is_countered = $3, buyer_comment = $4, seller_comment = $5, "+
		"status = $6, modified = now() WHERE id = $7", o.Price, o.Counter,
		o.IsCountered, o.BuyerComment, o.SellerComment, o.Status, o.ID)
	if err != nil {
		return false, &OfferError{Global: "An unexpected error occurred"}
	}
	numAffected, _ := res.RowsAffected()
	return numAffected == 1, nil
}

// Delete removes an offer from the database
func (o *Offer) Delete(db *sql.DB) bool {
	res, err := db.Exec("DELETE FROM offers WHERE id = $1", o.ID)
	if err != nil {
		return false
	}
	numAffected, _ := res.RowsAffected()
	return numAffected == 1
}

// GetOffers attaches a method to listings which gets all offers for a listing
func (l *Listing) GetOffers(db *sql.DB, pageNum, pageSize int) (
	[]Offer, error) {

	var offers = make([]Offer, 0, 20)
	var numFound = 0

	rows, err := db.Query("SELECT o.id, o.price, o.counter, o.is_countered, "+
		"o.buyer_comment, o.seller_comment, o.status, o.listing_id, o.buyer_id, "+
		"b.username, b.display_name, o.seller_id, s.username, s.display_name, "+
		"o.created, o.modified FROM offers o, users b, users s WHERE "+
		"o.buyer_id = b.id AND o.seller_id = s.id AND listing_id = $1 "+
		"LIMIT $2 OFFSET $3", l.ID, pageSize, pageNum*pageSize)
	if err != nil {
		return offers[:0], err
	}
	defer rows.Close()

	for rows.Next() {
		var offer Offer
		err = rows.Scan(&offer.ID, &offer.Price, &offer.Counter, &offer.IsCountered,
			&offer.BuyerComment, &offer.SellerComment, &offer.Status,
			&offer.Listing.ID, &offer.Buyer.ID, &offer.Buyer.Username,
			&offer.Buyer.DisplayName, &offer.Seller.ID, &offer.Seller.Username,
			&offer.Seller.DisplayName, &offer.Created, &offer.Modified)
		if err == nil {
			offer.PriceClient = utils.PriceServerToClient(offer.Price)
			if offer.IsCountered {
				offer.CounterClient = utils.PriceServerToClient(offer.Counter)
			}
			offers = append(offers, offer)
			numFound++
		}
	}

	return offers[:numFound], nil
}

// GetOfferByID is a utility method that can be used to get fields of an offer
// from a database to help with edits made to an offer
func GetOfferByID(db *sql.DB, id int) (*Offer, error) {
	row := db.QueryRow("SELECT id, price, counter, is_countered, buyer_comment, "+
		"seller_comment, status, listing_id, buyer_id, seller_id, created, "+
		"modified FROM offers WHERE id = $1", id)

	var offer Offer
	err := row.Scan(&offer.ID, &offer.Price, &offer.Counter, &offer.IsCountered,
		&offer.BuyerComment, &offer.SellerComment, &offer.Status, &offer.Listing.ID,
		&offer.Buyer.ID, &offer.Seller.ID, &offer.Created, &offer.Modified)
	if err != nil {
		return nil, err
	}
	offer.PriceClient = utils.PriceServerToClient(offer.Price)
	if offer.IsCountered {
		offer.CounterClient = utils.PriceServerToClient(offer.Counter)
	}
	return &offer, nil

}

// GetOffersAsSeller is attached to user and allows getting offers for a user
// on any listings they have posted
func (u *User) GetOffersAsSeller(db *sql.DB, pageNum, pageSize int) (
	[]Offer, error) {

	var offers = make([]Offer, 0, 50)
	var numFound = 0

	rows, err := db.Query("SELECT o.id, o.price, o.counter, o.is_countered, "+
		"o.buyer_comment, o.seller_comment, o.status, o.listing_id, l.name, "+
		"o.buyer_id, b.username, b.display_name, o.created, o.modified FROM "+
		"offers o, users b, listings l WHERE o.buyer_id = b.id AND "+
		"o.listing_id = l.id AND o.seller_id = $1 ORDER BY modified DESC "+
		"LIMIT $2 OFFSET $3", u.ID, pageSize, pageNum*pageSize)
	if err != nil {
		return offers[:0], err
	}
	defer rows.Close()

	for rows.Next() {
		var offer Offer
		err = rows.Scan(&offer.ID, &offer.Price, &offer.Counter, &offer.IsCountered,
			&offer.BuyerComment, &offer.SellerComment, &offer.Status,
			&offer.Listing.ID, &offer.Listing.Name, &offer.Buyer.ID,
			&offer.Buyer.Username, &offer.Buyer.DisplayName, &offer.Created,
			&offer.Modified)

		if err == nil {
			offer.PriceClient = utils.PriceServerToClient(offer.Price)
			if offer.IsCountered {
				offer.CounterClient = utils.PriceServerToClient(offer.Counter)
			}
			offers = append(offers, offer)
			numFound++
		}
	}

	return offers[:numFound], nil
}

// GetOffersAsBuyer is attached to user and allows getting offers for a user
// on any listings they have made offers for
func (u *User) GetOffersAsBuyer(db *sql.DB) ([]Offer, error) {
	var offers = make([]Offer, 0, 50)
	var numFound = 0

	rows, err := db.Query("SELECT o.id, o.price, o.counter, o.is_countered, "+
		"o.buyer_comment, o.seller_comment, o.status, o.listing_id, l.name, "+
		"l.price, i.url, o.seller_id, s.username, s.display_name, o.created, "+
		"o.modified FROM offers o JOIN users s ON o.seller_id = s.id JOIN "+
		"listings l ON o.listing_id = l.id LEFT JOIN images i ON i.media_id = "+
		"o.listing_id WHERE (i.id = (SELECT id FROM images WHERE media='"+
		MediaListing+"' AND media_id = o.listing_id ORDER BY ordinal ASC LIMIT 1)"+
		"OR i.id IS NULL) AND o.buyer_id = $1", u.ID)
	if err != nil {
		return offers[:0], err
	}
	defer rows.Close()

	for rows.Next() {
		var offer Offer
		err = rows.Scan(&offer.ID, &offer.Price, &offer.Counter, &offer.IsCountered,
			&offer.BuyerComment, &offer.SellerComment, &offer.Status,
			&offer.Listing.ID, &offer.Listing.Name, &offer.Listing.Price,
			&offer.Listing.ImageURL, &offer.Seller.ID, &offer.Seller.Username,
			&offer.Seller.DisplayName, &offer.Created, &offer.Modified)

		if err == nil {
			offer.PriceClient = utils.PriceServerToClient(offer.Price)
			offer.Listing.PriceClient = utils.PriceServerToClient(offer.Listing.Price)
			if offer.Listing.ImageURL == nil {
				offer.Listing.ImageURL = &ImageNotFound
			}
			if offer.IsCountered {
				offer.CounterClient = utils.PriceServerToClient(offer.Counter)
			}
			offers = append(offers, offer)
			numFound++
		}
	}

	return offers[:numFound], nil
}

// GetOfferOnListing attempts to get an offer that a user has made on a
// listing
func (u *User) GetOfferOnListing(db *sql.DB, id int) (*Offer, error) {
	var offer Offer

	row := db.QueryRow("SELECT o.id, o.price, o.counter, o.is_countered, "+
		"o.buyer_comment, o.seller_comment, o.status, o.listing_id, o.seller_id, "+
		"s.username, s.display_name, o.created, o.modified FROM offers o, users s "+
		"WHERE o.seller_id = s.id AND o.buyer_id = $1 AND o.listing_id = $2",
		u.ID, id)

	err := row.Scan(&offer.ID, &offer.Price, &offer.Counter, &offer.IsCountered,
		&offer.BuyerComment, &offer.SellerComment, &offer.Status, &offer.Listing.ID,
		&offer.Seller.ID, &offer.Seller.Username, &offer.Seller.DisplayName,
		&offer.Created, &offer.Modified)
	if err != nil {
		return nil, err
	}

	offer.PriceClient = utils.PriceServerToClient(offer.Price)
	if offer.IsCountered {
		offer.CounterClient = utils.PriceServerToClient(offer.Counter)
	}
	return &offer, nil
}

// GetConversationsForUser gets any offers that can be used for conversations
func (u *User) GetConversationsForUser(db *sql.DB) ([]Offer, error) {
	var offers = make([]Offer, 0, 10)
	var numFound = 0

	rows, err := db.Query("SELECT o.id, o.price, o.counter, "+
		"o.is_countered, o.listing_id, l.name, o.seller_id, s.username, "+
		"s.display_name, o.buyer_id, b.username, b.display_name, "+
		"(SELECT count(1) FROM messages WHERE offer_id = o.id AND "+
		"recepient_id = $1 AND seen = false) unread_count FROM offers o, "+
		"users b, users s, listings l WHERE o.listing_id = l.id AND "+
		"o.buyer_id = b.id AND o.seller_id = s.id AND "+
		"o.status = '"+OfferAccepted+"' AND (o.buyer_id = $1 OR "+
		"o.seller_id = $1)", u.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var offer Offer
		err = rows.Scan(&offer.ID, &offer.Price, &offer.Counter,
			&offer.IsCountered, &offer.Listing.ID, &offer.Listing.Name,
			&offer.Seller.ID, &offer.Seller.Username, &offer.Seller.DisplayName,
			&offer.Buyer.ID, &offer.Buyer.Username, &offer.Buyer.DisplayName,
			&offer.UnreadCount)
		if err != nil {
			continue
		} else {
			offer.PriceClient = utils.PriceServerToClient(offer.Price)
			if offer.IsCountered {
				offer.CounterClient = utils.PriceServerToClient(offer.Counter)
			}
			offers = append(offers, offer)
			numFound++
		}
	}
	return offers, nil
}
