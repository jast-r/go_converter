package handlers

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func (h *Handler) convertVideo(ctx *gin.Context) {
	path := ctx.Param("path")
	go func() {
		start := time.Now()
		test := ffmpeg.Input(path)
		log.Println(path)
		err := test.Output("C://Users//Roman//Desktop//conv_test//zzz.mp4").OverWriteOutput().Run()
		if err != nil {
			log.Println(err)
		}
		log.Println(time.Since(start))
	}()
}
