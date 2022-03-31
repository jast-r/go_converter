package main

import (
	"os"

	goconverter "github.com/jast-r/go_converter"
	"github.com/jast-r/go_converter/pkg/handlers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logrus.New()
	logrus.SetFormatter(new(logrus.JSONFormatter))
	logrus.SetOutput(os.Stderr)

	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializing configs: %s", err.Error())
	}

	handler := new(handlers.Handler)
	server := new(goconverter.Server)

	err := server.Run(viper.GetString("port"), handler.InitRoutes())
	if err != nil {
		logrus.Fatal(err)
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
