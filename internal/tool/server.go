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

package tool

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
)

const (
	defaultRootPath   = "/var/www/localhost/htdocs"
	defaultConfigPath = "/etc/lighttpd/lighttpd.conf"
	defaultConfig     = `
server.document-root = "/var/www/localhost/htdocs/" 
server.port = 5000
mimetype.assign = (
  ".html" => "text/html",
  ".json" => "application/json",
  ".txt" => "text/plain",
  ".png" => "image/png" 
)`
)

type LightServer struct {
	cmd *exec.Cmd
}

func NewLightServer() *LightServer {
	err := os.MkdirAll(defaultRootPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(defaultConfigPath, []byte(defaultConfig), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Kill, os.Interrupt)
	go func() {
		log.Println("signal: ", <-signals)
		cancel()
		os.Exit(0)
	}()
	server := &LightServer{}
	server.cmd = exec.CommandContext(ctx, "lighttpd", "-D", "-f", defaultConfigPath)
	//server.cmd.Stdout = os.Stdout
	//server.cmd.Stderr = os.Stderr
	return server
}

func (server *LightServer) Start() {
	//log.Println(server.cmd.String())
	err := server.cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}

func (server *LightServer) Stop() {
	if err := server.cmd.Process.Kill(); err != nil {
		log.Fatal("failed to kill process: ", err)
	}
}

func (server *LightServer) Publish(filename string, content string) {
	err := ioutil.WriteFile(filepath.Join(defaultRootPath, filename), []byte(content), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}
