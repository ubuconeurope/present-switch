[![Build Status](https://travis-ci.org/ubuconeurope/present-switch.svg?branch=master)](https://travis-ci.org/ubuconeurope/present-switch)
# Presenter Switch

The goal of this project is to allow automated triggering of content into HTTP clients listening on a [SSE](https://en.wikipedia.org/wiki/Server-sent_events) connection.

Ultimately, this allows having multiple clients listening to different URLs for specific content and allows you to change the room's content remotely or automatically (if you create a bot). 

At Ubucon Europe 2019, we will have 4 rooms to manage, each with their own schedule and non-overlapping terms. This server allows some clients to connect and we can change the content of the template by calling a POST method against the API with a body like `body.go`. Provided we will have an iCal calendar, we can automate this process by building another bot that will trigger the HTTP events on time, thus dynamically changing the room's displayed text.

## Running

Check out the repository and simply install the binaries (requires [Go](https://golang.org/) to be installed on your machine:

```
git clone https://github.com/ubuconeurope/present-switch
cd present-switch/
go run main.go
```
The server should then start listening on port `3000`.

Then: 
```
1) Open your browser at http://localhost:3000/rooms/1/ (do not forget the last /)
2) curl -X POST -H 'Content-Type: application/json'  "http://localhost:3000/rooms/1" -d '{"title": "Title Room 1", "speaker": "John Doe", "time": "15:00", "next": "Next title @ 16:00"}'
3) You may try with any other room number. They should stay independent
```



## Contributing

Check the issues page to know where your skills can be useful.
