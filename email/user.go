package email

import (
	"strconv"

	"github.com/anishmgoyal/calagora/models"
	"github.com/anishmgoyal/calagora/utils"
)

// SendActivationEmail sends an activation email for a new user
func SendActivationEmail(user models.User) {
	title := "Welcome to Calagora!"
	paragraphs := []interface{}{
		"Welcome to Calagora, " + user.DisplayName + "!",
		"You've successfully created an account with the username " +
			user.Username + ".",
		"In order to complete your registration, you just need to " +
			"activate your account.",
		"To do this, copy and paste this link into your address bar:",
		makeURLLink("https://www.calagora.com/user/activate/" +
			strconv.Itoa(user.ID) + "/" + user.Activation),
		"After you activate your account, you can make offers on " +
			"listings, create your own listings, and chat with sellers or " +
			"buyers through the site!",
	}
	email := &utils.Email{
		To:            []string{user.EmailAddress},
		From:          Base.AutomatedEmail,
		Subject:       title,
		FormattedText: GenerateHTML(title, paragraphs),
		PlainText:     GeneratePlain(title, paragraphs),
	}
	Base.EmailChannel <- email
}
