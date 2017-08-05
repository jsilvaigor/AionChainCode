/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// ============================================================================================================================
// Read - read a generic variable from ledger
//
// Shows Off GetState() - reading a key/value from the ledger
//
// Inputs - Array of strings
//  0
//  key
//  "abc"
// 
// Returns - string
// ============================================================================================================================
func read(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key, jsonResp string
	var err error
	fmt.Println("starting read")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting key of the var to query")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key) //get the var from ledger
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}

	fmt.Println("- end read")
	return shim.Success(valAsbytes) //send it onward
}

// ============================================================================================================================
// Get all certificates from university
//
// Inputs - university_id
//
// ----- Marbles ----- //
//type Certificate struct {
//	ObjectType string `json:"docType"`                //field for couchdb
//	Id         string `json:"id"`                     //UUID of certificate
//	Name       string `json:"name"`                   //Name of the graduate
//	Document   string `json:"document"`               //Personal identification document (cpf)
//	Body       string `json:"body"`                   //Principal content of certificate
//	City       string `json:"city"`                   //Emission city
//	Date       string `json:"date"`                   //Emission timestamp
//	University UniversityRelation `json:"university"` //University
//}
// Returns:
// {
//	"certificates": [{
//			"id": "c123",
//			"name": "Jose"
//			"document": "123"
//			"body": "certificate..."
//			"city": "Sao Paulo"
//			"date": "1501813348098"
//			"university": {
//			    "id": "u213",
//			    "name": "UniUni",
//			    "dean": "Joao"
//          }
//	}]
// }
// ============================================================================================================================
func read_all_certificates_from_university(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	type Everything struct {
		Certificates []Certificate   `json:"certificates"`
	}
	var everything Everything

	var err error
	fmt.Println("starting read_all_certificates_from_university")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting key of the var to query")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	// get university
	university, err := get_university(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	if len(university.Certificates) > 0 {
		for _, item := range university.Certificates {
			certificate, err := get_certificate(stub, item)
			if err != nil {
				return shim.Error("An error occurred while retrieving certificates")
			}
			everything.Certificates = append(everything.Certificates, certificate)
		}
	}

	fmt.Println("everything array - ", everything)

	//change to array of bytes
	everythingAsBytes, _ := json.Marshal(everything) //convert to array of bytes
	return shim.Success(everythingAsBytes)
}

// ============================================================================================================================
// Get history of university
//
// Shows Off GetHistoryForKey() - reading complete history of a key/value
//
// Inputs - Array of strings
//  0
//  id
//  "m01490985296352SjAyM"
// ============================================================================================================================
func getHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	type AuditHistory struct {
		TxId      string   `json:"txId"`
		Value     University   `json:"value"`
		Timestamp int64   `json:"timestamp"`
	}
	var history []AuditHistory
	var err error
	var university University

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	university_id := args[0]
	fmt.Printf("- start getHistoryForUniversity: %s\n", university_id)

	// Get History
	resultsIterator, err := stub.GetHistoryForKey(university_id)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		historicValue, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var tx AuditHistory
		tx.TxId = historicValue.TxId                     //copy transaction id over
		json.Unmarshal(historicValue.Value, &university) //un stringify it aka JSON.parse()
		if historicValue == nil { //university has been deleted
			var emptyUniversity University
			tx.Value = emptyUniversity //copy nil university
		} else {
			json.Unmarshal(historicValue.Value, &university) //un stringify it aka JSON.parse()
			tx.Value = university                            //copy university over
		}
		tx.Timestamp = historicValue.Timestamp.Seconds
		history = append(history, tx) //add this tx to the list
	}
	fmt.Printf("- getHistoryForUniversity returning:\n%s", university)

	//change to array of bytes
	historyAsBytes, _ := json.Marshal(history) //convert to array of bytes
	return shim.Success(historyAsBytes)
}
