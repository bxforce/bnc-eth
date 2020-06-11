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
    "log"
    "os"
    "time"
    "math/rand"
    "math/big"
    "encoding/hex"
    "crypto/ecdsa"
    "io/ioutil"
    "path/filepath"
    "github.com/ethereum/go-ethereum/accounts/keystore"
    "github.com/ethereum/go-ethereum/common/hexutil"
    "github.com/ethereum/go-ethereum/crypto"
)

var DefaultPrivateKey = "b45c2d049b489a5d7f5a1b5212a0c262472a28b241e73e3e465d3133036a1c2f"
var DefaultAddress = "0x1005388e1649240036d199b6ad71eafc0164edad"

// Create a random ASCII password
func createPassword(length int) string {
    rand.Seed(time.Now().UnixNano())
    chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
            "abcdefghijklmnopqrstuvwxyz" +
            "0123456789" +
            "~=+%^*/()[]{}/!@#$?|"
    
    buf := make([]byte, length)
    for i := range buf {
        buf[i] = chars[rand.Intn(len(chars))]
    }
    return string(buf)
}

// Get password from file and create it if neeeded
func getPassword(filename string) string {
    _, err := os.Stat(filename)
    if os.IsNotExist(err) {
        
        password := createPassword(16)

	    err := ioutil.WriteFile(filename, []byte(password), os.ModePerm)
	    if err != nil {
    		log.Fatal(err)
	    }
    }

    passwordBytes, err := ioutil.ReadFile(filename)
    if err != nil {
        log.Fatal(err)
    }
    return string(passwordBytes)
}

func DecodePrivateKey(privateKey string) *ecdsa.PrivateKey {
    privStr, err := hex.DecodeString(privateKey)
    if err != nil {
        log.Fatal(err)
    }
    priv := new(ecdsa.PrivateKey)
    priv.D = new(big.Int).SetBytes([]byte(privStr))
    priv.PublicKey.Curve = crypto.S256()
    priv.PublicKey.X, priv.PublicKey.Y = crypto.S256().ScalarBaseMult(priv.D.Bytes())
    return priv
}

func EncodePrivateKey(priv *ecdsa.PrivateKey) string {
    return hex.EncodeToString(priv.D.Bytes())
}

// Create a keystore
func createKeystore(backupFile string, ksPath string, password string) {
    ks := keystore.NewKeyStore(ksPath, keystore.StandardScryptN, keystore.StandardScryptP)
    
    privBytes, err := ioutil.ReadFile(backupFile)
    if err != nil {
        log.Fatal(err)
    }
    
    priv := DecodePrivateKey(string(privBytes))
    
    _, err = ks.ImportECDSA(priv, password)
    if err != nil {
        log.Fatal(err)
    }
}

// Create a key
func createKey(backupFile string) {
    priv, err := crypto.GenerateKey()
    if err != nil {
		log.Fatal(err)
    }
    
    err = ioutil.WriteFile(backupFile, []byte(EncodePrivateKey(priv)), os.ModePerm)
    if err != nil {
		log.Fatal(err)
    }
}

// Get privateKey from keystore and create it if neeeded
func GetKeystore(backupFile string, keystoreDir string, passwordFile string) (string, string, string) {
    
    password := getPassword(passwordFile)
    
    _, err := os.Stat(backupFile)
    if os.IsNotExist(err) {
        createKey(backupFile)
    }

    err = os.RemoveAll(keystoreDir)
    if err != nil {
		log.Fatal(err)
    }

    createKeystore(backupFile, keystoreDir, password)

    files, err := ioutil.ReadDir(keystoreDir)
    if err != nil {
        log.Fatal(err)
    }

    jsonBytes, err := ioutil.ReadFile(filepath.Join(keystoreDir, files[0].Name()))
    if err != nil {
        log.Fatal(err)
    }

    key, err := keystore.DecryptKey(jsonBytes, password)
	if err != nil {
		log.Fatal(err)
	}
	
    privateKeyBytes := crypto.FromECDSA(key.PrivateKey)
    privateKeyStr := hexutil.Encode(privateKeyBytes)[2:]

    publicKey := key.PrivateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
    }

    publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
    publicKeyStr := hexutil.Encode(publicKeyBytes)[4:]

    address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

    return privateKeyStr, publicKeyStr, address
}

