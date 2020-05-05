/*
Copyright 2020 IRT SystemX

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package com

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)

// This struct handles socket connection in the Hub
type Member struct {
	id string
	hub *Hub
	conn *websocket.Conn
	sock chan []byte
}

type Message struct {
	member *Member
	content []byte
}

// Close the member socket
func (member *Member) close() {
	member.hub.unregister <- member
	//log.Printf("close connection: %v", member.conn.RemoteAddr())
}

// Read member input
func (member *Member) read() {
	defer member.close()
	for {
		_, message, err := member.conn.ReadMessage()
		if err != nil {
			/*if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Fatal("error: %v", err)
			}*/
			return
		}
		member.hub.process <- &Message{member: member, content: message}
	}
}

// Send output to the member
func (member *Member) write() {
	defer member.close()
	for {
		select {
		case message, ok := <-member.sock:
			if !ok {
				member.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := member.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

// This struct handles multiple members socket to collect their infos
// and then send them back the output of the gathering (aka a genesis)
type Hub struct {
	members map[*Member]bool
	register chan *Member
	unregister chan *Member
	process chan *Message
	collect CollectHandler
	generate GenerateHandler
	complete CompleteHandler
}

// Collect member infos and return nil until the amount of members is reached
type CollectHandler func(message string) (string)
// Generate genesis if the amount of members is reached
type GenerateHandler func() (map[string][]byte)
// Complete the hub and end the listening 
type CompleteHandler func()

// Create a hub
func NewHub(collect CollectHandler, generate GenerateHandler, complete CompleteHandler) *Hub {
	return &Hub{
		members:    make(map[*Member]bool),
		register:   make(chan *Member),
		unregister: make(chan *Member),
		process:    make(chan *Message),
		collect:    collect,
		generate:    generate,
		complete:    complete,
	}
}

// Handle a new member socket connection
func (hub *Hub) HandleConnection(resp http.ResponseWriter, req *http.Request) {
    upgrader := websocket.Upgrader{ ReadBufferSize:  1024, WriteBufferSize: 1024 }
    conn, err := upgrader.Upgrade(resp, req, nil)
    if err != nil {
	    log.Fatal(err)
        return
    }
	member := &Member{hub: hub, conn: conn, sock: make(chan []byte, 256)}
	member.hub.register <- member
	//log.Printf("open connection: %v", member.conn.RemoteAddr())
	go member.write()
	go member.read()
}

// Listen to member actions
func (hub *Hub) Listen() {
	for {
		select {
		case member  := <- hub.register:
			// register
			hub.members[member] = true
		case message := <- hub.process:
			// process input
			message.member.id = hub.collect(string(message.content))
			outputs := hub.generate()
			if outputs != nil {
				// broadcast output
				for member := range hub.members {
					member.sock <- outputs[member.id]
				}
			}
		case member  := <- hub.unregister:
			// unregister
			if _, ok := hub.members[member]; ok {
				member.conn.Close()
				delete(hub.members, member)
				close(member.sock)
				if len(hub.members) == 0 {
					hub.complete()
				}
			}
		}
	}
}

