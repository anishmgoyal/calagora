package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/anishmgoyal/calagora/email"
	"github.com/anishmgoyal/calagora/models"
)

type resetPasswordViewData struct {
	User           models.User      `json:"user"`
	UserError      models.UserError `json:"user_error"`
	CodeError      string           `json:"code_error"`
	RecoveryCode   string           `json:"recovery_code"`
	RecoveryString string           `json:"recovery_string"`
}

// RecoverPassword handles the route '/recover/'
func RecoverPassword(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getRecoverPassword(w, r)
	case http.MethodPost:
		postRecoverPassword(w, r)
	default:
		BaseViewData(w, r).NotFound(w)
	}
}

func getRecoverPassword(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	viewData.Data = &models.PasswordRecoveryError{}
	RenderView(w, "recover#index", viewData)
}

func postRecoverPassword(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	emailAddr := r.FormValue("email_address")
	user, err := models.GetUserByEmailAddress(Base.Db, emailAddr)
	if err != nil {
		viewData.InternalError(w)
		return
	} else if user == nil {
		viewData.Data = &models.PasswordRecoveryError{
			UserNotFound: true,
			EmailAddress: emailAddr,
		}
		RenderView(w, "recover#index", viewData)
		return
	}

	err = user.DeleteExpiredRecoveryRequests(Base.Db)
	if err != nil {
		viewData.InternalError(w)
		return
	}

	if has, err2 := user.HasRecoveryRequest(Base.Db); err2 != nil {
		viewData.InternalError(w)
		return
	} else if has {
		viewData.Data = &models.PasswordRecoveryError{
			HasOtherRequest: true,
			EmailAddress:    emailAddr,
		}
		RenderView(w, "recover#index", viewData)
		return
	}

	prr := models.PasswordRecoveryRequest{User: *user}
	err = prr.Create(Base.Db)
	if err != nil {
		viewData.InternalError(w)
		return
	}

	email.PasswordRecoveryEmail(prr)

	viewData.RenderMessage(w, false, "Password Recovery", "We sent you an email "+
		"with your username, along with instructions explaining how to recover "+
		"your password.")
}

// ResetPassword handles the route '/recover/user/'
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	args := URIArgs(r)
	if len(args) != 2 {
		viewData.NotFound(w)
		return
	}

	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		viewData.NotFound(w)
		return
	}

	user := models.User{ID: id}
	prr, err := user.GetPasswordRecoveryRequest(Base.Db)
	if err != nil {
		viewData.InternalError(w)
		return
	} else if prr == nil {
		viewData.NotFound(w)
		return
	}

	if !prr.IsValid {
		viewData.RenderMessage(w, true, "Invalid Recovery Request", "You "+
			"previously asked us to prevent password recovery requests for your "+
			"account for 7 days. It has not yet been 7 days since this request. If "+
			"you believe this was done in error, please contact support@calagora.com "+
			"using the email address associated with your account for further "+
			"assistance.")
		return
	}

	recoveryStr := args[1]
	if strings.Compare(recoveryStr, prr.RecoveryString) != 0 {
		viewData.NotFound(w)
		return
	}

	switch r.Method {
	case http.MethodGet:
		viewData.Data = &resetPasswordViewData{
			User:           user,
			RecoveryString: recoveryStr,
			UserError:      models.UserError{},
		}
		RenderView(w, "recover#reset", viewData)
	case http.MethodPost:
		postResetPassword(w, r, user, *prr, recoveryStr, viewData)
	default:
		viewData.NotFound(w)
	}
}

func postResetPassword(w http.ResponseWriter, r *http.Request, u models.User,
	prr models.PasswordRecoveryRequest, recoveryString string, viewData ViewData) {

	user := models.GetUserByID(Base.Db, u.ID)
	if user == nil {
		// This shouldn't be the case. If there is a password recovery
		// request for a user, that user ought to exist
		viewData.InternalError(w)
		return
	}

	recoveryCode := strings.ToUpper(r.FormValue("recovery_code"))

	valid := true
	rpvd := resetPasswordViewData{
		User:           u,
		RecoveryString: recoveryString,
		RecoveryCode:   recoveryCode,
	}

	if strings.Compare(recoveryCode, prr.RecoveryCode) != 0 {
		rpvd.CodeError = "Your recovery code is invalid"
		valid = false
	}

	user.Password = r.FormValue("password")
	user.PasswordConfirmation = r.FormValue("password_confirmation")
	userValid, userErr := user.Validate(Base.Db, true, false)
	if !userValid {
		rpvd.UserError = userErr
		valid = false
	}

	if !valid {
		viewData.Data = &rpvd
		RenderView(w, "recover#reset", viewData)
		return
	}

	userValid, _ = user.Save(Base.Db)
	if !userValid {
		viewData.InternalError(w)
		return
	}

	// Ignore error: we just want to *try* to delete this. Chances are, a user
	// won't test this to see if it was actually deleted, and on top of that, it
	// will time out within 24 hours
	prr.Delete(Base.Db)

	viewData.RenderMessage(w, false, "Password Reset", "Your password has "+
		"successfully been reset. You may log in now.")
}
