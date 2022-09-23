package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"time"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset identifica la relazione di servizio che si stabilisce tra un Rider e il Driver a lui più vicino, al momento della richiesta del servizio stesso.
// @AssetID: identificativo dell'Asset
// @DriverID: identificativo del driver, entro l'applicazione
// @RiderID: identificativo del rider, entro l'applicazione
// @TimeStampServizio: timestamp generato nel momento in cui è associato un driver al rider
type Asset struct {
	AssetID           string `json:"AssetID"`
	DriverID          string `json:"DriverID"`
	RiderID           string `json:"RiderID"`
	TimeStampServizio string `json:"TimeStampServizio"`
}

// InitLedger inserisce una serie di Asset mock con cui interagire
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface, timestampMock string) error {
	assets := []Asset{
		{AssetID: "asset0", DriverID: "Driver91", RiderID: "Rider6", TimeStampServizio: timestampMock},
		{AssetID: "asset1", DriverID: "Driver2", RiderID: "Rider19", TimeStampServizio: timestampMock},
		{AssetID: "asset2", DriverID: "Driver41", RiderID: "Rider24", TimeStampServizio: timestampMock},
		{AssetID: "asset3", DriverID: "Driver32", RiderID: "Rider24", TimeStampServizio: timestampMock},
		{AssetID: "asset4", DriverID: "Driver14", RiderID: "Rider19", TimeStampServizio: timestampMock},
		{AssetID: "asset5", DriverID: "Driver53", RiderID: "Rider19", TimeStampServizio: timestampMock},
		{AssetID: "asset6", DriverID: "Driver6", RiderID: "Rider3", TimeStampServizio: timestampMock},
		{AssetID: "asset7", DriverID: "Driver27", RiderID: "Rider87", TimeStampServizio: timestampMock},
		{AssetID: "asset8", DriverID: "Driver18", RiderID: "Rider19", TimeStampServizio: timestampMock},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.AssetID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset inserisce un nuovo Asset di servizio
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, assetID string, driverID string, riderID string) error {
	exists, err := s.AssetExists(ctx, assetID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", assetID)
	}

	asset := Asset{
		AssetID:           assetID,
		DriverID:          driverID,
		RiderID:           riderID,
		TimeStampServizio: "null",
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(assetID, assetJSON)
}

// ReadAsset restituisce l'Asset corrispondente all'@assetID fornito
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, assetID string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", assetID)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset aggiorna lo stato di un Asset di servizio con i nuovi parametri forniti.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, assetID string, driverID string, riderID string, timeStampServizio time.Time) error {
	exists, err := s.AssetExists(ctx, assetID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", assetID)
	}

	// overwriting original asset with new asset
	asset := Asset{
		AssetID:           assetID,
		DriverID:          driverID,
		RiderID:           riderID,
		TimeStampServizio: timeStampServizio.String(),
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(assetID, assetJSON)
}

// DeleteAsset elimina l'Asset richiesto
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, assetID string) error {
	exists, err := s.AssetExists(ctx, assetID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", assetID)
	}

	return ctx.GetStub().DelState(assetID)
}

// AssetExists restituisce un booleano corrispondente all'esistenza dell'Asset di servizio
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, assetID string) (bool, error) {

	assetJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferAsset aggiorna l'Asset corrispondente a @assetID attribuendolo ad un nuovo Driver
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, assetID string, newDriver string) (string, error) {
	asset, err := s.ReadAsset(ctx, assetID)
	if err != nil {
		return "", err
	}

	oldDriver := asset.DriverID
	asset.DriverID = newDriver

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(assetID, assetJSON)
	if err != nil {
		return "", err
	}

	return oldDriver, nil
}

// GetAllAssets restituisce tutti i servizi erogati
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
