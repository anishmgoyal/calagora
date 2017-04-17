package models

import (
	"database/sql"
	"fmt"
	"time"
)

// Message is a single message sent under an offer
type Message struct {
	ID        int       `json:"id"`
	Message   string    `json:"message"`
	Seen      bool      `json:"seen"`
	Sender    User      `json:"sender"`
	Recepient User      `json:"recepient"`
	Offer     Offer     `json:"offer"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

// MessageError contains error messages from failures in validations
type MessageError struct {
	Message string `json:"message"`
	Global  string `json:"global"`
}

// Validate checks if a message is valid
func (m *Message) Validate() (bool, MessageError) {
	var err MessageError
	var valid = true
	if len(m.Message) == 0 {
		err.Message = "Can't send blank messages"
		valid = false
	} else if len(m.Message) > 255 {
		err.Message = "Messages can't be longer than 255 characters"
		valid = false
	}
	return valid, err
}

// Create inserts a message into the database (send)
func (m *Message) Create(db *sql.DB) (bool, *MessageError) {
	valid, validationError := m.Validate()
	if !valid {
		return valid, &validationError
	}

	row := db.QueryRow("INSERT INTO messages (message, sender_id, "+
		"recepient_id, offer_id, seen) VALUES ($1, $2, $3, $4, false) RETURNING id",
		m.Message, m.Sender.ID, m.Recepient.ID, m.Offer.ID)

	err := row.Scan(&m.ID)
	if err != nil {
		return false, &MessageError{Global: "Unexpected Error"}
	}
	return true, nil
}

// MarkRead attempts to mark a message as read; silently fails
// if an error occurs
func (m *Message) MarkRead(db *sql.DB, recepientID int) {
	m.Seen = true
	_, err := db.Exec("UPDATE messages SET seen = true WHERE id = $1 AND "+
		"recepient_id = $2", m.ID, recepientID)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// GetMessages is attached to an Offer and can be used to get all messages
// sent in a conversation
func (o *Offer) GetMessages(db *sql.DB, pageSize, page int, recepientID int) (
	[]Message, error) {

	var messages = make([]Message, 0, pageSize)
	var numFound = 0

	rows, err := db.Query("SELECT m.id, m.message, m.seen, m.sender_id, "+
		"s.username, s.display_name, m.recepient_id, m.created, m.modified FROM "+
		"messages m, users s WHERE m.sender_id = s.id AND m.offer_id = $1 ORDER BY "+
		"created DESC LIMIT $2 OFFSET $3", o.ID, pageSize, (page-1)*pageSize)

	if err != nil {
		return messages[:0], err
	}
	defer rows.Close()

	for rows.Next() {
		var message Message
		err = rows.Scan(&message.ID, &message.Message, &message.Seen, &message.Sender.ID,
			&message.Sender.Username, &message.Sender.DisplayName, &message.Recepient.ID,
			&message.Created, &message.Modified)
		if err != nil {
			continue
		}
		message.Offer.ID = o.ID

		if message.Recepient.ID == recepientID {
			message.MarkRead(db, recepientID)
		}

		messages = append(messages, message)
		numFound++
	}
	return messages[:numFound], nil
}

// GetLastMessage gets the newest message (if one exists) for a conversation
// held together by an Offer
func (o *Offer) GetLastMessage(db *sql.DB) (*Message, error) {
	row := db.QueryRow("SELECT id, message, seen, sender_id, created, modified "+
		"FROM messages WHERE offer_id = $1 ORDER BY created DESC LIMIT 1",
		o.ID)
	var message Message
	err := row.Scan(&message.ID, &message.Message, &message.Seen,
		&message.Sender.ID, &message.Created, &message.Modified)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// GetUnreadMessageCount attempts to determine how many unread
// messages a user has. Returns 0 if none, or on error.
func (u *User) GetUnreadMessageCount(db *sql.DB) int {
	row := db.QueryRow("SELECT COUNT(*) FROM messages WHERE recepient_id = $1 "+
		"AND seen = false", u.ID)
	var messageCount int
	err := row.Scan(&messageCount)
	if err != nil {
		return 0
	}
	return messageCount
}
