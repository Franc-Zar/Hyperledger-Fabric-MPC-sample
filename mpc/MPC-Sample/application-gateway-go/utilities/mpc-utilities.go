package utilities

import (
	"assetTransfer/model"
	"fmt"
	"github.com/fatih/color"
	"time"
)

func ObliviousRideMatching(nbDrivers int, riderID string) (string, string, bool) {
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

	startOblivious := time.Now()

	color.Yellow("============================================")
	color.Yellow("Homomorphic computations on batched integers")
	color.Yellow("============================================")
	fmt.Println()

	isClosestDriverFound := false

	start := time.Now()

	rider := model.NewRider(riderID)

	duration := time.Since(start)
	color.Cyan("setup time: %s", duration)

	start = time.Now()
	// generazione testo cifrato contenente la posizione del Rider
	riderCiphertext := rider.GetCipheredPosition(nbDrivers)

	duration = time.Since(start)
	color.Cyan("Random rider generation time: %s", duration)

	start = time.Now()
	// ricerca dei Driver vicini al Rider
	drivers := model.GetNearDrivers(rider, nbDrivers)

	duration = time.Since(start)
	color.Cyan("Random drivers generation time: %s", duration)

	start = time.Now()
	// ricerca del Driver più vicino al Rider: operazione di calcolo della distanza (calcolo sulle posizioni Homomorphic Encrypted)
	rider.FindClosestDriver(nbDrivers, &riderCiphertext, &drivers)

	duration = time.Since(start)
	color.Cyan("finding closest driver time: %s", duration)

	if drivers.ClosestDriverID != "" {
		isClosestDriverFound = true
	}

	duration = time.Since(startOblivious)
	color.Cyan("ObliviousRideMatching execution time: %s", duration)

	return drivers.ClosestDriverID, time.Now().String(), isClosestDriverFound

}
