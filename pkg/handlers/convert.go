package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type convertInput struct {
	Path   string `json:"path" binding:"required"`
	Status string `json:"status"`
}

const (
	statusAccepted       = "accepted"
	statusInProgress     = "in progress"
	statusInQueue        = "in queue"
	statusDone           = "conversion done"
	statusFailed         = "conversion failed"
	requestTemplate      = `{"status":"%s", "source_path": "%s", "output_path": "%s"}`
	requestErrorTemplate = `{"status":"%s", "error":"%s", "source_path": "%s", "output_path": "%s"}`
)

var (
	nextPath     string
	queueArray   []string
	mapConvArray = make(map[string]string)
	mutex        sync.RWMutex
)

func (h *Handler) convertVideo(ctx *gin.Context) {
	var input convertInput
	var err error

	if err = ctx.BindJSON(&input); err != nil {
		err = fmt.Errorf("source path can`t be empty")
		newErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	if _, err := os.Stat(os.Getenv("source_directory") + "/" + input.Path); err == nil {
		file := filepath.Base(input.Path)
		fileName := file[:strings.Index(file, ".")]
		outPath := fileName + ".mp4"

		mapConvArray[input.Path] = outPath
		ctx.JSON(http.StatusOK, map[string]string{
			"status":      statusAccepted,
			"output_path": outPath,
		})
		go handleRequest(input.Path, outPath, false)
		logrus.Printf("request %s accepted", input.Path)
		return
	} else if errors.Is(err, os.ErrNotExist) {
		newErrorResponse(ctx, http.StatusBadRequest, "file "+input.Path+" not exist")
		return
	} else {
		newErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}
}

func handleRequest(src_path, dst_path string, next bool) error {
	time.Sleep(1 * time.Second)
	worker_count, err := strconv.Atoi(os.Getenv("max_workers"))
	if err != nil {
		logrus.Fatal(err)
	}
	if (len(mapConvArray) < worker_count || next) && len(mapConvArray) > 0 {
		logrus.Printf("request %s in progress", src_path)
		go startConvertation(src_path, dst_path)
	} else {
		logrus.Println("request %s in queue", dst_path)
		strForPlatform := []byte(fmt.Sprintf(requestTemplate, statusInQueue, src_path, dst_path))
		if err := requestToPlatform([]byte(strForPlatform)); err != nil {
			logrus.Error(err)
		}
		queueArray = append(queueArray, src_path)
	}
	return nil
}

func requestToPlatform(request []byte) error {
	req, err := http.NewRequest("PATCH", os.Getenv("platform_endpoint"), bytes.NewBuffer(request))
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

func convertationFailed(err error, src_path, dst_path string) error {
	err = fmt.Errorf(requestErrorTemplate, statusFailed, err.Error(), src_path, dst_path)
	logrus.Error(err)
	strForPlatform := []byte(fmt.Sprintf(requestErrorTemplate, statusFailed, err.Error(), src_path, dst_path))
	if err = requestToPlatform([]byte(strForPlatform)); err != nil {
		newErrorResponse(&gin.Context{}, http.StatusInternalServerError, err.Error())
		return err
	}
	return nil
}

func startConvertation(src_path, dst_path string) {
	var err error
	if src_path == "" {
		err = fmt.Errorf("source path can`t be empty!")
		if reqErr := convertationFailed(err, src_path, dst_path); reqErr != nil {
			logrus.Error(reqErr)
		}
		return
	}
	if dst_path == "" {
		err = fmt.Errorf("output path can`t be empty!")
		if reqErr := convertationFailed(err, src_path, dst_path); reqErr != nil {
			logrus.Error(reqErr)
		}
		return
	}

	conv_path := os.Getenv("output_directory") + "/" + dst_path
	source_path := os.Getenv("source_directory") + "/" + src_path

	strForPlatform := []byte(fmt.Sprintf(requestTemplate, statusInProgress, src_path, dst_path))
	if err := requestToPlatform([]byte(strForPlatform)); err != nil {
		logrus.Error(err)
	}

	logrus.Printf("convertation start for %s", src_path)
	start := time.Now()
	converter := ffmpeg.Input(source_path)
	err = converter.Output(conv_path).OverWriteOutput().Run()

	mutex.Lock()
	defer mutex.Unlock()
	if len(mapConvArray) != 0 {
		delete(mapConvArray, src_path)
	}
	if len(queueArray) > 0 {
		nextPath, queueArray = queueArray[0], queueArray[1:]
		_, find := mapConvArray[nextPath]
		if find {
			handleRequest(nextPath, mapConvArray[nextPath], true)
		}
	}

	if err != nil {
		convertationFailed(err, src_path, dst_path)
	}
	logrus.Printf("convertation time for %s %s: ", src_path, time.Since(start).String())

	strForPlatform = []byte(fmt.Sprintf(requestTemplate, statusDone, src_path, dst_path))
	if err := requestToPlatform([]byte(strForPlatform)); err != nil {
		logrus.Error(err)
	}
}
