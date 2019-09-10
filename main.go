package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/alexandrevicenzi/go-sse"
)

const dbFilename string = "presentswitch.db"

var db *sql.DB

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
