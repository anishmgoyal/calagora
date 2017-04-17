package bootstrap

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anishmgoyal/calagora/cache"
	"github.com/anishmgoyal/calagora/constants"
	"github.com/anishmgoyal/calagora/controllers"
	"github.com/anishmgoyal/calagora/email"
	"github.com/anishmgoyal/calagora/utils"
	"github.com/anishmgoyal/calagora/wsock"
)

// GlobalStart begins initialization for the application,
// and notifies main() if an error occurrs.
func GlobalStart() bool {

	fmt.Println("[STARTUP] Loading Environment Settings")
	constants.LoadEnvironmentSettings()

	// Seed for random generators
	rand.Seed(time.Now().UTC().UnixNano())

	fmt.Println("[STARTUP] Loading Templates")
	templates := GetTemplates()

	fmt.Println("[STARTUP] Connecting to DB")
	db := GetDatabaseConnection()

	fmt.Println("[STARTUP] Initializing Services")
	cache.BaseInitialization(db)
	controllers.BaseInitialization(templates, db)
	email.BaseInitialization(templates)
	wsock.BaseInitialization(db)

	go utils.SessionEvicter(db)

	fmt.Println("[STARTUP] Creating Routes")
	CreateRoutes()

	var sslRedirect = func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if index := strings.Index(host, ":"); index > -1 {
			host = host[0:index]
		}

		nextHost := host
		if replaceHost, ok := constants.Domain.Map[nextHost]; ok {
			nextHost = replaceHost
		} else {
			nextHost = "https://" + nextHost
		}
		redirectURL := nextHost

		if constants.SSLPortNum != 443 {
			redirectURL += ":" + strconv.Itoa(constants.SSLPortNum)
		}
		http.Redirect(w, r, redirectURL+r.URL.RequestURI(),
			http.StatusMovedPermanently)
	}

	if constants.SSLEnable {
		fmt.Println("[STARTUP] Starting server on port " +
			strconv.Itoa(constants.SSLPortNum))
		fmt.Println("[STARTUP] Using redirect from port " +
			strconv.Itoa(constants.PortNum))

		go http.ListenAndServe(":"+strconv.Itoa(constants.PortNum),
			http.HandlerFunc(sslRedirect))
		http.ListenAndServeTLS(":"+strconv.Itoa(constants.SSLPortNum),
			constants.SSLCertificate, constants.SSLKeyFile, nil)
	} else {
		fmt.Println("[STARTUP] Starting server on port " +
			strconv.Itoa(constants.PortNum))

		http.ListenAndServe(":"+strconv.Itoa(constants.PortNum), nil)
	}

	// Shouldn't return, I don't believe..
	fmt.Println("[STARTUP] Startup failed.")
	return false
}
