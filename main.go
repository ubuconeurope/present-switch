package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/alexandrevicenzi/go-sse"
)

func handleRooms(h http.Handler, s *sse.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Based on the implementation of the http.StripPrefix()

		re := regexp.MustCompile(`^(/rooms/([0-9]+))/?.*`) // 3 groups
		reMatch := re.FindStringSubmatch(r.URL.Path)

		if len(reMatch) == 3 {
			prefix := reMatch[1]
			roomNumber := reMatch[2]

			switch r.Method {
			case "GET", "HEAD":

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
				// curl -X POST "http://localhost:3000/rooms/1" -d '{"title": "Title Room 1", "speaker": "John Doe", "time": "15:00", "next": "Next title @ 16:00"}'
				messageData, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Println("Error: could not parse post body")
					return
				}
				s.SendMessage("/events/room-"+roomNumber, sse.SimpleMessage(string(messageData)))

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

	// go func() {
	// 	for {
	// 		stringJSON := `{"title": "Title for Room 1", "next": "Next Presentation title @ 16:00", "time": "` + time.Now().String() + `"}`
	// 		s.SendMessage("/events/channel-1", sse.SimpleMessage(stringJSON))
	// 		time.Sleep(5 * time.Second)
	// 	}
	// }()

	// go func() {
	// 	i := 0
	// 	for {
	// 		i++
	// 		stringJSON := `{"title": "Title for Room 2", "next": "Next Presentation With a Relative Big Text Title, so I Can Test This big Text @ 16:00", "time": "` + strconv.Itoa(i) + `"}`
	// 		s.SendMessage("/events/channel-2", sse.SimpleMessage(stringJSON))
	// 		time.Sleep(5 * time.Second)
	// 	}
	// }()

	fmt.Println("============= Instructions ==============")
	fmt.Println("1) Open your browser at http://localhost:3000/rooms/1/ (do not forget the last /)")
	fmt.Println(`2) curl -X POST -H 'Content-Type: application/json'  "http://localhost:3000/rooms/1" -d '{"title": "Title Room 1", "speaker": "John Doe", "time": "15:00", "next": "Next title @ 16:00"}'`)
	fmt.Println("3) You may try with any other room number. They should stay independent")
	fmt.Println()

	log.Println("Listening at :3000")
	http.ListenAndServe(":3000", nil)
}
