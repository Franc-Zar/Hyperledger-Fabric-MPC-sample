package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Service
type SmartContract struct {
	contractapi.Contract
}

// Service identifica la relazione del servizio che si stabilisce tra un Rider e il Driver a lui più vicino, al momento della richiesta del servizio stesso:
// contiene informazioni di report che il Driver deve comunicare al RHS Provider per esporre i servizi erogati e i dati corrispondenti.
// @ServiceID: identificativo del Service erogato
// @DriverID: identificativo del driver, entro l'applicazione
// @TimeStampServizio: timestamp generato nel momento in cui è associato un driver al rider
// @Fare: contiene informazioni di report (banalmente il pagamento del servizio offerto) da riportare al RHS-Provider
type Service struct {
	ServiceID         string `json:"ServiceID"`
	DriverID          string `json:"DriverID"`
	TimeStampServizio string `json:"TimeStampServizio"`
	Fare              string `json:"Fare"`
}

// InitLedger inserisce una serie di Service mock con cui interagire
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface, timestampMock string) error {
	assets := []Service{
		{ServiceID: "service0", DriverID: "Driver91", TimeStampServizio: timestampMock, Fare: "21€"},
		{ServiceID: "service1", DriverID: "Driver2", TimeStampServizio: timestampMock, Fare: "25€"},
		{ServiceID: "service2", DriverID: "Driver41", TimeStampServizio: timestampMock, Fare: "57€"},
		{ServiceID: "service3", DriverID: "Driver32", TimeStampServizio: timestampMock, Fare: "12€"},
		{ServiceID: "service4", DriverID: "Driver14", TimeStampServizio: timestampMock, Fare: "7€"},
		{ServiceID: "service5", DriverID: "Driver53", TimeStampServizio: timestampMock, Fare: "30€"},
		{ServiceID: "service6", DriverID: "Driver6", TimeStampServizio: timestampMock, Fare: "9€"},
		{ServiceID: "service7", DriverID: "Driver27", TimeStampServizio: timestampMock, Fare: "24€"},
		{ServiceID: "service8", DriverID: "Driver18", TimeStampServizio: timestampMock, Fare: "15€"},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ServiceID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateService inserisce un nuovo Service, asset di servizio associato al Rider che richiede l'operazione
func (s *SmartContract) CreateService(ctx contractapi.TransactionContextInterface, serviceID string, driverID string, timestampServizio string, fare string) error {
	exists, err := s.ServiceExists(ctx, serviceID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", serviceID)
	}

	asset := Service{
		ServiceID:         serviceID,
		DriverID:          driverID,
		TimeStampServizio: timestampServizio,
		Fare:              fare,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(serviceID, assetJSON)
}

// ReadService restituisce l'Service corrispondente all'@assetID fornito
func (s *SmartContract) ReadService(ctx contractapi.TransactionContextInterface, serviceID string) (*Service, error) {
	assetJSON, err := ctx.GetStub().GetState(serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", serviceID)
	}

	var asset Service
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateService aggiorna lo stato di un Service di servizio con i nuovi parametri forniti.
func (s *SmartContract) UpdateService(ctx contractapi.TransactionContextInterface, serviceID string, driverID string, timeStampServizio string, fare string) error {
	exists, err := s.ServiceExists(ctx, serviceID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", serviceID)
	}

	// overwriting original asset with new asset
	asset := Service{
		ServiceID:         serviceID,
		DriverID:          driverID,
		TimeStampServizio: timeStampServizio,
		Fare:              fare,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(serviceID, assetJSON)
}

// DeleteService elimina l'Service richiesto
func (s *SmartContract) DeleteService(ctx contractapi.TransactionContextInterface, serviceID string) error {
	exists, err := s.ServiceExists(ctx, serviceID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", serviceID)
	}

	return ctx.GetStub().DelState(serviceID)
}

// ServiceExists restituisce un booleano corrispondente all'esistenza dell'Service di servizio
func (s *SmartContract) ServiceExists(ctx contractapi.TransactionContextInterface, serviceID string) (bool, error) {

	assetJSON, err := ctx.GetStub().GetState(serviceID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferService aggiorna il Service corrispondente a @assetID attribuendolo ad un nuovo Driver
func (s *SmartContract) TransferService(ctx contractapi.TransactionContextInterface, serviceID string, newDriver string) (string, error) {
	asset, err := s.ReadService(ctx, serviceID)
	if err != nil {
		return "", err
	}

	oldDriver := asset.DriverID
	asset.DriverID = newDriver

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(serviceID, assetJSON)
	if err != nil {
		return "", err
	}

	return oldDriver, nil
}

// GetAllServices restituisce tutti i servizi erogati
func (s *SmartContract) GetAllServices(ctx contractapi.TransactionContextInterface) ([]*Service, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Service
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Service
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
