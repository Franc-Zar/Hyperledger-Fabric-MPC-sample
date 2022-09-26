package model

import (
	"github.com/tuneinsight/lattigo/v3/bfv"
	"github.com/tuneinsight/lattigo/v3/ring"
	"math"
	"math/bits"
)

type Drivers struct {
	ClosestDriverID   string
	DriverCipherTexts []*bfv.Ciphertext
	DriversData       [][]uint64
}

// Restituisce una struct Driver, contenente le coordinate (x,y) cifrate di tutti i Driver vicini a Rider (in numero = @nbDrivers)
// trattandosi di una demo, la logica di business di reperimento dei dati viene simulata mediante una scelta casuale di (x,y)
func GetNearDrivers(rider Rider, nbDrivers int) Drivers {
	maxvalue := uint64(math.Sqrt(float64(rider.Params.T()))) // max values = floor(sqrt(plaintext modulus))
	mask := uint64(1<<bits.Len64(maxvalue) - 1)              // binary mask upper-bound for the uniform sampling

	// driversData coordinates [0, 0, ..., x, y, ..., 0, 0]
	driversData := make([][]uint64, nbDrivers)

	// generazione casuale dei dati di posizione
	driversPlaintexts := make([]*bfv.Plaintext, nbDrivers)
	for i := 0; i < nbDrivers; i++ {
		driversData[i] = make([]uint64, 1<<rider.Params.LogN())
		driversData[i][(i << 1)] = ring.RandUniform(rider.Prng, maxvalue, mask)
		driversData[i][(i<<1)+1] = ring.RandUniform(rider.Prng, maxvalue, mask)
		driversPlaintexts[i] = bfv.NewPlaintext(rider.Params)
		rider.encoder.Encode(driversData[i], driversPlaintexts[i])
	}

	// generazione cifrato, mediante la chiave pubblica del Rider, contenente le coordinate dei Drivers
	DriversCiphertexts := make([]*bfv.Ciphertext, nbDrivers)
	for i := 0; i < nbDrivers; i++ {
		DriversCiphertexts[i] = rider.EncryptorRiderPk.EncryptNew(driversPlaintexts[i])
	}

	return Drivers{DriverCipherTexts: DriversCiphertexts, DriversData: driversData}

}
