package email

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/anishmgoyal/calagora/constants"
	"github.com/anishmgoyal/calagora/utils"
)

type emailTemplateData struct {
	EmailTitle string
	Paragraphs []interface{}
}

// Base is the base struct for email-related code,
// containing any necessary data
var Base struct {
	// AutomatedEmail is used for any messages that do not warrant a reply
	AutomatedEmail string
	// SupportEmail is where all user questions are to be routed
	SupportEmail string
	// Templates contains any email templates
	Templates map[string]*template.Template
	// EmailChannel is a channel in which emails can be enqueued for delivery
	EmailChannel chan *utils.Email
	// linkRegex is a regex for finding links in an email
	linkRegex *regexp.Regexp
}

// BaseInitialization sets all base fields for emails
func BaseInitialization(templates map[string]*template.Template) {
	Base.Templates = templates
	Base.AutomatedEmail = constants.AutomatedEmail
	Base.SupportEmail = constants.SupportEmail
	Base.EmailChannel = utils.StartEmailService()
	Base.linkRegex = regexp.MustCompile("(https?://www.calagora.com/" +
		"[^\\s]*)::{{(.+)}}")
}

// GenerateHTML generates the HTML email template for a title and
// set of paragraphs
func GenerateHTML(title string, paragraphs []interface{}) string {
	richParagraphs := make([]interface{}, len(paragraphs))
	for i, paragraph := range paragraphs {
		if hasLink(paragraph.(string)) {
			richParagraphs[i] = template.HTML(replaceLinks(paragraph.(string)))
		} else {
			richParagraphs[i] = paragraph
		}
	}
	return generateEmail(title, richParagraphs, "email#default")
}

// GeneratePlain generates the plain email template for a title
// and set of paragraphs
func GeneratePlain(title string, paragraphs []interface{}) string {
	for i, paragraph := range paragraphs {
		paragraphs[i] = template.HTML(removeLinks(paragraph.(string)))
	}
	return generateEmail(title, paragraphs, "email#plain")
}

func generateEmail(title string, paragraphs []interface{}, template string) string {
	var buff bytes.Buffer
	tmpl, ok := Base.Templates[template]
	if !ok {
		fmt.Println("HTML Email Template not found")
		return ""
	}

	err := tmpl.ExecuteTemplate(&buff, "base", emailTemplateData{
		EmailTitle: title,
		Paragraphs: paragraphs,
	})

	if err != nil {
		fmt.Println("Failed to execute HTML email template")
		fmt.Println(err.Error())
		return ""
	}
	return buff.String()
}

func makeURLLink(url string) string {
	return makeLink(url, url)
}

func makeLink(url string, text string) string {
	return url + "::{{" + text + "}}"
}

func hasLink(paragraph string) bool {
	return Base.linkRegex.MatchString(paragraph)
}

func replaceLinks(paragraph string) string {
	paragraph = strings.Replace(
		strings.Replace(
			strings.Replace(paragraph, "<", "&lt;", -1), ">", "&gt;", -1),
		"&", "&amp;", -1)
	return Base.linkRegex.ReplaceAllString(paragraph, "<a href=\"$1\">$2</a>")
}

func removeLinks(paragraph string) string {
	return Base.linkRegex.ReplaceAllString(paragraph, "$2")
}
