package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/anishmgoyal/calagora/constants"
	"github.com/anishmgoyal/calagora/email"
	"github.com/anishmgoyal/calagora/models"
	"github.com/anishmgoyal/calagora/utils"
	"github.com/anishmgoyal/calagora/wsock"
)

const (
	offerPageSize = 50
)

type createOfferViewData struct {
	HasError bool
	Error    models.OfferError
	Listing  models.Listing
	Offer    models.Offer
}

type buyerListViewData struct {
	Offers []models.Offer
}

type sellerViewData struct {
	Offer []models.Offer
}

// OfferBuyer handles '/offer/buyer'
func OfferBuyer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getOfferBuyer(w, r)
	case http.MethodPost:
		postOfferBuyer(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func getOfferBuyer(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		viewData.ForceLogin(w, r)
		return
	}

	args := URIArgs(r)
	if len(args) != 1 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	listing, err := models.GetListingByID(Base.Db, id)
	if err != nil || listing == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if listing.User.PlaceID != viewData.Session.User.PlaceID {
		viewData.WrongPlace(w, listing.User.PlaceID, viewData.Session.User.PlaceID)
		return
	}

	// Trying to make an offer on own listing
	if listing.User.ID == viewData.Session.User.ID {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	covd := createOfferViewData{
		HasError: false,
		Listing:  *listing,
	}

	offer, _ := viewData.Session.User.GetOfferOnListing(Base.Db, listing.ID)
	if offer != nil {
		covd.Offer = *offer
	}

	viewData.Data = covd

	RenderView(w, "offer#buyer", viewData)
}

func postOfferBuyer(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		viewData.ForceLogin(w, r)
		return
	}

	if !viewData.ValidCsrf(r) {
		http.Redirect(w, r, r.RequestURI, http.StatusFound)
		return
	}

	id, err := strconv.Atoi(r.FormValue("listing_id"))
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	listing, err := models.GetListingByID(Base.Db, id)
	if err != nil || listing == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if listing.User.PlaceID != viewData.Session.User.PlaceID {
		viewData.WrongPlace(w, listing.User.PlaceID, viewData.Session.User.PlaceID)
		return
	}

	var offerErr *models.OfferError
	var ok bool
	var offer *models.Offer

	offer, _ = viewData.Session.User.GetOfferOnListing(Base.Db, listing.ID)

	price, err := utils.PriceClientToServer(r.FormValue("price"))
	if err != nil {
		ok = false
		offerErr = &models.OfferError{Price: "This must be a valid price"}
		if offer == nil {
			offer = &models.Offer{
				BuyerComment: r.FormValue("buyer_comment"),
				PriceClient:  r.FormValue("price"),
			}
		}
		offer.BuyerComment = r.FormValue("buyer_comment")
		offer.PriceClient = r.FormValue("price")
	} else {
		if offer == nil {
			offer = &models.Offer{
				Price:        price,
				PriceClient:  r.FormValue("price"),
				BuyerComment: r.FormValue("buyer_comment"),
				Status:       models.OfferOffered,
				Listing:      *listing,
				Buyer:        viewData.Session.User,
				Seller:       listing.User,
			}

			ok, offerErr = offer.Create(Base.Db)
			if ok {
				email.NewOfferEmail(*offer)
				offer.Listing = *listing
				offer.Buyer = viewData.Session.User
				Base.WebsockChannel <- wsock.UserJSONNotification(&listing.User,
					"NOTIF_NEW_OFFER", offer, true)
			}
		} else {
			offer.Price = price
			offer.PriceClient = r.FormValue("price")
			offer.BuyerComment = r.FormValue("buyer_comment")
			ok, offerErr = offer.Save(Base.Db)
			if ok {
				offer.Listing = *listing
				offer.Buyer = viewData.Session.User
				Base.WebsockChannel <- wsock.UserJSONNotification(&listing.User,
					"NOTIF_UPDATE_OFFER", offer, true)
			}
		}
	}

	if !ok {
		viewData.Data = createOfferViewData{
			HasError: true,
			Error:    *offerErr,
			Listing:  *listing,
			Offer:    *offer,
		}
		RenderView(w, "offer#buyer", viewData)
	} else {
		http.Redirect(w, r, "/buying", http.StatusFound)
	}
}

// OfferSeller handles the route '/offer/seller'
func OfferSeller(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getOfferSeller(w, r)
	case http.MethodPost:
		postOfferSeller(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func getOfferSeller(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		viewData.ForceLogin(w, r)
		return
	}

	args := URIArgs(r)
	if len(args) != 1 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	offer, err := models.GetOfferByID(Base.Db, id)
	if err != nil || offer == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Trying to counter an offer on someone else's listing
	if offer.Seller.ID != viewData.Session.User.ID {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	listing, err := models.GetListingByID(Base.Db, offer.Listing.ID)
	if err != nil || listing == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	viewData.Data = createOfferViewData{
		HasError: false,
		Listing:  *listing,
		Offer:    *offer,
	}

	RenderView(w, "offer#seller", viewData)
}

func postOfferSeller(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		viewData.ForceLogin(w, r)
		return
	}

	if !viewData.ValidCsrf(r) {
		http.Redirect(w, r, r.RequestURI, http.StatusFound)
		return
	}

	id, err := strconv.Atoi(r.FormValue("offer_id"))
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	offer, err := models.GetOfferByID(Base.Db, id)
	if err != nil || offer == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	listing, err := models.GetListingByID(Base.Db, offer.Listing.ID)
	if err != nil || listing == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	var offerErr *models.OfferError
	var ok bool

	counter, err := utils.PriceClientToServer(r.FormValue("counter"))
	if err != nil {
		ok = false
		offerErr = &models.OfferError{Counter: "This must be a valid price"}
		offer.CounterClient = r.FormValue("counter")
		offer.SellerComment = r.FormValue("seller_comment")
	} else {
		wasPreviouslyCountered := offer.IsCountered
		offer.Counter = counter
		offer.CounterClient = r.FormValue("counter")
		offer.SellerComment = r.FormValue("seller_comment")
		offer.IsCountered = true
		ok, offerErr = offer.Save(Base.Db)
		if ok {
			offer.Listing = *listing
			offer.Seller = viewData.Session.User
			Base.WebsockChannel <- wsock.UserJSONNotification(&offer.Buyer,
				"NOTIF_OFFER_COUNTER", offer, true)
			if !wasPreviouslyCountered {
				buyer := models.GetUserByID(Base.Db, offer.Buyer.ID)
				if buyer != nil {
					offer.Buyer = *buyer
					email.NewCounterEmail(*offer)
				}
			}
		}
	}

	if !ok {
		viewData.Data = createOfferViewData{
			HasError: true,
			Error:    *offerErr,
			Listing:  *listing,
			Offer:    *offer,
		}
		RenderView(w, "offer#seller", viewData)
	} else {
		Base.WebsockChannel <- wsock.UserMessage(&offer.Buyer, "-IOfferCountered")
		http.Redirect(w, r, "/selling", http.StatusFound)
	}
}

// BuyerList handles the route '/buying'
func BuyerList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getBuyerList(w, r)
	default:
		http.Error(w, constants.Error404, http.StatusNotFound)
	}
}

func getBuyerList(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		viewData.ForceLogin(w, r)
		return
	}

	offers, err := viewData.Session.User.GetOffersAsBuyer(Base.Db)
	if err != nil {
		http.Error(w, constants.Error500, http.StatusInternalServerError)
		return
	}

	viewData.Data = buyerListViewData{
		Offers: offers,
	}
	RenderView(w, "offer#buying", viewData)
}

type webAPIOffersAsBuyerResponse struct {
	HasError bool
	Error    string
	Offers   []models.Offer
}

// WebAPIOffersAsBuyer handles the route '/webapi/offers/buyer'
func WebAPIOffersAsBuyer(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	response := webAPIOffersAsBuyerResponse{
		HasError: true,
	}
	if viewData.Session == nil {
		response.Error = "Authentication Error"
		RenderJSON(w, response)
		return
	}

	offers, err := viewData.Session.User.GetOffersAsBuyer(Base.Db)
	if err != nil {
		response.Error = err.Error()
		RenderJSON(w, response)
		return
	}

	response.HasError = false
	response.Offers = offers
	RenderJSON(w, response)
}

type webAPIOffersAsSellerResponse struct {
	Successful bool           `json:"successful"`
	Error      string         `json:"error,omitempty"`
	Offers     []models.Offer `json:"offers"`
}

// WebAPIOffersAsSeller handles the route '/webapi/offers/seller'
func WebAPIOffersAsSeller(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	response := webAPIOffersAsSellerResponse{
		Successful: false,
	}

	if viewData.Session == nil {
		response.Error = constants.ErrorAuth
		RenderJSON(w, response)
		return
	}

	args := URIArgs(r)
	pageNum := 0
	if len(args) == 1 {
		var err error
		pageNumStr := args[0]
		pageNum, err = strconv.Atoi(pageNumStr)
		if err != nil {
			pageNum = 0
		}
	}

	offers, err := viewData.Session.User.GetOffersAsSeller(Base.Db, pageNum,
		offerPageSize)
	if err != nil {
		response.Error = constants.Error500
		fmt.Println(err.Error())
		RenderJSON(w, response)
		return
	}

	response.Successful = true
	response.Offers = offers
	RenderJSON(w, response)
}

type webAPIOfferDeleteResponse struct {
	Error      string `json:"error,omitempty"`
	Successful bool   `json:"successful"`
}

// WebAPIOfferDelete handles the route '/webapi/offer/delete/'
func WebAPIOfferDelete(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)

	response := webAPIOfferDeleteResponse{
		Successful: false,
	}
	if viewData.Session == nil {
		response.Error = "Not Logged In"
		RenderJSON(w, response)
		return
	}
	if !viewData.ValidCsrf(r) {
		response.Error = "Not Authorized"
		RenderJSON(w, response)
		return
	}

	args := URIArgs(r)

	if len(args) != 1 {
		response.Error = "Invalid Request"
		RenderJSON(w, response)
		return
	}

	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error = "Invalid ID"
		RenderJSON(w, response)
		return
	}

	offer, err := models.GetOfferByID(Base.Db, id)
	if err != nil {
		response.Error = "Not Found"
		RenderJSON(w, response)
		return
	}

	if offer.Seller.ID == viewData.Session.User.ID ||
		offer.Buyer.ID == viewData.Session.User.ID {
		if ok := offer.Delete(Base.Db); ok {

			if offer.Seller.ID == viewData.Session.User.ID {
				offer.Seller = viewData.Session.User
				listing, err := models.GetListingByID(Base.Db, offer.Listing.ID)
				if err == nil && listing != nil {
					offer.Listing = *listing
					Base.WebsockChannel <- wsock.UserJSONNotification(&offer.Buyer,
						"NOTIF_OFFER_REJECTED", offer, true)
				}
			} else {
				offer.Buyer = viewData.Session.User
				listing, err := models.GetListingByID(Base.Db, offer.Listing.ID)
				if err == nil && listing != nil {
					offer.Listing = *listing
					Base.WebsockChannel <- wsock.UserJSONNotification(&offer.Seller,
						"NOTIF_OFFER_REVOKED", offer, true)
				}
			}

			response.Successful = true
			RenderJSON(w, response)
			return
		}
		response.Error = "Unexpected Error"
		RenderJSON(w, response)
		return
	}

	response.Error = "Not Authorized"
	RenderJSON(w, response)
}

type webAPIListingOfferResponse struct {
	Successful bool          `json:"successful"`
	Error      string        `json:"error,omitempty"`
	Offer      *models.Offer `json:"offer"`
}

// WebAPIListingOffer handles the route '/webapi/listing/offer/'
func WebAPIListingOffer(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	response := webAPIListingOfferResponse{
		Successful: false,
	}
	if viewData.Session == nil {
		response.Error = constants.ErrorAuth
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

	offer, err := viewData.Session.User.GetOfferOnListing(Base.Db, id)
	if err != nil {
		response.Error = constants.Error500
		RenderJSON(w, response)
		return
	}

	if offer == nil {
		response.Error = constants.Error404
		RenderJSON(w, response)
		return
	}

	response.Successful = true
	response.Offer = offer
	RenderJSON(w, response)
}

type webAPIListingOfferListResponse struct {
	Successful bool           `json:"successful"`
	Error      string         `json:"error,omitempty"`
	Offers     []models.Offer `json:"offers"`
}

// WebAPIListingOfferList handles the route '/webapi/listing/offers/'
func WebAPIListingOfferList(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	response := webAPIListingOfferListResponse{
		Successful: false,
	}
	if viewData.Session == nil {
		response.Error = constants.ErrorAuth
		RenderJSON(w, response)
		return
	}

	args := URIArgs(r)
	if len(args) < 1 || len(args) > 2 {
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

	if listing.User.ID != viewData.Session.User.ID {
		response.Error = constants.Error403
		RenderJSON(w, response)
		return
	}

	pageNum := 0
	if len(args) == 2 {
		pageNumStr := args[1]
		pageNum, err = strconv.Atoi(pageNumStr)
		if err != nil {
			pageNum = 0
		}
	}

	offers, err := listing.GetOffers(Base.Db, pageNum, offerPageSize)
	if err != nil {
		response.Error = constants.Error500
		RenderJSON(w, response)
		return
	}

	response.Successful = true
	response.Offers = offers
	RenderJSON(w, response)
}

type webAPIOfferAcceptResponse struct {
	Error      string `json:"error,omitempty"`
	Successful bool   `json:"successful"`
}

// WebAPIOfferAccept handles the route '/webapi/offer/accept/'
func WebAPIOfferAccept(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)

	response := webAPIOfferAcceptResponse{
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

	offer, err := models.GetOfferByID(Base.Db, id)
	if err != nil || offer == nil {
		response.Error = constants.Error404
		RenderJSON(w, response)
		return
	}

	if offer.Seller.ID != viewData.Session.User.ID {
		response.Error = constants.Error403
		RenderJSON(w, response)
		return
	}

	offer.Status = models.OfferAccepted
	if ok, _ := offer.Save(Base.Db); !ok {
		response.Error = constants.Error500
		RenderJSON(w, response)
		return
	}

	listing, err := models.GetListingByID(Base.Db, offer.Listing.ID)
	if err == nil && listing != nil {
		offer.Seller = viewData.Session.User
		offer.Listing = *listing
		Base.WebsockChannel <- wsock.UserJSONNotification(&offer.Buyer,
			"OFFER_ACCEPTED", offer, true)
	}

	response.Successful = true
	RenderJSON(w, response)
}

type webAPIOfferFinalizeResponse struct {
	Error      string `json:"error,omitempty"`
	Successful bool   `json:"successful"`
}

// WebAPIOfferFinalize handles the route '/webapi/offer/finalize'
func WebAPIOfferFinalize(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)

	response := webAPIOfferFinalizeResponse{
		Successful: false,
	}
	if viewData.Session == nil {
		response.Error = "Not logged in"
		RenderJSON(w, response)
		return
	}
	if !viewData.ValidCsrf(r) {
		response.Error = "Invalid CSRF"
		RenderJSON(w, response)
		return
	}

	args := URIArgs(r)
	if len(args) != 1 {
		response.Error = "Invalid Request"
		RenderJSON(w, response)
		return
	}

	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error = "Invalid ID"
		RenderJSON(w, response)
		return
	}

	offer, err := models.GetOfferByID(Base.Db, id)
	if err != nil || offer == nil {
		response.Error = "Offer does not exist"
		RenderJSON(w, response)
		return
	}

	if offer.Seller.ID != viewData.Session.User.ID {
		response.Error = "Not Permitted"
		RenderJSON(w, response)
		return
	}

	offer.Status = models.OfferCompleted
	if ok, _ := offer.Save(Base.Db); !ok {
		response.Error = "Unexpected Error"
		RenderJSON(w, response)
		return
	}

	if ok, _ := offer.Listing.MarkSold(Base.Db); !ok {
		response.Error = "Unexpected Error While Marking Listing"
		RenderJSON(w, response)
		return
	}

	response.Successful = true
	RenderJSON(w, response)
}

type webAPIConversationListResponse struct {
	Offers   []models.Offer `json:"offers"`
	HasError bool           `json:"has_error"`
	Error    string         `json:"error,omitempty"`
}

// WebAPIConversationList handles the route '/api/conversation/list'
func WebAPIConversationList(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	response := webAPIConversationListResponse{}
	if viewData.Session == nil {
		response.HasError = true
		response.Error = "Not Authenticated"
		RenderJSON(w, response)
		return
	}

	offers, err := viewData.Session.User.GetConversationsForUser(Base.Db)
	if err != nil {
		response.HasError = true
		response.Error = "Internal Error"
		RenderJSON(w, response)
		return
	}

	response.HasError = false
	response.Offers = offers
	RenderJSON(w, response)
}
