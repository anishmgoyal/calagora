package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anishmgoyal/calagora/utils"
)

//RandStringLen is the length of an unencoded random string
const RandStringLen = 45

// Session defines the fields used to keep track of a user's presence
// on the webapp
type Session struct {
	User          User
	SessionID     string
	SessionSecret string
	CsrfToken     string
	BrowserAgent  string
	Created       time.Time
	Modified      time.Time
}

func randomString() (string, error) {
	bytes := make([]byte, RandStringLen)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		return "", err
	}

	encodedString := base64.StdEncoding.EncodeToString(bytes)
	return encodedString, nil
}

// Create inserts a session into the database
// DO NOT CALL unless the user has already been authenticated
// in a controller.
func (s *Session) Create(db *sql.DB) (bool, error) {
	var sessionID string
	var err error

	var validSessionToken = false
	for i := 0; i < 10; i++ {
		sessionID, err = randomString()
		if err != nil {
			return false, err
		}
		if !isSessionIDTaken(db, sessionID) {
			validSessionToken = true
			break
		}
	}

	if !validSessionToken {
		return false, errors.New("Failed to generate session id for new session.")
	}

	s.SessionID = sessionID

	sessionSecret, err := randomString()
	if err != nil {
		return false, err
	}
	s.SessionSecret = sessionSecret

	csrfToken, err := randomString()
	if err != nil {
		return false, err
	}
	s.CsrfToken = csrfToken

	if len(s.BrowserAgent) > 200 {
		s.BrowserAgent = s.BrowserAgent[:200]
	}

	res, err := db.Exec("INSERT INTO sessions (session_id, session_secret, "+
		"csrf_token, browser_agent, user_id) VALUES ($1, $2, $3, $4, $5)",
		s.SessionID, s.SessionSecret, s.CsrfToken, s.BrowserAgent, s.User.ID)

	if err != nil {
		fmt.Println("ERROR!")
		fmt.Println(err.Error())
		return false, err
	}

	affected, _ := res.RowsAffected()
	return affected == 1, errors.New("An unexpected error occurred.")
}

// Update updates the modified timestamp of a session so that it doesn't
// time out if it is active
func (s *Session) Update(db *sql.DB) {
	db.Exec("UPDATE sessions SET modified = now() WHERE " +
		"session_id = $1 AND session_secret = $2")
}

func isSessionIDTaken(db *sql.DB, sessionID string) bool {

	rows, err := db.Query("SELECT COUNT(*) FROM sessions WHERE session_id = $1",
		sessionID)
	if err != nil {
		return false
	}
	defer rows.Close()

	var numFound int
	if rows.Next() {
		rows.Scan(&numFound)
	} else {
		return false
	}
	if numFound > 0 {
		return true
	}

	return false
}

// GetSessionFromRequest checks if the current request
// has session data; if so, attempts to get a session
func GetSessionFromRequest(db *sql.DB, w http.ResponseWriter,
	r *http.Request) *Session {

	sessionID := utils.GetCookie(r, "session_id")
	if len(sessionID) == 0 {
		return nil
	}

	sessionSecret := utils.GetCookie(r, "session_secret")
	if len(sessionSecret) == 0 {
		utils.DeleteCookie(w, "session_id")
		return nil
	}

	session := GetSession(db, sessionID, sessionSecret, r.Header.Get("User-Agent"))
	if session == nil {
		utils.DeleteCookie(w, "session_id")
		utils.DeleteCookie(w, "session_secret")
	} else {
		// Update the cookies
		utils.SetCookie(w, "session_id", sessionID, 14)
		utils.SetCookie(w, "session_secret", sessionSecret, 14)
	}
	return session
}

// GetSession attempts to retrieve a session with the given userid, sessionid,
// and browser agent string
func GetSession(db *sql.DB, sessionID string, sessionSecret string,
	browserAgent string) *Session {

	if len(browserAgent) > 200 {
		browserAgent = browserAgent[:200]
	}
	rows, err := db.Query("SELECT u.id, u.username, u.display_name, "+
		"u.email_address, u.place_id, s.csrf_token, s.created, s.modified FROM "+
		"sessions s, users u WHERE s.user_id = u.id AND s.session_id = $1 AND "+
		"s.session_secret = $2 AND s.browser_agent = $3", sessionID, sessionSecret,
		browserAgent)
	if err != nil {
		fmt.Println("ERROR!")
		fmt.Println(err.Error())
		return nil
	}
	defer rows.Close()

	session := Session{SessionID: sessionID, SessionSecret: sessionSecret,
		BrowserAgent: browserAgent}

	if rows.Next() {
		err = rows.Scan(&session.User.ID, &session.User.Username,
			&session.User.DisplayName, &session.User.EmailAddress,
			&session.User.PlaceID, &session.CsrfToken, &session.Created,
			&session.Modified)
	} else {
		return nil
	}

	currentTime := time.Now()
	evictionTime := currentTime.AddDate(0, 0, -14)
	if evictionTime.After(session.Modified) {
		fmt.Println("Session eviction")
		session.Delete(db)
		return nil
	}

	if err != nil {
		return nil
	}

	return &session
}

// Delete deletes a session
func (s *Session) Delete(db *sql.DB) bool {

	res, err := db.Exec("DELETE FROM sessions WHERE session_id = $1",
		s.SessionID)

	if err != nil {
		fmt.Println("ERROR!")
		fmt.Println(err.Error())
		return false
	}

	numAffected, _ := res.RowsAffected()
	return numAffected == 1
}
