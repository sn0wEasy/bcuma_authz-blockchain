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
	"strings"

	// "github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric-chaincode-go/shim"

	// pb "github.com/hyperledger/fabric/protos/peer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// チケットに関する情報
type TicketInfo struct {
	Phase       string
	Permissions []Permission
}

// 許可を与えるスコープの構造
type Permission struct {
	ResourceId     string
	ResourceScopes []string
	Expire         uint
}

// timestamp/timestamp.go を呼び出して timestamp を検証する
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

// client_id が登録されているか確認する
// return: flag bool
//func callClientIdFlag(stub shim.ChaincodeStubInterface, clientId string) string {
func callClientIdFlag(clientId string) bool {
	/* 別 CC などに client_id が登録されているか確認する */
	calledClientId := "client_id"
	if clientId != calledClientId {
		return false
	}
	return true
}

// claims_redirect_uri が登録されているか確認する
// return: flag bool
//func callClaimsRedirectUriFlag(stub shim.ChaincodeStubInterface, uri string) string {
func callClaimsRedirectUriFlag(uri string) bool {
	/* 別 CC などに claims_redirect_uri が登録されているか確認する */
	//calledUri := "https://client.example.com/redirect_claims"
	calledUri := "http://fl-client.ctiport.net:8888/redirect-claims" // デモ用
	if uri != calledUri {
		return false
	}
	return true
}

// ticket のフェーズを返す
// return: phase string, argsErrMsg string
func callTicketPhase(stub shim.ChaincodeStubInterface, ticket string) (string, string) {
	argBytes := [][]byte{[]byte("callTicketPhase"), []byte(ticket)}
	mycc := "perm"
	mychannel := "mychannel"
	res := stub.InvokeChaincode(mycc, argBytes, mychannel)
	var argsErrMsg string
	if res.Status != 200 {
		argsErrMsg = "Status of Invoke Chaincode is " + string(res.Status)
	}
	phase := string(res.Payload)
	return phase, argsErrMsg
}

// ticket 及びフェーズを更新し，新しい ticket と ticket_info を返す
// return: newTicket string, newTicketInfo TicketInfo, argsErrMsg string
func updateTicketAndPhase(stub shim.ChaincodeStubInterface, ticket string, newPhase string, timestamp string, timeSig string) (string, TicketInfo, string) {
	argBytes := [][]byte{[]byte("updateTicketAndPhase"), []byte(ticket), []byte(newPhase), []byte(timestamp), []byte(timeSig)}
	mycc := "perm"
	mychannel := "mychannel"
	res := stub.InvokeChaincode(mycc, argBytes, mychannel)
	var argsErrMsg string
	if res.Status != 200 {
		argsErrMsg = "Status of Invoke Chaincode is " + string(res.Status)
	}
	//fmt.Println("updateTicketAndPhase at token endpoint")
	//fmt.Println(string(res.Payload))
	resStr := string(res.Payload)
	newTicket := strings.Split(resStr, ":")[0]
	newTicketInfoStr := resStr[strings.Index(resStr, ":")+1:]
	var newTicketInfo TicketInfo
	json.Unmarshal([]byte(newTicketInfoStr), &newTicketInfo)
	return newTicket, newTicketInfo, argsErrMsg
}

// ticket を失効する
// return: argsErrMsg string
func revokeTicket(stub shim.ChaincodeStubInterface, ticket string) string {
	argBytes := [][]byte{[]byte("revokeTicket"), []byte(ticket)}
	mycc := "perm"
	mychannel := "mychannel"
	res := stub.InvokeChaincode(mycc, argBytes, mychannel)
	var argsErrMsg string
	if res.Status != 200 {
		argsErrMsg = "Status of Invoke Chaincode is " + string(res.Status)
	}
	return argsErrMsg
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	//fmt.Println("claim interactive Init")
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	//fmt.Println("claim interactive Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// claim interactive request
		return t.invoke(stub, args)
	} else if function == "invokeAuthen" {
		// ticket update request at the authentication process
		return t.invokeAuthen(stub, args)
	}
	return shim.Error("Invalid invoke function name. Expecting \"invoke\"")
}

// return: ticket(redirected) []byte
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Claim Interactive Endpoint
	// Request Body : { client_id, ticket(need_info), claims_redirect_uri, timestamp, timeSig }
	// Response Body: { ticket(redirected) }

	//fmt.Println("claim interactive request")

	var clientId, ticket, claimsRedirectUri, timestamp, timeSig string
	var argsErrMsg string

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	clientId = args[0]
	ticket = args[1]
	claimsRedirectUri = args[2]

	timestamp = args[3]
	timeSig = args[4]
	// timestamp を検証
	ret := checkTimestamp(stub, timestamp, timeSig)
	if ret != "" {
		return shim.Error(ret + ": The signature may be invalid.")
	}

	/* client_id が登録されてなければエラー */
	flagClientId := callClientIdFlag(clientId)
	if !(flagClientId) {
		return shim.Error("400 Bad Request (error: invalid client_id")
	}

	/* claims_redirect_uri が登録されてなければエラー */
	flagClaimsRedirectUri := callClaimsRedirectUriFlag(claimsRedirectUri)
	if !(flagClaimsRedirectUri) {
		return shim.Error("400 Bad Request (error: invalid claims_redirect_uri")
	}

	/* ticket のフェーズが need_info でないならエラー */
	phase, argsErrMsg := callTicketPhase(stub, ticket)
	if argsErrMsg != "" {
		return shim.Error("callTicketPhase" + argsErrMsg)
	}
	if phase != "need_info" {
		return shim.Error("400 Bad Request (error: invalid_grant)")
	}

	/* 新たな ticket を発行し，フェーズを redirected とする */
	newPhase := "redirected"
	newTicket, _, argsErrMsg := updateTicketAndPhase(stub, ticket, newPhase, timestamp, timeSig)
	if argsErrMsg != "" {
		return shim.Error("updateTicketAndPhase" + argsErrMsg)
	}
	//fmt.Println(newTicketInfo)

	/* 古い ticket のフェーズを revoked へ更新する */
	argsErrMsg = revokeTicket(stub, ticket)
	if argsErrMsg != "" {
		return shim.Error("revokeTicket" + argsErrMsg)
	}

	// claims_redirect_uri のパラメータを返す
	//fmt.Println("ticket(redirected): ", newTicket)

	// 評価用ログ出力(ticket(need_info):ticket(redirected))
	fmt.Println(ticket + ":" + newTicket)

	return shim.Success([]byte(newTicket))
}

func (t *SimpleChaincode) invokeAuthen(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Ticket Update Endpoint for Authentication Server
	// Request Body : { ticket(redirected), timestamp, timeSig }
	// Response Body: { ticket(authentication) }

	var ticket, timestamp, timeSig string
	var argsErrMsg string

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	ticket = args[0]
	timestamp = args[1]
	timeSig = args[2]

	// timestamp を検証
	ret := checkTimestamp(stub, timestamp, timeSig)
	if ret != "" {
		return shim.Error(ret + ": The signature may be invalid.")
	}

	/* ticket のフェーズが redirected でないならエラー */
	phase, argsErrMsg := callTicketPhase(stub, ticket)
	if argsErrMsg != "" {
		return shim.Error("callTicketPhase" + argsErrMsg)
	}
	if phase != "redirected" {
		return shim.Error("400 Bad Request (error: invalid_grant)")
	}

	/* 新たな ticket を発行し，フェーズを authenticated とする */
	newPhase := "authenticated"
	newTicket, _, argsErrMsg := updateTicketAndPhase(stub, ticket, newPhase, timestamp, timeSig)
	if argsErrMsg != "" {
		return shim.Error("updateTicketAndPhase" + argsErrMsg)
	}
	//fmt.Println(newTicketInfo)

	/* 古い ticket のフェーズを revoked へ更新する */
	argsErrMsg = revokeTicket(stub, ticket)
	if argsErrMsg != "" {
		return shim.Error("revokeTicket" + argsErrMsg)
	}

	return shim.Success([]byte(newTicket))
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
