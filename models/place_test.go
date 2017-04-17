package models

import (
	"strconv"
	"testing"
)

func TestFindDomain(t *testing.T) {
	db := getDBConnection()
	id, err := FindPlaceID(db, "test@rutgers.edu")
	if err != nil {
		t.Error("Error: " + err.Error())
		t.Fail()
	}
	if id != 1 {
		t.Error("Got ID " + strconv.Itoa(id) + ", expected 1")
		t.Fail()
	}

	id, err = FindPlaceID(db, "amg380@scarletmail.rutgers.edu")
	if err != nil {
		t.Error("Error: " + err.Error())
		t.Fail()
	}

	if id != 1 {
		t.Error("Got ID " + strconv.Itoa(id) + ", expected 1")
		t.Fail()
	}

	id, err = FindPlaceID(db, "test@cuny.edu")
	if err != nil {
		t.Error("Error: " + err.Error())
		t.Fail()
	}

	if id != 2 {
		t.Error("Got ID " + strconv.Itoa(id) + ", expected 2")
		t.Fail()
	}
}
