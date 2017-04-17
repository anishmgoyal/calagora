package wsock

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/anishmgoyal/calagora/models"
)

type jsonMessage struct {
	NotifType    string      `json:"notif_type"`
	Notification interface{} `json:"notification"`
}

// UserMessage wraps initialization of a message struct for a user
func UserMessage(u *models.User, message string) *Message {
	return &Message{
		Target:  UserChannelPrefix + strconv.Itoa(u.ID),
		Message: message,
	}
}

// UserJSONNotification attempts to render a message as JSON and
// return it. Nil is returned if JSON cannot be created
func UserJSONNotification(u *models.User, notifType string,
	notification interface{}, createRecord bool) *Message {

	jsm := jsonMessage{
		NotifType:    notifType,
		Notification: notification,
	}
	b, err := json.Marshal(jsm)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	notificationValue := string(b)

	notificationRecord := models.Notification{
		User:  *u,
		Value: notificationValue,
	}

	if createRecord {
		notificationRecord.Create(Base.Db)
	}

	b, err = json.Marshal(notificationRecord)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	notificationValue = string(b)

	return &Message{
		Target:  UserChannelPrefix + strconv.Itoa(u.ID),
		Message: notificationValue,
	}
}

// UserTarget generates the target for a user based on a user model
func UserTarget(u *models.User) string {
	return UserChannelPrefix + strconv.Itoa(u.ID)
}

// SessionMessage wraps initialization of a message struct for a session
func SessionMessage(s *models.Session, message string) *Message {
	return &Message{
		Target:  SessionChannelPrefix + s.SessionID,
		Message: message,
	}
}

// SessionTarget generates the target for a session based on a session model
func SessionTarget(s *models.Session) string {
	return SessionChannelPrefix + s.SessionID
}

// BroadcastMessage wraps initalization of a message struct for all users
func BroadcastMessage(message string) *Message {
	return &Message{
		Target:  BroadcastChannel,
		Message: message,
	}
}
