package controllers

import (
	"net/http"
	"strconv"

	"github.com/anishmgoyal/calagora/models"
)

type webAPINotificationsResponse struct {
	Notifications []models.Notification `json:"notifications"`
}

// WebAPINotifications handles the route '/notifications/'
func WebAPINotifications(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		RenderJSON(w, webAPINotificationsResponse{
			Notifications: []models.Notification{},
		})
		return
	}

	args := URIArgs(r)
	page := 0
	if len(args) == 1 {
		pageArg, err := strconv.Atoi(args[0])
		if err != nil {
			page = pageArg
		}
	}

	notifications := viewData.Session.User.GetRecentNotifications(Base.Db, page)
	response := webAPINotificationsResponse{
		Notifications: notifications,
	}
	RenderJSON(w, response)
}

type webAPINotificationCountsResponse struct {
	MessageCount      int `json:"message_count"`
	NotificationCount int `json:"notification_count"`
}

// WebAPINotificationCounts handles the route '/webapi/notification/counts/'
func WebAPINotificationCounts(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		viewData.NotFound(w)
		return
	}
	user := viewData.Session.User
	response := webAPINotificationCountsResponse{}
	response.MessageCount = user.GetUnreadMessageCount(Base.Db)
	response.NotificationCount = user.GetUnreadNotificationCount(Base.Db)
	RenderJSON(w, response)
}
