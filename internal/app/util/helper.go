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
    "net/http"
    "errors"
    "io/ioutil"
    "encoding/json"
)

func JsonPrettyDump(obj interface{}) string {
    jsonBytes, err := json.MarshalIndent(obj, "", "\t")
    if err != nil {
        log.Fatal(err)
    }
    return string(jsonBytes)+"\n"
}

func JsonDump(obj interface{}) string {
    jsonBytes, err := json.Marshal(obj)
    if err != nil {
        log.Fatal(err)
    }
    return string(jsonBytes)
}

func JsonLoad(file string) (map[string]interface{}, error) {
    if len(file) == 0 {
        return nil, nil
    }
    _, err := os.Stat(file)
    if os.IsNotExist(err) {
        return nil, errors.New("'"+file+"': no such file or directory")
    }
	buf, err := ioutil.ReadFile(file)
    if err != nil {
        log.Fatal(err)
    }
	content := make(map[string]interface{})
	err = json.Unmarshal([]byte(buf), &content)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
    return content, nil
}

func fetch(url string) string {
    res, err := http.Get(url)
    if err != nil {
        log.Fatal(err)
    }
    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        log.Fatal(err)
    }    
    return string(body)
}

func FetchInfos(host string) *Infos {
    return ParseInfos(fetch("http://"+host+"/infos.json"))
}

func FetchGenesis(host string) string {
    return fetch("http://"+host+"/genesis.json")
}
