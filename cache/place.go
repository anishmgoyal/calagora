package cache

import "github.com/anishmgoyal/calagora/models"

// The implementation for this cache is simple, since we can
// hold information about every place in memory for now. There are only
// two that'll be present at launch

var places map[int]*models.Place

func initPlaceCache() {
	places = make(map[int]*models.Place)
}

// GetPlaceByID gets a place in a cacheable manner
func GetPlaceByID(id int) (*models.Place, error) {
	var err error
	var place *models.Place
	var ok bool
	if place, ok = places[id]; !ok {
		place, err = models.GetPlaceByID(Base.Db, id)
		if place != nil {
			places[id] = place
		}
	}
	return place, err
}

// MapPlaceToListing maps a single place to a single listing
func MapPlaceToListing(listing *models.Listing) {
	if listing == nil {
		return
	}
	place, err := GetPlaceByID(listing.User.PlaceID)
	if place != nil && err == nil {
		listing.User.PlaceName = place.Name
	}
}

// MapPlaceToListings uses the cache to set place info for a listing
func MapPlaceToListings(listings []models.Listing) {
	for i := 0; i < len(listings); i++ {
		MapPlaceToListing(&listings[i])
	}
}
