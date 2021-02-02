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

package eth

import (
	"bytes"
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

type Contract struct {
	Name    string
	Address string
	parser  abi.ABI
}

func NewContract(name string, address string, buf []byte) (*Contract, error) {
	parser, err := abi.JSON(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	return &Contract{Name: name, Address: address, parser: parser}, nil
}

func LoadContract(solcPath string, name string) (*Contract, error) {
	buf, err := ioutil.ReadFile(filepath.Join(solcPath, name+".abi"))
	if err != nil {
		return nil, err
	}
	return NewContract(name, "", buf)
}

func LoadContracts(abi string, config string) (map[string]*Contract, error) {
	_, err := os.Stat(config)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(config)
	if err != nil {
		return nil, err
	}
	registry := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(data), &registry)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	contracts := make(map[string]*Contract)
	err = filepath.Walk(abi, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".abi" {
			buf, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			name := filepath.Base(path)
			name = name[:len(name)-4]
			if address, ok := registry[name]; ok {
				contract, err := NewContract(name, address.(string), buf)
				if err != nil {
					return err
				}
				contracts[contract.Address] = contract
				log.Printf("Load contract %s %s", contract.Name, contract.Address)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return contracts, nil
}

func (web3 *Web3) LoadContracts(name string) (map[string]*Contract, error) {
	contracts := make(map[string]*Contract)
	return contracts, nil
}

func (contract *Contract) ParseData(data []byte) (map[string]interface{}, error) {
	if data != nil && len(data) > 4 {
		functionId := string(hexutil.Encode(data[:4]))
		for method := range contract.parser.Methods {
			hash := crypto.Keccak256Hash([]byte(contract.parser.Methods[method].Sig())).Hex()[:10]
			if functionId == hash {
				buf := make(map[string]interface{})
				err := contract.parser.Methods[method].Inputs.UnpackIntoMap(buf, data[4:])
				if err != nil {
					log.Fatal(err)
				}
				inputs := make([]map[string]interface{}, len(contract.parser.Methods[method].Inputs))
				for i, arg := range contract.parser.Methods[method].Inputs.NonIndexed() {
					inputs[i] = make(map[string]interface{})
					inputs[i]["name"] = arg.Name
					inputs[i]["type"] = arg.Type.String()
					inputs[i]["value"] = buf[arg.Name]
				}
				event := make(map[string]interface{})
				event["name"] = method
				event["params"] = inputs
				return event, err
			}
		}
	}
	return nil, errors.New("Data mismatch with contract functions")
}

func (contract *Contract) ParseLogs(logs []*types.Log) []map[string]interface{} {
	var events []map[string]interface{}
	for eventName := range contract.parser.Events {
		eventSignature := contract.parser.Events[eventName].Sig()
		re := regexp.MustCompile(`([a-zA-Z]+)?`)
		eventLabel := re.FindString(eventSignature)
		hash := crypto.Keccak256Hash([]byte(eventSignature)).Hex()
		for _, vLog := range logs {
			for i := range vLog.Topics {
				if vLog.Topics[i].Hex() == hash {
					buf := make(map[string]interface{})
					err := contract.parser.UnpackIntoMap(buf, eventName, vLog.Data)
					if err != nil {
						log.Fatal(err)
					}
					event := make(map[string]interface{})
					event["name"] = eventLabel
					eventsParams := make([]map[string]interface{}, len(contract.parser.Events[eventName].Inputs))
					for j, arg := range contract.parser.Events[eventName].Inputs.NonIndexed() {
						eventsParams[j] = make(map[string]interface{})
						eventsParams[j]["name"] = arg.Name
						eventsParams[j]["type"] = arg.Type.String()
						eventsParams[j]["value"] = buf[arg.Name]
					}
					event["params"] = eventsParams
					events = append(events, event)
				}
			}
		}
	}
	return events
}

func (contract *Contract) Get() abi.ABI {
	return contract.parser
}

func (contract *Contract) Deploy(web3 *Web3, bin []byte, private string, val int64, params ...interface{}) (string, error) {
	input, err := contract.parser.Pack("", params...)
	if err != nil {
		return "", err
	}
	bytecode := common.FromHex(string(bin))
	data := append(bytecode, input...)
	privateKey, fromAddress := getAddress(private)
	tx := web3.prepareTransaction(fromAddress, nil, val, data)
	signedTx := web3.signTransaction(privateKey, tx)
	receipt, err := web3.sendTransaction(signedTx)
	if err != nil {
		return "", err
	}
	contractAddress := receipt.ContractAddress.Hex()
	if contractAddress == Empty || receipt.Status == 0 {
		return "", errors.New("Contract deploy failure")
	} else {
		return contractAddress, nil
	}
}

func (contract *Contract) Transact(web3 *Web3, private string, val int64, contractAddress string, method string, params ...interface{}) (*types.Transaction, *types.Receipt, error) {
	privateKey, fromAddress := getAddress(private)
	address := common.HexToAddress(contractAddress)
	input, err := contract.parser.Pack(method, params...)
	if err != nil {
		return nil, nil, err
	}
	tx := web3.prepareTransaction(fromAddress, &address, val, input)
	signedTx := web3.signTransaction(privateKey, tx)
	receipt, err := web3.sendTransaction(signedTx)
	return signedTx, receipt, err
}

func (contract *Contract) Call(web3 *Web3, contractAddress string, method string, params ...interface{}) (interface{}, error) {
	input, err := contract.parser.Pack(method, params...)
	if err != nil {
		return nil, err
	}
	address := common.HexToAddress(contractAddress)
	msg := ethereum.CallMsg{To: &address, Data: input}
	output, err := web3.Client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = contract.parser.Unpack(&result, method, output)
	return result, err
}
