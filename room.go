package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"log"
	"github.com/m-sugawara/go-tracer"
	"github.com/stretchr/objx"
)

type room struct {
	forward chan *message
	join    chan *client
	leave   chan *client
	clients map[*client]bool
	tracer  trace.Tracer
}

func newRoom() *room {
	return &room{
		forward: make(chan *message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// Attend
			r.clients[client] = true
			r.tracer.Trace("New client joined.")
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("client left.")
		case msg := <-r.forward:
			r.tracer.Trace("New message received: ", msg.Message)
			// Send the message to all clients
			for client := range r.clients {
				select {
				case client.send <- msg:
					// Send message
					r.tracer.Trace(" -- A message was sent to client")
				default:
					// failed to send
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace(" -- Failed to send a message. Clean up client")
				}
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatal("Failed to get the cookie:", err)
		return
	}
	client := &client{
		socket: socket,
		send: make(chan *message, messageBufferSize),
		room: r,
		userData: objx.MustFromBase64(authCookie.Value),
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
