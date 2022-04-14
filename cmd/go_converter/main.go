package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/gin-gonic/gin"
	goconverter "github.com/jast-r/go_converter"
	"github.com/jast-r/go_converter/pkg/handlers"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func checkSudo() {
	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()

	if err != nil {
		logrus.Fatal(err)
	}

	// 0 = root, 501 = non-root user
	i, err := strconv.Atoi(string(output[:len(output)-1]))

	if err != nil {
		logrus.Fatal(err)
	}

	if i == 0 {
		logrus.Println("Awesome! You are now running converter with root permissions!")
	} else {
		fmt.Println("This program must be run as root! (sudo)")
		logrus.Fatal("This program must be run as root! (sudo)")
	}
}

func main() {
	checkSudo()
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
