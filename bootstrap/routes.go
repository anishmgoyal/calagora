package bootstrap

import (
	"net/http"

	"golang.org/x/net/websocket"

	"github.com/anishmgoyal/calagora/constants"
	"github.com/anishmgoyal/calagora/controllers"
	"github.com/anishmgoyal/calagora/resources"
	"github.com/anishmgoyal/calagora/wsock"
)

// CreateRoutes maps URI's to the corresponding controller method
func CreateRoutes() {
	resources.MapCSSHandler()
	resources.MapImageHandler()
	resources.MapJSHandler()

	if !constants.DoUploadAWS {
		resources.MapUploadHandler()
	}

	http.Handle(route("/unsupported", controllers.HomeUnsupported))

	http.Handle(route("/buying/", controllers.BuyerList))
	http.Handle(route("/selling/", controllers.SellerList))

	http.Handle(route("/info/about/", controllers.InfoAbout))
	http.Handle(route("/info/contact/", controllers.InfoContact))
	http.Handle(route("/info/help/", controllers.InfoHelp))
	http.Handle(route("/info/tos/", controllers.InfoTos))

	http.Handle(route("/listing/create/", controllers.ListingCreate))
	http.Handle(route("/listing/delete/", controllers.ListingDelete))
	http.Handle(route("/listing/edit/", controllers.ListingEdit))
	http.Handle(route("/listing/view/", controllers.ListingView))
	http.Handle(route("/listing/section/", controllers.ListingSection))

	http.Handle(route("/message/client/", controllers.MessageClient))
	http.Handle(route("/message/read/", controllers.MessageRead))

	http.Handle(route("/offer/buyer/", controllers.OfferBuyer))
	http.Handle(route("/offer/seller/", controllers.OfferSeller))

	http.Handle(route("/recover/user/", controllers.ResetPassword))
	http.Handle(route("/recover/", controllers.RecoverPassword))

	http.Handle(route("/search/", controllers.Search))

	http.Handle(route("/upload/", controllers.Upload))

	http.Handle(route("/user/activate/", controllers.UserActivate))
	http.Handle(route("/user/login/", controllers.UserLogin))
	http.Handle(route("/user/logout/", controllers.UserLogout))
	http.Handle(route("/user/profile/", controllers.UserProfile))
	http.Handle(route("/user/register/", controllers.UserRegister))

	http.Handle(route("/webapi/conversation/list/", controllers.WebAPIConversationList))

	http.Handle(route("/webapi/image/delete/", controllers.WebAPIImageDelete))
	http.Handle(route("/webapi/image/meta/", controllers.WebAPIImageMeta))

	http.Handle(route("/webapi/listings/user/", controllers.WebAPIListingsUser))
	http.Handle(route("/webapi/listings/", controllers.WebAPIListings))
	http.Handle(route("/webapi/listing/offer/", controllers.WebAPIListingOffer))
	http.Handle(route("/webapi/listing/offers/", controllers.WebAPIListingOfferList))
	http.Handle(route("/webapi/listing/delete/", controllers.WebAPIListingDelete))

	http.Handle(route("/webapi/messages/", controllers.WebAPIMessages))
	http.Handle(route("/webapi/message/send/", controllers.WebAPIMessageSend))

	http.Handle(route("/webapi/notification/counts/", controllers.WebAPINotificationCounts))
	http.Handle(route("/webapi/notifications/", controllers.WebAPINotifications))

	http.Handle(route("/webapi/offer/delete/", controllers.WebAPIOfferDelete))
	http.Handle(route("/webapi/offer/accept/", controllers.WebAPIOfferAccept))
	http.Handle(route("/webapi/offer/finalize/", controllers.WebAPIOfferFinalize))
	http.Handle(route("/webapi/offers/buyer/", controllers.WebAPIOffersAsBuyer))
	http.Handle(route("/webapi/offers/seller/", controllers.WebAPIOffersAsSeller))

	http.Handle(route("/webapi/upload/progress/", controllers.WebAPIUploadProgress))

	http.Handle(route("/test/email/", controllers.TestEmail))

	http.Handle("/ws/", websocket.Handler(wsock.Connect))

	http.Handle(route("/", controllers.Home))
}

// Quick wrapper for StripPrefix which prevents typos
func route(path string, callback http.HandlerFunc) (string, http.Handler) {
	fn := callback
	handler := http.StripPrefix(path, fn)
	if constants.SSLEnable {
		fn = func(w http.ResponseWriter, r *http.Request) {
			if redirect, ok := constants.Domain.Map[r.Host]; ok {
				http.Redirect(w, r, redirect+r.URL.RequestURI(),
					http.StatusMovedPermanently)
				return
			}
			http.StripPrefix(path, callback).ServeHTTP(w, r)
		}
		handler = http.HandlerFunc(fn)
	}
	return path, handler
}
