package email

import (
	"strconv"

	"github.com/anishmgoyal/calagora/models"
	"github.com/anishmgoyal/calagora/utils"
)

// NewOfferEmail is sent when a user gets a new offer on a listing
func NewOfferEmail(offer models.Offer) {
	title := "Calagora - New Offer"
	paragraphs := []interface{}{
		offer.Buyer.DisplayName + " offered you $" +
			offer.PriceClient + " for your listing, " + makeLink(
			"https://www.calagora.com/listing/view/"+
				strconv.Itoa(offer.Listing.ID), offer.Listing.Name),
		"You can accept, reject, or counter this offer by logging in and going " +
			"to the page for your listing, which is at:",
		makeURLLink("https://www.calagora.com/listing/view/" +
			strconv.Itoa(offer.Listing.ID)),
		"Or you can view this offer, as well as offers for other listings, from " +
			"the seller tab at:",
		makeURLLink("https://www.calagora.com/selling/"),
	}
	email := &utils.Email{
		To:            []string{offer.Seller.EmailAddress},
		From:          Base.AutomatedEmail,
		Subject:       title,
		FormattedText: GenerateHTML(title, paragraphs),
		PlainText:     GeneratePlain(title, paragraphs),
	}
	Base.EmailChannel <- email
}

// NewCounterEmail is sent when a user's offer is first countered
func NewCounterEmail(offer models.Offer) {
	title := "Calagora - Offer Countered"
	paragraphs := []interface{}{
		offer.Seller.DisplayName + " countered your offer of $" +
			offer.PriceClient + " for " + makeLink(
			"https://www.calagora.com/listing/view/"+
				strconv.Itoa(offer.Listing.ID), offer.Listing.Name) + " with $" +
			offer.CounterClient,
		"You can edit or revoke this offer from the buyer tab directly at:",
		makeURLLink("https://www.calagora.com/offer/buyer/" +
			strconv.Itoa(offer.Listing.ID)),
		"Or, from the listing's page at:",
		makeURLLink("https://www.calagora.com/listing/view/" +
			strconv.Itoa(offer.Listing.ID)),
		"You can also edit or revoke other offers you've made or received from " +
			"the buyer tab at:",
		makeURLLink("https://www.calagora.com/buying/"),
	}
	email := &utils.Email{
		To:            []string{offer.Buyer.EmailAddress},
		From:          Base.AutomatedEmail,
		Subject:       title,
		FormattedText: GenerateHTML(title, paragraphs),
		PlainText:     GeneratePlain(title, paragraphs),
	}
	Base.EmailChannel <- email
}

// OfferAcceptedEmail is sent when a user's offer is accepted
func OfferAcceptedEmail(offer models.Offer) {
	title := "Calagora - Offer Accepted"
	paragraphs := []interface{}{
		offer.Seller.DisplayName + " has accepted your offer of $" +
			offer.PriceClient + " for " + makeLink(
			"https://www.calagora.com/listing/view/"+
				strconv.Itoa(offer.Listing.ID), offer.Listing.Name),
		"You can now chat with this seller about finishing your transaction in " +
			"the Messages tab at:",
		makeURLLink("https://www.calagora.com/message/client/#conversation") +
			strconv.Itoa(offer.ID),
	}
	email := &utils.Email{
		To:            []string{offer.Buyer.EmailAddress},
		From:          Base.AutomatedEmail,
		Subject:       title,
		FormattedText: GenerateHTML(title, paragraphs),
		PlainText:     GeneratePlain(title, paragraphs),
	}
	Base.EmailChannel <- email
}
