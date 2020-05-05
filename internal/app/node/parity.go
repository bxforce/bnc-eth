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
    "os"
    "os/exec"
    "os/signal"
    "context"
    "path/filepath"
    "io/ioutil"
    "encoding/json"
    "strings"
    eth "bnc-eth/internal/eth"
    box "bnc-eth/internal/box"
)

const (
    RootConfigDir = "/chain"
)

var (
    configDir = filepath.Join(RootConfigDir, "init")
    dataDir = filepath.Join(RootConfigDir, "data")
    
    parityDir = filepath.Join(dataDir, "parity")
    networkDir = filepath.Join(parityDir, "network")
    keystoreDir = filepath.Join(parityDir, "keys")
    backupFile = filepath.Join(dataDir, "backup_key")
    passwordFile = filepath.Join(configDir, "node.pwd")
    keyFile = filepath.Join(networkDir, "key")
    signerFile = filepath.Join(dataDir, "signer")
    peersFile = filepath.Join(dataDir, "peers.txt")
    
    configFile = filepath.Join(dataDir, "config.toml")
    genesisFile = filepath.Join(configDir, "genesis.json")
    infosFile = filepath.Join(RootConfigDir, "infos.json")
)

type parity struct {
    node *node
    discover bool
}

/**
 * METHODS
 */
func (parity *parity) initialize() (string, string) {
    err := os.MkdirAll(dataDir, os.ModePerm)
    if err != nil {
		log.Fatal(err)
    }
    err = os.MkdirAll(configDir, os.ModePerm)
    if err != nil {
		log.Fatal(err)
    }
    err = os.MkdirAll(networkDir, os.ModePerm)
    if err != nil {
		log.Fatal(err)
    }

    // Create keystore and write file /chain/init/node.pwd
    privateKey, publicKey, address := eth.GetKeystore(backupFile, keystoreDir, passwordFile)

    // Write file /chain/data/parity/chain/network/key
    err = ioutil.WriteFile(keyFile, []byte(privateKey), os.ModePerm)
    if err != nil {
		log.Fatal(err)
    }
    
    // Write file /chain/data/parity/signer
    err = ioutil.WriteFile(signerFile, []byte(address), os.ModePerm)
    if err != nil {
		log.Fatal(err)
    }

    // Write file /chain/infos.json
    infos := make(map[string]interface{})
    infos["pubkey"] = publicKey
    infos["address"] = address
    jsonBytes, err := json.Marshal(infos)
    if err != nil {
		log.Fatal(err)
    }
    err = ioutil.WriteFile(infosFile, jsonBytes, os.ModePerm)
    if err != nil {
		log.Fatal(err)
    }
    
    // Write empty peers
    _, err = os.Stat(peersFile)
    if os.IsNotExist(err) {
        parity.setPeers("")
    }

    return publicKey, address
}

func (parity *parity) configure(genesis string) {

    // Write file /chain/init/genesis.json
    err := ioutil.WriteFile(genesisFile, []byte(genesis), os.ModePerm)
    if err != nil {
		log.Fatal(err)
    }
    
    // Write file /chain/data/parity/config.toml
    err = ioutil.WriteFile(configFile, []byte(box.Get("/config.toml")), os.ModePerm)
    if err != nil {
		log.Fatal(err)
    }
}

func (parity *parity) run() {
    buf, err := ioutil.ReadFile(peersFile)
    if err != nil {
        log.Fatal(err)
    }
    peers := strings.Split(string(buf), ",")

    buf, err = ioutil.ReadFile(signerFile)
    if err != nil {
        log.Fatal(err)
    }
    signer := string(buf)
    
    ctx, cancel := context.WithCancel(context.Background())
    signals := make(chan os.Signal, 1)
    signal.Notify(signals, os.Kill, os.Interrupt)
    go func() {
		log.Println("signal: ", <-signals)
		cancel()
		os.Exit(0)
	}()
	
	identity := parity.node.infos.Name
	
    cmd := exec.CommandContext(ctx, "parity", "--base-path", parityDir, "--config", configFile, "--engine-signer", signer, "--author", signer)
    if len(peers) >= 1 {
        args := make([]string , len(cmd.Args))
        copy(args, cmd.Args)
        if !parity.discover {
            args = append(args, "--no-discovery")
        }
        if len(identity) > 0 {
            args = append(args, "--identity", identity)
        }
        args = append(args, "--bootnodes", strings.Join(peers, ","))
        cmd.Args = args
    }
    cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
    
    log.Println(cmd.String())
    
    err = cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}

/**
 * GETTERS & SETTERS
 */
func (parity *parity) setPrivateKey(privateKey string) {
    if len(privateKey) > 0 {
        if privateKey == "0x0" {
            privateKey = eth.DefaultPrivateKey
        } else if privateKey[:2] == "0x" {
            privateKey = privateKey[2:len(privateKey)]
        }
        priv := eth.DecodePrivateKey(privateKey)
        err := os.MkdirAll(dataDir, os.ModePerm)
        if err != nil {
    		log.Fatal(err)
        }
        err = ioutil.WriteFile(backupFile, []byte(eth.EncodePrivateKey(priv)), os.ModePerm)
        if err != nil {
    		log.Fatal(err)
        }
    }
}

func (parity *parity) setPeers(peers string) {
    err := ioutil.WriteFile(peersFile, []byte(peers), os.ModePerm)
    if err != nil {
		log.Fatal(err)
    }
}

func (parity *parity) setDiscover(discover bool) {
    parity.discover = discover
}
