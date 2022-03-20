package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type convertInput struct {
	Path   string `json:"path" binding:"required"`
	Output string `json:"out" binding:"required"`
}

func (h *Handler) convertVideo(ctx *gin.Context) {
	var input convertInput
	var outPath string
	if err := ctx.BindJSON(&input); err != nil {
		log.Println(err)
	}
	go func() {
		start := time.Now()
		separatedPath := strings.Split(input.Path, string(os.PathSeparator))
		fmt.Println(separatedPath[len(separatedPath)-1])
		test := ffmpeg.Input(input.Path)
		log.Println(input.Path)
		fileName := separatedPath[len(separatedPath)-1]
		outPath = input.Output + "/" + fileName[:strings.LastIndex(fileName, ".")] + ".mp4"
		err := test.Output(outPath).OverWriteOutput().Run()
		if err != nil {
			log.Println(err)
		}
		log.Println(time.Since(start))
	}()

	ctx.JSON(http.StatusOK, map[string]string{
		"out_path": outPath,
	})
}
