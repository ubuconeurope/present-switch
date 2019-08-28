package main

import (
	"fmt"
	"log"
	"net/http"
)

// Broker class is created in this program. It is responsible
// for keeping a list of which clients (browsers) are currently attached
// and broadcasting events (messages) to those clients.
type Broker struct {
	// Create a map of clients, the keys of the map are the channels
	// over which we can push messages to attached clients.  (The values
	// are just booleans and are meaningless.)
	clients map[chan string]bool
	// Channel into which new clients can be pushed
	newClients chan chan string
	// Channel into which disconnected clients should be pushed
	defunctClients chan chan string
	// Channel into which messages are pushed to be broadcast out
	// to attahed clients.
	messages chan string
}

// Start will start a new goroutine.  It handles the addition & removal of clients, as well as the broadcasting
// of messages out to clients that are currently attached.
//
func (b *Broker) Start() {
	// Start a goroutine to handle all interactions between the SSE bridge.
	go func() {
		for {
			// Block until we receive from one of the three following channels.
			select {
			case s := <-b.newClients:
				b.clients[s] = true
				log.Println("Added new client")
			case s := <-b.defunctClients:
				delete(b.clients, s)
				close(s)
				log.Println("Removed client")
			case msg := <-b.messages:
				for s := range b.clients {
					s <- msg
				}
				log.Printf("Broadcast message to %d clients", len(b.clients))
			}
		}
	}()
}

// This Broker method handles and HTTP request at the "/events/" URL.
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Make sure that the writer supports flushing.
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	// Create a new channel, over which the broker can send this client messages.
	messageChan := make(chan string)
	// Add this client to the map of those that should receive updates
	b.newClients <- messageChan
	// Listen to the closing of the http connection via the CloseNotifier
	notify := w.(http.CloseNotifier).CloseNotify()

	go func() {
		<-notify
		// Remove this client from the map of attached clients when `EventHandler` exits.
		b.defunctClients <- messageChan
		log.Println("HTTP connection just closed.")
	}()

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Read from our messageChan.
	for {
		msg, open := <-messageChan
		if !open {
			// If our messageChan was closed, this means that the client has disconnected.
			break
		}
		// Write to the ResponseWriter, `w`.
		fmt.Fprintf(w, "%s\n\n", msg)
		// Flush the response.  This is only possible if the repsonse supports streaming.
		f.Flush()
	}

	// Done.
	log.Println("Finished HTTP request at ", r.URL.Path)
}
