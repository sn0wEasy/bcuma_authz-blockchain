/*
Copyright IBM Corp. 2016 All Rights Reserved.

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

package main

//WARNING - this chaincode's ID is hard-coded in chaincode_example04 to illustrate one way of
//calling chaincode from a chaincode. If this example is modified, chaincode_example04.go has
//to be modified as well with the new ID of chaincode_example02.
//chaincode_example05 show's how chaincode ID can be passed in as a parameter instead of
//hard-coding.

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"

	// "github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric-chaincode-go/shim"

	// pb "github.com/hyperledger/fabric/protos/peer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// 署名検証
func verifySig(msg string, pkStr string, sigStr string) error {

	pkBytes, err := base64.StdEncoding.DecodeString(pkStr)
	if err != nil {
		return err
	}
	pk, err := x509.ParsePKIXPublicKey(pkBytes)
	if err != nil {
		return err
	}

	sigBytes, err := base64.StdEncoding.DecodeString(sigStr)
	if err != nil {
		return err
	}

	// SHA-256 のハッシュ関数を使って受信データのハッシュ値を算出する
	h := crypto.Hash.New(crypto.SHA256)
	h.Write([]byte(msg))
	hashed := h.Sum(nil)

	// 署名の検証．有効な署名はnilを返すことによって示される．
	// 1. 送信者のデータ（署名データ）を公開鍵で複合し，ハッシュ値を算出
	// 2. 受信側で算出したハッシュ値と，1のハッシュ値を比較し，一致すれば「送信者が正しい」「データが改ざんされていない」ことを確認できる
	err = rsa.VerifyPKCS1v15(pk.(*rsa.PublicKey), crypto.SHA256, hashed, sigBytes)
	if err != nil {
		return err
	}

	return nil
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	//fmt.Println("pat Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "checkTimestamp" {
		return t.checkTimestamp(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"checkTimestamp\"")
}

// timestampの署名検証
// 本来は過去に一度使用していた場合再使用不可とするのが望ましい？が，評価用の入力を容易にするために除外
// return: OK or error
func (t *SimpleChaincode) checkTimestamp(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var timestamp string
	var timeSig string

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	timestamp = args[0]
	timeSig = args[1]

	//pk := oracleAPI.getPubKey()
	pk := "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA52trlypHnYCqRSQIwv0YFSKdVHSZWBqMcwLEB3kOX/X42hlHcbc6nNn5Tukjg3DRErHgyqtvt2Gv3vrS5VHAlJkhUrXRoGhHOzHSh6Tw8AQfdgL36RbF3kR3HZXaWDnXhjFTVdGhdXYYqPM8ex6UsAROhjDXnaD3tTFPu0uYe65q/RDTQqTm1+NeSRD8bh3C78TsUT9ybwiJyb8RD1f6/ZhM0yNvCXEjDJ5Q13tTsASnJO/OiBPfVshdjckbnio6v3zss+HO4S6/K6Un+EQvBxcCWdunAs8cv8yoR3gUQ3SQ+Fo2wRCGC9DheMazWjemhAU/q8xPtihF3y3+NSG0awIDAQAB"

	// timestamp が過去に使用されていないか確認する．
	/*if stub.GetState(timestamp) == []byte("used") {
		return shim.Error("Timestamp has already been used")
	}*/

	// 署名検証
	ret := verifySig(timestamp, pk, timeSig)
	if ret != nil {
		return shim.Error("Verification Failed")
	}

	// 一度使用された timestamp は再度使用不可とする．
	/*err = stub.PutState(timestamp, []byte("used"))
	if err != nil {
		return shim.Error("Failed to put state")
	}*/

	return shim.Success([]byte("OK"))
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
