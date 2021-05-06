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
	"fmt"

	// "github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric-chaincode-go/shim"

	// pb "github.com/hyperledger/fabric/protos/peer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// pat.go から pat がアクティブか確認する
// return: argsErrMsg
func checkPAT(stub shim.ChaincodeStubInterface, pat string) string {
	argBytes := [][]byte{[]byte("queryactivated"), []byte(pat)}
	mycc := "pat"
	mychannel := "mychannel"
	res := stub.InvokeChaincode(mycc, argBytes, mychannel) // peer.Response型
	var argsErrMsg string
	if res.Status != 200 {
		argsErrMsg = "Status of Invoke Chaincode is " + string(res.Status)
	}
	return argsErrMsg
}

// token.go から RPT に紐づく情報を取得する
// return: rptInfoBytes []byte, argsErrMsg string
func getRptInfo(stub shim.ChaincodeStubInterface, token string) ([]byte, string) {
	argBytes := [][]byte{[]byte("query"), []byte(token)}
	mycc := "token"
	mychannel := "mychannel"
	res := stub.InvokeChaincode(mycc, argBytes, mychannel) // peer.Response型
	var argsErrMsg string
	if res.Status != 200 {
		argsErrMsg = "Status of Invoke Chaincode is " + string(res.Status)
	}
	rptInfoBytes := res.Payload
	return rptInfoBytes, argsErrMsg
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	//fmt.Println("intro Init")
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	//fmt.Println("intro Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// introspection request
		return t.invoke(stub, args)
	}
	return shim.Error("Invalid invoke function name. Expecting \"invoke\"")
}

// RPT に紐づく情報を返す
// return: Token{} []byte
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Token Introspection Endpoint
	// Authorization: Bearer <<PAT>>
	// Request Body : { token }
	// Response Body: { active, exp, permissions[ { resource_id, resource_scopes[], exp } ] }

	//fmt.Println("introspection request")

	var pat, token string
	var argsErrMsg string

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	pat = args[0]
	token = args[1]
	// ブロック番号を取得（取得方法わからないのでとりあえず7）
	//blockNum := 7

	/* pat が正しいか確認 */
	argsErrMsg = checkPAT(stub, pat)
	if argsErrMsg != "" {
		return shim.Error(argsErrMsg)
	}

	/* RPT に紐づく情報を呼び出す */
	rptInfoBytes, argsErrMsg := getRptInfo(stub, token)
	if argsErrMsg != "" {
		return shim.Error(argsErrMsg)
	}

	fmt.Println("rptInfo: ", string(rptInfoBytes))
	return shim.Success(rptInfoBytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
