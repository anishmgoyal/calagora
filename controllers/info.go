package controllers

import "net/http"

// InfoAbout handles the route '/info/about/'
func InfoAbout(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	RenderView(w, "info#about", viewData)
}

// InfoContact handles the route '/info/contact/'
func InfoContact(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	RenderView(w, "info#contact", viewData)
}

// InfoHelp handles the route '/info/help/'
func InfoHelp(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	RenderView(w, "info#help", viewData)
}

// InfoTos handles the route '/info/tos/'
func InfoTos(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	RenderView(w, "info#tos", viewData)
}
