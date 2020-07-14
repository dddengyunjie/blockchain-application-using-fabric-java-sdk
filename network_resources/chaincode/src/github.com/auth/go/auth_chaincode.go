/*
 SPDX-License-Identifier: Apache-2.0
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// ==== Invoke marbles ====
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble1","blue","35","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble2","red","50","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble3","blue","70","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["transferMarble","marble2","jerry"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["transferMarblesBasedOnColor","blue","jerry"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["delete","marble1"]}'

// ==== Query marbles ====
// peer chaincode query -C myc1 -n marbles -c '{"Args":["readMarble","marble1"]}'
// peer chaincode query -C myc1 -n marbles -c '{"Args":["getMarblesByRange","marble1","marble3"]}'
// peer chaincode query -C myc1 -n marbles -c '{"Args":["getHistoryForMarble","marble1"]}'

// Rich Query (Only supported if CouchDB is used as state database):
// peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarblesByOwner","tom"]}'
// peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"owner\":\"tom\"}}"]}'

// Rich Query with Pagination (Only supported if CouchDB is used as state database):
// peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarblesWithPagination","{\"selector\":{\"owner\":\"tom\"}}","3",""]}'

// INDEXES TO SUPPORT COUCHDB RICH QUERIES
//
// Indexes in CouchDB are required in order to make JSON queries efficient and are required for
// any JSON query with a sort. As of Hyperledger Fabric 1.1, indexes may be packaged alongside
// chaincode in a META-INF/statedb/couchdb/indexes directory. Each index must be defined in its own
// text file with extension *.json with the index definition formatted in JSON following the
// CouchDB index JSON syntax as documented at:
// http://docs.couchdb.org/en/2.1.1/api/database/find.html#db-index
//
// This marbles02 example chaincode demonstrates a packaged
// index which you can find in META-INF/statedb/couchdb/indexes/indexOwner.json.
// For deployment of chaincode to production environments, it is recommended
// to define any indexes alongside chaincode so that the chaincode and supporting indexes
// are deployed automatically as a unit, once the chaincode has been installed on a peer and
// instantiated on a channel. See Hyperledger Fabric documentation for more details.
//
// If you have access to the your peer's CouchDB state database in a development environment,
// you may want to iteratively test various indexes in support of your chaincode queries.  You
// can use the CouchDB Fauxton interface or a command line curl utility to create and update
// indexes. Then once you finalize an index, include the index definition alongside your
// chaincode in the META-INF/statedb/couchdb/indexes directory, for packaging and deployment
// to managed environments.
//
// In the examples below you can find index definitions that support marbles02
// chaincode queries, along with the syntax that you can use in development environments
// to create the indexes in the CouchDB Fauxton interface or a curl command line utility.
//

//Example hostname:port configurations to access CouchDB.
//
//To access CouchDB docker container from within another docker container or from vagrant environments:
// http://couchdb:5984/
//
//Inside couchdb docker container
// http://127.0.0.1:5984/

// Index for docType, owner.
//
// Example curl command line to define index in the CouchDB channel_chaincode database
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[\"docType\",\"owner\"]},\"name\":\"indexOwner\",\"ddoc\":\"indexOwnerDoc\",\"type\":\"json\"}" http://hostname:port/myc1_marbles/_index
//

// Index for docType, owner, size (descending order).
//
// Example curl command line to define index in the CouchDB channel_chaincode database
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[{\"size\":\"desc\"},{\"docType\":\"desc\"},{\"owner\":\"desc\"}]},\"ddoc\":\"indexSizeSortDoc\", \"name\":\"indexSizeSortDesc\",\"type\":\"json\"}" http://hostname:port/myc1_marbles/_index

// Rich Query with index design doc and index name specified (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":\"marble\",\"owner\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

// Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":{\"$eq\":\"marble\"},\"owner\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"owner\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
        "time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type AuthChaincode struct {
}
//授权数据项使用场景
type AuthScene struct{
	AuthId	   	string 		`json:"authId"`		//授权Id
	AuthName        string          `json:"authName"`       //授权名称
	GrantorType	string		`json:"grantorType"`	//授权人类型
	GrantorId  	string 		`json:"grantorId"`	//授权人Id
	GrantorName    	string 		`json:"grantorName"`	//授权人名字
	GrantorIdCard	string		`json:"grantorIdCard"`	//授权人身份证号
	CnName		string		`json:"cnname"`		//组织名称
	SocialCreditCode string		`json:"socialCreditCode"`//统一社会信用代码	
	GranteeId  	string 		`json:"granteeId"`	//被授权人Id
	GranteeName	string 		`json:"granteeName"`	//被授权人
	SceneId		string		`json:"sceneId"`	//场景Id
	SceneName	string		`json:"sceneName"`	//场景名称
	Period		string		`json:"period"`		//业务阶段
	AuthProtoFileId string 		`json:"protocolFile"`   //授权协议文件
	AuthDate 	string		`json:"authDate"`	//授权日期
	AuthBeginTime	string	 	`json:"authBeginTime"`	//授权开始时间
	AuthEndTime     string		`json:"authEndTime"`	//授权结束时间
	AuthStatus	string 		`json:"authStatus"`	//授权状态
}

//授权数据项
type AuthDataItem struct{
	FingerId        string          `json:"fingerId"`       //数据指纹Id
        AuthId          string          `json:"authId"`         //授权Id
	SourceDeptId    string          `json:"sourceDeptId"`   //数据生产单位Id
        SourceDept      string          `json:"sourceDept"`     //数据生产单位
	GranteeId       string          `json:"granteeId"`      //被授权人Id
        GranteeName     string          `json:"granteeName"`    //被授权人
	SceneId         string          `json:"sceneId"`        //场景Id
        SceneName       string          `json:"sceneName"`      //场景名称
	ModelId		string		`json:"modelId"`	//模型Id
	ModelName	string		`json:"modelName"`	//模型名称
	ObjectId	string		`json:"objectId"`	//对象Id
	ObjectName      string          `json:"objectName"`     //对象名称
	FieldId		string		`json:"fieldId"`	//数据项Id
	FieldName	string		`json:"fieldName"`	//数据项名称
	UseRule		string		`json:"useRule"`	//使用规则
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(AuthChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *AuthChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *AuthChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

        // Handle different functions
        if function == "createAuthScene" { //create a new auth scene
                return t.createAuthScene(stub, args)
        }else if function == "createAuthDataItem" { //create a new auth data item
                return t.createAuthDataItem(stub, args)
        }else if function == "readAuthScene" { //read a auth scene
                return t.readAuthScene(stub, args)
        }else if function == "queryAuthSceneByGrantorId" { //read a auth info by grantorId
                return t.queryAuthSceneByGrantorId(stub, args)
        }else if function == "queryAuthSceneBySceneId" { //read a auth info by sceneId
                return t.queryAuthSceneBySceneId(stub, args)
        }else if function == "queryAuthSceneByGranteeId" { //read a auth info by granteeId
                return t.queryAuthSceneByGranteeId(stub, args)
        }else if function == "queryAuthSceneWithPagination" { //read a auth info with pagination
                return t.queryAuthSceneWithPagination(stub, args)
        }else if function == "getHistoryForAuthScene" { //get a auth info change history
                return t.getHistoryForAuthScene(stub, args)
        }

	fmt.Println("invoke did not find func: " + function) //error
        return shim.Error("Received unknown function invocation")
}

// ============================================================
// createAuthScene - create a new auth scene, store into chaincode state
// ============================================================
func (t *AuthChaincode) createAuthScene(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 18 {
		return shim.Error("Incorrect number of arguments. Expecting 18")
	}

	// ==== Input sanitation ====
	fmt.Println("- start create auth scene")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
                return shim.Error("4th argument must be a non-empty string")
        }
	if len(args[5]) <= 0 {
                return shim.Error("4th argument must be a non-empty string")
        }
	authId := args[0]
	authName := args[1]
	grantorType := args[2]
	grantorId := args[3]
	grantorName := args[4]
	grantorIdCard := args[5]
	cnname := args[6]
	socialCreditCode := args[7]
	granteeId := args[8]
	granteeName := args[9]
	sceneId := args[10]
	sceneName := args[11]
 	period := args[12]
	protocolFile := args[13]
	authDate := args[14]
	authBeginTime := args[15]
	authEndTime := args[16]
	authStatus := args[17]

	// ==== Check if auth scene already exists ====
	authSceneAsBytes, err := stub.GetState(authId)
	if err != nil {
		return shim.Error("Failed to get auth scene: " + err.Error())
	} else if authSceneAsBytes != nil {
		fmt.Println("This auth scene already exists: " + authId)
		return shim.Error("This auth scene already exists: " + authId)
	}

	// ==== Create auth scene object and marshal to JSON ====
	authScene := &AuthScene{authId, authName, grantorType, grantorId, grantorName, grantorIdCard, cnname, socialCreditCode, granteeId, granteeName, sceneId, sceneName, period, protocolFile, authDate, authBeginTime, authEndTime, authStatus}
	authSceneJSONasBytes, err := json.Marshal(authScene)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save auth scene to state ===
	err = stub.PutState(authId, authSceneJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the marble to enable color-based range queries, e.g. return all blue marbles ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	indexName := "grantorId~granteeId~sceneId"
	indexKey, err := stub.CreateCompositeKey(indexName, []string{authScene.GrantorId, authScene.GranteeId, authScene.SceneId})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the marble.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(indexKey, value)

	// ==== Auth info saved and indexed. Return success ====
	fmt.Println("- end create auth scene")
	return shim.Success(nil)
}

// ============================================================
// createAuthDataItem - create a new auth data item, store into chaincode state
// ============================================================
func (t *AuthChaincode) createAuthDataItem(stub shim.ChaincodeStubInterface, args []string) pb.Response {
        var err error

        if len(args) != 5 {
                return shim.Error("Incorrect number of arguments. Expecting 4")
        }

        // ==== Input sanitation ====
        fmt.Println("- start create auth data item")
        if len(args[0]) <= 0 {
                return shim.Error("1st argument must be a non-empty string")
        }
        if len(args[1]) <= 0 {
                return shim.Error("2nd argument must be a non-empty string")
        }
        if len(args[2]) <= 0 {
                return shim.Error("3rd argument must be a non-empty string")
        }
        if len(args[3]) <= 0 {
                return shim.Error("4th argument must be a non-empty string")
        }
        if len(args[4]) <= 0 {
                return shim.Error("4th argument must be a non-empty string")
        }
        fingerId := args[0]
        authId := args[1]
        sourceDeptId := args[2]
        sourceDept := args[3]
        granteeId := args[4]
        granteeName := args[5]
        sceneId := args[6]
        sceneName := args[7]
        modelId := args[8]
        modelName := args[9]
        objectId := args[10]
        objectName := args[11]
        fieldId := args[12]
        fieldName := args[13]
        useRule := args[14]

        // ==== Check if auth data item already exists ====
        authDataItemAsBytes, err := stub.GetState(fingerId)
        if err != nil {
                return shim.Error("Failed to get auth data item: " + err.Error())
        } else if authDataItemAsBytes != nil {
                fmt.Println("This auth data item already exists: " + fingerId)
                return shim.Error("This auth data item already exists: " + fingerId)
        }

        // ==== Create auth data item object and marshal to JSON ====
        authDataItem := &AuthDataItem{fingerId, authId, sourceDeptId, sourceDept, granteeId, granteeName, sceneId, sceneName, modelId, modelName, objectId, objectName, fieldId, fieldName, useRule}
        authDataItemJSONasBytes, err := json.Marshal(authDataItem)
        if err != nil {
                return shim.Error(err.Error())
        }

        // === Save auth scene to state ===
        err = stub.PutState(authId, authDataItemJSONasBytes)
        if err != nil {
                return shim.Error(err.Error())
        }

        //  ==== Index the marble to enable color-based range queries, e.g. return all blue marbles ====
        //  An 'index' is a normal key/value entry in state.
        //  The key is a composite key, with the elements that you want to range query on listed first.
        //  In our case, the composite key is based on indexName~color~name.
        //  This will enable very efficient state range queries based on composite keys matching indexName~color~*
        indexName := "authId~sourceDeptId~granteeId~sceneId~modelId~objectId~fieldId"
        indexKey, err := stub.CreateCompositeKey(indexName, []string{authDataItem.AuthId, authDataItem.SourceDeptId, authDataItem.GranteeId, authDataItem.SceneId, authDataItem.ModelId, authDataItem.ObjectId, authDataItem.FieldId})
        if err != nil {
                return shim.Error(err.Error())
        }
        //  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the marble.
        //  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
        value := []byte{0x00}
        stub.PutState(indexKey, value)

        // ==== Auth info saved and indexed. Return success ====
        fmt.Println("- end create auth data item")
        return shim.Success(nil)
}

// ===============================================
// readAuthScene - read a authScene from chaincode state
// ===============================================
func (t *AuthChaincode) readAuthScene(stub shim.ChaincodeStubInterface, args []string) pb.Response {
        var authId, jsonResp string
        var err error

        if len(args) != 1 {
                return shim.Error("Incorrect number of arguments. Expecting id of the authInfo to query")
        }

        authId = args[0]
        valAsbytes, err := stub.GetState(authId) //get the authScene from chaincode state
        if err != nil {
                jsonResp = "{\"Error\":\"Failed to get state for " + authId + "\"}"
                return shim.Error(jsonResp)
        } else if valAsbytes == nil {
                jsonResp = "{\"Error\":\"Auth info does not exist: " + authId + "\"}"
                return shim.Error(jsonResp)
        }

        return shim.Success(valAsbytes)
}

// ===== Example: Parameterized rich query =================================================
// queryAuthSceneByGrantorId queries for authInfo based on a passed in owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *AuthChaincode) queryAuthSceneByGrantorId(stub shim.ChaincodeStubInterface, args []string) pb.Response {
        if len(args) < 1 {
                return shim.Error("Incorrect number of arguments. Expecting 1")
        }

        grantorId := args[0]

        queryString := fmt.Sprintf("{\"selector\":{\"grantorId\":\"%s\"}}", grantorId)

        queryResults, err := getQueryResultForQueryString(stub, queryString)
        if err != nil {
                return shim.Error(err.Error())
        }
        return shim.Success(queryResults)
}

// ===== Example: Parameterized rich query =================================================
// queryAuthSceneBySceneId queries for authInfo based on a passed in owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *AuthChaincode) queryAuthSceneBySceneId(stub shim.ChaincodeStubInterface, args []string) pb.Response {
        if len(args) < 1 {
                return shim.Error("Incorrect number of arguments. Expecting 1")
        }

        sceneId := args[0]

        queryString := fmt.Sprintf("{\"selector\":{\"sceneId\":\"%s\"}}", sceneId)

        queryResults, err := getQueryResultForQueryString(stub, queryString)
        if err != nil {
                return shim.Error(err.Error())
        }
        return shim.Success(queryResults)
}

// ===== Example: Parameterized rich query =================================================
// queryAuthSceneByGranteeId queries for authInfo based on a passed in owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *AuthChaincode) queryAuthSceneByGranteeId(stub shim.ChaincodeStubInterface, args []string) pb.Response {
        if len(args) < 1 {
                return shim.Error("Incorrect number of arguments. Expecting 1")
        }

        granteeId := args[0]

        queryString := fmt.Sprintf("{\"selector\":{\"granteeId\":\"%s\"}}", granteeId)

        queryResults, err := getQueryResultForQueryString(stub, queryString)
        if err != nil {
                return shim.Error(err.Error())
        }
        return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

        fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

        resultsIterator, err := stub.GetQueryResult(queryString)
        if err != nil {
                return nil, err
        }
        defer resultsIterator.Close()

        buffer, err := constructQueryResponseFromIterator(resultsIterator)
        if err != nil {
                return nil, err
        }

        fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

        return buffer.Bytes(), nil
}

// ===========================================================================================
// constructQueryResponseFromIterator constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
        // buffer is a JSON array containing QueryResults
        var buffer bytes.Buffer
        buffer.WriteString("[")

        bArrayMemberAlreadyWritten := false
        for resultsIterator.HasNext() {
                queryResponse, err := resultsIterator.Next()
                if err != nil {
                        return nil, err
                }
                // Add a comma before array members, suppress it for the first array member
                if bArrayMemberAlreadyWritten == true {
                        buffer.WriteString(",")
                }
                buffer.WriteString("{\"Key\":")
                buffer.WriteString("\"")
                buffer.WriteString(queryResponse.Key)
                buffer.WriteString("\"")

                buffer.WriteString(", \"Record\":")
                // Record is a JSON object, so we write as-is
                buffer.WriteString(string(queryResponse.Value))
                buffer.WriteString("}")
                bArrayMemberAlreadyWritten = true
        }
        buffer.WriteString("]")

        return &buffer, nil
}
// ===== Example: Pagination with Ad hoc Rich Query ========================================================
// queryAuthSceneWithPagination uses a query string, page size and a bookmark to perform a query
// for marbles. Query string matching state database syntax is passed in and executed as is.
// The number of fetched records would be equal to or lesser than the specified page size.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryMarblesForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// Paginated queries are only valid for read only transactions.
// =========================================================================================
func (t *AuthChaincode) queryAuthSceneWithPagination(stub shim.ChaincodeStubInterface, args []string) pb.Response {

        //   0
        // "queryString"
        if len(args) < 3 {
                return shim.Error("Incorrect number of arguments. Expecting 3")
        }

        queryString := args[0]
        //return type of ParseInt is int64
        pageSize, err := strconv.ParseInt(args[1], 10, 32)
        if err != nil {
                return shim.Error(err.Error())
        }
        bookmark := args[2]

        queryResults, err := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookmark)
        if err != nil {
                return shim.Error(err.Error())
        }
        return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryStringWithPagination executes the passed in query string with
// pagination info. Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryStringWithPagination(stub shim.ChaincodeStubInterface, queryString string, pageSize int32, bookmark string) ([]byte, error) {

        fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

        resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(queryString, pageSize, bookmark)
        if err != nil {
                return nil, err
        }
        defer resultsIterator.Close()

        buffer, err := constructQueryResponseFromIterator(resultsIterator)
        if err != nil {
                return nil, err
        }

        bufferWithPaginationInfo := addPaginationMetadataToQueryResults(buffer, responseMetadata)

        fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", bufferWithPaginationInfo.String())

        return buffer.Bytes(), nil
}

// ===========================================================================================
// addPaginationMetadataToQueryResults adds QueryResponseMetadata, which contains pagination
// info, to the constructed query results
// ===========================================================================================
func addPaginationMetadataToQueryResults(buffer *bytes.Buffer, responseMetadata *pb.QueryResponseMetadata) *bytes.Buffer {

        buffer.WriteString("[{\"ResponseMetadata\":{\"RecordsCount\":")
        buffer.WriteString("\"")
        buffer.WriteString(fmt.Sprintf("%v", responseMetadata.FetchedRecordsCount))
        buffer.WriteString("\"")
        buffer.WriteString(", \"Bookmark\":")
        buffer.WriteString("\"")
        buffer.WriteString(responseMetadata.Bookmark)
        buffer.WriteString("\"}}]")

        return buffer
}

func (t *AuthChaincode) getHistoryForAuthScene(stub shim.ChaincodeStubInterface, args []string) pb.Response {

        if len(args) < 1 {
                return shim.Error("Incorrect number of arguments. Expecting 1")
        }

        authId := args[0]

        fmt.Printf("- start getHistoryForAuthScene: %s\n", authId)

        resultsIterator, err := stub.GetHistoryForKey(authId)
        if err != nil {
                return shim.Error(err.Error())
        }
        defer resultsIterator.Close()

        // buffer is a JSON array containing historic values for the auth scene
        var buffer bytes.Buffer
        buffer.WriteString("[")

        bArrayMemberAlreadyWritten := false
        for resultsIterator.HasNext() {
                response, err := resultsIterator.Next()
                if err != nil {
                        return shim.Error(err.Error())
                }
                // Add a comma before array members, suppress it for the first array member
                if bArrayMemberAlreadyWritten == true {
                        buffer.WriteString(",")
                }
                buffer.WriteString("{\"TxId\":")
                buffer.WriteString("\"")
                buffer.WriteString(response.TxId)
                buffer.WriteString("\"")

                buffer.WriteString(", \"Value\":")
                // if it was a delete operation on given key, then we need to set the
                //corresponding value null. Else, we will write the response.Value
                //as-is (as the Value itself a JSON marble)
                if response.IsDelete {
                        buffer.WriteString("null")
                } else {
                        buffer.WriteString(string(response.Value))
                }

                buffer.WriteString(", \"Timestamp\":")
                buffer.WriteString("\"")
                buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
                buffer.WriteString("\"")

                buffer.WriteString(", \"IsDelete\":")
                buffer.WriteString("\"")
                buffer.WriteString(strconv.FormatBool(response.IsDelete))
                buffer.WriteString("\"")

                buffer.WriteString("}")
                bArrayMemberAlreadyWritten = true
        }
        buffer.WriteString("]")

        fmt.Printf("- getHistoryForAuthInfo returning:\n%s\n", buffer.String())

        return shim.Success(buffer.Bytes())
}

