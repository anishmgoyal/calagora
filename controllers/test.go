package controllers

import (
	"bytes"
	"net/http"
)

type emailData struct {
	EmailTitle string
	Paragraphs []string
}

// TestEmail tests the email layouts
func TestEmail(w http.ResponseWriter, r *http.Request) {
	testEmailData := emailData{
		EmailTitle: "Test Email",
		Paragraphs: []string{
			"Hello, world!",
			"This is an email!",
			"It seems to be working OK!",
		},
	}
	var buff bytes.Buffer
	tmpl, ok := Base.Templates["email#plain"]
	if !ok {
		http.Error(w, "Bad Template", http.StatusNotFound)
		return
	}

	err := tmpl.ExecuteTemplate(&buff, "base", testEmailData)
	if err != nil {
		http.Error(w, "Execution Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buff.Bytes())
}
