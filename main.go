package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alexandrevicenzi/go-sse"
)

const dbFilename string = "presentswitch.db"

var db *sql.DB

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

				if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
					r2 := new(http.Request)
					*r2 = *r
					r2.URL = new(url.URL)
					*r2.URL = *r.URL
					r2.URL.Path = p

					roomNumber, err := strconv.Atoi(roomNumberStr)
					if err != nil {
						http.Error(w, "Error: room number cannot be converted to int", http.StatusBadRequest)
						return
					}

					// If there is a persisted entry for this room, use it
					if roomInfo, err := ReadRoomInfo(db, roomNumber); err == nil {
						if roomInfoJSON, err2 := json.Marshal(roomInfo); err2 == nil {
							go func() {
								time.Sleep(1000 * time.Millisecond) // FixMe: this is wrong, but works (occasionally)
								s.SendMessage("/events/room-"+roomNumberStr, sse.SimpleMessage(string(roomInfoJSON)))
							}()
						}
					}

					h.ServeHTTP(w, r2)
				} else {
					http.NotFound(w, r)
				}

			case "POST", "PUT":
				messageData, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Error: could not parse json data", 400)
					return
				}

				err = persistRoomInfo(messageData, roomNumberStr)
				if err != nil {
					http.Error(w, err.Error(), 400)
					return
				}
				s.SendMessage("/events/room-"+roomNumberStr, sse.SimpleMessage(string(messageData)))

			}
		} else {
			http.NotFound(w, r)
		}

	})

	// prefixToRemove := fmt.Sprintf("/rooms/%s/", numberToRemove[1])
	// fmt.Println("removing prefix: ", prefixToRemove)

	// http.StripPrefix(prefixToRemove, http.FileServer(http.Dir("html_template")))
	// http.StripPrefix("/rooms/", http.FileServer(http.Dir("html_template")))
}

func main() {
	s := sse.NewServer(&sse.Options{
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, OPTIONS",
			"Access-Control-Allow-Headers": "Keep-Alive,X-Requested-With,Cache-Control,Content-Type,Last-Event-ID",
		},
		// Custom channel name generator
		ChannelNameFunc: func(request *http.Request) string {
			return request.URL.Path
		},
		Logger: log.New(os.Stdout, "go-sse: ", log.Ldate|log.Ltime|log.Lshortfile),
	})
	defer s.Shutdown()

	// Create internal redirect server to render static files
	http.Handle("/rooms/", handleRooms(http.FileServer(http.Dir("html_template")), s))

	// Register /events endpoint
	http.Handle("/events/", s)

	log.Println("Opening Database")
	db = InitDB(dbFilename)
	CreateTables(db)

	log.Println("Listening at :3000")
	http.ListenAndServe(":3000", nil)
}
