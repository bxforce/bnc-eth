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
	box "github.com/bxforce/bnc-eth/internal/box"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

type config struct {
	name       string
	difficulty string
	gasLimit   string
	accounts   map[string]string
	topology   map[string][]string
	genesis    map[string]interface{}
	configFile string
}

func NewConfig() (*config, error) {
	genesisStr := string(box.Get("/genesis.json"))
	genesis := make(map[string]interface{})
	err := json.Unmarshal([]byte(genesisStr), &genesis)
	return &config{accounts: make(map[string]string), topology: make(map[string][]string), genesis: genesis}, err
}

func (config *config) ParseConfig(configFile string) {
	_, err := os.Stat(configFile)
	if err == nil {
		config.configFile = configFile
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Fatal(err)
		}
		raw := make(map[interface{}]interface{})
		err = yaml.Unmarshal([]byte(data), &raw)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		name, ok := raw["name"]
		if ok {
			config.name = name.(string)
		}
		gasLimit, ok := raw["gasLimit"]
		if ok {
			config.gasLimit = gasLimit.(string)
		}
		difficulty, ok := raw["difficulty"]
		if ok {
			config.difficulty = difficulty.(string)
		}
		accounts, ok := raw["accounts"]
		if ok {
			for key, value := range accounts.(map[interface{}]interface{}) {
				config.accounts[key.(string)] = value.(string)
			}
		}
	}
}

func (config *config) setParams() {
	if len(config.name) > 0 {
		config.genesis["name"] = config.name
	}
	if len(config.gasLimit) > 0 {
		config.genesis["genesis"].(map[string]interface{})["gasLimit"] = config.gasLimit
	}
	if len(config.difficulty) > 0 {
		config.genesis["genesis"].(map[string]interface{})["difficulty"] = config.difficulty
	}
}

func (config *config) setValidators(validators interface{}) {
	config.genesis["engine"].(map[string]interface{})["authorityRound"].(map[string]interface{})["params"].(map[string]interface{})["validators"].(map[string]interface{})["list"] = validators
}

func (config *config) setAccounts(accounts map[string]string) {
	for key, value := range accounts {
		config.genesis["accounts"].(map[string]interface{})[key] = map[string]string{"balance": value}
	}
}
