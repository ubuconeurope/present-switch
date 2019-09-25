package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/alexandrevicenzi/go-sse"
)

// PresentationEvent is a wrapper for a Data (RoomInfo os some kind of action)
type PresentationEvent struct {
	Kind string      `json:"kind"`
	Data interface{} `json:"data"`
}

// ControlEventField object for PresentationEvent.Data
type ControlEventField struct {
	Action string `json:"action"`
}

func persistRoomInfo(message []byte, roomNumberStr string) (RoomInfo, error) {

	roomNumber, err := strconv.Atoi(roomNumberStr)
	if err != nil {
		return RoomInfo{}, err
	}

	var roomInfo RoomInfo
	if err = json.Unmarshal(message, &roomInfo); err != nil {
		return RoomInfo{}, err
	}
	roomInfo.ID = roomNumber
	log.Println("persisting ", roomInfo)

	StoreItem(db, roomInfo)

	return roomInfo, err
}

func handleRoomsGET(h http.Handler, w http.ResponseWriter, r *http.Request, s *sse.Server, prefix, roomNumberStr string) {
	// handle missing /
	if r.URL.Path == "/rooms/"+roomNumberStr {
		http.Redirect(w, r, "/rooms/"+roomNumberStr+"/", http.StatusFound)
		return
	}

	if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = p

		h.ServeHTTP(w, r2)
	} else {
		http.NotFound(w, r)
	}
}

func handleRoomsPOST(w http.ResponseWriter, r *http.Request, s *sse.Server, roomNumberStr string) {

	// If both ADMIN_AUTH_USERNAME and ADMIN_AUTH_PASSWORD env variables are defined,
	//     ensure BasicAuth
	if os.Getenv("ROOMS_AUTH_USERNAME") != "" && os.Getenv("ROOMS_AUTH_PASSWORD") != "" {
		username, password, ok := r.BasicAuth()
		if !ok || username != os.Getenv("ROOMS_AUTH_USERNAME") || password != os.Getenv("ROOMS_AUTH_PASSWORD") {
			w.Header().Set("WWW-Authenticate", "Basic realm=rooms")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	}

	messageData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error: could not parse data", http.StatusBadRequest)
		return
	}

	roomInfo, err := persistRoomInfo(messageData, roomNumberStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newEvent := PresentationEvent{"room-info", roomInfo}
	newEventStr, err := json.Marshal(newEvent)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error - marshal response", http.StatusInternalServerError)
	}
	s.SendMessage("/events/room-"+roomNumberStr, sse.SimpleMessage(string(newEventStr)))
}

func handleAdminUpdatePOST(w http.ResponseWriter, r *http.Request, s *sse.Server, roomNumberStr string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error: could not parse data", http.StatusBadRequest)
		return
	}

	roomNumber, err := strconv.Atoi(roomNumberStr)
	if err != nil {
		http.Error(w, "Error: room number cannot be converted to int", http.StatusBadRequest)
		return
	}

	roomInfo := RoomInfo{
		roomNumber,
		r.FormValue("room-name"),
		r.FormValue("current-title"),
		r.FormValue("current-speaker"),
		r.FormValue("current-time"),
		r.FormValue("next-title"),
		r.FormValue("next-speaker"),
		r.FormValue("next-time"),
	}
	StoreItem(db, roomInfo)

	newEvent := PresentationEvent{"room-info", roomInfo}
	newEventStr, err := json.Marshal(newEvent)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error - marshal response", http.StatusInternalServerError)
	}
	s.SendMessage("/events/room-"+roomNumberStr, sse.SimpleMessage(string(newEventStr)))
}

func handleAdminControlPOST(w http.ResponseWriter, r *http.Request, s *sse.Server, roomNumberStr string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error: could not parse data", http.StatusBadRequest)
		return
	}

	newEvent := PresentationEvent{"control", ControlEventField{r.FormValue("action")}}
	newEventStr, err := json.Marshal(newEvent)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error - marshal response", http.StatusInternalServerError)
	}
	s.SendMessage("/events/room-"+roomNumberStr, sse.SimpleMessage(string(newEventStr)))
}

func handleRooms(h http.Handler, s *sse.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Based on the implementation of the http.StripPrefix()

		defer func() {
			if err := recover(); err != nil {
				log.Println("ERROR: ", err)
				http.Error(w, "Internal error", http.StatusInternalServerError)
			}
		}()

		re := regexp.MustCompile(`^(/rooms/([0-9]+))/?.*`) // 3 groups
		reMatch := re.FindStringSubmatch(r.URL.Path)

		if len(reMatch) == 3 {
			prefix := reMatch[1]
			roomNumberStr := reMatch[2]

			switch r.Method {
			case "GET", "HEAD":
				handleRoomsGET(h, w, r, s, prefix, roomNumberStr)
			case "POST", "PUT":
				handleRoomsPOST(w, r, s, roomNumberStr)
			}
		} else {
			http.NotFound(w, r)
		}

	})
}

func handleAdmin(h http.Handler, s *sse.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Based on the implementation of the http.StripPrefix()

		defer func() {
			if err := recover(); err != nil {
				log.Println("ERROR: ", err)
				http.Error(w, "Internal error", http.StatusInternalServerError)
			}
		}()

		// If both ADMIN_AUTH_USERNAME and ADMIN_AUTH_PASSWORD env variables are defined,
		//     ensure BasicAuth
		if os.Getenv("ADMIN_AUTH_USERNAME") != "" && os.Getenv("ADMIN_AUTH_PASSWORD") != "" {
			username, password, ok := r.BasicAuth()
			if !ok || username != os.Getenv("ADMIN_AUTH_USERNAME") || password != os.Getenv("ADMIN_AUTH_PASSWORD") {
				w.Header().Set("WWW-Authenticate", "Basic realm=admin")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		}

		re := regexp.MustCompile(`^(/admin/([0-9]+))/?(.*)?`) // 4 groups
		reMatch := re.FindStringSubmatch(r.URL.Path)

		if len(reMatch) == 4 {
			prefix := reMatch[1]
			roomNumberStr := reMatch[2]
			actionStr := reMatch[3]

			switch r.Method {
			case "GET", "HEAD":
				// handle missing /
				if r.URL.Path == "/admin/"+roomNumberStr {
					http.Redirect(w, r, "/admin/"+roomNumberStr+"/", http.StatusFound)
					return
				}

				if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
					r2 := new(http.Request)
					*r2 = *r
					r2.URL = new(url.URL)
					*r2.URL = *r.URL
					r2.URL.Path = p

					h.ServeHTTP(w, r2)
				} else {
					http.NotFound(w, r)
				}
			case "POST", "PUT":
				switch actionStr {
				case "update":
					handleAdminUpdatePOST(w, r, s, roomNumberStr)
				case "control":
					handleAdminControlPOST(w, r, s, roomNumberStr)
				default:
					http.Error(w, "Invalid Action", http.StatusBadRequest)
				}
				// after handle the POST, redirect to GET
				http.Redirect(w, r, "/admin/"+roomNumberStr+"/", http.StatusFound)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.NotFound(w, r)
		}

	})
}

func handleRoomInfoSync(s *sse.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				log.Println("ERROR: ", err)
				http.Error(w, "Internal error", http.StatusInternalServerError)
			}
		}()

		re := regexp.MustCompile(`^(/room-info/([0-9]+))/?.*`) // 3 groups
		reMatch := re.FindStringSubmatch(r.URL.Path)

		if len(reMatch) == 3 {
			// prefix := reMatch[1]
			roomNumberStr := reMatch[2]

			roomNumber, err := strconv.Atoi(roomNumberStr)
			if err != nil {
				http.Error(w, "Error: room number cannot be converted to int", http.StatusBadRequest)
				return
			}

			// If there is a persisted entry for this room, use it
			if roomInfo, err := ReadRoomInfo(db, roomNumber); err == nil {
				if roomInfoJSON, err2 := json.Marshal(roomInfo); err2 == nil {
					w.WriteHeader(200)
					w.Header().Set("Content-Type", "application/json")
					w.Write(roomInfoJSON)
				}
			}
		}
	})
}
