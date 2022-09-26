package utilities

import (
	"github.com/google/uuid"
	"math/rand"
	"strconv"
	"time"
)

// metodo mock di comunicazione e instaurazione servizio tra @riderID e @driverID: restituisce i dati necessari a inserire un Service nel Ledger,
// ossia le informazioni relative al servizio stesso (nella logica demo l'id univoco di transazione e l'importo della Fare)
func SetRide(riderID string, driverID string) (string, string) {

	rand.Seed(time.Now().UnixNano())

	return uuid.New().String(), strconv.Itoa(rand.Intn(100-5)+5) + "€"

}
