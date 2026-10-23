package main

import (
	_ "embed"
	"github.com/gin-gonic/gin"
	oasviewer "github.com/iseki0/go-oas-viewer"
	"log"
)

//go:embed openapi.yaml
var spec []byte

func main() {
	router := gin.Default()
	handler, err := oasviewer.New(oasviewer.Config{
		BasePath: "/docs",
		Title:    "示例 API",
		Spec:     oasviewer.Bytes(spec),
	})
	if err != nil {
		log.Fatal(err)
	}
	router.GET("/docs/*path", handler)
	router.HEAD("/docs/*path", handler)
	router.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })
	log.Fatal(router.Run("127.0.0.1:8080"))
}
