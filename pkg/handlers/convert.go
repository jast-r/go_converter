package handlers

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
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
	queueArray       []string
	mapConvArray     = make(map[string]string)
	pathArray        = make(chan map[string]string)
	statusInProgress = "in progress"
	statusInQueue    = "in queue"
	statusDone       = "convertation done"
	requestTemplate  = `{"status":"%s", "output_path": "%s"}`
	mutex            sync.RWMutex
)

func init() {
	rand.Seed(time.Now().UnixNano())

	// var val map[string]string
	// var open bool
	// var stringForPlatformReq []byte
	// var counter int
	// var outPath string
	// counter = 0

	// go func() {
	// 	for {

	// 	}
	// }()
}

func (h *Handler) convertVideo(ctx *gin.Context) {
	var input convertInput
	var err error

	if err = ctx.BindJSON(&input); err != nil {
		err = fmt.Errorf("source path can`t be empty")
		newErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	fmt.Println("hello")

	// id := ctx.Param("id")

	file := filepath.Base(input.Path)
	fileName := file[:strings.Index(file, ".")]
	outPath := viper.GetString("output_directory") + "/" + fileName + ".mp4"
	// sourceArray = append(sourceArray, input.Path)
	// convertArray = append(convertArray, outPath)
	mapConvArray[input.Path] = outPath
	handleRequest(input.Path, outPath, false)
}

func handleRequest(src_path, dst_path string, next bool) {
	if (len(mapConvArray) < 3 || next) && len(mapConvArray) > 0 {
		strForPlatform := []byte(fmt.Sprintf(requestTemplate, statusInProgress, dst_path))
		if err := requestToPlatform([]byte(strForPlatform)); err != nil {
			newErrorResponse(&gin.Context{}, http.StatusInternalServerError, err.Error())
			return
		}
		logrus.Println(src_path, dst_path)
		go startConvertation(src_path, dst_path)
	} else {
		strForPlatform := []byte(fmt.Sprintf(requestTemplate, statusInQueue, dst_path))
		if err := requestToPlatform([]byte(strForPlatform)); err != nil {
			newErrorResponse(&gin.Context{}, http.StatusInternalServerError, err.Error())
			return
		}
		queueArray = append(queueArray, src_path)
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

func startConvertation(src_path, dst_path string) error {
	start := time.Now()
	converter := ffmpeg.Input(src_path)
	err := converter.Output(dst_path).OverWriteOutput().Run()
	if err != nil {
		logrus.Error(err)
	}
	logrus.Println("convertation time: " + time.Since(start).String())

	strForPlatform := []byte(fmt.Sprintf(requestTemplate, statusDone, dst_path))
	if err := requestToPlatform([]byte(strForPlatform)); err != nil {
		newErrorResponse(nil, http.StatusInternalServerError, err.Error())
		return err
	}

	if len(queueArray) > 0 {
		var next string
		mutex.Lock()
		logrus.Println(queueArray)
		next = queueArray[0]
		queueArray = queueArray[1:]
		delete(mapConvArray, dst_path)
		mutex.Unlock()
		handleRequest(next, mapConvArray[next], true)
	}

	return nil
}
