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
	"errors"
	com "github.com/bxforce/bnc-eth/com"
	"log"
	"path/filepath"
)

var (
	defaultParamsFile   = filepath.Join(com.RootServerDir, "params.json")
	defaultAccountsFile = filepath.Join(com.RootServerDir, "accounts.json")
	defaultTopologyFile = filepath.Join(com.RootServerDir, "topology.json")
)

type manager struct {
	config     *config
	hub        *com.Hub
	nodes      []*Infos
	validators []string
	count      int
}

func NewManager(handler com.CompleteHandler) (*manager, error) {
	config, err := NewConfig()
	manager := &manager{config: config}
	manager.hub = com.NewHub(manager.collect, manager.generate, handler)
	return manager, err
}

func (manager *manager) Serve() {
	if manager.validators == nil {
		log.Fatal(errors.New("Error: limit must be set"))
	}
	log.Printf("Start hub (waiting for %x validators)", len(manager.validators))
	go manager.hub.Listen()
}

func (manager *manager) SetLimit(limit int) {
	manager.validators = make([]string, limit)
}

func (manager *manager) SetConfig(config string) {
	manager.config.ParseConfig(config)
}

func (manager *manager) GetGenesis() map[string]interface{} {
	return manager.config.genesis
}

func (manager *manager) GetNodes() []*Infos {
	return manager.nodes
}

func (manager *manager) GetHandleConnection() com.WebsocketHandler {
	return manager.hub.HandleConnection
}

func (manager *manager) collect(message string) string {
	infos := ParseInfos(message)
	if infos.Faucets != nil {
		for key, value := range infos.Faucets {
			manager.config.accounts[key] = value
		}
	}
	manager.nodes = append(manager.nodes, infos)
	if len(infos.Name) > 0 {
		if infos.IsValidator {
			manager.validators[manager.count] = infos.Address
			log.Printf("New validator: %s", infos.Address)
			manager.count++
		} else {
			log.Printf("New node: %s", infos.Address)
		}
	}
	return infos.Name
}

func (manager *manager) generate() map[string][]byte {
	if manager.count == len(manager.validators) {
		if len(manager.config.configFile) == 0 {
			manager.SetConfig(defaultParamsFile)
		}
		accounts := manager.getAccounts()
		peers := manager.getTopology()
		manager.config.setParams()
		manager.config.setValidators(manager.validators)
		manager.config.setAccounts(accounts)
		log.Printf("Broadcast genesis & topology")
		outputs := make(map[string][]byte)
		for _, node := range manager.nodes {
			output := make(map[string]interface{})
			output["genesis"] = manager.config.genesis
			output["peers"] = peers[node.Name]
			outputs[node.Name] = []byte(JsonDump(output))
		}
		return outputs
	}
	return nil
}

func (manager *manager) getAccounts() map[string]string {
	accounts, _ := JsonLoad(defaultAccountsFile)
	if accounts != nil {
		for key, value := range accounts {
			manager.config.accounts[key] = value.(string)
		}
	}
	return manager.config.accounts
}

func (manager *manager) getTopology() map[string][]string {
	topology, _ := JsonLoad(defaultTopologyFile)
	if topology != nil {
		for key, value := range topology {
			manager.config.topology[key] = make([]string, len(value.([]interface{})))
			for i, val := range value.([]interface{}) {
				manager.config.topology[key][i] = val.(string)
			}
		}
	}
	peers := make(map[string][]string)
	for _, node := range manager.nodes {
		peers[node.Name] = make([]string, 0)
		for _, peer := range manager.nodes {
			if find(peer.Name, manager.config.topology[node.Name]) || manager.config.topology == nil {
				peers[node.Name] = append(peers[node.Name], peer.Enode())
			}
		}
	}
	return peers
}

func find(val string, slice []string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func ConnectMember(host string, infos string) (map[string]interface{}, []string) {
	client := com.Connect(host)
	client.Send(infos)
	genesis, peers := parseGenerate(client.Receive())
	client.Close()
	return genesis, peers
}

func parseGenerate(input map[string]interface{}) (map[string]interface{}, []string) {
	_, ok := input["genesis"]
	if !ok {
		log.Fatalf("error: field not found")
	}
	genesis := input["genesis"].(map[string]interface{})
	_, ok = input["peers"]
	if !ok {
		log.Fatalf("error: field not found")
	}
	topology := input["peers"].([]interface{})
	peers := make([]string, len(topology))
	for i, node := range topology {
		peers[i] = node.(string)
	}
	//log.Printf("Peers: %s", peers)
	return genesis, peers
}
