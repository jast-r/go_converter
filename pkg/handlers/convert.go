package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type convertInput struct {
	Path string `json:"path" binding:"required"`
}

var (
	convertArray     []string
	statusInProgress = "in progress"
	statusInQueue    = "in queue"
	statusDone       = "convertation done"
	requestTemplate  = `{"status":"%s", "output_path": "%s"}`
)

func (h *Handler) convertVideo(ctx *gin.Context) {
	var input convertInput
	var stringForReq []byte
	if err := ctx.BindJSON(&input); err != nil {
		err = fmt.Errorf("source path can`t be empty")
		newErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	file := filepath.Base(input.Path)
	fileName := file[:strings.Index(file, ".")]
	converter := ffmpeg.Input(input.Path)
	outPath := viper.GetString("output_directory") + "/" + fileName + ".mp4"
	fmt.Println(viper.GetString("output_directory"))

	if len(convertArray) < 2 {
		go func() {
			start := time.Now()
			convertArray = append(convertArray, outPath)
			err := converter.Output(outPath).OverWriteOutput().Run()
			if err != nil {
				logrus.Println(err)
				return
			}
			stringForReq = []byte(fmt.Sprintf(requestTemplate, statusDone, outPath))
			_, err = http.NewRequest("PATCH", viper.GetString("platform_endpoint"), bytes.NewBuffer(stringForReq))
			if err != nil {
				logrus.Println(err)
				return
			}
			logrus.Println(time.Since(start))
		}()
	} else {
		stringForReq = []byte(fmt.Sprintf(requestTemplate, statusInQueue, outPath))
		_, err := http.NewRequest("PATCH", viper.GetString("platform_endpoint"), bytes.NewBuffer(stringForReq))
		if err != nil {
			logrus.Println(err)
			return
		}
	}
}
