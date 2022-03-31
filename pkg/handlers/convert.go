package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type convertInput struct {
	Path   string `json:"path" binding:"required"`
	Status string `json:"status"`
}

var (
	convertArray     []string
	sourceArray      []string
	idsArray         []string
	statusInProgress = "in progress"
	statusInQueue    = "in queue"
	statusDone       = "convertation done"
	requestTemplate  = `{"status":"%s", "output_path": "%s"}`
	mutex            sync.RWMutex
)

func (h *Handler) convertVideo(ctx *gin.Context) {
	var input convertInput
	var stringForPlatformReq []byte
	var currentPath string
	var err error

	if err = ctx.BindJSON(&input); err != nil {
		err = fmt.Errorf("source path can`t be empty")
		newErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	id := ctx.Param("id")
	file := filepath.Base(input.Path)
	fileName := file[:strings.Index(file, ".")]
	converter := ffmpeg.Input(input.Path)
	outPath := viper.GetString("output_directory") + "/" + id + fileName + ".mp4"
	fmt.Println(viper.GetString("output_directory"))

	idsArray = append(idsArray, id)
	convertArray = append(convertArray, outPath)
	sourceArray = append(sourceArray, input.Path)

	fmt.Println(len(convertArray))
	if len(convertArray) < 3 || input.Status == statusDone {
		go func() {
			mutex.Lock()
			currentPath, convertArray = convertArray[0], convertArray[1:]
			mutex.Unlock()

			if _, err := os.Stat(currentPath); err == os.ErrNotExist {
				err = fmt.Errorf("%s does`t exist", currentPath)
				newErrorResponse(ctx, http.StatusInternalServerError, err.Error())
				return
			}

			start := time.Now()
			err = converter.Output(currentPath).OverWriteOutput().Run()
			if err != nil {
				newErrorResponse(ctx, http.StatusInternalServerError, err.Error())
				return
			}
			stringForPlatformReq = []byte(fmt.Sprintf(requestTemplate, statusDone, currentPath))
			if err := requestToPlatform(stringForPlatformReq); err != nil {
				newErrorResponse(ctx, http.StatusInternalServerError, err.Error())
				return
			}
			logrus.Println(time.Since(start))

			// Посылаем запрос на обработку следующего элемента очереди
			mutex.Lock()
			if len(convertArray) > 2 {
				stringForQueueReq := []byte(fmt.Sprintf(`{"status":"%s", "path": "%s"}`, statusDone, sourceArray[0]))
				url := fmt.Sprintf("http://localhost:%s/api/convert/%s", viper.GetString("port"), idsArray[0])
				http.Post(url, "application/json", bytes.NewBuffer(stringForQueueReq))
			}
			fmt.Println(idsArray, convertArray)
			idsArray = idsArray[1:]
			sourceArray = convertArray[1:]
			mutex.Unlock()
		}()
	} else {
		stringForPlatformReq = []byte(fmt.Sprintf(requestTemplate, statusInQueue, outPath))
		if err := requestToPlatform(stringForPlatformReq); err != nil {
			newErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		}
	}
}

func requestToPlatform(request []byte) error {
	fmt.Println(string(request))
	req, err := http.NewRequest("PATCH", viper.GetString("platform_endpoint"), bytes.NewBuffer(request))
	if err != nil {
		return err
	}
	req.Header.Set("Content-type", "application/json")
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return nil
}
