package main

import (
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

func handleRooms(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		re := regexp.MustCompile(`.*(/rooms/[0-9]+)/.*`)
		prefix := re.FindStringSubmatch(r.URL.Path)

		if len(prefix) > 0 {
			if p := strings.TrimPrefix(r.URL.Path, prefix[1]); len(p) < len(r.URL.Path) {
				r2 := new(http.Request)
				*r2 = *r
				r2.URL = new(url.URL)
				*r2.URL = *r.URL
				r2.URL.Path = p
				h.ServeHTTP(w, r2)
			} else {
				http.NotFound(w, r)
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
	http.Handle("/rooms/", handleRooms(http.FileServer(http.Dir("html_template"))))

	// Register /events endpoint
	http.Handle("/events/", s)

	go func() {
		for {
			s.SendMessage("/events/channel-1", sse.SimpleMessage(time.Now().String()))
			time.Sleep(10 * time.Second)
		}
	}()

	go func() {
		i := 0
		for {
			i++
			s.SendMessage("/events/channel-2", sse.SimpleMessage(strconv.Itoa(i)))
			time.Sleep(10 * time.Second)
		}
	}()

	log.Println("Listening at :3000")
	http.ListenAndServe(":3000", nil)
}
