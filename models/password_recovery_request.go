package models

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	mrand "math/rand"
	"time"
)

const (
	// RecoveryStringLength is how long random recovery strings will be
	RecoveryStringLength = 48
	// RecoveryCodeLength is how long random recovery codes will be
	RecoveryCodeLength = 7

	recoveryCodeSource = "ABCDEFGHIJKLMNOPQRSTUVWXYZ23456789"
)

// PasswordRecoveryRequest is the encapsulation for a password recovery
// attempt
type PasswordRecoveryRequest struct {
	User           User      `json:"-"`
	IsValid        bool      `json:"is_valid"`
	RecoveryString string    `json:"-"`
	RecoveryCode   string    `json:"-"`
	Created        time.Time `json:"created"`
	Modified       time.Time `json:"modified"`
}

// PasswordRecoveryError encapsulates errors that occur during password
// recovery
type PasswordRecoveryError struct {
	UserNotFound    bool   `json:"user_not_found"`
	HasOtherRequest bool   `json:"has_other_request"`
	EmailAddress    string `json:"email_address"`
	HasError        bool   `json:"has_error"`
	Global          string `json:"global"`
}

func generateRecoveryString() (string, error) {
	bytes := make([]byte, RecoveryStringLength)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		return "", err
	}

	encodedRecoveryString := base64.URLEncoding.EncodeToString(bytes)
	return encodedRecoveryString, nil
}

func generateRecoveryCode() (string, error) {
	var buff bytes.Buffer
	for i := 0; i < RecoveryCodeLength; i++ {
		index := mrand.Intn(len(recoveryCodeSource))
		buff.WriteByte(recoveryCodeSource[index])
	}
	return buff.String(), nil
}

// DeleteExpiredRecoveryRequests removes any recovery requests that
// are older than a day.
func (u *User) DeleteExpiredRecoveryRequests(db *sql.DB) error {
	statement := "DELETE FROM password_recovery_requests WHERE user_id = $1 " +
		"AND modified < now() - interval '1 DAY'"
	_, err := db.Exec(statement, u.ID)
	return err
}

// ExtendRecoveryRequest invalidates a prior recovery request, and sets
// the expiry time to 7 days from now
func (u *User) ExtendRecoveryRequest(db *sql.DB) error {
	statement := "UPDATE password_recovery_requests SET modified = now() + " +
		"interval '7 DAY', is_valid = false WHERE user_id = $1"
	_, err := db.Exec(statement)
	return err
}

// HasRecoveryRequest determines whether or not a user has made a recovery
// request
func (u *User) HasRecoveryRequest(db *sql.DB) (bool, error) {
	query := "SELECT COUNT(1) FROM password_recovery_requests WHERE user_id = $1"
	res := db.QueryRow(query, u.ID)

	var numRecords int
	err := res.Scan(&numRecords)
	if err != nil {
		return false, err
	}

	return numRecords > 0, nil
}

// Create takes a user's information and populates fields of a password
// recovery request, IF there is no existing request.
func (prr *PasswordRecoveryRequest) Create(db *sql.DB) error {
	recoveryString, err := generateRecoveryString()
	if err != nil {
		return err
	}
	prr.RecoveryString = recoveryString
	recoveryCode, err := generateRecoveryCode()
	if err != nil {
		return err
	}
	prr.RecoveryCode = recoveryCode

	statement := "INSERT INTO password_recovery_requests (user_id, " +
		"recovery_string, recovery_code, is_valid) VALUES ($1, $2, $3, true)"
	_, err = db.Exec(statement, prr.User.ID, prr.RecoveryString,
		prr.RecoveryCode)

	return err
}

// Delete removes a password recovery request after it has successfully been
// used to reset a password
func (prr *PasswordRecoveryRequest) Delete(db *sql.DB) error {
	statement := "DELETE FROM password_recovery_requests WHERE user_id = $1"
	_, err := db.Exec(statement, prr.User.ID)
	if err != nil {
		fmt.Println("PasswordRecoveryRequest.Delete(): " + err.Error())
	}
	return err
}

// Invalidate marks a password recovery request as invalid, and sets its
// modified date into the future to keep it in the database for seven days
// from current date
func (prr *PasswordRecoveryRequest) Invalidate(db *sql.DB) error {
	statement := "UPDATE password_recovery_requests SET is_valid = false, " +
		"modified = now() + interval '6 DAY' WHERE user_id = $1"
	_, err := db.Exec(statement, prr.User.ID)
	return err
}

// GetPasswordRecoveryRequest gets the last recovery request made for a user
func (u *User) GetPasswordRecoveryRequest(db *sql.DB) (*PasswordRecoveryRequest,
	error) {

	query := "SELECT recovery_string, recovery_code, is_valid, created, " +
		"modified FROM password_recovery_requests WHERE user_id = $1"
	res := db.QueryRow(query, u.ID)

	var prr PasswordRecoveryRequest
	prr.User = *u
	err := res.Scan(&prr.RecoveryString, &prr.RecoveryCode, &prr.IsValid,
		&prr.Created, &prr.Modified)
	if err != nil {
		return nil, err
	}
	return &prr, nil
}
