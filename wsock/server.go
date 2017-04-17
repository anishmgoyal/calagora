package wsock

import (
	"database/sql"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anishmgoyal/calagora/models"
	"golang.org/x/net/websocket"
)

var gDB *sql.DB
var wSockets map[string]map[string]*websocket.Conn
var wSocketID = uint64(0)
var connectionMutex sync.Mutex

const (
	// BroadcastChannel can be used to send a message to all active users
	BroadcastChannel = "broadcast##"
	// SessionChannelPrefix can be used as a prefix for sending a user messages
	// only on devices using the given session
	SessionChannelPrefix = "session##"
	// UserChannelPrefix can be used as a prefix for sending a user messages
	UserChannelPrefix = "user##"

	websockThreadCount = 5
	websockChannelSize = 200
)

// Message contains fields necessary to send websocket messages on
// a channel
type Message struct {
	Target  string
	Message string
}

func init() {
	wSockets = make(map[string]map[string]*websocket.Conn)
}

// StartWebsocketService creates goroutines for sending messages via websockets
func StartWebsocketService(db *sql.DB) chan *Message {
	ch := make(chan *Message, websockChannelSize)
	gDB = db
	for i := 0; i < websockThreadCount; i++ {
		go websocketSender(ch)
	}
	return ch
}

func websocketSender(ch chan *Message) {
	for {
		wsmsg := <-ch
		if target, ok := wSockets[wsmsg.Target]; ok {
			for id, ws := range target {
				n, err := ws.Write([]byte(wsmsg.Message))
				if err != nil || n != len(wsmsg.Message) {
					delete(target, id)
				}
			}
		}
	}
}

func waitForAuthentication(ws *websocket.Conn, ch chan []byte) {
	var buff [330]byte
	n, err := ws.Read(buff[:])
	if err != nil {
		return
	}
	ch <- buff[:n]
}

func authenticationTimeout(ch chan bool) {
	time.Sleep(time.Duration(30) * time.Second)
	ch <- false
}

// Connect handles a new websocket connection, then essentially becomes
// the keep-alive loop
// Also: does not permit writing back to the server from the client side
func Connect(ws *websocket.Conn) {
	defer ws.Close()

	var buff []byte

	ch1 := make(chan []byte)
	ch2 := make(chan bool)
	go waitForAuthentication(ws, ch1)
	go authenticationTimeout(ch2)

	select {
	case buff = <-ch1:
		break
	case <-ch2:
		ws.Write([]byte("-EAuthentication Timeout"))
		return
	}

	if session, ok := checkCredentials(string(buff)); ok {
		ws.Write([]byte("-IConnected"))

		connectionMutex.Lock()

		id := strconv.FormatUint(wSocketID, 16)
		wSocketID++
		if wSocketID == 0 {
			panic("Overflow in websocket ID. Server restart required.")
		}

		userChannelMap, ok := wSockets[UserTarget(&session.User)]
		if !ok {
			userChannelMap = make(map[string]*websocket.Conn)
			wSockets[UserTarget(&session.User)] = userChannelMap
		}
		sessionChannelMap, ok := wSockets[SessionTarget(session)]
		if !ok {
			sessionChannelMap = make(map[string]*websocket.Conn)
			wSockets[SessionTarget(session)] = sessionChannelMap
		}
		broadcastChannelMap, ok := wSockets[BroadcastChannel]
		if !ok {
			broadcastChannelMap = make(map[string]*websocket.Conn)
			wSockets[BroadcastChannel] = broadcastChannelMap
		}

		userChannelMap[id] = ws
		sessionChannelMap[id] = ws
		broadcastChannelMap[id] = ws
		buff = make([]byte, 1048)

		connectionMutex.Unlock()

		for {
			count, err := ws.Read(buff)
			if err != nil || count == 0 {
				break
			}
			if strings.Index(string(buff), "-R") == 0 {
				idStr := string(buff)[2:count]
				id, err := strconv.Atoi(idStr)
				if err == nil {
					session.User.MarkNotificationsRead(gDB, id)
				}
			} else if strings.Index(string(buff), "-r") == 0 {
				idStr := string(buff)[2:count]
				id, err := strconv.Atoi(idStr)
				if err == nil {
					session.User.MarkNotificationRead(gDB, id)
				}
			}
		}

		ws.Write([]byte("-EDisconnecting"))

		connectionMutex.Lock()

		delete(userChannelMap, id)
		if len(userChannelMap) == 0 {
			delete(wSockets, UserChannelPrefix+session.User.Username)
		}
		delete(sessionChannelMap, id)
		if len(sessionChannelMap) == 0 {
			delete(wSockets, SessionChannelPrefix+session.SessionID)
		}
		delete(broadcastChannelMap, id)
		if len(broadcastChannelMap) == 0 {
			delete(wSockets, BroadcastChannel)
		}

		connectionMutex.Unlock()
	} else {
		ws.Write([]byte("-EBad Credentials"))
	}
}

func checkCredentials(credentials string) (*models.Session, bool) {
	credentialList := strings.Split(credentials, "~")
	if len(credentialList) != 3 {
		return nil, false
	}

	sessionID := credentialList[0]
	sessionSecret := credentialList[1]
	browserAgent := credentialList[2]

	session := models.GetSession(gDB, sessionID, sessionSecret, browserAgent)
	if session == nil {
		return nil, false
	}

	return session, true
}
