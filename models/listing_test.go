package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

var gDb *sql.DB

var cleanupIDs []int

func getDBConnection() *sql.DB {
	if gDb != nil {
		return gDb
	}

	db, err := sql.Open("postgres",
		"postgres://calagorauser:calagorapassword@localhost:5432/calagora?"+
			"sslmode=disable")

	if err != nil {
		fmt.Println(err)
		return nil
	}
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	gDb = db
	return db
}

func TestCreateListing(t *testing.T) {
	db := getDBConnection()

	listing := Listing{
		Name:        "Test Listing",
		Type:        "textbook",
		Status:      "listed",
		Condition:   "na",
		PriceClient: "30.44",
		Description: "desc",
		Published:   true,
		User: User{
			ID:      2,
			PlaceID: 1,
		},
	}

	if succ, err := listing.Create(db); !succ {
		t.Errorf("An error was reported in listing.")
		bytes, err := json.Marshal(err)
		if err != nil {
			t.Errorf("Couldn't form string for error returned")
		} else {
			t.Errorf(string(bytes))
		}
	}

	cleanupIDs = append(cleanupIDs, listing.ID)

	rows, err := db.Query("SELECT listings.id, listings.name, type, status, "+
		"price, description, condition, published, users.place_id, user_id FROM "+
		"listings, users WHERE listings.user_id = users.id AND listings.id = $1",
		listing.ID)
	if err != nil {
		t.Errorf("Error occurred while attempting to query for created listing: %s",
			err)
		t.Fail()
		return
	}

	fmt.Println("Created listing with ID", listing.ID)

	if rows.Next() {
		var found Listing
		err = rows.Scan(&found.ID, &found.Name, &found.Type, &found.Status,
			&found.Price, &found.Description, &found.Condition, &found.Published,
			&found.User.PlaceID, &found.User.ID)
		if err != nil {
			t.Fatalf("Error occurred while attempting to get created listing: %s",
				err)
		}
		var comparisons bool
		comparisons = listing.ID == found.ID
		comparisons = comparisons && strings.Compare(found.Type, "textbook") == 0
		comparisons = comparisons && strings.Compare(found.Status, "listed") == 0
		comparisons = comparisons && strings.Compare(found.Description, "desc") == 0
		comparisons = comparisons && found.Published
		comparisons = comparisons && found.User.PlaceID == 1
		comparisons = comparisons && found.User.ID == 2
		comparisons = comparisons && found.Price == 3044
		if !comparisons {
			t.Error("Created listing was incorrect")
			jsn, err := json.Marshal(listing)
			if err == nil {
				t.Error("Expected: ", string(jsn))
			}
			jsn, err = json.Marshal(found)
			if err == nil {
				t.Error("Got: ", string(jsn))
			}
		}
	} else {
		t.Error("No listing was created.")
	}
}

func Cleanup(db *sql.DB) {
	for _, id := range cleanupIDs {
		var delete = Listing{
			ID: id,
		}
		retval, err := delete.Delete(db)
		if !retval {
			fmt.Fprintln(os.Stderr, "Error deleting listing with id "+
				strconv.Itoa(id))
			fmt.Fprintln(os.Stderr, err)
		} else {
			fmt.Println("Deleted test listing " + strconv.Itoa(id))
		}
	}
}

func TestMain(m *testing.M) {
	db := getDBConnection()
	if db == nil {
		fmt.Fprintln(os.Stderr, "Could not connect to database.")
		os.Exit(1)
	}
	returnCode := m.Run()
	Cleanup(db)
	os.Exit(returnCode)
}
