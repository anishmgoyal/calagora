package controllers

import (
	"net/http"
	"strconv"

	"github.com/anishmgoyal/calagora/models"
	"github.com/anishmgoyal/calagora/wsock"
)

const (
	notifNewMessage = "NEW_MESSAGE"
)

type webAPIMessageSendResponse struct {
	HasError     bool                `json:"has_error"`
	Error        string              `json:"error,omitempty"`
	MessageError models.MessageError `json:"message_error,omitempty"`
	Successful   bool                `json:"successful"`
}

type webAPIMessageNotification struct {
	OfferID          int            `json:"offer_id"`
	Message          models.Message `json:"message"`
	NotificationType string         `json:"notification_type"`
}

// WebAPIMessageSend handles the route '/webapi/message/'
func WebAPIMessageSend(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	response := webAPIMessageSendResponse{Successful: false}

	if len(r.FormValue("message")) == 0 {
		response.HasError = true
		response.Error = "No Message"
		RenderJSON(w, response)
		return
	}

	if viewData.Session == nil {
		response.HasError = true
		response.Error = "Not Logged In"
		RenderJSON(w, response)
		return
	}

	if !viewData.ValidCsrf(r) {
		response.HasError = true
		response.Error = "CSRF Error"
		RenderJSON(w, response)
		return
	}

	args := URIArgs(r)
	if len(args) != 1 {
		response.HasError = true
		response.Error = "Invalid Arguments"
		RenderJSON(w, response)
		return
	}

	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.HasError = true
		response.Error = "Invalid conversation ID"
		RenderJSON(w, response)
		return
	}

	offer, err := models.GetOfferByID(Base.Db, id)
	if err != nil || offer == nil {
		response.HasError = true
		response.Error = "Not Found"
		RenderJSON(w, response)
		return
	}

	if offer.Status != models.OfferAccepted {
		response.HasError = true
		response.Error = "Not a Conversation"
		RenderJSON(w, response)
		return
	}

	message := models.Message{
		Message: r.FormValue("message"),
		Offer:   *offer,
		Sender:  viewData.Session.User,
	}
	var otherUser *models.User
	if offer.Buyer.ID == viewData.Session.User.ID {
		message.Recepient.ID = offer.Seller.ID
		otherUser = &offer.Seller
	} else {
		message.Recepient.ID = offer.Buyer.ID
		otherUser = &offer.Buyer
	}

	ok, messageErr := message.Create(Base.Db)
	if err != nil {
		response.HasError = true
		response.MessageError = *messageErr
		RenderJSON(w, response)
		return
	}

	Base.WebsockChannel <- wsock.UserJSONNotification(otherUser,
		notifNewMessage,
		message, true)
	Base.WebsockChannel <- wsock.UserJSONNotification(&viewData.Session.User,
		notifNewMessage,
		message, false)

	if ok {
		response.Successful = true
	}
	RenderJSON(w, response)
}

type webAPIMessagesResponse struct {
	Messages []models.Message `json:"messages"`
	HasError bool             `json:"has_error"`
	Error    string           `json:"error,omitempty"`
}

// WebAPIMessages handles '/webapi/messages/'
func WebAPIMessages(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	response := webAPIMessagesResponse{}
	if viewData.Session == nil {
		response.HasError = true
		response.Error = "Not Authenticated"
		RenderJSON(w, response)
		return
	}

	args := URIArgs(r)
	if len(args) != 2 {
		response.HasError = true
		response.Error = "Invalid arg count; should be /<offer-id>/<pagenum>"
		RenderJSON(w, response)
		return
	}

	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.HasError = true
		response.Error = "Invalid offer id."
		RenderJSON(w, response)
		return
	}

	pageStr := args[1]
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		response.HasError = true
		response.Error = "Invalid page number."
		RenderJSON(w, response)
		return
	}

	offer, err := models.GetOfferByID(Base.Db, id)
	if err != nil || offer == nil {
		response.HasError = true
		response.Error = "Failed to get conversation/offer."
		RenderJSON(w, response)
		return
	}

	if offer.Buyer.ID != viewData.Session.User.ID &&
		offer.Seller.ID != viewData.Session.User.ID {

		response.HasError = true
		response.Error = "Unauthorized"
		RenderJSON(w, response)
		return
	}

	messages, err := offer.GetMessages(Base.Db, 100, page, viewData.Session.User.ID)
	if err != nil {
		response.HasError = true
		response.Error = "Failed to get messages."
		RenderJSON(w, response)
		return
	}

	response.HasError = false
	response.Messages = messages
	RenderJSON(w, response)
}

// MessageRead handles the route '/message/read/'
func MessageRead(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	args := URIArgs(r)
	if len(args) != 1 {
		http.Error(w, "Invalid Args", http.StatusBadRequest)
		return
	}
	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	message := models.Message{ID: id}
	message.MarkRead(Base.Db, viewData.Session.User.ID)
	http.Error(w, "Command Confirmed", http.StatusOK)
}

// MessageClient handles the route '/message/client'
func MessageClient(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getMessageClient(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func getMessageClient(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		viewData.ForceLogin(w, r)
		return
	}

	RenderView(w, "message#client", viewData)
}
