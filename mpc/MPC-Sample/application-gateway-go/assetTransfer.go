/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"time"

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

func main() {
	log.Println("============ application-golang starts ============")

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

	fmt.Println("*************************************************")
	fmt.Println("initLedger:")
	initLedger(contract)
	fmt.Println()

	fmt.Println("*************************************************")
	fmt.Println("createAsset:")
	createAsset(contract, "Rider787")
	fmt.Println()

	fmt.Println("*************************************************")
	fmt.Println("readAssetByID:")
	readAssetByID(contract, "asset543")
	fmt.Println()

	fmt.Println("*************************************************")
	fmt.Println("getAllAssets:")
	getAllAssets(contract)
	fmt.Println()

	log.Println("============ application-golang ends ============")
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
	fmt.Printf("Submit Transaction: InitLedger, function creates the initial set of assets on the ledger \n")

	_, err := contract.SubmitTransaction("InitLedger", time.Now().String())
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Printf("*** Transaction committed successfully\n")
}

// Evaluate a transaction to query ledger state.
func getAllAssets(contract *client.Contract) {
	fmt.Println("Evaluate Transaction: GetAllAssets, function returns all the current assets on the ledger")

	evaluateResult, err := contract.EvaluateTransaction("GetAllAssets")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("*** Result:%s\n", result)
}

// Submit a transaction synchronously, blocking until it has been committed to the ledger.
func createAsset(contract *client.Contract, riderID string) {
	fmt.Printf("Submit Transaction: CreateAsset, creates new asset with ID, Driver, Rider, timestamp \n")

	_, timestampServizio, isClosestDriverFound := ObliviousRiding(defaultnbDrivers, riderID)

	if isClosestDriverFound {
		// transazione mock che non dipende dal risultato di ObliviousRiding poichè la logica con cui è implementata contiene funzionni di generazione casuale
		// dei dati di input relativi alle coordinate: non è possibile garantire l'esecuzione deterministica dello smart contract e quindi fallirebbe nella maggior
		// parte dei casi. Si considera tale transazione effettivamente corrispondente alla logica di esecuzione, ciò è accettabile poichè richiede la sola
		// sosituzione della generazione casuale dei dati con la opportuna logica di reperimento delle applicazioni
		_, err := contract.SubmitTransaction("CreateAsset", "asset543", riderID, "CloserRiderFoundID", timestampServizio)
		if err != nil {
			panic(fmt.Errorf("failed to submit transaction: %w", err))
		}

		fmt.Printf("*** Transaction committed successfully\n")

	} else {
		fmt.Printf("failed to submit transiction: driver not found")
	}
}

// Evaluate a transaction by assetID to query ledger state.
func readAssetByID(contract *client.Contract, assetID string) {
	fmt.Printf("Evaluate Transaction: ReadAsset, function returns asset attributes\n")

	evaluateResult, err := contract.EvaluateTransaction("ReadAsset", assetID)
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("*** Result:%s\n", result)
}

// Format JSON data
func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, " ", ""); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()
}
