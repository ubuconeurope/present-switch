package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/alexandrevicenzi/go-sse"
)

const dbFilename string = "presentswitch.db"

var db *sql.DB

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

	// Endpoint for static content and roomInfo updates
	http.Handle("/rooms/", handleRooms(http.FileServer(http.Dir("html_template")), s))

	// Endpoint for SSE events
	http.Handle("/events/", s)

	log.Println("Opening Database")
	db = InitDB(dbFilename)
	CreateTables(db)

	log.Println("Listening at :3000")
	http.ListenAndServe(":3000", nil)
}
