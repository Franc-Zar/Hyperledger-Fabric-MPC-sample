# MPC-Sample

The MPC-Sample is a simple Hyperledger Fabric application gateway with mpc functionalities:

- built starting from [fabric-samples](https://github.com/hyperledger/fabric-samples) and implementing mpc functionalities through [lattigo](https://github.com/tuneinsight/lattigo)
- inspired by [ORide](https://oride.epfl.ch/)


## About

This sample includes smart contract and application code in Go and is thought to be a simple demo. 
This sample shows creation of an asset representing a result of Oblivious Homomorphic Encryption based computation. 
Application's main goal is presenting a decentralized solution to the problem described by [ORide](https://oride.epfl.ch/) developers, 
built upon Hyperledger Fabric Blockchain Infrastructure.


### Application

Follow the execution flow in the client application code, and corresponding output on running the application. Pay attention to the sequence of:

- Transaction invocations (console output like "**--> Submit Transaction**" and "**--> Evaluate Transaction**").
- Results returned by transactions (console output like "**\*\*\* Result**").

### Smart Contract

The smart contract (in folder `chaincode-go`) implements the following functions to support the application:

- CreateService
- ReadService
- UpdateService
- DeleteService
- TransferService
- GetAllServices
- ServiceExists

Note that the asset transfer implemented by the smart contract is a simplified scenario, without ownership validation, meant only to demonstrate how to invoke transactions.

## Running the sample

The Fabric test network is used to deploy and run this sample:

Create test network, deploy chaincode and run application through its Gateway 
```
./run_mpc_app.sh --up  
```
Once and until the network is up, chaincode is deployed and application is executable (from the `MPC-Sample/application-gateway-go` folder):
```
go run .
```

## Clean up

When you are finished, you can bring down the test network (from the `test-network` folder). The command will remove all the nodes of the test network, and delete any ledger data that you created.

```
./run_mpc_app.sh --down  
```
