package order

import (
	"github.com/scalog/logger"
	"github.com/spf13/viper"
)

func Start() {
	logger.Printf("Ordering layer server %d started on %d\n", viper.Get("id"), viper.Get("port"))
}
