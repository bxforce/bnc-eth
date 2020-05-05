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
    "log"
    "os"
    "io/ioutil"
    "encoding/json"
    "gopkg.in/yaml.v2"
    box "bnc-eth/internal/box"
)

type config struct {
    accounts map[string]string
    topology map[string][]string
    genesis map[string]interface{}
}

func NewConfig() (*config, error) {
    genesisStr := string(box.Get("/genesis.json"))
	genesis := make(map[string]interface{})
	err := json.Unmarshal([]byte(genesisStr), &genesis)
    return &config{accounts: make(map[string]string), topology: make(map[string][]string), genesis: genesis}, err
}

func unmarshal(raw map[interface{}]interface{}, field string) map[string]string {
    _, ok := raw[field]
    if !ok {
        log.Fatalf("error: field not found")
    }
    tab := raw[field].(map[interface{}]interface{})
    output := make(map[string]string)
    for key, value := range tab {
        output[key.(string)] = value.(string)
    }
    return output
}

func (config *config) ParseConfig(configFile string)  {
	_, err := os.Stat(configFile)
	if err != nil {
		log.Fatal("Config file is missing: ", configFile)
	}
    data, err := ioutil.ReadFile(configFile)
    if err != nil {
        log.Fatal(err)
    }
    raw := make(map[interface{}]interface{})
    err = yaml.Unmarshal([]byte(data), &raw)
    if err != nil {
        log.Fatalf("error: %v", err)
    }
    config.accounts = unmarshal(raw, "accounts")
}

func (config *config) setValidators(validators interface{}) {
    config.genesis["engine"].(map[string]interface{})["authorityRound"].(map[string]interface{})["params"].(map[string]interface{})["validators"].(map[string]interface{})["list"] = validators
}

func (config *config) setAccounts(accounts map[string]string) {
    for key, value := range accounts {
        config.genesis["accounts"].(map[string]interface{})[key] = map[string]string{"balance": value}
    }
}
