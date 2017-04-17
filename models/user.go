package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/anishmgoyal/calagora/constants"

	"golang.org/x/crypto/scrypt"
)

const (
	// SaltLen is how long random salt strings will be
	SaltLen = 32
	// PwLen is how long the password hash will be
	PwLen = 64
	// ActivationLen is how long activation strings will be
	ActivationLen = 48
)

// User contains fields that pertain to a specific
// ... well, user ... of Calagora
type User struct {
	ID                   int    `json:"id"`
	Username             string `json:"username"`
	DisplayName          string `json:"display_name"`
	EmailAddress         string `json:"email_address"`
	Password             string `json:"-"`
	PasswordConfirmation string `json:"-"`
	Salt                 string `json:"-"`
	Activation           string `json:"-"`
	PlaceID              int    `json:"-"`
	PlaceName            string `json:"place"`
}

// UserLimited is a version of user which is rendered
// without some fields in user when rendered as JSON
type UserLimited struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	DisplayName  string `json:"display_name"`
	EmailAddress string `json:"-"`
	Password     string `json:"-"`
	Salt         string `json:"-"`
}

// UserError contains error messages for each field in
// User if validation fails
// ID and Salt cannot be invalid, they're handled by the server
type UserError struct {
	Username             string `json:"username,omitempty"`
	DisplayName          string `json:"display_name,omitempty"`
	EmailAddress         string `json:"email_address,omitempty"`
	Password             string `json:"password,omitempty"`
	PasswordConfirmation string `json:"password_confirmation,omitempty"`
	Global               string `json:"global,omitempty"`
}

// GetDisplayName gets a display name
func (user *User) GetDisplayName() string {
	return user.Username
}

// GetEmailAddress gets an email address
func (user *User) GetEmailAddress() string {
	return user.EmailAddress
}

// GetID gets an ID
func (user *User) GetID() int {
	return user.ID
}

// GetUsername gets a username
func (user *User) GetUsername() string {
	return user.Username
}

func encryptPassword(sourcePassword string, salt []byte) (
	string, string, error) {

	if salt == nil {
		salt = make([]byte, SaltLen)
		_, err := io.ReadFull(rand.Reader, salt)
		if err != nil {
			return "", "", errors.New("Failed to generate salt")
		}
	}

	encryptedPassword, err := scrypt.Key([]byte(sourcePassword), salt, 16384, 8,
		1, 32)
	if err != nil {
		return "", "", errors.New("Failed to encrypt")
	}

	encodedPassword := base64.StdEncoding.EncodeToString(encryptedPassword)
	encodedSalt := base64.StdEncoding.EncodeToString(salt)

	return encodedPassword, encodedSalt, nil
}

func generateActivationString() (string, error) {

	bytes := make([]byte, ActivationLen)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		return "", err
	}

	encodedActivation := base64.URLEncoding.EncodeToString(bytes)
	return encodedActivation, nil
}

func comparePassword(sourcePassword string,
	encodedPassword string,
	encodedSalt string,
	forceFail bool) bool {

	salt, err := base64.StdEncoding.DecodeString(encodedSalt)
	if err != nil {
		return false
	}

	encryptedPassword, err := scrypt.Key([]byte(sourcePassword), salt, 16384, 8,
		1, 32)
	if err != nil {
		return false
	}

	encodedAttempt := base64.StdEncoding.EncodeToString(encryptedPassword)

	return strings.Compare(encodedAttempt, encodedPassword) == 0 && !forceFail
}

// Authenticate checks if a given password is the one supplied by a user
func (user *User) Authenticate(password string) (bool, error) {

	var forceFail bool
	if user != nil {
		forceFail = false
	} else {
		user = &User{}
		forceFail = true
	}

	if comparePassword(password, user.Password, user.Salt, forceFail) {
		if strings.Compare(user.Activation, "ACTIVATION_ACTIVE") != 0 {
			return false, errors.New("Your account has not been activated yet.")
		}
		return true, nil
	}

	return false, errors.New("Invalid username or password")
}

// Validate checks if the fields in user are valid.
// If not, returns a UserError object containing error messages
func (user *User) Validate(db *sql.DB, validatePassword bool,
	validateUserExists bool) (bool, UserError) {

	var userError UserError
	var valid = true

	usernameMatch, _ :=
		regexp.MatchString("^[a-zA-Z0-9_ ]{6,20}$", user.Username)

	if !usernameMatch {
		userError.Username = "Usernames must be between 6 and 20 characters, and " +
			"can only contain letters, numbers, and spaces."
		valid = false
	} else if validateUserExists && user.checkExistsUsername(db) {
		userError.Username = "That username is taken, sorry!"
		valid = false
	}

	if len(user.DisplayName) < 1 || len(user.DisplayName) > 100 {
		userError.DisplayName = "Display names may be between 1 and 100 characters."
		valid = false
	}

	// This regex is not quite meant to prevent invalid email addresses
	// but rather to help ensure the user doesn't accidentally mistype something
	// After all, email validation is required
	emailMatch, _ := regexp.MatchString(
		"^[[:alnum:]\\._%+\\-!#$%&'\\*/=\\?\\^_`{\\|}~\"]+@"+
			"(([[:alnum:]_\\-]+\\.)*[[:alnum:]_\\-]+\\.[[:alnum:]_\\-]{1,10}|"+
			"\\[.+\\])$",
		user.EmailAddress)

	if !emailMatch {
		userError.EmailAddress = "The email address you provided is invalid."
		valid = false
	}

	place, err := FindPlaceID(db, user.EmailAddress)
	if err != nil {
		userError.Global = "An unexpected error occurred."
		valid = false
	}

	if place == -1 {
		userError.Global = "We don't have a school on record for your " +
			"email address. Please contact " + constants.SupportEmail + " if you " +
			"would like your school to be added to Calagora."
		valid = false
	}
	user.PlaceID = place

	if validatePassword {
		if len(user.Password) < 8 || len(user.Password) > 90 {
			userError.Password = "Passwords may be between 8 and 90 characters."
			valid = false
		}

		if strings.Compare(user.Password, user.PasswordConfirmation) != 0 {
			userError.PasswordConfirmation = "Your passwords didn't match."
			valid = false
		}
	}

	return valid, userError
}

// Normalize alters members of User, making them conform to
// the following standards (may be expanded):
// Username: lowercase
func (user *User) Normalize() {
	if len(user.Username) < 1000 && len(user.DisplayName) < 1000 &&
		len(user.EmailAddress) < 1000 {

		user.Username = strings.TrimSpace(strings.ToLower(user.Username))
		user.DisplayName = strings.TrimSpace(user.DisplayName)
		user.EmailAddress = strings.TrimSpace(strings.ToLower(user.EmailAddress))
	}
}

func (user *User) checkExistsUsername(db *sql.DB) bool {
	rows, err := db.Query("SELECT COUNT(1) AS c FROM users WHERE username = $1",
		user.Username)
	if err != nil {
		return true // Just in case
	}
	defer rows.Close()

	if rows.Next() {
		var count int
		rows.Scan(&count)
		return (count > 0)
	}

	return true // Should't be reached, but safe fallback

}

// Create for User attempts to insert a new user into
// the database
func (user *User) Create(db *sql.DB) (bool, *UserError) {
	user.Normalize()
	valid, validationError := user.Validate(db, true, true)
	if !valid {
		return false, &validationError
	}

	encryptedPassword, salt, error := encryptPassword(user.Password, nil)
	if error != nil {
		return false, &UserError{
			Global: "An unexpected error occurred.",
		}
	}

	activationString, error := generateActivationString()
	if error != nil {
		return false, &UserError{
			Global: "An unexpected error occurred.",
		}
	}

	user.Password = encryptedPassword
	user.Salt = salt

	rows, err := db.Query("INSERT INTO users (username, display_name, "+
		"email_address, password, salt, activation, place_id) VALUES "+
		"($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		user.Username, user.DisplayName, user.EmailAddress, user.Password,
		user.Salt, activationString, user.PlaceID)

	if err != nil {
		fmt.Println("ERROR!")
		fmt.Println(err)
		return false, &UserError{
			Global: "Your account could not be added. Please try again later.",
		}
	}

	defer rows.Close()
	if rows.Next() {
		rows.Scan(&user.ID)
	}

	user.Activation = activationString
	return true, nil
}

// Activate sets the activation for a user to active
func (user *User) Activate(db *sql.DB) error {
	if user == nil {
		return errors.New("No user specified.")
	}
	_, err := db.Exec("UPDATE users SET activation = 'ACTIVATION_ACTIVE' "+
		"WHERE id = $1", user.ID)
	return err
}

// Save attempts to save changes made to a user profile
func (user *User) Save(db *sql.DB) (bool, *UserError) {
	if user == nil {
		return false, &UserError{
			Global: "User not found",
		}
	}

	user.Normalize()
	validatePassword := len(user.Password) > 0
	valid, validationErr := user.Validate(db, validatePassword, false)
	if !valid {
		return valid, &validationErr
	}

	statement := "UPDATE users SET display_name = $1"
	argCount := 2
	args := make([]interface{}, 0, 4)
	args = append(args, user.DisplayName)

	if len(user.Password) > 0 {
		encryptedPassword, salt, err := encryptPassword(user.Password, nil)
		if err != nil {
			return false, &UserError{
				Global: "An unexpected error occurred.",
			}
		}

		args = append(args, encryptedPassword, salt)

		statement = statement + ", password = $2, salt = $3"
		argCount = 4
	}

	statement = statement + " WHERE id = $" + strconv.Itoa(argCount)
	args = append(args, user.ID)

	res, err := db.Exec(statement, args[:argCount]...)
	if err != nil {
		fmt.Println("ERROR!")
		fmt.Println(err.Error())
		return false, &UserError{
			Global: "An unexpected error occurred, and your profile could not be " +
				"saved. Please refresh the page or try again later. If this issue " +
				"persists, please contact support at " + constants.SupportEmail + ".",
		}
	}

	affected, _ := res.RowsAffected()
	return affected == 1, &UserError{
		Global: "An unexpected error occurred.",
	}
}

// GetUserByUsername pulls a user by username
// from the database
func GetUserByUsername(db *sql.DB, username string) *User {
	username = strings.TrimSpace(strings.ToLower(username))
	rows, err := db.Query("SELECT id, display_name, "+
		"email_address, password, salt, activation, place_id FROM users "+
		"WHERE username = $1", username)
	if err != nil {
		fmt.Println("ERROR!")
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	user := User{Username: username}

	if rows.Next() {
		err = rows.Scan(&user.ID, &user.DisplayName, &user.EmailAddress,
			&user.Password, &user.Salt, &user.Activation, &user.PlaceID)
	} else {
		return nil
	}

	if err != nil {
		return nil
	}

	return &user
}

// GetUserByEmailAddress gets a Calagora user by his/her email address
func GetUserByEmailAddress(db *sql.DB, emailAddress string) (*User, error) {
	emailAddress = strings.TrimSpace(strings.ToLower(emailAddress))
	query := "SELECT id, username, display_name, place_id FROM users WHERE " +
		"email_address = $1"
	rows, err := db.Query(query, emailAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var user User
		user.EmailAddress = emailAddress
		err := rows.Scan(&user.ID, &user.Username, &user.DisplayName, &user.PlaceID)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}
	return nil, nil
}

// GetUserByID gets a Calagora user by his/her
// user id
func GetUserByID(db *sql.DB, ID int) *User {
	rows, err := db.Query("SELECT username, display_name, "+
		"email_address, password, salt, activation FROM users WHERE"+
		" id = $1", ID)
	if err != nil {
		fmt.Println("ERROR!")
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	user := User{ID: ID}
	if rows.Next() {
		err = rows.Scan(&user.Username, &user.DisplayName, &user.EmailAddress,
			&user.Password, &user.Salt, &user.Activation)
	} else {
		return nil
	}

	if err != nil {
		return nil
	}

	return &user
}
