package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Loads the template
// TODO add ways to dynamically set next conf name and etup
func loadTemplate(content ...string) *template.Template {
	// Read in the template with our SSE JavaScript code.
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal("Errors parsing your template file. Fix it and try again.")
	}
	return t
}

// Handler for the main page, which we wire up to the route at "/" below in `main`.
// TODO ensure it loads only by ID
func contentGetter(w http.ResponseWriter, r *http.Request, b *Broker) {
	var rid = strings.TrimPrefix(r.URL.Path, "/rooms/")
	if _, err := strconv.ParseInt(rid, 10, 8); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	t := loadTemplate(rid)
	t.Execute(w, "friend")
	pushMessage(b, *t, rid)
	log.Println("Finished HTTP request at", r.URL.Path)
}

// Method that controls modifying the content of a template.
// First it attempts to modify it and then it will push the message to the output.
func changeContent(w http.ResponseWriter, r *http.Request, b *Broker) {
	var rid = strings.TrimPrefix(r.URL.Path, "/rooms/")
	var body = parseBody(r)
	ReplaceInTemplate(rid, body.convertToMap())

	t := loadTemplate(rid)
	t.Execute(w, "friend")
	pushMessage(b, *t, rid)
	log.Println("Finished HTTP Request at", r.URL.Path)
}

// Method to parse body and handle its errors separately.
func parseBody(r *http.Request) RequestBody {
	var rb RequestBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&rb)
	if err != nil {
		panic(err)
	}

	return rb
}

func pushMessage(b *Broker, t template.Template, rid string) {
	var tpl bytes.Buffer
	t.Execute(&tpl, strings.Join([]string{"room"}, rid))
	b.messages <- tpl.String()
}

// Main routine
func main() {
	// Make a new Broker instance
	b := &Broker{
		make(map[chan string]bool),
		make(chan (chan string)),
		make(chan (chan string)),
		make(chan string),
	}

	// Start processing events
	b.Start()

	// Make b the HTTP handler for "/events/".  It can do
	// this because it has a ServeHTTP method.  That method
	// is called in a separate goroutine for each
	// request to "/events/".
	http.Handle("/events/", b)

	// Routing handler
	http.HandleFunc("/rooms/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			contentGetter(w, r, b)
			return
		case http.MethodPost:
			changeContent(w, r, b)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})

	// This function sends data for the message chan every 5seconds.
	go func() {
		for i := 0; ; i++ {

			// Create a little message to send to clients,
			// including the current time.
			b.messages <- fmt.Sprintf("TEST MESSAGE: %d - the time is: <br>%v", i, time.Now())

			// Print a nice log message and sleep for 5s.
			log.Printf("Sent message %d ", i)
			time.Sleep(5e9)

		}
	}()

	// Routing handler
	http.Handle("/", http.FileServer(http.Dir("static_files")))

	// Start the server and listen forever on port 8000.
	log.Println("Starting...")
	http.ListenAndServe(":8000", nil)
	log.Print("Server staretd listening to new clients.")
}
