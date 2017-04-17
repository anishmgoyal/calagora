package controllers

import (
	"net/http"
	"strings"

	"github.com/anishmgoyal/calagora/models"
)

type homeData struct {
	Error      bool
	FlashTitle string
	Flash      interface{}
}

// NotFound is a method called by a different controller
// handler and displays a 404 page.
func (vd ViewData) NotFound(w http.ResponseWriter) {
	vd.Data = &homeData{
		Error:      true,
		FlashTitle: "404 Not Found",
		Flash: "Uh-oh... we couldn't find that page. If you think this is a " +
			"mistake, please contact support at " + Base.SupportEmail,
	}
	RenderView(w, "home#index", vd)
}

// InternalError is a method called by a different controller
// handler and displays a 500 page.
func (vd ViewData) InternalError(w http.ResponseWriter) {
	vd.Data = &homeData{
		Error:      true,
		FlashTitle: "500 Internal Error",
		Flash: "This is embarassing... We made a mistake. We have no clue " +
			"what it is, but we have a particular set of skills that make us " +
			"particularly dangerous to bugs like this and blah blah blah. Try " +
			"refreshing the page or trying again later. If that doesn't work, " +
			"please contact support at " + Base.SupportEmail + ", and we'll " +
			"try to figure out what the issue is as quickly as we can.",
	}
	RenderView(w, "home#index", vd)
}

// Forbidden is a method called by a different controller
// handler and displays a 403 page.
func (vd ViewData) Forbidden(w http.ResponseWriter) {
	vd.Data = &homeData{
		Error:      true,
		FlashTitle: "403 Forbidden",
		Flash: "You're not allowed to see that! If you think this is a " +
			"mistake, please contact support at " + Base.SupportEmail,
	}
	RenderView(w, "home#index", vd)
}

// WrongPlace is a method called by a different controller
// handler, and lets a user know that the listing they're trying to
// look at is for a different school than they're set up for
func (vd ViewData) WrongPlace(w http.ResponseWriter, aPlaceID,
	bPlaceID int) {

	aPlace, err := models.GetPlaceByID(Base.Db, aPlaceID)
	if err != nil {
		vd.InternalError(w)
		return
	}
	bPlace, err := models.GetPlaceByID(Base.Db, bPlaceID)
	if err != nil {
		vd.InternalError(w)
		return
	}

	message := "The listing you are trying to view was not posted by a student " +
		"in your school. It was posted by a student from " + aPlace.Name +
		", but " + "you are from " + bPlace.Name + ". You can only make offers " +
		"for listings posted in your school."
	vd.Data = &homeData{
		Error:      true,
		FlashTitle: "That Listing is for a Different University",
		Flash:      message,
	}
	RenderView(w, "home#index", vd)
}

// Home handles any requests on the '/' route
func Home(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getHome(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func getHome(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)

	if len(r.RequestURI) > 0 && strings.Compare(r.RequestURI, "/") != 0 {
		viewData.NotFound(w)
		return
	}

	if viewData.Session != nil {
		viewData.Data = &homeData{
			Error:      false,
			FlashTitle: "Welcome",
			Flash: "You are currently logged in as " +
				viewData.Session.User.DisplayName,
		}
	}
	RenderView(w, "home#index", viewData)
}

// HomeUnsupported is displayed if a user's browser is not supported
func HomeUnsupported(w http.ResponseWriter, r *http.Request) {
	RenderView(w, "home#unsupported", BaseViewData(w, r))
}
