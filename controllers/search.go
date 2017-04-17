package controllers

import (
	"net/http"
	"strconv"

	"github.com/anishmgoyal/calagora/models"
	"github.com/anishmgoyal/calagora/utils"
)

type searchViewData struct {
	Listings    []models.Listing `json:"listings"`
	Query       string           `json:"query"`
	Page        int              `json:"page"`
	StartOffset int              `json:"start_offset"`
	EndOffset   int              `json:"end_offset"`
	MaxTotal    int              `json:"max_total"`
	OutOf       int              `json:"out_of"`
}

// Search handles the route '/search/'
func Search(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)

	termMap := utils.GetSearchTermsForString(r.FormValue("q"), true)
	terms := make([]string, len(termMap))
	i := 0
	for term := range termMap {
		terms[i] = term
		i++
	}

	pageNumStr := "1"
	if len(r.FormValue("page")) > 0 {
		pageNumStr = r.FormValue("page")
	}

	page, err := strconv.Atoi(pageNumStr)
	if err != nil {
		viewData.NotFound(w)
		return
	}
	// Correct for the human readable format for page numbers used
	// by the client here
	page = page - 1

	placeID := -1
	if viewData.Session != nil {
		placeID = viewData.Session.User.PlaceID
	}

	listings := []models.Listing{}
	if len(terms) > 0 {
		listings, err = models.DoSearchForTerms(Base.Db, terms, page, placeID)
		if err != nil {
			viewData.InternalError(w)
			return
		}
	}

	numPages := models.GetPageCountForTerms(Base.Db, terms, placeID)

	viewData.Data = searchViewData{
		Listings:    listings,
		Query:       r.FormValue("q"),
		Page:        page + 1,
		StartOffset: page*50 + 1,
		EndOffset:   page*50 + len(listings),
		MaxTotal:    numPages * 50,
		OutOf:       numPages,
	}
	RenderView(w, "search#search", viewData)
}
