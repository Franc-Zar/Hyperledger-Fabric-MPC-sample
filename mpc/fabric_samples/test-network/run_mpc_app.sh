#!/bin/bash
#Script di avvio dell'applicazione:

case $1 in 
    "-u" | "--up")
        #1) Network setup
        ./network.sh up createChannel -c mychannel -ca

        #2) Chaincode deployment
        ./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go

        export PATH=${PWD}/../bin:$PATH
        export FABRIC_CFG_PATH=$PWD/../config/

        # Environment variables for Org1

        export CORE_PEER_TLS_ENABLED=true
        export CORE_PEER_LOCALMSPID="Org1MSP"
        export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
        export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
        export CORE_PEER_ADDRESS=localhost:7051

        #4) Moving to app dir 
        cd ../asset-transfer-basic/application-gateway-go/

        #4) Run
        go run .
    	;;
    
    "-d" |" --down") 
        ./network.sh down
	;;     
	
	*) 
        echo "Automatically configure the Hyperledger Fabric test-network, deploy the sample mpc chaincode and run it with Gateway 
        Usage:
        ./run_mpc_app.sh [flags] 
        
        flags:
        -h --help       help for usage
        -u --up         start network, deploy chaincode and run
        -d --down       clean environment"
    	;;
	
esac
