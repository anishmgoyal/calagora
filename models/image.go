package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/anishmgoyal/calagora/utils"
)

const (
	// MediaListing denotes an image that relates to a listing
	MediaListing = "listing"
	// MaxListingImages is the number of images allowed per listing
	MaxListingImages = 8
)

// ImageNotFound is the url to the notfound.jpg image to be displayed
// when no other image is available
var ImageNotFound = "/img/notfound"

// Image defines an image struct which holds information about
// uploaded images
type Image struct {
	ID       int       `json:"id"`
	Media    string    `json:"media"`
	MediaID  int       `json:"media_id"`
	Ordinal  int       `json:"ordinal"`
	URL      string    `json:"url"`
	User     User      `json:"user"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// Validate ensures that a user doesn't exceed their maximum allotted
// images
func (i *Image) Validate(db *sql.DB) bool {
	if strings.Compare(i.Media, MediaListing) == 0 {
		listing := Listing{ID: i.MediaID}
		count, err := listing.GetImageCount(db)
		if err != nil || count > 8 {
			return false
		}
	}

	return true
}

// Create saves an image model to the database
func (i *Image) Create(db *sql.DB) (bool, error) {
	rows, err := db.Query("INSERT INTO images (media, media_id, ordinal, url, "+
		"user_id) VALUES ($1, $2, $3, $4, $5) RETURNING id", i.Media, i.MediaID,
		i.Ordinal, i.URL, i.User.ID)

	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&i.ID)
		if err != nil {
			return true, err
		}

		if !i.Validate(db) {
			i.Delete(db)
			return false, errors.New("Exceeded image count limit")
		}

		return true, nil
	}

	return true, errors.New("Failed to fetch ID")
}

// Save saves changes to an image model
func (i *Image) Save(db *sql.DB) (bool, error) {
	res, err := db.Exec("UPDATE images SET media = $1, media_id = $2, "+
		"ordinal = $3, url = $4, modified = now() WHERE id = $5", i.Media,
		i.MediaID, i.Ordinal, i.URL, i.ID)

	if err != nil {
		return false, err
	}

	numAffected, _ := res.RowsAffected()
	return numAffected == 1, errors.New("Updated an unexpected number of rows")
}

// Delete delets an image model from the database
func (i *Image) Delete(db *sql.DB) (bool, error) {
	if len(i.URL) > 0 && !utils.DeleteImage(i.URL) {
		return false, errors.New("Failed to delete image from S3")
	}

	res, err := db.Exec("DELETE FROM images WHERE id = $1", i.ID)
	if err != nil {
		return false, err
	}
	numAffected, _ := res.RowsAffected()
	return numAffected == 1, errors.New("Deleted unexpected number of rows")
}

// GetImageByID attempts to get an image by its ID
func GetImageByID(db *sql.DB, id int) (*Image, error) {
	row := db.QueryRow("SELECT id, media, media_id, ordinal, url, "+
		"user_id, created, modified FROM images WHERE id = $1", id)
	var image Image
	err := row.Scan(&image.ID, &image.Media, &image.MediaID, &image.Ordinal,
		&image.URL, &image.User.ID, &image.Created, &image.Modified)
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// GetImages gets all of a listing's images from the database
func (l *Listing) GetImages(db *sql.DB) ([]Image, error) {
	images := make([]Image, 0, 8)
	numFound := 0
	rows, err := db.Query("SELECT id, media, media_id, ordinal, url, "+
		"user_id, created, modified FROM images WHERE media = '"+
		MediaListing+"' AND media_id = $1 ORDER BY ordinal, created ASC",
		l.ID)
	if err != nil {
		return images[:0], err
	}
	defer rows.Close()

	for rows.Next() {
		image := Image{}
		rows.Scan(&image.ID, &image.Media, &image.MediaID, &image.Ordinal, &image.URL,
			&image.User.ID, &image.Created, &image.Modified)
		images = append(images, image)
		numFound++
	}

	return images[:numFound], nil
}

// GetImageCount gets the number of images associated with a listing
func (l *Listing) GetImageCount(db *sql.DB) (int, error) {
	count := 0
	row := db.QueryRow("SELECT COUNT(*) FROM images WHERE media = '"+MediaListing+
		"' AND media_id = $1", l.ID)
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// UpdatePrimaryImage sets the ordinal for a given image to 1, and all others
// to 0
func (l *Listing) UpdatePrimaryImage(db *sql.DB, id int) (bool, error) {
	_, err := db.Exec("UPDATE images SET ordinal = 0 WHERE media = $1 AND "+
		"media_id = $2", MediaListing, l.ID)
	if err != nil {
		return false, err
	}
	_, err = db.Exec("UPDATE images SET ordinal = -1 WHERE media = $1 AND "+
		"media_id = $2 AND id = $3", MediaListing, l.ID, id)
	if err != nil {
		return false, err
	}
	return true, nil
}
