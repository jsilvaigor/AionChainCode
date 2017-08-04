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
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type AionCertsChainCode struct {
}

// ============================================================================================================================
// Asset Definitions - The ledger will store certificates and institutions
// ============================================================================================================================

// ----- Marbles ----- //
type Certificate struct {
	ObjectType string `json:"docType"`                //field for couchdb
	Id         string `json:"id"`                     //UUID of certificate
	Name       string `json:"name"`                   //Name of the graduate
	Document   string `json:"document"`               //Personal identification document (cpf)
	Body       string `json:"body"`                   //Principal content of certificate
	City       string `json:"city"`                   //Emission city
	Date       string `json:"date"`                   //Emission timestamp
	University UniversityRelation `json:"university"` //University
}

// ----- University ----- //
type University struct {
	ObjectType     string `json:"docType"`        //field for couchdb
	Id             string `json:"id"`             //UUID
	Dean           string `json:"dean"`           //dean of university (reitor)
	UniversityName string `json:"universityname"` //Name of university
	Document       string `json:"document"`       //University national document (cnpj)
	Certificates   []string `json:"certifcates"`  //Id of all certificates emitted
}

type UniversityRelation struct {
	Id   string `json:"id"`
	Dean string `json:"dean"` //cosmetic/handy, the real relation is by Id
	Name string `json:"name"` //cosmetic/handy, the real relation is by Id
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(AionCertsChainCode))
	if err != nil {
		fmt.Printf("Error starting AionCerts chaincode - %s", err)
	}
}

// ============================================================================================================================
// Init - initialize the chaincode - runs a simple test
// ============================================================================================================================
func (t *AionCertsChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("AionCerts Is Starting Up")
	_, args := stub.GetFunctionAndParameters()
	var Aval int
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// convert numeric string to integer
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return shim.Error("Expecting a numeric string argument to Init()")
	}

	// this is a very simple dumb test.  let's write to the ledger and error on any errors
	err = stub.PutState("selftest", []byte(strconv.Itoa(Aval))) //making a test var "selftest", its handy to read this right away to test the network
	if err != nil {
		return shim.Error(err.Error()) //self-test fail
	}

	fmt.Println(" - ready for action") //self-test pass
	return shim.Success(nil)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *AionCertsChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println(" ")
	fmt.Println("starting invoke, for - " + function)

	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub)
	} else if function == "read" { //generic read ledger
		return read(stub, args)
	} else if function == "write" { //generic writes to ledger
		return write(stub, args)
	} else if function == "init_cert" { //create a new certificate
		return init_cert(stub, args)
	} else if function == "set_dean" { //change owner of a marble
		return set_dean(stub, args)
	} else if function == "init_university" { //create a new marble owner
		return init_university(stub, args)
	} else if function == "read_everything" { //read all certificates from university
		return read_all_certificates_from_university(stub, args)
	} else if function == "getUniversityHistory" { //read history of a university
		return getHistory(stub, args)
	}

	// error out
	fmt.Println("Received unknown invoke function name - " + function)
	return shim.Error("Received unknown invoke function name - '" + function + "'")
}

// ============================================================================================================================
// Query - legacy function
// ============================================================================================================================
func (t *AionCertsChainCode) Query(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("Unknown supported call - Query()")
}
