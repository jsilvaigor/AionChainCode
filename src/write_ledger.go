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
// write() - generic write variable into ledger
//
// Shows Off PutState() - writing a key/value into the ledger
//
// Inputs - Array of strings
//    0   ,    1
//   key  ,  value
//  "abc" , "test"
// ============================================================================================================================
func write(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key, value string
	var err error
	fmt.Println("starting write")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2. key of the variable and value to set")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	key = args[0] //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value)) //write the variable into the ledger
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end write")
	return shim.Success(nil)
}

// ============================================================================================================================
// Init Certificate - create a new certificate, store into chaincode state
//
// Shows off building a key's JSON value manually
//
// Inputs - Array of strings
//   0    | 1      |  2        | 3                | 4           | 5               | 6             | 7
//  id    | name   |  document | body             | city        | date            | university_id | university_doc
// "c123" | "José" |  "123456" | "Certificate..." | "Sao Paulo" | "1501810298042" | "u123"        | "456789"
// ============================================================================================================================
func init_cert(stub shim.ChaincodeStubInterface, args []string) (pb.Response) {
	var err error
	fmt.Println("starting init_cert")

	if len(args) != 8 {
		return shim.Error("Incorrect number of arguments. Expecting 8")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	id := args[0]
	name := args[1]
	document := args[2]
	body := args[3]
	city := args[4]
	date := args[5]
	university_id := args[6]
	university_doc := args[7]

	//check if university exists
	university, err := get_university(stub, university_id)
	if err != nil {
		fmt.Println("Failed to find university - " + university_id)
		return shim.Error(err.Error())
	}

	//check university document
	if university.Document != university_doc {
		return shim.Error("The university '" + university.UniversityName + "' cannot authorize creation for university '" + university_doc + "'.")
	}

	//check if certificate id already exists
	certificate, err := get_certificate(stub, id)
	if err == nil {
		fmt.Println("This certificate already exists - " + id)
		fmt.Println(certificate)
		return shim.Error("This certificate already exists - " + id) //all stop a certificate by this id exists
	}

	//build the certificate json string manually
	str := `{
		"docType":"certificate",
		"id": "` + id + `",
		"name": "` + name + `",
		"document": ` + document + `,
		"body": ` + body + `,
		"city": ` + city + `,
		"date": ` + date + `,
		"university": {
			"id": "` + university_id + `",
			"dean": "` + university.Dean + `",
			"name": "` + university.UniversityName + `"
		}
	}`
	err = stub.PutState(id, []byte(str)) //store certificate with id as key
	if err != nil {
		fmt.Println("Could not store certificate")
		return shim.Error(err.Error())
	}

	//add current certificate to known certificates
	university.Certificates = append(university.Certificates, id)
	//store university
	universityAsBytes, _ := json.Marshal(university)      //convert to array of bytes
	err = stub.PutState(university.Id, universityAsBytes) //store university by its Id
	if err != nil {
		fmt.Println("Could not store university")
		return shim.Error(err.Error())
	}

	fmt.Println("- end init_certificate")
	return shim.Success(nil)
}

// ============================================================================================================================
// Init University - create a new university, store into chaincode state
//
// Shows off building key's value from GoLang Structure
//
//type University struct {
//	ObjectType     string `json:"docType"`        //field for couchdb
//	Id             string `json:"id"`             //UUID
//	Dean           string `json:"dean"`           //dean of university (reitor)
//	UniversityName string `json:"universityname"` //Name of university
//	Document       string `json:"document"`       //University national document (cnpj)
//}
// Inputs - Array of Strings
// 0      | 1      | 2        | 3
// id     | dean   | name     | document
// "u123" | "joao" | "uniuni" | "123456"
// ============================================================================================================================
func init_university(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting init_university")

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var university University
	university.ObjectType = "university"
	university.Id = args[0]
	university.Dean = args[1]
	university.UniversityName = args[2]
	university.Document = args[3]
	fmt.Println(university)

	//check if university already exists
	_, err = get_university(stub, university.Id)
	if err == nil {
		fmt.Println("This university already exists - " + university.Id)
		return shim.Error("This university already exists - " + university.Id)
	}

	//store university
	universityAsBytes, _ := json.Marshal(university)      //convert to array of bytes
	err = stub.PutState(university.Id, universityAsBytes) //store university by its Id
	if err != nil {
		fmt.Println("Could not store university")
		return shim.Error(err.Error())
	}

	fmt.Println("- end init_university")
	return shim.Success(nil)
}

// ============================================================================================================================
// Set Dean on University
//
// Shows off GetState() and PutState()
//
// Inputs - Array of Strings
//  0             | 1        | 2
//  university_id | old_dean | new_dean
// "u123"         | "Joao"   | "Jose"
// ============================================================================================================================
func set_dean(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	var university University
	fmt.Println("starting set_owner")

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var university_id = args[0]
	var old_dean = args[1]
	var new_dean = args[2]
	fmt.Println(university_id + "->" + old_dean + " - |" + new_dean)

	// check if new_dean is equals to old_dean
	if old_dean == new_dean {
		return shim.Error("Old dean (" + old_dean + ") is equal to new dean (" + new_dean + ")")
	}

	// retrieves university
	university, err = get_university(stub, university_id)
	if err != nil {
		return shim.Error(err.Error())
	}

	// check if provided old dean is diferent from state saved current dean
	if university.Dean != old_dean {
		return shim.Error("Provided old dean isn't the current dean")
	}
	// check if provided new dean is equal to state saved current dean
	if university.Dean == new_dean {
		return shim.Error("Provided new dean is the current dean")
	}

	// change dean
	university.Dean = new_dean
	jsonAsBytes, _ := json.Marshal(university) //convert to array of bytes
	err = stub.PutState(args[0], jsonAsBytes)  //rewrite the university with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end set dean")
	return shim.Success(nil)
}
