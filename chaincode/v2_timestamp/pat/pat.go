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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"

	// "github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric-chaincode-go/shim"

	// pb "github.com/hyperledger/fabric/protos/peer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type Token struct {
	Active bool
	Expire uint
}

// timestamp/verify.go を呼び出して timestamp を検証する
// return: argsErrMsg string
func checkTimestamp(stub shim.ChaincodeStubInterface, msg string, sig string) string {
	argBytes := [][]byte{[]byte("checkTimestamp"), []byte(msg), []byte(sig)}
	mycc := "timestamp"
	mychannel := "mychannel"
	res := stub.InvokeChaincode(mycc, argBytes, mychannel) // peer.Response型
	var argsErrMsg string
	if res.Status != 200 {
		argsErrMsg = "Status of Invoke Chaincode is " + string(res.Status)
	}
	return argsErrMsg
}

func uint8Toint64(u [32]byte) []int64 {
	f := make([]int64, len(u))
	for n := range u {
		f[n] = int64(u[n])
	}
	return f
}

func int64Tostring(i []int64) string {
	s := ""
	for n := range i {
		s = s + strconv.FormatInt(i[n], 16)
	}
	return s
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	/*
		fmt.Println("pat Init executed")
		fmt.Println("本 CC では PAT は RO 及び RS の名前を\":\"で結合した文字列のハッシュから生成される．")
		fmt.Println("PAT の使用可否は RO が指定できるが，そのための厳密な仕組みは未実装である．署名検証などが望ましい．")
		fmt.Println("また，有効期限は未実装であるが，設定できることが望ましい．")
		fmt.Println("トークン偽装防止のために PAT 要求者の署名検証ができる仕組みも未実装であるが，設定できることが望ましい．")
	*/
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	//fmt.Println("pat Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// Make payment of X units from A to B
		return t.invoke(stub, args)
	} else if function == "revoke" {
		// Revokes an PAT from its state
		return t.revoke(stub, args)
	} else if function == "queryactivated" {
		// the old "Query" is now implemtned in invoke
		return t.queryactivated(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}

// return: pat []byte
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	/*
		:param ro string:
		:param rs string:
		:param timestamp string:
		:param timeSig string:
	*/

	//fmt.Println("pat invoke")
	var ro, rs, timestamp, timeSig string
	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	ro = args[0]
	rs = args[1]
	timestamp = args[2]
	timeSig = args[3]

	ret := checkTimestamp(stub, timestamp, timeSig)
	if ret != "" {
		return shim.Error("Failed to use timestamp. The signature may be invalid." + ret)
	}

	// Initialize the chaincode
	str := ro + ":" + rs + ":" + timestamp
	//fmt.Println("str", str)
	hashRaw := sha256.Sum256([]byte(str))
	//fmt.Println(hash)
	//fmt.Println(reflect.TypeOf(hash))
	hashInt := uint8Toint64(hashRaw)
	//fmt.Println(hashInt)
	pat := "0x" + int64Tostring(hashInt)
	//fmt.Println(pat)
	//fmt.Println(reflect.TypeOf(pat))

	//fmt.Println("new pat -> ", pat)

	// 評価用ログ出力(rs:pat)
	fmt.Println(rs + ":" + pat)

	/* PAT に紐付ける構造体を宣言 */
	var patInfo Token
	patInfo.Active = true
	intTimestamp, _ := strconv.Atoi(timestamp)
	patInfo.Expire = uint(intTimestamp + 86400)

	/* pat -> token{} を台帳に記録 */
	patInfoBytes, _ := json.Marshal(patInfo)
	err = stub.PutState(pat, patInfoBytes)
	if err != nil {
		return shim.Error("Failed to put pat")
	}

	return shim.Success([]byte(pat))
}

// return: nil
func (t *SimpleChaincode) revoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("pat revoke")
	var pat string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	pat = args[0]

	patInfoBytes, err := stub.GetState(pat)
	if err != nil {
		return shim.Error("Failed to get state")
	}

	var patInfo Token
	json.Unmarshal(patInfoBytes, &patInfo)

	patInfo.Active = false

	newPatInfoBytes, _ := json.Marshal(patInfo)
	err = stub.PutState(pat, newPatInfoBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("revoke pat -> ", pat)

	return shim.Success(nil)
}

// return: nil
func (t *SimpleChaincode) queryactivated(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	var err error

	// pat がアクティベートされているか確認

	pat := args[0]
	//fmt.Println(pat)

	patInfoBytes, err := stub.GetState(pat)
	if err != nil {
		return shim.Error("Failed to get state")
	}

	var patInfo Token
	json.Unmarshal(patInfoBytes, &patInfo)

	if patInfo.Active != true {
		fmt.Println("PAT is not activated")
		return shim.Error("PAT is not activated")
	}

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
