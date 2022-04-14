package main

import (
	"os"

	"github.com/gin-gonic/gin"
	goconverter "github.com/jast-r/go_converter"
	"github.com/jast-r/go_converter/pkg/handlers"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))
	gin.SetMode(gin.ReleaseMode)
	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error loading env variables: %s", err.Error())
	}

	file, err := os.OpenFile(os.Getenv("log_path"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logrus.SetOutput(os.Stderr)
	}
	logrus.SetOutput(file)

	handler := new(handlers.Handler)
	server := new(goconverter.Server)

	err = server.Run(os.Getenv("port"), handler.InitRoutes())
	if err != nil {
		logrus.Fatal(err)
	}
}
