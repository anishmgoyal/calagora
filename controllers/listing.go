package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/anishmgoyal/calagora/cache"
	"github.com/anishmgoyal/calagora/constants"
	"github.com/anishmgoyal/calagora/models"
)

type listingCreateData struct {
	HasError bool
	Error    models.ListingError
	Listing  models.Listing
}

type listingEditData struct {
	HasError bool
	Error    models.ListingError
	Listing  models.Listing
	Images   []models.Image
}

type listingViewData struct {
	Listing  models.Listing
	Offer    *models.Offer
	Images   []models.Image
	IsSeller bool
}

type listingSectionData struct {
	ActiveLink string
	Type       string
	TypeStr    string
}

// WebAPIListings handles the route '/webapi/listings/'
func WebAPIListings(w http.ResponseWriter, r *http.Request) {
	opts := models.ListingQueryOpts{}

	viewData := BaseViewData(w, r)
	if viewData.Session != nil {
		opts.RestrictByPlace = true
		opts.PlaceID = viewData.Session.User.PlaceID
	}

	pageSizeStr := r.FormValue("pageSize")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = 50
	} else if pageSize > 50 {
		pageSize = 50
	}

	pageNumStr := r.FormValue("pageNum")
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil {
		pageNum = 0
	}

	typeStr := r.FormValue("type")
	if len(typeStr) > 0 {
		opts.Type = typeStr
		opts.RestrictByType = true
	}

	opts.Status = models.ListingListed
	opts.RestrictByStatus = true

	opts.HideDraft = true

	opts.PageSize = pageSize
	opts.PageNum = pageNum
	opts.UsePaging = true

	listings := models.GetListingList(Base.Db, opts)
	cache.MapPlaceToListings(listings)
	RenderJSON(w, listings)
}

// WebAPIListingsUser handles the route '/webapi/listings/user/'
func WebAPIListingsUser(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	args := URIArgs(r)
	if len(args) != 1 {
		RenderJSON(w, nil)
		return
	}

	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RenderJSON(w, nil)
		return
	}

	opts := models.ListingQueryOpts{}
	opts.HideDraft = viewData.Session == nil || viewData.Session.User.ID != id

	opts.UserID = id
	opts.RestrictByUser = true

	status := r.FormValue("status")
	if len(status) > 0 {
		opts.Status = status
		opts.RestrictByStatus = true
	}

	pageSizeStr := r.FormValue("pageSize")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = 50
	} else if pageSize > 50 {
		pageSize = 50
	}

	pageNumStr := r.FormValue("pageNum")
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil {
		pageNum = 0
	}

	opts.PageSize = pageSize
	opts.PageNum = pageNum
	opts.UsePaging = true

	listings := models.GetListingList(Base.Db, opts)
	cache.MapPlaceToListings(listings)
	RenderJSON(w, listings)
}

type webAPIListingDeleteResponse struct {
	Successful bool   `json:"successful"`
	Error      string `json:"error,omitempty"`
}

// WebAPIListingDelete handles the route '/webapi/listing/delete/'
func WebAPIListingDelete(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	response := webAPIListingDeleteResponse{
		Successful: false,
	}
	if viewData.Session == nil {
		response.Error = constants.ErrorAuth
		RenderJSON(w, response)
		return
	}

	if !viewData.ValidCsrf(r) {
		response.Error = constants.ErrorCSRF
		RenderJSON(w, response)
		return
	}

	args := URIArgs(r)
	if len(args) != 1 {
		response.Error = constants.ErrorArguments
		RenderJSON(w, response)
		return
	}

	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error = constants.Error404
		RenderJSON(w, response)
		return
	}

	listing, err := models.GetListingByID(Base.Db, id)
	if err != nil || listing == nil {
		response.Error = constants.Error404
		RenderJSON(w, response)
		return
	}
	cache.MapPlaceToListing(listing)

	if listing.User.ID != viewData.Session.User.ID {
		response.Error = constants.Error403
		RenderJSON(w, response)
		return
	}

	if ok, _ := listing.Delete(Base.Db); !ok {
		response.Error = constants.Error500
		RenderJSON(w, response)
		return
	}

	response.Successful = true
	RenderJSON(w, response)
}

// ListingCreate handles the route '/listings/create'
func ListingCreate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getCreate(w, r)
	case http.MethodPost:
		postCreate(w, r)
	default:
		BaseViewData(w, r).NotFound(w)
	}
}

func getCreate(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		viewData.ForceLogin(w, r)
	} else {
		viewData.Data = &listingCreateData{HasError: false}
		RenderView(w, "listing#create", viewData)
	}
}

func postCreate(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if !viewData.ValidCsrf(r) {
		http.Redirect(w, r, "/", http.StatusFound)
	} else {

		listing := models.Listing{
			Name:        r.FormValue("name"),
			PriceClient: r.FormValue("price"),
			Status:      models.ListingListed,
			Type:        r.FormValue("type"),
			Condition:   r.FormValue("condition"),
			Description: r.FormValue("description"),
			User:        viewData.Session.User,
		}

		if strings.Compare(r.FormValue("submissionType"), "publish") == 0 {
			listing.Published = true
		}

		valid, listingErr := listing.Create(Base.Db)

		if valid {
			if (strings.Compare(r.FormValue("submissionType"), "addim")) == 0 {
				// Here we set the checkbox to true for the user so that they
				// don't accidentally leave new listings in draft state
				listing.Published = true
				viewData.Data = &listingEditData{
					HasError: false,
					Listing:  listing,
					Images:   []models.Image{},
				}
				RenderView(w, "listing#edit", viewData)
			} else {
				http.Redirect(w, r, "/listing/view/"+strconv.Itoa(listing.ID),
					http.StatusFound)
			}
		} else {
			viewData.Data = &listingCreateData{
				HasError: true,
				Error:    *listingErr,
				Listing:  listing,
			}
			RenderView(w, "listing#create", viewData)
		}
	}
}

// ListingEdit handles the route '/listing/edit'
func ListingEdit(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getEdit(w, r)
	case http.MethodPost:
		postEdit(w, r)
	default:
		BaseViewData(w, r).NotFound(w)
	}
}

func getEdit(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		viewData.ForceLogin(w, r)
		return
	}

	args := URIArgs(r)
	if len(args) != 1 {
		viewData.NotFound(w)
		return
	}

	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		viewData.NotFound(w)
		return
	}

	listing, err := models.GetListingByID(Base.Db, id)
	if err != nil || listing == nil {
		viewData.NotFound(w)
		return
	}
	cache.MapPlaceToListing(listing)

	images, err := listing.GetImages(Base.Db)
	if err != nil {
		viewData.InternalError(w)
		return
	}

	if listing.User.ID != viewData.Session.User.ID {
		viewData.Forbidden(w)
		return
	}

	viewData.Data = &listingEditData{
		HasError: false,
		Listing:  *listing,
		Images:   images,
	}
	RenderView(w, "listing#edit", viewData)
}

func postEdit(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		viewData.ForceLogin(w, r)
		return
	}

	if !viewData.ValidCsrf(r) {
		http.Redirect(w, r, "/listing/edit/"+r.FormValue("listing_id"),
			http.StatusFound)
		return
	}

	idStr := r.FormValue("listing_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		viewData.NotFound(w)
		return
	}

	listing, err := models.GetListingByID(Base.Db, id)
	if listing == nil && err != nil {
		viewData.NotFound(w)
		return
	}
	cache.MapPlaceToListing(listing)

	primaryImage := r.FormValue("primaryImage")
	primaryImageID, err := strconv.Atoi(primaryImage)
	if err == nil {
		// If this fails, it will currently silently fail. Need an
		// unintrusive way of notifying the user, since the rest of the
		// edit could succeed
		listing.UpdatePrimaryImage(Base.Db, primaryImageID)
	}

	listing.Name = r.FormValue("name")
	listing.PriceClient = r.FormValue("price")
	listing.Type = r.FormValue("type")
	listing.Condition = r.FormValue("condition")
	listing.Description = r.FormValue("description")
	listing.Published = strings.Compare(r.FormValue("published"), "1") == 0

	valid, listingErr := listing.Save(Base.Db)

	if valid {
		http.Redirect(w, r, "/listing/view/"+strconv.Itoa(listing.ID),
			http.StatusFound)
	} else {
		images, err := listing.GetImages(Base.Db)
		if err != nil {
			viewData.InternalError(w)
			return
		}

		viewData.Data = &listingEditData{
			HasError: true,
			Error:    *listingErr,
			Listing:  *listing,
			Images:   images,
		}
		RenderView(w, "listing#edit", viewData)
	}
}

// ListingView handles '/listing/view/#id'
func ListingView(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)

	args := URIArgs(r)
	if len(args) != 1 {
		viewData.NotFound(w)
		return
	}

	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		viewData.NotFound(w)
		return
	}

	listing, err := models.GetListingByID(Base.Db, id)
	if listing == nil || err != nil {
		viewData.NotFound(w)
		return
	}
	cache.MapPlaceToListing(listing)

	if viewData.Session != nil &&
		listing.User.PlaceID != viewData.Session.User.PlaceID {

		viewData.WrongPlace(w, listing.User.PlaceID, viewData.Session.User.PlaceID)
		return
	}

	images, err := listing.GetImages(Base.Db)
	if err != nil {
		viewData.InternalError(w)
		return
	}

	isSeller := viewData.Session != nil &&
		viewData.Session.User.ID == listing.User.ID

	if !isSeller && !listing.Published {
		viewData.RenderMessage(w, true, "This is a Draft",
			"The seller has not published this listing yet, or has reverted this "+
				"listing to a draft, so you cannot see it. If you have previously "+
				"placed an offer on this listing, the seller can still see your "+
				"offer, and you can still chat about this listing after the seller "+
				"accepts your offer.")
		return
	}

	lvd := &listingViewData{
		Listing:  *listing,
		Images:   images,
		IsSeller: isSeller,
	}

	if viewData.Session != nil && !isSeller && listing != nil {
		offer, err := viewData.Session.User.GetOfferOnListing(Base.Db, listing.ID)
		if err == nil && offer != nil {
			lvd.Offer = offer
		}
	}

	viewData.Data = lvd
	RenderView(w, "listing#view", viewData)
}

// ListingDelete handles the route '/listing/delete/'
func ListingDelete(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		postListingDelete(w, r)
	default:
		BaseViewData(w, r).NotFound(w)
	}
}

func postListingDelete(w http.ResponseWriter, r *http.Request) {

	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		viewData.ForceLogin(w, r)
		return
	}

	args := URIArgs(r)
	if len(args) != 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if !viewData.ValidCsrf(r) {
		http.Redirect(w, r, "/listing/view/"+args[0], http.StatusFound)
		return
	}

	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		viewData.NotFound(w)
		return
	}

	listing, err := models.GetListingByID(Base.Db, id)
	if err != nil || listing == nil ||
		listing.User.ID != viewData.Session.User.ID {

		viewData.NotFound(w)
		return
	}
	cache.MapPlaceToListing(listing)

	ok, err := listing.Delete(Base.Db)
	if !ok {
		if err != nil {
			fmt.Println(err.Error())
			viewData.InternalError(w)
			return
		}
		fmt.Println("E")
		fmt.Println(!ok)
		viewData.InternalError(w)
		return
	}

	http.Redirect(w, r, "/selling", http.StatusFound)
}

// ListingSection handles the route '/listing/section/'
func ListingSection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getListingSection(w, r)
	default:
		BaseViewData(w, r).NotFound(w)
	}
}

func getListingSection(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	args := URIArgs(r)
	if len(args) == 0 {
		viewData.Data = listingSectionData{
			ActiveLink: "lnk_all",
			Type:       "",
			TypeStr:    "All Listings",
		}
	} else {
		typeNm := args[0]
		typeStr, ok := models.ListingTypes[typeNm]
		if !ok {
			viewData.NotFound(w)
			return
		}
		viewData.Data = listingSectionData{
			ActiveLink: "lnk_" + typeNm,
			Type:       typeNm,
			TypeStr:    typeStr,
		}
	}
	RenderView(w, "listing#section", viewData)
}

type sellerListViewData struct {
	Listings []models.Listing
}

// SellerList handles the route '/selling'
func SellerList(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		viewData.ForceLogin(w, r)
		return
	}

	opts := models.ListingQueryOpts{
		UserID:           viewData.Session.User.ID,
		RestrictByUser:   true,
		Status:           models.ListingListed,
		RestrictByStatus: true,
		UsePaging:        false,
	}

	listings := models.GetListingList(Base.Db, opts)
	cache.MapPlaceToListings(listings)
	viewData.Data = &sellerListViewData{
		Listings: listings,
	}

	RenderView(w, "listing#selling", viewData)
}
