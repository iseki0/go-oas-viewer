package main

import (
	"embed"
	"github.com/gin-gonic/gin"
	oasviewer "github.com/iseki0/go-oas-viewer"
	"log"
	"net/http"
)

//go:embed public/*
var public embed.FS

func main() {
	router := gin.Default()
	// 扩展由调用方托管；此路径独立于 /docs/*path。
	router.StaticFS("/custom", http.FS(public))
	handler, err := oasviewer.New(oasviewer.Config{
		BasePath: "/docs",
		Title:    "带登录扩展的 API",
		Spec:     oasviewer.URL("/custom/public/openapi.yaml"),
		Extension: &oasviewer.Extension{
			ModuleURL:      "/custom/public/login.js",
			StylesheetURLs: []string{"/custom/public/login.css"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	router.GET("/docs/*path", handler)
	// 演示登录跳转：实际业务提供这个端点并在验证成功后设置 HttpOnly Cookie。
	// 不包含 mock 登录成功、账号口令或生产认证实现。
	router.GET("/login", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", []byte("<p>在此接入业务的登录页或 SSO，再返回 /docs/。</p><a href='/docs/'>返回文档</a>"))
	})
	router.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })
	log.Fatal(router.Run("127.0.0.1:8080"))
}
