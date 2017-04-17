package email

import (
	"strconv"

	"github.com/anishmgoyal/calagora/models"
	"github.com/anishmgoyal/calagora/utils"
)

// PasswordRecoveryEmail is sent if a user forgets his/her password
func PasswordRecoveryEmail(prr models.PasswordRecoveryRequest) {

	title := "Calagora - Account Recovery"
	paragraphs := []interface{}{
		"You are receiving this email because you requested recovery information " +
			"for your account with the username: " + prr.User.Username,
		"If you did not make this request, please ignore this email.",
		"If you would like to recover your password, please copy and paste the " +
			"link at the bottom of this email into your address bar in your " +
			"browser, then create a new password in the form " +
			"provided. You will also have to provide the following recovery code " +
			"in that form:",
		"Recovery Code: " + prr.RecoveryCode,
		"You can view the form to reset your password at:",
		makeURLLink("https://www.calagora.com/recover/user/" +
			strconv.Itoa(prr.User.ID) + "/" + prr.RecoveryString),
	}
	email := &utils.Email{
		To:            []string{prr.User.EmailAddress},
		From:          Base.AutomatedEmail,
		Subject:       title,
		FormattedText: GenerateHTML(title, paragraphs),
		PlainText:     GeneratePlain(title, paragraphs),
	}
	Base.EmailChannel <- email
}
