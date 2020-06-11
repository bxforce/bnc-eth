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

package node

import (
    "log"
    "time"
    "io/ioutil"
    "strings"
    util "github.com/bxforce/bnc-eth/internal/app/util"
    tool "github.com/bxforce/bnc-eth/internal/tool"
    com "github.com/bxforce/bnc-eth/internal/com"
)

type engine interface {
    setPrivateKey(string)
    setPeers(string)
    setDiscover(bool)
    initialize() (string, string)
    configure(string)
    run()
}

type server interface {
    Serve()
    SetHandler(com.WebsocketHandler)
    Close()
}

type manager interface {
    Serve()
    SetLimit(int)
    SetConfig(config string)
    GetGenesis() map[string]interface{}
    GetNodes() []*util.Infos
    GetHandleConnection() com.WebsocketHandler
}

/**
 * CONSTRUCTOR
 */
type node struct {
    engine engine
    server server
    manager manager
    infos *util.Infos
    light *tool.LightServer
}

func NewNode() (*node, error) {
    node := &node{}
    node.engine = engine( &parity{node: node} )
    obj, err := util.NewManager(func () {
        log.Printf("Close hub")
        time.Sleep(2 * time.Second) // wait api send response
        node.server.Close()
    })
    node.manager = manager(obj)
    return node, err
}

func (node *node) Initialize() {
    publicKey, address := node.engine.initialize()
    node.infos = util.NewInfos(publicKey, address)
    node.server = server( com.NewServer(node) )
    node.server.SetHandler(node.manager.GetHandleConnection())
    node.light = tool.NewLightServer()
    node.light.Start()
}

func (node *node) SetConfig(config string) {
    if len(config) > 0 {
        node.manager.SetConfig(config)
    }
}

/**
 * METHODS
 */
func (node *node) Serve() {
    node.manager.Serve()
    node.server.Serve()
    node.Run()
}

func (node *node) Enroll(host string, name string, isNode bool) {
    node.infos.Name = name
    node.infos.IsValidator = !isNode
    infosStr := util.JsonPrettyDump(node.infos)
    node.light.Publish("infos.json", infosStr)
    log.Printf("Enroll as validator=%t %s : %s with %s", !isNode, node.infos.Name, node.infos.Address, host)
    genesis, peers := util.ConnectMember(host, infosStr)
	log.Printf("Genesis '%s' and %x peers received", genesis["name"].(string), len(peers))
	genesisStr := util.JsonPrettyDump(genesis)
	node.light.Publish("genesis.json", genesisStr)
	node.engine.configure(genesisStr)
    node.engine.setPeers(strings.Join(peers, ","))
}

func (node *node) Run(params ...string) {
    if len(params) > 0 {
        buf, err := ioutil.ReadFile(params[0])
        if err != nil {
            log.Fatal(err)
        }
        node.engine.configure(string(buf))
    }
    node.engine.run()
}

func (node *node) Connect(host string) {
    node.light.Publish("infos.json", node.GetInfos())
    log.Printf("Enroll as validator=%t %s : %s with %s", false, "anonymous", node.infos.Address, host)
    genesisStr := util.FetchGenesis(host)
	node.light.Publish("genesis.json", genesisStr)
	node.engine.configure(genesisStr)
    infos := util.FetchInfos(host)
    peers := []string{infos.Enode()}
    node.engine.setPeers(strings.Join(peers, ","))
}

/**
 * GETTERS & SETTERS
 */
func (node *node) SetLimit(limit int) {
    node.manager.SetLimit(limit)
}

func (node *node) SetPrivateKey(privateKey string) {
    node.engine.setPrivateKey(privateKey)
}

func (node *node) SetPeers(peers string) {
    node.engine.setPeers(peers)
}

func (node *node) SetDiscover(discover bool) {
    node.engine.setDiscover(discover)
}

func (node *node) GetServerPort() string {
    return com.ServerPort
}

func (node *node) GetInfos() string {
    return util.JsonPrettyDump(node.infos)
}

func (node *node) GetGenesis() string {
    return util.JsonPrettyDump(node.manager.GetGenesis())
}

func (node *node) GetNodes() string {
    return util.JsonPrettyDump(node.manager.GetNodes())
}

