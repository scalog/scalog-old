package discovery

import (
	"github.com/scalog/scalog/logger"
	"github.com/spf13/viper"
)

// Start serving requests for the data layer
func Start() {
	logger.Printf("Discovery server starting on %d\n", viper.Get("port"))
}
