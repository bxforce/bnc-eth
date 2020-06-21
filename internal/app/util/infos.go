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

package util

import (
	"encoding/json"
	"log"
	"os"
)

var DefaultFaucetValue = "10000000000000000000000000000000000"

type Infos struct {
	Name        string            `json:"name"`
	Address     string            `json:"address"`
	PublicKey   string            `json:"publicKey"`
	Host        string            `json:"host"`
	IsValidator bool              `json:"validator"`
	Faucets     map[string]string `json:"faucets"`
}

func NewInfos(publicKey string, address string) *Infos {
	infos := &Infos{Address: address, PublicKey: publicKey}
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	host, exists := os.LookupEnv("ENODE_HOST")
	if exists {
		hostname = host
	}
	port, exists := os.LookupEnv("ENODE_PORT")
	if !exists {
		port = "30303"
	}
	infos.Host = hostname + ":" + port
	infos.Faucets = map[string]string{address: DefaultFaucetValue}
	return infos
}

func (infos *Infos) Enode() string {
	return "enode://" + infos.PublicKey + "@" + infos.Host
}

func ParseInfos(message string) *Infos {
	infos := Infos{}
	err := json.Unmarshal([]byte(message), &infos)
	if err != nil {
		log.Fatal(err)
	}
	return &infos
}
