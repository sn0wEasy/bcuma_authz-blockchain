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
	"strconv"
	"strings"

	"github.com/google/uuid"

	// "github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric-chaincode-go/shim"

	// pb "github.com/hyperledger/fabric/protos/peer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// PATに紐づくリソースIDの構造
type ResourceID struct {
	Id []string
}

// リソースIDに紐づくリソース記述の構造
type ResourceDescription struct {
	ResourceScopes []string
	Description    string
	IconUri        string
	Name           string
	Type           string
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

// pat.go から pat がアクティブか確認する
// return: nil
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

// rreg.go から resource_id のリストを呼び出す
// return: resource_id[] []string, argsErrMsg string
func getIdList(stub shim.ChaincodeStubInterface, pat string) ([]string, string) {
	// return [id1, id2, ...], argsErrMsg
	argBytes := [][]byte{[]byte("list"), []byte(pat)}
	mycc := "rreg"
	mychannel := "mychannel"
	res := stub.InvokeChaincode(mycc, argBytes, mychannel) // peer.Response型
	var argsErrMsg string
	if res.Status != 200 {
		argsErrMsg = "Status of Invoke Chaincode is " + string(res.Status)
	}
	resslice := strings.Split(string(res.Payload), ":")
	// fmt.Println("res.Payload: ", string(res.Payload))
	return resslice, argsErrMsg
}

// rreg.go から resource_description を呼び出す
// return: resource_description{} ResourceDescription, argsErrMsg string
func queryDescription(stub shim.ChaincodeStubInterface, pat string, resourceId string) (ResourceDescription, string) {
	// return ResourceDescription, argsErrMsg
	argBytes := [][]byte{[]byte("query"), []byte(pat), []byte(resourceId)}
	mycc := "rreg"
	mychannel := "mychannel"
	res := stub.InvokeChaincode(mycc, argBytes, mychannel) // peer.Response型
	var argsErrMsg string
	if res.Status != 200 {
		argsErrMsg = "Status of Invoke Chaincode is " + string(res.Status)
	}
	resBytes := res.Payload
	var ret ResourceDescription
	json.Unmarshal(resBytes, &ret)
	return ret, argsErrMsg
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	//fmt.Println("perm Init")
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	//fmt.Println("perm Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// permission request
		return t.invoke(stub, args)
	} else if function == "callTicketPhase" {
		return t.callTicketPhase(stub, args)
	} else if function == "callTicketInfo" {
		return t.callTicketInfo(stub, args)
	} else if function == "callResourceId" {
		return t.callResourceId(stub, args)
	} else if function == "updateTicketAndPhase" {
		return t.updateTicketAndPhase(stub, args)
	} else if function == "revokeTicket" {
		return t.revokeTicket(stub, args)
	}
	return shim.Error("Invalid invoke function name. Expecting \"invoke\"")
}

// パーミッションを確認し，スコープが正しければチケットを発行
// return: ticket []byte
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Permission Endpoint
	// Authorization: Bearer <<PAT>>
	// Request Body : { permissions[pat, resource_id, resource_scopes], timestamp, timeSig }
	// Response Body: { ticket }
	// 要求された resource_scopes[] のうち登録されているものを確認してチケットを返す
	// チケットは別のエンドポイントで確認できるように状態を記録する
	// ticket の流れ： perm(init) -> token(need_info) -> claim(redirected)

	//fmt.Println("Permission Request")

	var perms, pat, resourceId, timestamp, timeSig string
	var resourceScopes []string
	var err error
	var argsErrMsg string

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	pat = args[0]

	// perms: {{resource_id,\"scope01:scope02...\"},{...},...}
	perms = args[1]
	perms = strings.TrimLeft(perms, "{{")
	perms = strings.TrimRight(perms, "}}")
	slice := strings.Split(perms, "},{")

	timestamp = args[2]
	timeSig = args[3]
	// timestamp を検証
	ret := checkTimestamp(stub, timestamp, timeSig)
	if ret != "" {
		return shim.Error(ret + ": The signature may be invalid.")
	}

	// pat が正しいか確認
	argsErrMsg = checkPAT(stub, pat)
	if argsErrMsg != "" {
		return shim.Error(argsErrMsg)
	}

	// ticket_info を初期化
	var ticketInfo TicketInfo
	ticketInfo.Phase = "init"

	for k := 0; k < len(slice); k++ {
		perm := strings.Split(slice[k], ",")
		resourceId = perm[0]
		_scopes := perm[1]
		_scopes = strings.TrimLeft(_scopes, "\"")
		_scopes = strings.TrimRight(_scopes, "\"")
		resourceScopes = strings.Split(_scopes, ":")

		// pat に紐づく resource_id を呼び出す
		var idlist ResourceID
		idlist.Id, argsErrMsg = getIdList(stub, pat)
		if argsErrMsg != "" {
			return shim.Error(argsErrMsg)
		}
		// 台帳に id が１件も登録されていなければエラー
		//fmt.Println("idlist: ", idlist.Id)
		if len(idlist.Id) == 0 {
			return shim.Error("invalid_resource_id: At least one of the provided resource identifiers was not found at the blockhchain")
		}

		// resource_id に紐づく resource_description{} を台帳から呼び出し(rreg.query)
		var stateResourceScopes ResourceDescription
		stateResourceScopes, argsErrMsg = queryDescription(stub, pat, resourceId)
		if argsErrMsg != "" {
			return shim.Error(argsErrMsg)
		}

		// リクエストの resource_scopes[] と台帳に登録される resource_scopes[] を比較
		var validScopes []string
		for i := 0; i < len(resourceScopes); i++ {
			for j := 0; j < len(stateResourceScopes.ResourceScopes); j++ {
				if resourceScopes[i] == stateResourceScopes.ResourceScopes[j] {
					validScopes = append(validScopes, resourceScopes[i])
				}
			}
		}
		//fmt.Println("validScopes: ", validScopes)

		// 一致する resource_scopes[] がなければエラー
		if len(validScopes) == 0 {
			return shim.Error("invalid_scope: At least one of the scopes included in the request was not registered previously by this resource server for the referenced resource")
		}

		// permission を生成
		var p Permission
		p.ResourceId = resourceId
		p.ResourceScopes = validScopes
		intTimestamp, _ := strconv.Atoi(timestamp)
		p.Expire = uint(intTimestamp + 86400)

		// ticket_info にパーミッションを追加
		ticketInfo.Permissions = append(ticketInfo.Permissions, p)
	}
	fmt.Println("ticketInfo: ", ticketInfo)

	// ticket を生成
	nsDNS := uuid.NameSpaceDNS
	seed := pat + timestamp
	ticket := uuid.NewSHA1(nsDNS, []byte(seed)).String()
	//fmt.Println("ticket: ", ticket)

	// 評価用ログ出力（pat:ticket(init)）
	fmt.Println(pat + ":" + ticket)

	// ticket -> ticket_info{ phase, permissions[] } を台帳に記録
	ticketInfoBytes, _ := json.Marshal(ticketInfo)
	err = stub.PutState(ticket, ticketInfoBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	//fmt.Println("Permission success: ")
	//fmt.Println(ticket + " -> " + string(ticketInfoBytes))

	// ticket を返す
	return shim.Success([]byte(ticket))
}

// ticket がどの処理段階を示すか呼び出す
// return: phase []byte
func (t *SimpleChaincode) callTicketPhase(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var ticket string
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	ticket = args[0]
	ticketInfoBytes, err := stub.GetState(ticket)
	var ticketInfo TicketInfo
	json.Unmarshal(ticketInfoBytes, &ticketInfo)
	fmt.Println(ticketInfo)

	if err != nil {
		return shim.Error("Failed to get state")
	}

	phase := ticketInfo.Phase
	if phase == "init" {
		fmt.Println("init")
		return shim.Success([]byte("init"))
	} else if phase == "need_info" {
		fmt.Println("need_info")
		return shim.Success([]byte("need_info"))
	} else if phase == "redirected" {
		fmt.Println("redirected")
		return shim.Success([]byte("redirected"))
	} else if phase == "authenticated" {
		fmt.Println("authenticated")
		return shim.Success([]byte("authenticated"))
	} else if phase == "revoked" {
		fmt.Println("revoked")
		return shim.Success([]byte("revoked"))
	}

	return shim.Error("Failed to get ticket info")
}

// ticket_info を呼び出す
// return: ticket_info []byte
func (t *SimpleChaincode) callTicketInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var ticket string
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	ticket = args[0]
	ticketInfoBytes, err := stub.GetState(ticket)

	if err != nil {
		return shim.Error("Failed to get state")
	}

	return shim.Success(ticketInfoBytes)
}

// ticket に紐づく resource_id[] を呼び出す
// return: resource_id[0]:resource_id[1]:... []byte
func (t *SimpleChaincode) callResourceId(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var ticket string
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	ticket = args[0]
	ticketInfoBytes, err := stub.GetState(ticket)
	var ticketInfo TicketInfo
	json.Unmarshal(ticketInfoBytes, &ticketInfo)
	fmt.Println(ticketInfo)

	if err != nil {
		return shim.Error("Failed to get state")
	}

	var resourceId string
	for i := 0; i < len(ticketInfo.Permissions); i++ {
		resourceId = resourceId + ":" + ticketInfo.Permissions[i].ResourceId
	}
	resourceId = strings.TrimLeft(resourceId, ":")
	fmt.Println("Resource id called: ", resourceId)
	return shim.Success([]byte(resourceId))
}

// ticket 及び処理フェーズを更新し，古い ticket を失効する
// return: ticket:ticket_info string
func (t *SimpleChaincode) updateTicketAndPhase(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var ticket, newPhase, timestamp, timeSig string
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	ticket = args[0]
	newPhase = args[1]
	ticketInfoBytes, err := stub.GetState(ticket)
	var ticketInfo TicketInfo
	json.Unmarshal(ticketInfoBytes, &ticketInfo)
	fmt.Println(ticketInfo)

	timestamp = args[2]
	timeSig = args[3]
	// timestamp を検証
	argsErrMsg := checkTimestamp(stub, timestamp, timeSig)
	if argsErrMsg != "" {
		return shim.Error(argsErrMsg + ": The signature may be invalid.")
	}

	if err != nil {
		return shim.Error("Failed to get ticket info")
	}

	// 現在の ticket を失効
	ticketInfo.Phase = "revoked"
	oldTicketInfoBytes, _ := json.Marshal(ticketInfo)
	err = stub.PutState(ticket, oldTicketInfoBytes)
	if err != nil {
		return shim.Error("Failed to put ticket info")
	}
	// ticket を更新
	var uuidObj uuid.UUID
	nsDNS := uuid.NameSpaceDNS
	seed := ticket + timestamp
	uuidObj = uuid.NewSHA1(nsDNS, []byte(seed))
	newTicket := uuidObj.String()
	// フェーズを更新
	ticketInfo.Phase = newPhase
	newTicketInfoBytes, _ := json.Marshal(ticketInfo)
	err = stub.PutState(newTicket, newTicketInfoBytes)
	if err != nil {
		return shim.Error("Failed to put ticket info")
	}

	fmt.Println(newTicket + " -> " + string(newTicketInfoBytes))
	ret := []byte(string(newTicket + ":" + string(newTicketInfoBytes)))
	return shim.Success(ret)

	if ticketInfo.Phase == "revoked" {
		fmt.Println("ticket is already revoked")
		return shim.Error("ticket is already revoked")
	}

	return shim.Error("Update Error: Invalid ticket phase.")
}

// ticket を失効する
// return: nil
func (t *SimpleChaincode) revokeTicket(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var ticket string
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	ticket = args[0]
	ticketInfoBytes, err := stub.GetState(ticket)
	var ticketInfo TicketInfo
	json.Unmarshal(ticketInfoBytes, &ticketInfo)
	fmt.Println(ticketInfo)

	if err != nil {
		return shim.Error("Failed to get ticket info")
	}

	if ticketInfo.Phase == "revoked" {
		fmt.Println("ticket is already revoked")
		return shim.Error("ticket is already revoked")
	}

	ticketInfo.Phase = "revoked"
	newTicketInfoBytes, _ := json.Marshal(ticketInfo)
	err = stub.PutState(ticket, newTicketInfoBytes)
	if err != nil {
		return shim.Error("Failed to put ticket info")
	}

	fmt.Println("Ticket revoked")
	fmt.Println(ticket + " -> " + string(newTicketInfoBytes))
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
