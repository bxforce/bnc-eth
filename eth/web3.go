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
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"time"
)

const (
	Empty               = "0x0000000000000000000000000000000000000000"
	Wei                 = 1000000000000000000
	gasLimit            = uint64(4700000) // TODO: use config
	retry           int = 30
	connectionRetry     = time.Duration(5)
)

type Web3 struct {
	Url    string
	Client *ethclient.Client
}

func NewWeb3(url string) *Web3 {
	return &Web3{Url: url}
}

func (web3 *Web3) Connect() *ethclient.Client {
	for {
		client, err := ethclient.Dial(web3.Url)
		if err != nil {
			time.Sleep(connectionRetry * time.Second)
		} else {
			web3.Client = client
			break
		}
	}
	return web3.Client
}

func (web3 *Web3) GetBalance(receiver string) (string, error) {
	address := common.HexToAddress(receiver)
	balance, err := web3.Client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return "", err
	}
	balanceEth := new(big.Int).Div(balance, big.NewInt(Wei))
	return balanceEth.String(), nil
}

func (web3 *Web3) prepareTransaction(fromAddress *common.Address, toAddress *common.Address, val int64, data []byte) *types.Transaction {
	nonce, err := web3.Client.PendingNonceAt(context.Background(), *fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	gasPrice, err := web3.Client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	value := new(big.Int).Mul(big.NewInt(val), big.NewInt(Wei))
	var rawTx *types.Transaction
	if toAddress == nil {
		rawTx = types.NewContractCreation(nonce, value, gasLimit, gasPrice, data)
	} else {
		rawTx = types.NewTransaction(nonce, *toAddress, value, gasLimit, gasPrice, data)
	}
	return rawTx
}

func (web3 *Web3) signTransaction(privateKey *ecdsa.PrivateKey, tx *types.Transaction) *types.Transaction {
	chainID, err := web3.Client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	return signedTx
}

func (web3 *Web3) sendTransaction(signedTx *types.Transaction) (*types.Receipt, error) {
	err := web3.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}
	var receipt *types.Receipt
	for i := 0; i < retry; i++ {
		receipt, err = web3.Client.TransactionReceipt(context.Background(), signedTx.Hash())
		if err != nil {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	if receipt == nil {
		err = errors.New("Receipt not received in " + string(retry) + " s")
	}
	return receipt, err
}

func (web3 *Web3) Transact(private string, receiver string, val int64) (*types.Transaction, *types.Receipt, error) {
	privateKey, fromAddress := getAddress(private)
	toAddress := common.HexToAddress(receiver)
	data := make([]byte, 0)
	tx := web3.prepareTransaction(fromAddress, &toAddress, val, data)
	signedTx := web3.signTransaction(privateKey, tx)
	receipt, err := web3.sendTransaction(signedTx)
	return signedTx, receipt, err
}

func getAddress(private string) (*ecdsa.PrivateKey, *common.Address) {
	privateKey, err := crypto.HexToECDSA(private)
	if err != nil {
		log.Fatal(err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return privateKey, &address
}
