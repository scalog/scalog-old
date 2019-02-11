package order

import (
	"log"

	"github.com/spf13/viper"
)

func Start() {
	log.Printf("Ordering layer started on %d\n", viper.Get("port"))
}
