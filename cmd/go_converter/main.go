package main

import (
	"log"

	goconverter "github.com/jast-r/go_converter"
	"github.com/jast-r/go_converter/pkg/handlers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializing configs: %s", err.Error())
	}

	handler := new(handlers.Handler)

	server := new(goconverter.Server)
	err := server.Run(viper.GetString("port"), handler.InitRoutes())
	if err != nil {
		log.Fatalln(err)
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
