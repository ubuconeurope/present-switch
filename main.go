package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alexandrevicenzi/go-sse"
)

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
	fs := http.FileServer(http.Dir("templates"))
	http.Handle("/rooms/", http.StripPrefix("/rooms/", fs))

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
