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
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
)

// This struct handles socket connection to the hub
type Client struct {
	conn *websocket.Conn
}

// Connect to the hub
func Connect(host string) *Client {
	client := &Client{}
	client.setConnection(host)
	return client
}

// Connect to websocket
func (client *Client) setConnection(host string) {
	url := url.URL{Scheme: "ws", Host: host, Path: "/ws/enroll"}
	//log.Printf("connecting to %s", url.String())
	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	client.conn = conn
}

// Send infos
func (client *Client) Send(message string) {
	err := client.conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Fatal("write:", err)
	}
}

// Get genesis
func (client *Client) Receive() map[string]interface{} {
	_, message, err := client.conn.ReadMessage()
	if err != nil {
		log.Fatal("read:", err)
	}
	input := make(map[string]interface{})
	err = json.Unmarshal([]byte(message), &input)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return input
}

// Close websocket
func (client *Client) Close() {
	err := client.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Fatal("write close:", err)
	}
	//client.conn.Close()
}
