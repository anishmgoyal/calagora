package controllers

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/anishmgoyal/calagora/constants"
	"github.com/anishmgoyal/calagora/models"
	"github.com/anishmgoyal/calagora/utils"
	"github.com/anishmgoyal/calagora/wsock"
)

// ViewData defines the base properties used by a view
type ViewData struct {
	Session    *models.Session
	Data       interface{}
	Constants  map[string]interface{}
	CurrentURI string
	header     http.Header
}

// MessageViewData defines properties used to display a message
// to a user with an optional redirect
type MessageViewData struct {
	Header      string
	Body        string
	DoRedirect  bool
	RedirectURL string
}

// Base contains all variables shared by controllers
var Base struct {
	// Templates contains all views available to controllers
	Templates map[string]*template.Template
	// DB is the database connection to be used by controllers and passed to models
	Db *sql.DB
	// SupportEmail is the email address to provide users with on errors
	SupportEmail string
	// ImageChannel is a channel in which image process requests can be enqueued
	ImageChannel chan *utils.ImageProcessRequest
	// WebsockChannel is a channel in which notifications can be pushed to users
	WebsockChannel chan *wsock.Message
}

// BaseInitialization initializes all controllers
func BaseInitialization(templates map[string]*template.Template, db *sql.DB) {
	Base.Templates = templates
	Base.Db = db
	Base.SupportEmail = constants.SupportEmail
	Base.ImageChannel = utils.StartImageService()
	Base.WebsockChannel = wsock.StartWebsocketService(db)
}

// BaseViewData gets any fields necessary for rendering a basic view
func BaseViewData(w http.ResponseWriter, r *http.Request) ViewData {
	session := models.GetSessionFromRequest(Base.Db, w, r)
	if session != nil {
		// Keeps active sessions alive over time
		go session.Update(Base.Db)
	}
	return ViewData{
		Session: session,
		Data:    nil,
		Constants: map[string]interface{}{
			"listing.conditionnames": models.ListingConditionNames,
			"listing.conditions":     models.ListingConditions,
			"listing.typenames":      models.ListingTypeNames,
			"listing.types":          models.ListingTypes,
		},
		CurrentURI: r.RequestURI,
		header:     r.Header,
	}
}

// ValidCsrf returns true if the user has a valid session and CSRFToken submitted
// by form, false if one of the conditions is not true
func (vd ViewData) ValidCsrf(r *http.Request) bool {
	if vd.Session == nil {
		return false
	}
	return strings.Compare(vd.Session.CsrfToken, r.FormValue("csrfToken")) == 0
}

// RenderPlainView attempts to render a view with only the base view data
func RenderPlainView(w http.ResponseWriter, r *http.Request,
	templateName string) {

	RenderView(w, templateName, BaseViewData(w, r))
}

// RenderView attempts to render a view. Gives a 404 error on failure
func RenderView(w http.ResponseWriter, templateName string, data ViewData) {
	var buff bytes.Buffer

	tmpl, ok := Base.Templates[templateName]
	if !ok {
		errStr := fmt.Sprintf("Unknown Template: %s", templateName)
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	}

	err := tmpl.ExecuteTemplate(&buff, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if strings.Contains(data.header.Get("Accept-Encoding"), "gzip") {
		gz := gzip.NewWriter(w)
		defer gz.Close()
		w.Header().Set("Content-Encoding", "gzip")
		gz.Write(buff.Bytes())
	} else {
		w.Write(buff.Bytes())
	}
}

// RenderJSON attempts to render an object as JSON. Gives a 404 error on failure
func RenderJSON(w http.ResponseWriter, value interface{}) {
	b, err := json.Marshal(value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(b)
}

// RenderMessage renders the message template.
func (vd ViewData) RenderMessage(w http.ResponseWriter,
	isError bool, header, body string) {

	vd.Data = homeData{
		Error:      isError,
		FlashTitle: header,
		Flash:      body,
	}
	RenderView(w, "home#index", vd)
}

// RenderTextJSON attempts to render an object as JSON with mime type text/html.
// Gives a 404 error on failure
func RenderTextJSON(w http.ResponseWriter, value interface{}) {
	b, err := json.Marshal(value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(b)
}

// RenderTextErrorJSON attempts to render an object as JSON with mime type
// text/html and a 404 error.
func RenderTextErrorJSON(w http.ResponseWriter, value interface{}) {
	b, err := json.Marshal(value)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err != nil {
		http.Error(w, "{}", http.StatusNotFound)
		return
	}

	http.Error(w, string(b), http.StatusBadRequest)
}

// URIArgs gets arguments from the current page URI
func URIArgs(r *http.Request) []string {
	args := make([]string, 0, 10)
	numFound := 0

	uri := r.URL.RequestURI()

	var currArg bytes.Buffer
	for i := 0; i < len(uri); i++ {
		if uri[i] == '?' {
			if currArg.Len() > 0 {
				args = append(args, currArg.String())
				currArg.Reset()
				numFound++
			}
			break
		} else if uri[i] == '/' {
			if currArg.Len() > 0 {
				args = append(args, currArg.String())
				numFound++
			}
			currArg.Reset()
		} else {
			currArg.Write([]byte{uri[i]})
		}
	}
	if currArg.Len() > 0 {
		args = append(args, currArg.String())
		numFound++
	}
	return args[:numFound]
}
