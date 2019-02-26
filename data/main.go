package data

import (
	"github.com/scalog/scalog/logger"
	"github.com/spf13/viper"
)

func Start() {
	logger.Printf("Data layer server %d started on %d\n", viper.Get("id"), viper.Get("port"))
}
