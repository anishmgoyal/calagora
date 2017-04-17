package models

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	// MaxNotificationsToKeep determines how many notifications we save for
	// a user before starting to delete old ones
	MaxNotificationsToKeep = 500
	// NotificationsPerPage determines how many notifications can be loaded
	// in one go
	NotificationsPerPage = 50
)

// Notification encapsulates data needed to create notification
// blocks on the UI end
type Notification struct {
	ID      int       `json:"id"`
	User    User      `json:"user"`
	Value   string    `json:"value"`
	Read    bool      `json:"read"`
	Created time.Time `json:"created"`
}

// Create saves a notification to the database
func (n *Notification) Create(db *sql.DB) (bool, error) {
	rows, err := db.Query("INSERT INTO notifications (user_id, "+
		"is_read, notification_value) VALUES ($1, $2, $3) RETURNING id", n.User.ID,
		false, n.Value)
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&n.ID)
	}

	_, err = db.Exec("DELETE FROM notifications WHERE user_id = $1 AND "+
		"id = any(array(SELECT id FROM notifications WHERE user_id = $1 "+
		"ORDER BY id DESC OFFSET $2))",
		n.User.ID, MaxNotificationsToKeep)
	if err != nil {
		fmt.Println(err.Error())
	}
	return true, nil
}

// GetRecentNotifications gets a page of notifications for a user,
// and marks them all as read
func (u *User) GetRecentNotifications(db *sql.DB, page int) []Notification {
	rows, err := db.Query("SELECT id, notification_value, is_read, created FROM "+
		"notifications WHERE user_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3",
		u.ID, NotificationsPerPage, NotificationsPerPage*page)
	if err != nil {
		fmt.Println(err.Error())
		return []Notification{}
	}
	defer rows.Close()
	res := make([]Notification, 0, 50)
	for rows.Next() {
		var notification Notification
		notification.User = *u
		err = rows.Scan(&notification.ID, &notification.Value, &notification.Read,
			&notification.Created)
		if err == nil {
			res = append(res, notification)
		} else {
			fmt.Println(err.Error())
		}
	}
	return res
}

// GetUnreadNotificationCount gets the number of notifications a user
// has yet to acknowledge
func (u *User) GetUnreadNotificationCount(db *sql.DB) int {
	rows, err := db.Query("SELECT COUNT(*) FROM notifications WHERE "+
		"user_id = $1 AND is_read = false", u.ID)
	if err != nil {
		return 0
	}
	defer rows.Close()
	if rows.Next() {
		var unreadCount int
		err = rows.Scan(&unreadCount)
		if err != nil {
			return 0
		}
		return unreadCount
	}
	return 0
}

// MarkNotificationRead attempts to mark a single notification read
// for a user
func (u *User) MarkNotificationRead(db *sql.DB, id int) error {
	_, err := db.Exec("UPDATE notifications SET is_read = true WHERE "+
		"user_id = $1 AND id = $2", u.ID, id)
	if err != nil {
		fmt.Println(err.Error())
	}
	return err
}

// MarkNotificationsRead attempts to mark a set of notifications as read
// for a user
func (u *User) MarkNotificationsRead(db *sql.DB, id int) error {
	_, err := db.Exec("UPDATE notifications SET is_read = true WHERE "+
		"user_id = $1 AND id <= $2", u.ID, id)
	if err != nil {
		fmt.Println(err.Error())
	}
	return err
}
