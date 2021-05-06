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
	"encoding/json"
	"fmt"

	//"strconv"
	//"reflect"

	// "github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric-chaincode-go/shim"

	// pb "github.com/hyperledger/fabric/protos/peer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// resource_id に紐づく policy_condition の構造
type PolicyCondition struct {
	Issuer   string // ID token 発行者
	Subject  string // ID token の エンドユーザID
	Audience string // ID token の発行先クライアントID
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	//fmt.Println("policy Init")
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	//fmt.Println("policy Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// Create Policy
		return t.invoke(stub, args)
	} else if function == "query" {
		// Query Policy
		return t.query(stub, args)
	}
	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"query\"")
}

// ポリシーを生成
// return: nil
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Create Policy
	// Request Body : { resource_id, issuer, subject, audience }
	// Response Body: { }

	//fmt.Println("policy invoke")
	var resourceId, iss, sub, aud string
	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	resourceId = args[0]
	iss = args[1]
	sub = args[2]
	aud = args[3]

	/* ポリシーを生成 */
	var policy PolicyCondition
	policy.Issuer = iss
	policy.Subject = sub
	policy.Audience = aud

	/* ポリシーを台帳に記録 */
	policyBytes, _ := json.Marshal(policy)
	err = stub.PutState(resourceId, policyBytes)
	if err != nil {
		return shim.Error("Failed to register policy")
	}

	/* 結果を返す */
	//fmt.Println("Register policy: ", policy)

	return shim.Success([]byte("OK"))

}

// ポリシーを呼び出す
// return: policy{} []byte
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Query Policy
	// Request Body : { resource_id }
	// Response Body: { access_policy }

	fmt.Println("policy query")
	var resourceId string

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	resourceId = args[0]

	/* 台帳からポリシーを呼び出す */
	policyBytes, err := stub.GetState(resourceId)
	if err != nil {
		return shim.Error("Failed to get policy")
	}
	var policy PolicyCondition
	json.Unmarshal(policyBytes, &policy)

	/* 結果を返す */
	fmt.Println("policy called: ", policy)
	return shim.Success(policyBytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
