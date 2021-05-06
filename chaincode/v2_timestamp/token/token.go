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

// claim_token の構造
type ClaimToken struct {
	Issuer   string
	Subject  string
	Audience string
}

// resource_id に紐づく policy_condition の構造
type PolicyCondition struct {
	Issuer   string // ID token 発行者
	Subject  string // ID token の エンドユーザID
	Audience string // ID token の発行先クライアントID
}

// チケットに関する情報
type TicketInfo struct {
	Phase       string
	Permissions []Permission
}

// ケース１のエラー文構造
type NeedInfoReqClaims struct {
	Error          string
	Ticket         string
	RequiredClaims RequiredClaims
}
type RequiredClaims struct {
	ClaimTokenFormat string
	Issuer           string
}

// ケース２のエラー文構造
type NeedInfoRedirectUser struct {
	Error        string
	Ticket       string
	RedirectUser string
}

// 許可を与えるスコープの構造
type Permission struct {
	ResourceId     string
	ResourceScopes []string
	Expire         uint
}

// RPT の構造
type Token struct {
	Active      bool
	Expire      uint
	Permissions []Permission
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

// ticket_info を返す
// return: ticket_info TicketInfo, argsErrMsg string
func callTicketInfo(stub shim.ChaincodeStubInterface, ticket string) (TicketInfo, string) {
	argBytes := [][]byte{[]byte("callTicketInfo"), []byte(ticket)}
	mycc := "perm"
	mychannel := "mychannel"
	res := stub.InvokeChaincode(mycc, argBytes, mychannel)
	var argsErrMsg string
	if res.Status != 200 {
		argsErrMsg = "Status of Invoke Chaincode is " + string(res.Status)
	}
	var ticketInfo TicketInfo
	json.Unmarshal(res.Payload, &ticketInfo)
	return ticketInfo, argsErrMsg
}

// ticket に紐づく resource_id[] を返す
// return: resource_id[] []string, argsErrMsg string
func callResourceId(stub shim.ChaincodeStubInterface, ticket string) ([]string, string) {
	argBytes := [][]byte{[]byte("callResourceId"), []byte(ticket)}
	mycc := "perm"
	mychannel := "mychannel"
	res := stub.InvokeChaincode(mycc, argBytes, mychannel)
	var argsErrMsg string
	if res.Status != 200 {
		argsErrMsg = "Status of Invoke Chaincode is " + string(res.Status)
	}
	resourceIdStr := string(res.Payload)
	resslice := strings.Split(resourceIdStr, ":")
	return resslice, argsErrMsg
}

// resource_id に紐づく access_claim を返す
// return: policy PolicyCondition, argsErrMsg string
func callAccessPolicy(stub shim.ChaincodeStubInterface, resourceId string) (PolicyCondition, string) {
	argBytes := [][]byte{[]byte("query"), []byte(resourceId)}
	mycc := "policy"
	mychannel := "mychannel"
	res := stub.InvokeChaincode(mycc, argBytes, mychannel)
	var argsErrMsg string
	if res.Status != 200 {
		argsErrMsg = "Status of Invoke Chaincode is " + string(res.Status)
	}
	var policy PolicyCondition
	json.Unmarshal(res.Payload, &policy)
	//fmt.Println("called policy: ", policy)
	return policy, argsErrMsg
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

// ケース１
// return: _retErrBytes
func ReqClaims(stub shim.ChaincodeStubInterface, args []string) ([]byte, string) {
	// :param ticket:
	// :param newPhase:
	// :param claimTokenFormat:
	// :param issuer:
	// :param timestamp:
	// :param timeSig:

	ticket := args[0]
	newPhase := args[1]
	claimTokenFormat := args[2]
	issuer := args[3]

	timestamp := args[4]
	timeSig := args[5]

	if len(args) != 6 {
		return []byte(""), "Incorrect number of arguments. Expecting 6"
	}

	newTicket, newTicketInfo, argsErrMsg := updateTicketAndPhase(stub, ticket, newPhase, timestamp, timeSig)
	if argsErrMsg != "" {
		return []byte(""), argsErrMsg
	}
	// エラー内容を生成
	var requiredClaims RequiredClaims
	requiredClaims.ClaimTokenFormat = claimTokenFormat
	requiredClaims.Issuer = issuer
	var _retErr NeedInfoReqClaims
	_retErr.Error = newTicketInfo.Phase
	_retErr.Ticket = newTicket
	_retErr.RequiredClaims = requiredClaims
	_retErrBytes, _ := json.Marshal(_retErr)
	return _retErrBytes, ""
}

// ケース２
// return: _retErrBytes
func RedirectUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, string) {
	// :param ticket:
	// :param newPhase:
	// :param redirectUser:
	// :param timestamp:
	// :param timestamp:

	ticket := args[0]
	newPhase := args[1]
	redirectUser := args[2]

	timestamp := args[3]
	timeSig := args[4]

	if len(args) != 5 {
		return []byte(""), "Incorrect number of arguments. Expecting 5"
	}

	newTicket, newTicketInfo, argsErrMsg := updateTicketAndPhase(stub, ticket, newPhase, timestamp, timeSig)
	if argsErrMsg != "" {
		return []byte(""), argsErrMsg
	}
	// エラー内容を生成
	var _retErr NeedInfoRedirectUser
	_retErr.Error = newTicketInfo.Phase
	_retErr.Ticket = newTicket
	_retErr.RedirectUser = redirectUser
	_retErrBytes, _ := json.Marshal(_retErr)
	return _retErrBytes, ""
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	//fmt.Println("token Init")
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	//fmt.Println("token Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// rpt issue request
		return t.invoke(stub, args)
	} else if function == "query" {
		// rpt info request
		return t.query(stub, args)
	}
	return shim.Error("Invalid invoke function name. Expecting \"invoke\"")
}

// return: RPT []byte
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Token Endpoint
	// Authorization: Basic(?)
	// Request Body : { grant_type, ticket, claim_token, claim_token_format, timestamp, timeSig }
	// Response Body: { access_token, token_type } or { error: need_info, ticket, required_claims[] or redirect_user }
	// ticket を確認して正しければアクセストークンを，情報が足りていなければ ticket を更新してエラー文を返す

	//fmt.Println("token Request")

	var grantType, ticket, claimTokenStr, claimTokenFormat, timestamp, timeSig string
	var err error
	var argsErrMsg string

	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}

	grantType = args[0]
	ticket = args[1]
	claimTokenStr = args[2]
	claimTokenFormat = args[3]

	timestamp = args[4]
	timeSig = args[5]
	// timestamp を検証
	ret := checkTimestamp(stub, timestamp, timeSig)
	if ret != "" {
		return shim.Error(ret + ": The signature may be invalid.")
	}

	// ケースを固定（1 or 2）
	bcumaCase := 2
	// grant_type を固定
	bcumaGrantType := "urn:ietf:params:oauth:grant-type:uma-ticket"
	// claim_token_format を固定
	bcumaClaimTokenFormat := "http://openid.net/specs/openid-connect-core-1_0.html#IDToken"
	// issuer を固定
	//bcumaIssuer := "https://example.com/idp"
	bcumaIssuer := "http://authz-blockchain.ctiport.net:8888/authen" // デモ用 for Interactive Claim Gathering
	//bcumaIssuer := "http://authz-blockchain.ctiport.net:8888/idp" // デモ用 for OpenID Connect
	// claims_redirect_uri を固定
	//bcumaRedirectUser := "https://as.example.com/redirect_claims"
	bcumaRedirectUser := "http://authz-blockchain.ctiport.net:8888/rqp-claims" // デモ用

	// claim_token がない場合
	if claimTokenStr == "" {
		if bcumaCase == 1 {
			// ケース１
			//fmt.Println("403 Forbidden(error: need_info)")
			newPhase := "need_info"
			_args := []string{ticket, newPhase, bcumaClaimTokenFormat, bcumaIssuer}
			_retErrBytes, argsErrMsg := ReqClaims(stub, _args)
			if argsErrMsg != "" {
				return shim.Error("ReqClaims: " + argsErrMsg)
			}
			// エラーを返す
			//fmt.Println(string(_retErrBytes))

			// 評価用ログ出力(ticket(init):ticket(need_info))
			var _retErr NeedInfoReqClaims
			json.Unmarshal(_retErrBytes, &_retErr)
			fmt.Println(ticket + ":" + _retErr.Ticket)

			return shim.Success(_retErrBytes)
		} else if bcumaCase == 2 {
			// ケース２
			//fmt.Println("403 Forbidden(error: need_info)")
			newPhase := "need_info"
			_args := []string{ticket, newPhase, bcumaRedirectUser, timestamp, timeSig}
			_retErrBytes, argsErrMsg := RedirectUser(stub, _args)
			if argsErrMsg != "" {
				return shim.Error("RedirectUser: " + argsErrMsg)
			}
			// エラーを返す
			//fmt.Println(string(_retErrBytes))

			// 評価用ログ出力(ticket(need_info))
			var _retErr NeedInfoRedirectUser
			json.Unmarshal(_retErrBytes, &_retErr)
			fmt.Println(ticket + ":" + _retErr.Ticket)

			return shim.Success(_retErrBytes)
		} else {
			return shim.Error("error: invalid bcumaCase")
		}
	}

	// claimTokenStr から claim_token を構成
	claimTokenStr = strings.TrimLeft(claimTokenStr, "{")
	claimTokenStr = strings.TrimRight(claimTokenStr, "}")
	_claim := strings.Split(claimTokenStr, ",")

	var claimToken ClaimToken
	claimToken.Issuer = _claim[0][strings.Index(_claim[0], ":")+1:]
	claimToken.Subject = _claim[1][strings.Index(_claim[1], ":")+1:]
	claimToken.Audience = _claim[2][strings.Index(_claim[2], ":")+1:]
	//fmt.Println("claim_token iss: ", claimToken.Issuer)

	/* grant_type が指定のものと異なればエラーを返す */
	if grantType != bcumaGrantType {
		return shim.Error("400 Bad Request (invalid_grant)")
	}

	/* claim_token_format がなければエラーを返す */
	if claimTokenFormat == "" {
		return shim.Error("403 Forbidden (error: Invalid claim_token_format)")
	}

	/* claim_token を評価 */
	// claim_token_format によって評価方法を変える．今回は指定したフォーマット以外はエラーを返す仕様にする
	if claimTokenFormat != bcumaClaimTokenFormat {
		return shim.Error("Invalid claim_token_format")
	}
	// ticket に紐づく permissions を呼び出す
	var ticketInfo TicketInfo
	ticketInfo, argsErrMsg = callTicketInfo(stub, ticket)
	if argsErrMsg != "" {
		return shim.Error("callTicketInfo: " + argsErrMsg)
	}
	perms := ticketInfo.Permissions
	//fmt.Println("perms: ", perms)
	// permission を全探索で評価
	for i := 0; i < len(perms); i++ {
		// resource_id に紐づくポリシーを呼び出す
		_id := perms[i].ResourceId
		policy, argsErrMsg := callAccessPolicy(stub, _id)
		fmt.Println("policy_iss: ", policy.Issuer)
		fmt.Println("claim_iss: ", claimToken.Issuer)
		fmt.Println("policy_sub: ", policy.Subject)
		fmt.Println("claim_sub: ", claimToken.Subject)
		fmt.Println("claim_aud: ", claimToken.Audience)
		fmt.Println("policy_aud: ", policy.Audience)
		if argsErrMsg != "" {
			return shim.Error("callAccessPolicy: " + argsErrMsg)
		}
		// claim_token がポリシーで指定されるものと一致するか確認する
		var _issBool, _subBool, _audBool bool = true, true, true
		if claimToken.Issuer != policy.Issuer {
			_issBool = false
		}
		if claimToken.Subject != policy.Subject {
			_subBool = false
		}
		if claimToken.Audience != policy.Audience {
			_audBool = false
		}
		// 一致しなければケースに応じてエラー処理
		if !(_issBool && _subBool && _audBool) {
			if bcumaCase == 1 {
				// ケース１
				//fmt.Println("403 Forbidden(error: need_info)")
				newPhase := "need_info"
				_args := []string{ticket, newPhase, bcumaClaimTokenFormat, bcumaIssuer, timestamp, timeSig}
				_retErrBytes, argsErrMsg := ReqClaims(stub, _args)
				if argsErrMsg != "" {
					return shim.Error("ReqClaims: " + argsErrMsg)
				}
				// エラーを返す
				//fmt.Println(string(_retErrBytes))

				// 評価用ログ出力(ticket(init):ticket(need_info))
				var _retErr NeedInfoReqClaims
				json.Unmarshal(_retErrBytes, &_retErr)
				fmt.Println(ticket + ":" + _retErr.Ticket)

				return shim.Success(_retErrBytes)
			} else {
				// ケース２
				//fmt.Println("403 Forbidden(error: need_info)")
				newPhase := "need_info"
				_args := []string{ticket, newPhase, bcumaRedirectUser, timestamp, timeSig}
				_retErrBytes, argsErrMsg := RedirectUser(stub, _args)
				if argsErrMsg != "" {
					return shim.Error("RedirectUser: " + argsErrMsg)
				}
				// エラーを返す
				//fmt.Println(string(_retErrBytes))

				// 評価用ログ出力(ticket(need_info))
				var _retErr NeedInfoRedirectUser
				json.Unmarshal(_retErrBytes, &_retErr)
				fmt.Println(ticket + ":" + _retErr.Ticket)

				return shim.Success(_retErrBytes)
			}
		}
	}

	/* RPTを生成する */
	nsDNS := uuid.NameSpaceDNS
	seed := ticket + timestamp
	rpt := uuid.NewSHA1(nsDNS, []byte(seed)).String()
	//fmt.Println("Requesting Party Token: ", rpt)

	// 評価用ログ出力(ticket(redirected):token)
	fmt.Println(ticket + ":" + rpt)

	/* RPT に紐付ける構造体を宣言 */
	var rptInfo Token
	rptInfo.Active = true
	intTimestamp, _ := strconv.Atoi(timestamp)
	rptInfo.Expire = uint(intTimestamp + 86400)
	rptInfo.Permissions = perms

	/* rpt -> token{} を台帳に記録 */
	rptInfoBytes, _ := json.Marshal(rptInfo)
	err = stub.PutState(rpt, rptInfoBytes)
	if err != nil {
		return shim.Error("Failed to put rpt")
	}

	/* ticket のフェーズを "revoked" へ更新する */
	argsErrMsg = revokeTicket(stub, ticket)
	if argsErrMsg != "" {
		return shim.Error("Failed to revoke ticket")
	}

	/* RPTを発行する */
	return shim.Success([]byte(rpt))
}

// return: RPT{} []byte
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("token query")

	var token string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	token = args[0]

	// ケース１の場合
	// idp が同ドメインの場合，ticket(need_info) はIdPドメインで失効
	// idp が別ドメインの場合，ticket(need_info) はそもそも発行しない？

	rptInfoBytes, err := stub.GetState(token)
	if err != nil {
		return shim.Error("Failed to get rpt info")
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
