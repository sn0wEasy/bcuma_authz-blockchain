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

	"github.com/google/uuid"

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

// PATに紐づくリソースIDの構造
type ResourceID struct {
	Id []string
}

// resource_id に紐づく resource_description の構造
type ResourceDescription struct {
	ResourceScopes []string
	Description    string
	IconUri        string
	Name           string
	Type           string
}

// resource_id に紐づく policy_condition の構造
type PolicyCondition struct {
	Issuer   string // ID token 発行者
	Subject  string // ID token の エンドユーザID
	Audience string // ID token の発行先クライアントID
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
		argsErrMsg = "checkTimestamp: Status of Invoke Chaincode is " + string(res.Status)
	}
	return argsErrMsg
}

// pat.go を呼び出して pat がアクティブか確認する
// return: argsErrMsg string
func checkPAT(stub shim.ChaincodeStubInterface, pat string) string {
	argBytes := [][]byte{[]byte("queryactivated"), []byte(pat)}
	mycc := "pat"
	mychannel := "mychannel"
	res := stub.InvokeChaincode(mycc, argBytes, mychannel) // peer.Response型
	var argsErrMsg string
	if res.Status != 200 {
		argsErrMsg = "checkPAT: Status of Invoke Chaincode is " + string(res.Status)
	}
	return argsErrMsg
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	//fmt.Println("Resource Registration Init")
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	//fmt.Println("rreg Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// Create Endpoint
		return t.invoke(stub, args)
	} else if function == "list" {
		// List Endpoint
		return t.list(stub, args)
	} else if function == "query" {
		// Read Endpoint
		return t.query(stub, args)
	} else if function == "update" {
		// Update Endpoint
		return t.update(stub, args)
	} else if function == "delete" {
		// Delete Endpoint
		return t.delete(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}

// Transaction makes a resource description
// return: resource_id []byte
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Create Endpoint
	// Authorization: Bearer <<PAT>>
	// Request Body : { resource_scopes[], description, icon_uri, name, type, timestamp, timeSig }
	// Response Body: { resource_id, (user_access_policy_uri) }

	// リソース登録時に必要なパラメータ．スコープは","で区切って複数登録可能

	//fmt.Println("Resource Registration invoke")
	var pat, resourceScopes, description, iconUri, name, _type, timestamp, timeSig string

	if len(args) != 8 {
		return shim.Error("Incorrect number of arguments. Expecting 8")
	}

	pat = args[0]
	resourceScopes = args[1]
	slice := strings.Split(resourceScopes, ":")
	description = args[2]
	iconUri = args[3]
	name = args[4]
	_type = args[5]

	timestamp = args[6]
	timeSig = args[7]
	// timestamp を検証
	ret := checkTimestamp(stub, timestamp, timeSig)
	if ret != "" {
		return shim.Error(ret + ": The signature may be invalid.")
	}

	// PAT が正しいか確認
	argsErrMsg := checkPAT(stub, pat)
	if argsErrMsg != "" {
		return shim.Error(argsErrMsg)
	}

	// resource_id を生成
	// resource_id 初回登録であれば，pat をシードとして uuid を生成
	// resource_id を一度でも登録していれば，前回登録した resource_id をシードとしてuuidを生成
	idpat := "id_" + pat // resource_id 登録用の key 文字列
	_idlistBytes, err := stub.GetState(idpat)
	var idlist ResourceID
	json.Unmarshal(_idlistBytes, &idlist)
	var uuidObj uuid.UUID
	nsDNS := uuid.NameSpaceDNS
	var seed string
	if len(idlist.Id) == 0 {
		seed = pat + timestamp
		uuidObj = uuid.NewSHA1(nsDNS, []byte(seed))
	} else {
		seed = idlist.Id[len(idlist.Id)-1] + timestamp
		uuidObj = uuid.NewSHA1(nsDNS, []byte(seed))
	}

	// 新規 resource_id を追加
	idlist.Id = append(idlist.Id, uuidObj.String())
	idlistBytes, _ := json.Marshal(idlist)

	// pat -> resource_id[] を台帳に記録
	err = stub.PutState(idpat, idlistBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	//fmt.Println("Resource registration success:")
	//fmt.Println(idpat + " -> " + string(idlistBytes))

	// resource_description を格納する構造体を作成
	info := ResourceDescription{ResourceScopes: slice,
		Description: description,
		IconUri:     iconUri,
		Name:        name,
		Type:        _type}

	// JSON文字列をバイト列に変換
	infoBytes, _ := json.Marshal(info)

	// resource_id -> resource_description{} を台帳に記録
	err = stub.PutState(uuidObj.String(), infoBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	//fmt.Println("Success:")
	//fmt.Println(uuidObj.String() + " -> " + string(infoBytes))

	// resource_id を表示
	//fmt.Println("New Resource ID -> ", uuidObj.String())

	// 評価用ログ出力(pat:resource_id)
	fmt.Println(pat + ":" + uuidObj.String())

	return shim.Success([]byte(uuidObj.String()))
}

// return: resource_id[0]:resource_id[1]:... []byte
func (t *SimpleChaincode) list(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// List Endpoint
	// Authorization: Bearer <<PAT>>
	// Request Body : None
	// Response Body: { <<resource id list>> }

	fmt.Println("Resource Registration list")

	var pat string

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	pat = args[0]

	// PAT が正しいか確認
	argsErrMsg := checkPAT(stub, pat)
	if argsErrMsg != "" {
		return shim.Error(argsErrMsg)
	}

	// list 処理
	idpat := "id_" + pat
	newidlistBytes, err := stub.GetState(idpat)
	if err != nil {
		return shim.Error("Failed to list resource_id")
	}
	var newidlist ResourceID
	json.Unmarshal(newidlistBytes, &newidlist)
	fmt.Println(newidlist.Id)

	var ret string
	for i := 0; i < len(newidlist.Id); i++ {
		ret = ret + ":" + newidlist.Id[i]
	}
	ret = strings.TrimLeft(ret, ":")

	return shim.Success([]byte(ret))

}

// query callback representing the query of a chaincode
// return: resource_description{} []byte
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Read Endpoint
	// Authorization: Bearer <<PAT>>
	// Request Body : { resource_id }
	// Response Body: { resource_id, resource_scopes[], icon_uri, name, type }

	var pat, resourceId string

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	pat = args[0]
	resourceId = args[1]

	// PAT が正しいか確認
	argsErrMsg := checkPAT(stub, pat)
	if argsErrMsg != "" {
		return shim.Error(argsErrMsg)
	}

	// リソースが登録されていない時
	idpat := "id_" + pat // resource_id 登録の key 文字列
	idlistBytes, err := stub.GetState(idpat)
	var idlist ResourceID
	json.Unmarshal(idlistBytes, &idlist)
	if len(idlist.Id) == 0 {
		return shim.Error("No resource registered")
	}

	// リソースバイト列を台帳から呼び出し
	newinfoBytes, err := stub.GetState(resourceId)
	if err != nil {
		return shim.Error("Failed to query state")
	}

	return shim.Success(newinfoBytes)
}

// Update Endpoint
// return: resource_description{} []byte
func (t *SimpleChaincode) update(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Update Endpoint
	// Authorization: Bearer <<PAT>>
	// Request Body : { resource_id, resource_scopes[], description, icon_uri, name, type }
	// Response Body: { resource_id }

	fmt.Println("Resource Description update")

	var pat, resourceId, resourceScopes, description, iconUri, name, _type string

	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}

	pat = args[0]
	resourceId = args[1]
	resourceScopes = args[2]
	slice := strings.Split(resourceScopes, ",")
	description = args[3]
	iconUri = args[4]
	name = args[5]
	_type = args[6]

	// PAT が正しいか確認
	argsErrMsg := checkPAT(stub, pat)
	if argsErrMsg != "" {
		return shim.Error(argsErrMsg)
	}

	// resource_description を格納する構造体を初期化
	newDescript := ResourceDescription{ResourceScopes: slice,
		Description: description,
		IconUri:     iconUri,
		Name:        name,
		Type:        _type}

	newDescriptBytes, _ := json.Marshal(newDescript)

	// resource_id に紐づく resource_description を呼び出す
	ret, err := stub.GetState(resourceId)
	if len(ret) == 0 {
		fmt.Println("No resource registered")
		return shim.Error("Failed to update")
	}

	// resource_id -> new resource_description{} を台帳に記録
	err = stub.PutState(resourceId, newDescriptBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("Success:")
	fmt.Println(resourceId + " -> " + string(newDescriptBytes))

	return shim.Success([]byte(newDescriptBytes))
}

// Deletes an entity from state
// return: nil
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Delete Endpoint
	// Authorization: Bearer <<PAT>>
	// Request Body : resource_id
	// Response Body: None

	var pat, resourceId string

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	pat = args[0]
	resourceId = args[1]

	// PAT が正しいか確認
	argsErrMsg := checkPAT(stub, pat)
	if argsErrMsg != "" {
		return shim.Error(argsErrMsg)
	}

	// Delete the key from the state in ledger
	err := stub.DelState(resourceId)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	fmt.Println("Delete success")

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
