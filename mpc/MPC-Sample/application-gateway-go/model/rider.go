package model

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/tuneinsight/lattigo/v3/bfv"
	"github.com/tuneinsight/lattigo/v3/ring"
	"github.com/tuneinsight/lattigo/v3/rlwe"
	"github.com/tuneinsight/lattigo/v3/utils"
	"math"
	"math/bits"
	"strconv"
)

// Rider è la struttura contenente gli attributi necessari ad eseguire le operazioni di calcolo sicuro del Driver più vicino
type Rider struct {
	RiderID          string
	Params           bfv.Parameters
	publicKey        *rlwe.PublicKey
	secretKey        *rlwe.SecretKey
	decryptor        bfv.Decryptor
	encoder          bfv.Encoder
	EncryptorRiderPk bfv.Encryptor
	encryptorRiderSk bfv.Encryptor
	evaluator        bfv.Evaluator
	Prng             *utils.KeyedPRNG
	riderPosX        uint64
	riderPosY        uint64
}

// Istanzia un nuovo struct Rider
func NewRider(riderID string) Rider {

	// BFV parameters (128 bit security) with plaintext modulus 65929217
	paramDef := bfv.PN13QP218
	paramDef.T = 0x3ee0001

	params, err := bfv.NewParametersFromLiteral(paramDef)
	if err != nil {
		panic(err)
	}

	prng, err := utils.NewPRNG()
	if err != nil {
		panic(err)
	}

	color.Yellow("Parameters : N=%d, T=%d, Q = %d bits, sigma = %f \n",
		1<<params.LogN(), params.T(), params.LogQP(), params.Sigma())
	fmt.Println()

	// Rider's keygen
	kgen := bfv.NewKeyGenerator(params)
	riderSk, riderPk := kgen.GenKeyPair()

	encoder := bfv.NewEncoder(params)

	decryptor := bfv.NewDecryptor(params, riderSk)
	encryptorRiderPk := bfv.NewEncryptor(params, riderPk)
	encryptorRiderSk := bfv.NewEncryptor(params, riderSk)

	evaluator := bfv.NewEvaluator(params, rlwe.EvaluationKey{})

	return Rider{
		RiderID:          riderID,
		Params:           params,
		publicKey:        riderPk,
		secretKey:        riderSk,
		decryptor:        decryptor,
		encoder:          encoder,
		evaluator:        evaluator,
		Prng:             prng,
		EncryptorRiderPk: encryptorRiderPk,
		encryptorRiderSk: encryptorRiderSk,
	}
}

// Restituisce le coordinate (x,y) cifrate del Rider
// trattandosi di una demo, la logica di business di reperimento dei dati viene simulata mediante una scelta casuale di (x,y)
func (rider *Rider) GetCipheredPosition(nbDrivers int) bfv.Ciphertext {
	maxvalue := uint64(math.Sqrt(float64(rider.Params.T()))) // max values = floor(sqrt(plaintext modulus))
	mask := uint64(1<<bits.Len64(maxvalue) - 1)              // binary mask upper-bound for the uniform sampling

	color.Yellow("Generating %d driversData and 1 Rider randomly positioned on a grid of %d x %d units \n",
		nbDrivers, maxvalue, maxvalue)
	fmt.Println()

	// Rider coordinates [x, y, x, y, ....., x, y]
	rider.riderPosX, rider.riderPosY = ring.RandUniform(rider.Prng, maxvalue, mask), ring.RandUniform(rider.Prng, maxvalue, mask)

	Rider := make([]uint64, 1<<rider.Params.LogN())
	for i := 0; i < nbDrivers; i++ {
		Rider[(i << 1)] = rider.riderPosX
		Rider[(i<<1)+1] = rider.riderPosY
	}

	riderPlaintext := bfv.NewPlaintext(rider.Params)
	rider.encoder.Encode(Rider, riderPlaintext)
	riderCiphertext := rider.encryptorRiderSk.EncryptNew(riderPlaintext)

	return *riderCiphertext

}

// calcolo del Driver più vicino, tra tutti quelli forniti come argomento
func (rider *Rider) FindClosestDriver(nbDrivers int, riderCiphertext *bfv.Ciphertext, drivers *Drivers) {
	color.Yellow("Computing encrypted distance = ((CtD1 + CtD2 + CtD3 + CtD4...) - CtR)^2 ...")
	fmt.Println()

	rider.evaluator.Neg(riderCiphertext, riderCiphertext)
	for i := 0; i < nbDrivers; i++ {
		rider.evaluator.Add(riderCiphertext, drivers.DriverCipherTexts[i], riderCiphertext)
	}

	// result contiene le coppie (driverPosXi - riderPosX)^2 e (driverPosYi - riderPosY)^2 con i = 0...nbDrivers-1
	result := rider.encoder.DecodeUintNew(rider.decryptor.DecryptNew(rider.evaluator.MulNew(riderCiphertext, riderCiphertext)))

	minIndex, minPosX, minPosY, minDist := 0, rider.Params.T(), rider.Params.T(), rider.Params.T()

	errors := 0

	for i := 0; i < nbDrivers; i++ {

		driverPosX, driverPosY := drivers.DriversData[i][i<<1], drivers.DriversData[i][(i<<1)+1]

		computedDist := result[i<<1] + result[(i<<1)+1]
		expectedDist := distance(driverPosX, driverPosY, rider.riderPosX, rider.riderPosY)

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
			color.Yellow("Distance with Driver %d : %8d = (%4d - %4d)^2 + (%4d - %4d)^2 --> correct: %t\n",
				i, computedDist, driverPosX, rider.riderPosX, driverPosY, rider.riderPosY, computedDist == expectedDist)
		}

		if i == nbDrivers>>1 {
			color.Yellow("...")
		}
	}

	drivers.ClosestDriverID = "Driver" + strconv.Itoa(minIndex)

	color.Yellow("\nFinished with %.2f%% errors\n\n", 100*float64(errors)/float64(nbDrivers))
	color.Green("Closest Driver to %s is %s (%d, %d) with a distance of %d units\n", rider.RiderID, drivers.ClosestDriverID, minPosX, minPosY, int(math.Sqrt(float64(minDist))))

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
