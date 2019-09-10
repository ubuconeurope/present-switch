package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/alexandrevicenzi/go-sse"
)

func persistRoomInfo(message []byte, roomNumberStr string) error {

	roomNumber, err := strconv.Atoi(roomNumberStr)
	if err != nil {
		return err
	}

	var roomInfo RoomInfo
	if err = json.Unmarshal(message, &roomInfo); err != nil {
		return err
	}
	roomInfo.ID = roomNumber
	log.Println(roomInfo)

	StoreItem(db, roomInfo)

	return err
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
	messageData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error: could not parse json data", http.StatusBadRequest)
		return
	}

	err = persistRoomInfo(messageData, roomNumberStr)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.SendMessage("/events/room-"+roomNumberStr, sse.SimpleMessage(string(messageData)))
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
