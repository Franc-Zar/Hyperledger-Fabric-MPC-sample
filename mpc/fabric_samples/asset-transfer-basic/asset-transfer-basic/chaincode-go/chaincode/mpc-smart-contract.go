package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/tuneinsight/lattigo/v3/bfv"
	"github.com/tuneinsight/lattigo/v3/ring"
	"github.com/tuneinsight/lattigo/v3/rlwe"
	"github.com/tuneinsight/lattigo/v3/utils"
	"math"
	"math/bits"
	"time"
)

// SmartContract provides functions for managing an Service
type SmartContract struct {
	contractapi.Contract
}

// Service identifica la relazione di servizio che si stabilisce tra un Rider e il Driver a lui più vicino, al momento della richiesta del servizio stesso.
// @ServiceID: identificativo dell'Service
// @DriverID: identificativo del driver, entro l'applicazione
// @RiderID: identificativo del rider, entro l'applicazione
// @TimeStampServizio: timestamp generato nel momento in cui è associato un driver al rider
type Service struct {
	ServiceID         string `json:"ServiceID"`
	DriverID          string `json:"DriverID"`
	RiderID           string `json:"RiderID"`
	TimeStampServizio string `json:"TimeStampServizio"`
}

func distance(a, b, c, d uint64) uint64 {
	if a > c {
		a, c = c, a
	}
	if b > d {
		b, d = d, b
	}
	x, y := a-c, b-d
	return x*x + y*y
}

func (s *SmartContract) ObliviousRiding(ctx contractapi.TransactionContextInterface, riderID string) {
	// This example simulates a situation where an anonymous rider
	// wants to find the closest available rider within a given area.
	// The application is inspired by the paper https://oride.epfl.ch/
	//
	// 		A. Pham, I. Dacosta, G. Endignoux, J. Troncoso-Pastoriza,
	//		K. Huguenin, and J.-P. Hubaux. ORide: A Privacy-Preserving
	//		yet Accountable Ride-Hailing Service. In Proceedings of the
	//		26th USENIX Security Symposium, Vancouver, BC, Canada, August 2017.
	//
	// Each area is represented as a rectangular grid where each driver
	// anyonymously signs in (i.e. the server only knows the driver is located
	// in the area).
	//
	// First, the rider generates an ephemeral key pair (riderSk, riderPk), which she
	// uses to encrypt her coordinates. She then sends the tuple (riderPk, enc(coordinates))
	// to the server handling the area she is in.
	//
	// Once the public key and the encrypted rider coordinates of the rider
	// have been received by the server, the rider's public key is transferred
	// to all the drivers within the area, with a randomized different index
	// for each of them, that indicates in which coefficient each driver must
	// encode her coordinates.
	//
	// Each driver encodes her coordinates in the designated coefficient and
	// uses the received public key to encrypt her encoded coordinates.
	// She then sends back the encrypted coordinates to the server.
	//
	// Once the encrypted coordinates of the drivers have been received, the server
	// homomorphically computes the squared distance: (x0 - x1)^2 + (y0 - y1)^2 between
	// the rider and each of the drivers, and sends back the encrypted result to the rider.
	//
	// The rider decrypts the result and chooses the closest driver.

	// Number of drivers in the area
	nbDrivers := 2048 //max is N

	// BFV parameters (128 bit security) with plaintext modulus 65929217
	paramDef := bfv.PN13QP218
	paramDef.T = 0x3ee0001

	params, err := bfv.NewParametersFromLiteral(paramDef)
	if err != nil {
		panic(err)
	}

	encoder := bfv.NewEncoder(params)

	// Rider's keygen
	kgen := bfv.NewKeyGenerator(params)

	riderSk, riderPk := kgen.GenKeyPair()

	decryptor := bfv.NewDecryptor(params, riderSk)

	encryptorRiderPk := bfv.NewEncryptor(params, riderPk)

	encryptorRiderSk := bfv.NewEncryptor(params, riderSk)

	evaluator := bfv.NewEvaluator(params, rlwe.EvaluationKey{})

	fmt.Println("============================================")
	fmt.Println("Homomorphic computations on batched integers")
	fmt.Println("============================================")
	fmt.Println()
	fmt.Printf("Parameters : N=%d, T=%d, Q = %d bits, sigma = %f \n",
		1<<params.LogN(), params.T(), params.LogQP(), params.Sigma())
	fmt.Println()

	maxvalue := uint64(math.Sqrt(float64(params.T()))) // max values = floor(sqrt(plaintext modulus))
	mask := uint64(1<<bits.Len64(maxvalue) - 1)        // binary mask upper-bound for the uniform sampling

	fmt.Printf("Generating %d driversData and 1 Rider randomly positioned on a grid of %d x %d units \n",
		nbDrivers, maxvalue, maxvalue)
	fmt.Println()

	prng, err := utils.NewPRNG()
	if err != nil {
		panic(err)
	}
	// Rider coordinates [x, y, x, y, ....., x, y]
	riderPosX, riderPosY := ring.RandUniform(prng, maxvalue, mask), ring.RandUniform(prng, maxvalue, mask)

	Rider := make([]uint64, 1<<params.LogN())
	for i := 0; i < nbDrivers; i++ {
		Rider[(i << 1)] = riderPosX
		Rider[(i<<1)+1] = riderPosY
	}

	riderPlaintext := bfv.NewPlaintext(params)
	encoder.Encode(Rider, riderPlaintext)

	// driversData coordinates [0, 0, ..., x, y, ..., 0, 0]
	driversData := make([][]uint64, nbDrivers)

	driversPlaintexts := make([]*bfv.Plaintext, nbDrivers)
	for i := 0; i < nbDrivers; i++ {
		driversData[i] = make([]uint64, 1<<params.LogN())
		driversData[i][(i << 1)] = ring.RandUniform(prng, maxvalue, mask)
		driversData[i][(i<<1)+1] = ring.RandUniform(prng, maxvalue, mask)
		driversPlaintexts[i] = bfv.NewPlaintext(params)
		encoder.Encode(driversData[i], driversPlaintexts[i])
	}

	fmt.Printf("Encrypting %d driversData (x, y) and 1 Rider (%d, %d) \n",
		nbDrivers, riderPosX, riderPosY)
	fmt.Println()

	RiderCiphertext := encryptorRiderSk.EncryptNew(riderPlaintext)

	DriversCiphertexts := make([]*bfv.Ciphertext, nbDrivers)
	for i := 0; i < nbDrivers; i++ {
		DriversCiphertexts[i] = encryptorRiderPk.EncryptNew(driversPlaintexts[i])
	}

	fmt.Println("Computing encrypted distance = ((CtD1 + CtD2 + CtD3 + CtD4...) - CtR)^2 ...")
	fmt.Println()

	evaluator.Neg(RiderCiphertext, RiderCiphertext)
	for i := 0; i < nbDrivers; i++ {
		evaluator.Add(RiderCiphertext, DriversCiphertexts[i], RiderCiphertext)
	}

	result := encoder.DecodeUintNew(decryptor.DecryptNew(evaluator.MulNew(RiderCiphertext, RiderCiphertext)))

	minIndex, minPosX, minPosY, minDist := 0, params.T(), params.T(), params.T()

	errors := 0

	for i := 0; i < nbDrivers; i++ {

		driverPosX, driverPosY := driversData[i][i<<1], driversData[i][(i<<1)+1]

		computedDist := result[i<<1] + result[(i<<1)+1]
		expectedDist := distance(driverPosX, driverPosY, riderPosX, riderPosY)

		if computedDist == expectedDist {
			if computedDist < minDist {
				minIndex = i
				minPosX, minPosY = driverPosX, driverPosY
				minDist = computedDist
			}
		} else {
			errors++
		}

		if i < 4 || i > nbDrivers-5 {
			fmt.Printf("Distance with Driver %d : %8d = (%4d - %4d)^2 + (%4d - %4d)^2 --> correct: %t\n",
				i, computedDist, driverPosX, riderPosX, driverPosY, riderPosY, computedDist == expectedDist)
		}

		if i == nbDrivers>>1 {
			fmt.Println("...")
		}
	}

	fmt.Printf("\nFinished with %.2f%% errors\n\n", 100*float64(errors)/float64(nbDrivers))
	fmt.Printf("Closest Driver to Rider is n°%d (%d, %d) with a distance of %d units\n",
		minIndex, minPosX, minPosY, int(math.Sqrt(float64(minDist))))
}

// InitLedger inserisce una serie di Service mock con cui interagire
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface, timestampMock string) error {
	assets := []Service{
		{ServiceID: "asset0", DriverID: "Driver91", RiderID: "Rider6", TimeStampServizio: timestampMock},
		{ServiceID: "asset1", DriverID: "Driver2", RiderID: "Rider19", TimeStampServizio: timestampMock},
		{ServiceID: "asset2", DriverID: "Driver41", RiderID: "Rider24", TimeStampServizio: timestampMock},
		{ServiceID: "asset3", DriverID: "Driver32", RiderID: "Rider24", TimeStampServizio: timestampMock},
		{ServiceID: "asset4", DriverID: "Driver14", RiderID: "Rider19", TimeStampServizio: timestampMock},
		{ServiceID: "asset5", DriverID: "Driver53", RiderID: "Rider19", TimeStampServizio: timestampMock},
		{ServiceID: "asset6", DriverID: "Driver6", RiderID: "Rider3", TimeStampServizio: timestampMock},
		{ServiceID: "asset7", DriverID: "Driver27", RiderID: "Rider87", TimeStampServizio: timestampMock},
		{ServiceID: "asset8", DriverID: "Driver18", RiderID: "Rider19", TimeStampServizio: timestampMock},
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

// CreateAsset inserisce un nuovo Service di servizio
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, assetID string, driverID string, riderID string) error {
	exists, err := s.AssetExists(ctx, assetID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", assetID)
	}

	asset := Service{
		ServiceID:         assetID,
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

// ReadAsset restituisce l'Service corrispondente all'@assetID fornito
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, assetID string) (*Service, error) {
	assetJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", assetID)
	}

	var asset Service
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset aggiorna lo stato di un Service di servizio con i nuovi parametri forniti.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, assetID string, driverID string, riderID string, timeStampServizio time.Time) error {
	exists, err := s.AssetExists(ctx, assetID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", assetID)
	}

	// overwriting original asset with new asset
	asset := Service{
		ServiceID:         assetID,
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

// DeleteAsset elimina l'Service richiesto
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

// AssetExists restituisce un booleano corrispondente all'esistenza dell'Service di servizio
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, assetID string) (bool, error) {

	assetJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferAsset aggiorna l'Service corrispondente a @assetID attribuendolo ad un nuovo Driver
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
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Service, error) {
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
