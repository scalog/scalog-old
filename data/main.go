package data

import (
	"github.com/scalog/scalog/logger"
	"github.com/spf13/viper"
)

func Start() {
	logger.Printf("Data layer started on %d\n", viper.Get("port"))
}
