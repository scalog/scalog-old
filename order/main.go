package order

import (
	"fmt"

	"github.com/spf13/viper"
)

func Start() {
	fmt.Printf("Ordering layer started on %s\n", viper.Get("port"))
}
