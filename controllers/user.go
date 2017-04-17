package controllers

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/anishmgoyal/calagora/email"
	"github.com/anishmgoyal/calagora/models"
	"github.com/anishmgoyal/calagora/utils"
)

type loginData struct {
	HasError bool
	Error    string
	Username string
	Redirect string
}

type registerData struct {
	HasError bool
	Error    models.UserError
	User     models.User
}

type profileData struct {
	HasError bool
	Error    models.UserError
	User     models.User
}

// ForceLogin redirects a user to the login page with a redirect back
// to whatever page they were looking at
func (vd *ViewData) ForceLogin(w http.ResponseWriter, r *http.Request) {
	location := "/user/login/"
	args := "?return=" + url.QueryEscape(r.RequestURI)
	http.Redirect(w, r, location+args, http.StatusFound)
}

// UserLogin handles '/user/login'
func UserLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getLogin(w, r)
	case http.MethodPost:
		postLogin(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func getLogin(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		redirect := r.FormValue("return")
		if len(redirect) == 0 {
			redirect = "/"
		}
		viewData.Data = &loginData{HasError: false, Redirect: redirect}
		RenderView(w, "user#login", viewData)
	}
}

func postLogin(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		username := strings.TrimSpace(strings.ToLower(r.FormValue("username")))
		password := r.FormValue("password")
		user := models.GetUserByUsername(Base.Db, username)
		valid, err := user.Authenticate(password)
		if err != nil || !valid {

			viewData.Data = &loginData{HasError: true,
				Error:    err.Error(),
				Username: username,
				Redirect: r.FormValue("redirect"),
			}

			RenderView(w, "user#login", viewData)
		} else {
			session := models.Session{
				User:         *user,
				BrowserAgent: r.Header.Get("User-Agent"),
			}
			created, _ := session.Create(Base.Db)
			if !created {
				viewData.Data = &loginData{
					HasError: true,
					Error:    "Failed to create session. Please try again later.",
					Username: username,
					Redirect: r.FormValue("redirect"),
				}
				RenderView(w, "user#login", viewData)
			} else {
				// Create cookie
				utils.SetCookie(w, "session_id", session.SessionID, 14)
				utils.SetCookie(w, "session_secret", session.SessionSecret, 14)
				redirect := r.FormValue("redirect")

				if len(redirect) == 0 ||
					redirect[0] != '/' || strings.HasPrefix(redirect, "/user") {

					// Not comprehensive, but quick safe-guard against someone having the
					// login page redirect elsewhere by providing a false link
					// Also, many user pages will cause their own redirect
					// for a user that is logged in
					redirect = "/"
				}
				http.Redirect(w, r, redirect, http.StatusFound)
			}
		}
	}
}

// UserProfile handles '/user/profile/'
func UserProfile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getUserProfile(w, r)
	case http.MethodPost:
		postUserProfile(w, r)
	default:
		BaseViewData(w, r).NotFound(w)
	}
}

func getUserProfile(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		viewData.ForceLogin(w, r)
		return
	}

	viewData.Data = &profileData{
		HasError: false,
		User:     viewData.Session.User,
	}

	RenderView(w, "user#profile", viewData)
}

func postUserProfile(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		viewData.ForceLogin(w, r)
		return
	}

	viewData.Session.User.DisplayName = r.FormValue("display_name")
	viewData.Session.User.Password = r.FormValue("password")
	viewData.Session.User.PasswordConfirmation = r.FormValue("password_confirmation")

	valid, userErr := viewData.Session.User.Save(Base.Db)
	if !valid {
		viewData.Data = &profileData{
			HasError: true,
			Error:    *userErr,
			User:     viewData.Session.User,
		}
		RenderView(w, "user#profile", viewData)
		return
	}
	viewData.RenderMessage(w, false, "Profile Saved", "You have successfully "+
		"modified your profile.")
}

// UserRegister handles '/user/register/'
func UserRegister(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getRegister(w, r)
	case http.MethodPost:
		postRegister(w, r)
	default:
		BaseViewData(w, r).NotFound(w)
	}
}

func getRegister(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		viewData.Data = &registerData{HasError: false}
		RenderView(w, "user#register", viewData)
	}
}

func postRegister(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		user := models.User{
			Username:             r.FormValue("username"),
			DisplayName:          r.FormValue("display_name"),
			EmailAddress:         r.FormValue("email_address"),
			Password:             r.FormValue("password"),
			PasswordConfirmation: r.FormValue("password_confirmation"),
		}
		created, err := user.Create(Base.Db)
		if !created {
			viewData.Data = &registerData{
				HasError: true,
				Error:    *err,
				User:     user,
			}
			RenderView(w, "user#register", viewData)
		} else {
			email.SendActivationEmail(user)
			viewData.Data = &homeData{
				Error:      false,
				FlashTitle: "Account created successfully",
				Flash: "Welcome to Calagora! Your account has been " +
					"created. Please check your email for instructions about " +
					"how to activate your account.",
			}
			RenderView(w, "home#index", viewData)
		}
	}
}

// UserActivate handles '/user/activate/'
func UserActivate(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	args := URIArgs(r)
	if len(args) != 2 {
		viewData.NotFound(w)
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		viewData.NotFound(w)
		return
	}

	activation := args[1]

	user := models.GetUserByID(Base.Db, id)
	if user == nil {
		viewData.NotFound(w)
		return
	}

	if strings.Compare(activation, user.Activation) == 0 {
		if user.Activate(Base.Db) == nil {
			viewData.RenderMessage(w, false, "Account Activated",
				"Your account has successfully been activated! You can now log "+
					"in, post listings, make offers, and chat with other users.")
		} else {
			viewData.InternalError(w)
		}
	} else {
		viewData.NotFound(w)
	}
}

// UserLogout handles '/user/logout'
func UserLogout(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getLogout(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

// Slightly different pattern from the other methods
// Since no view is rendered, no BaseViewData is necessary
func getLogout(w http.ResponseWriter, r *http.Request) {
	session := models.GetSessionFromRequest(Base.Db, w, r)
	utils.DeleteCookie(w, "session_id")
	utils.DeleteCookie(w, "session_secret")

	if session != nil {
		session.Delete(Base.Db)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
