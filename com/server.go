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
    "fmt"
    "log"
    "os"
    "encoding/json"
    "path/filepath"
    "io/ioutil"
    "net/http"
)

const (
    ServerPort = "8080"
    RootServerDir = "/chain/config"
)

type WebsocketHandler func(resp http.ResponseWriter, req *http.Request)

type resources interface {
    GetServerPort() string
    GetInfos() string
    GetGenesis() string
    GetNodes() string
    Enroll(host string, name string, isNode bool)
}

type server struct {
    handler WebsocketHandler
    resources resources
    httpServer *http.Server
}

func NewServer(resources resources) *server {
    addr, exists := os.LookupEnv("SERVER_URL")
    if ! exists {
    	addr = "0.0.0.0:"+resources.GetServerPort()
    }
    httpServer := &http.Server{Addr: addr, Handler: http.DefaultServeMux}
    return &server{resources: resources, httpServer: httpServer}
}

func (server *server) SetHandler(handler WebsocketHandler) {
    server.handler = handler
}

func (server *server) Close() {
    server.httpServer.Close()
}

func (server *server) Serve() {
    err := os.MkdirAll(RootServerDir, os.ModePerm)
    if err != nil {
		log.Fatal(err)
    }
	http.HandleFunc("/ws/enroll", server.handler)
    http.HandleFunc("/upload", func (resp http.ResponseWriter, req *http.Request) {
        req.ParseMultipartForm(10 << 20) // maximum upload of 10 MB files
        file, handler, err := req.FormFile("file") // get file entry
        if err != nil {
            fmt.Println("Error Retrieving File")
            fmt.Println(err)
            return
        }
        defer file.Close()
        fileBytes, err := ioutil.ReadAll(file)
        if err != nil {
            fmt.Println(err)
            return
        }
        filename := filepath.Join(RootServerDir, handler.Filename)
        err = ioutil.WriteFile(filename, fileBytes, os.ModePerm)
        if err != nil {
    		fmt.Println(err)
    		return
        }
        fmt.Fprintf(resp, "Successfully Uploaded File %s", filename)
    })
	http.HandleFunc("/", func (resp http.ResponseWriter, req *http.Request) {
	    if req.Method != "GET" {
		    http.Error(resp, "Method not allowed", http.StatusMethodNotAllowed)
		    return
	    }
	    filename := filepath.Join(RootServerDir, req.URL.EscapedPath())
	    log.Println(filename)
	    http.ServeFile(resp, req, filename)
    })
    http.HandleFunc("/enroll", func (resp http.ResponseWriter, req *http.Request) {
        if req.Method != "POST" {
            http.Error(resp, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        decoder := json.NewDecoder(req.Body)
        input := make(map[string]interface{})
        err := decoder.Decode(&input)
        if err != nil {
    		log.Fatal(err)
        }
        if _, exists := input["name"]; !exists {
            fmt.Println("Error Retrieving Field")
            return
        }
        if _, exists := input["bootstrap"]; !exists {
            fmt.Println("Error Retrieving Field")
            return
        }
        if _, exists := input["validate"]; !exists {
            fmt.Println("Error Retrieving Field")
            return
        }
        name := input["name"].(string)
        bootstrap := input["bootstrap"].(string)
        validate := input["validate"].(bool)
        resp.Header().Set("Content-Type", "application/json")
        log.Printf("Bootstrap = %s; Name = %s\n", bootstrap, name)
        go func() {
            server.resources.Enroll(bootstrap, name, !validate)
        }()
        //fmt.Fprintf(resp, "Name = %s\n", name)
        jsonBytes, err := json.MarshalIndent(input, "", "\t")
        if err != nil {
            log.Fatal(err)
        }
        fmt.Fprintf(resp, string(jsonBytes)+"\n")
    })
    http.HandleFunc("/infos", func (resp http.ResponseWriter, req *http.Request) {
        if req.Method != "GET" {
            http.Error(resp, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        resp.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(resp, server.resources.GetInfos())
    })
    http.HandleFunc("/genesis", func (resp http.ResponseWriter, req *http.Request) {
        if req.Method != "GET" {
            http.Error(resp, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        resp.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(resp, server.resources.GetGenesis())
    })
    http.HandleFunc("/nodes", func (resp http.ResponseWriter, req *http.Request) {
        if req.Method != "GET" {
            http.Error(resp, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        resp.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(resp, server.resources.GetNodes())
    })
    //log.Printf("Listening %s", server.httpServer.Addr)
    err = server.httpServer.ListenAndServe()
    if err != nil {
        //log.Fatal("ListenAndServe: ", err)
    }
}
