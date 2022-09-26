/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"assetTransfer/utilities"
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/fatih/color"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	mspID            = "Org1MSP"
	cryptoPath       = "../../test-network/organizations/peerOrganizations/org1.example.com"
	certPath         = cryptoPath + "/users/User1@org1.example.com/msp/signcerts/cert.pem"
	keyPath          = cryptoPath + "/users/User1@org1.example.com/msp/keystore/"
	tlsCertPath      = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint     = "localhost:7051"
	gatewayPeer      = "peer0.org1.example.com"
	channelName      = "secure-rider-driver"
	chaincodeName    = "mpc-app"
	defaultnbDrivers = 2048
)

// serve al reperimento dell'id della transazione creata, dal momento che è generato casualmente mediante il package uuid di Google
var createdServiceID string

func main() {
	color.Blue("============ application-golang starts ============")

	// The gRPC client connection should be shared by all Gateway connections to this endpoint
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gateway.Close()

	network := gateway.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	color.Cyan("*************************************************")
	color.Cyan("initLedger:")
	initLedger(contract)
	fmt.Println()

	color.Cyan("*************************************************")
	color.Cyan("createServices:")
	createServices(contract, "Rider787")
	fmt.Println()

	color.Cyan("*************************************************")
	color.Cyan("readServiceByID sul servizio appena creato:")
	readServiceByID(contract, createdServiceID)
	fmt.Println()

	color.Cyan("*************************************************")
	color.Cyan("getAllServices:")
	getAllServices(contract)
	fmt.Println()

	color.Blue("============ application-golang ends ============")
}

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() *grpc.ClientConn {
	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() *identity.X509Identity {
	certificate, err := loadCertificate(certPath)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign() identity.Sign {
	files, err := ioutil.ReadDir(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key directory: %w", err))
	}
	privateKeyPEM, err := ioutil.ReadFile(path.Join(keyPath, files[0].Name()))

	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}

// This type of transaction would typically only be run once by an application the first time it was started after its
// initial deployment. A new version of the chaincode deployed later would likely not need to run an "init" function.
func initLedger(contract *client.Contract) {
	color.Cyan("Submit Transaction: InitLedger, function creates the initial set of assets on the ledger \n")

	_, err := contract.SubmitTransaction("InitLedger", time.Now().String())
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	color.Green("*** Transaction committed successfully\n")
}

// Evaluate a transaction to query ledger state.
func getAllServices(contract *client.Contract) {
	color.Cyan("Evaluate Transaction: GetAllServices, function returns all the current assets on the ledger")

	evaluateResult, err := contract.EvaluateTransaction("GetAllServices")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	color.Green("*** Result:%s\n", result)
}

// Submit a transaction synchronously, blocking until it has been committed to the ledger.
func createServices(contract *client.Contract, riderID string) {
	color.Cyan("Submit Transaction: CreateService, creates new asset with ID, Driver, Timestamp, Fare \n")

	// reperimento Driver più vicino
	closestDriverID, timestampServizio, isClosestDriverFound := utilities.ObliviousRiding(defaultnbDrivers, riderID)

	if isClosestDriverFound {
		// mock di servizio
		var fare string
		createdServiceID, fare = utilities.SetRide(riderID, closestDriverID)

		_, err := contract.SubmitTransaction("CreateService", createdServiceID, closestDriverID, timestampServizio, fare)
		if err != nil {
			panic(fmt.Errorf("failed to submit transaction: %w", err))
		}

		color.Green("*** Transaction committed successfully\n")

	} else {
		color.Red("failed to submit Transaction: driver not found")
	}
}

// Evaluate a transaction by serviceID to query ledger state.
func readServiceByID(contract *client.Contract, serviceID string) {
	color.Cyan("Evaluate Transaction: ReadService, function returns asset attributes\n")

	evaluateResult, err := contract.EvaluateTransaction("ReadService", serviceID)
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	color.Green("*** Result:%s\n", result)
}

// Format JSON data
func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, " ", ""); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()
}
