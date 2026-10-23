// Demo server: go run ./examples/demo -addr 127.0.0.1:18080
package main

import (
	"crypto/rand"
	"crypto/subtle"
	"flag"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	oasviewer "github.com/iseki0/go-oas-viewer"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:18080", "demo HTTP 监听地址")
	flag.Parse()
	if info, ok := debug.ReadBuildInfo(); ok {
		log.Printf("Go build info:\n%s", info.String())
	} else {
		log.Print("Go build info: unavailable")
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Fatal(err)
	}

	// 仅演示 Cookie 会话：不验证账号口令。重启服务后旧会话失效。
	token := rand.Text()
	signedIn := func(c *gin.Context) bool {
		value, err := c.Cookie("oas_viewer_demo")
		return err == nil && subtle.ConstantTimeCompare([]byte(value), []byte(token)) == 1
	}
	r.GET("/", func(c *gin.Context) { c.Data(200, "text/html; charset=utf-8", []byte(home)) })
	r.GET("/demo/toolbar.js", func(c *gin.Context) { c.Data(200, "text/javascript; charset=utf-8", []byte(toolbarJS)) })
	r.GET("/demo/toolbar.css", func(c *gin.Context) { c.Data(200, "text/css; charset=utf-8", []byte(toolbarCSS)) })
	r.GET("/demo/openapi.yaml", func(c *gin.Context) { c.Data(200, "application/yaml; charset=utf-8", []byte(spec)) })

	ext := &oasviewer.Extension{
		ModuleURL:      "/demo/toolbar.js",
		StylesheetURLs: []string{"/demo/toolbar.css"},
	}
	for _, item := range []struct {
		path   string
		source oasviewer.Source
	}{
		{"/docs", oasviewer.Bytes([]byte(spec))},
		{"/url-docs", oasviewer.URL("../demo/openapi.yaml")},
	} {
		handler, err := oasviewer.New(oasviewer.Config{
			BasePath: item.path, Title: "go-oas-viewer Demo", Spec: item.source, Extension: ext,
		})
		if err != nil {
			log.Fatal(err)
		}
		r.GET(item.path+"/*path", handler)
		r.HEAD(item.path+"/*path", handler)
	}
	r.GET("/api/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })
	r.POST("/api/echo", func(c *gin.Context) {
		var input struct {
			Message string `json:"message" binding:"required"`
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024*1024)
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "请提交包含非空 message 的 JSON"})
			return
		}
		c.JSON(200, gin.H{"message": input.Message})
	})
	r.GET("/api/me", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		if !signedIn(c) {
			c.JSON(401, gin.H{"error": "尚未建立演示会话，请点击右上角演示登录"})
			return
		}
		c.JSON(200, gin.H{"user": "demo", "message": "HttpOnly Cookie 会话有效"})
	})
	r.POST("/api/demo-session", func(c *gin.Context) {
		http.SetCookie(c.Writer, &http.Cookie{Name: "oas_viewer_demo", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
		c.Status(204)
	})
	r.DELETE("/api/demo-session", func(c *gin.Context) {
		http.SetCookie(c.Writer, &http.Cookie{Name: "oas_viewer_demo", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
		c.Status(204)
	})
	server := &http.Server{Addr: *addr, Handler: r, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Demo: http://%s/", *addr)
	log.Fatal(server.ListenAndServe())
}

const home = `<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>go-oas-viewer Demo</title>
<style>body{max-width:900px;margin:64px auto;padding:0 24px;font:16px/1.7 system-ui;color:#263f5d;background:#fafbfc}h1{font-size:32px}section{background:white;border:1px solid #d8dee6;padding:20px 24px;margin:20px 0}a{color:#215cad;margin-right:24px}code{background:#edf1f5;padding:2px 6px}li{margin:10px 0}</style></head><body>
<h1>go-oas-viewer Demo</h1><p>三个查看器共用一份 OpenAPI YAML，所有查看器静态资源均由 Go 内嵌提供。</p>
<section><h2>内存 Bytes 来源</h2><p><a href="/docs/">Scalar</a><a href="/docs/elements.html">Stoplight Elements</a><a href="/docs/swagger.html">Swagger UI</a></p><p>规格由库提供：<a href="/docs/openapi.yaml">openapi.yaml</a></p></section>
<section><h2>相对 URL 来源</h2><p><a href="/url-docs/">Scalar</a><a href="/url-docs/elements.html">Stoplight Elements</a><a href="/url-docs/swagger.html">Swagger UI</a></p><p>配置为 <code>../demo/openapi.yaml</code>，由 demo server 单独托管。</p></section>
<section><h2>手动验证</h2><ol><li>打开任一查看器，检查顶部切换链接和右上角工具栏。</li><li>在 Try it 中请求 <code>GET /api/ping</code>，应返回 pong。</li><li>请求 <code>POST /api/echo</code>，提交 JSON，检查回显。</li><li>未登录时请求 <code>GET /api/me</code> 应返回 401。</li><li>点击右上角“演示登录”，再请求 /api/me 应返回 200。</li><li>切换查看器或刷新，演示会话应保持；点击退出后 /api/me 应恢复 401。</li></ol><p>演示登录不验证账号口令，仅用于验证扩展模块、HttpOnly Cookie 和查看器请求的配合。无需输入任何真实凭证。</p></section>
</body></html>`

const toolbarCSS = `#viewer-toolbar{gap:10px}#viewer-toolbar a{color:#263f5d;text-decoration:none}#viewer-toolbar button{height:30px;padding:0 12px;background:white;border:1px solid #a9b5c4;color:#263f5d;cursor:pointer;font:inherit}#viewer-toolbar span{font-size:12px;white-space:nowrap}`

const toolbarJS = `export async function mount({toolbar,viewer}) {
  const home=document.createElement('a');home.href='/';home.textContent='Demo 首页';
  const status=document.createElement('span');status.setAttribute('role','status');status.setAttribute('aria-live','polite');
  const login=document.createElement('button');login.type='button';login.textContent='演示登录';
  const logout=document.createElement('button');logout.type='button';logout.textContent='退出';
  toolbar.append(home,status,login,logout);toolbar.dataset.viewer=viewer;
  async function refresh(){
    const response=await fetch('/api/me',{credentials:'include',cache:'no-store',signal:AbortSignal.timeout(5000)});
    if(response.status!==200&&response.status!==401)throw new Error('会话查询失败');
    const active=response.status===200;status.textContent=active?'demo 已登录':'未登录';login.hidden=active;logout.hidden=!active;
  }
  async function change(method){
    login.disabled=logout.disabled=true;
    try{const response=await fetch('/api/demo-session',{method,credentials:'include',signal:AbortSignal.timeout(5000)});if(!response.ok)throw new Error('会话操作失败');await refresh();}
    catch(error){status.textContent=error.message;}
    finally{login.disabled=logout.disabled=false;}
  }
  login.addEventListener('click',()=>change('POST'));logout.addEventListener('click',()=>change('DELETE'));
  try{await refresh();}catch(error){status.textContent=error.message;}
}`

const spec = `openapi: 3.0.3
info:
  title: go-oas-viewer Demo API
  version: 1.0.0
  description: 用于验证三个查看器、JSON 请求回显，以及右上角扩展与 Cookie 会话。
servers:
  - url: /
paths:
  /api/ping:
    get:
      summary: 连通性检查
      operationId: ping
      responses:
        '200':
          description: pong
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Message'
              example:
                message: pong
  /api/echo:
    post:
      summary: 回显 JSON 消息
      operationId: echo
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Message'
            example:
              message: 你好，API Viewer
      responses:
        '200':
          description: 原样返回消息
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Message'
        '400':
          description: JSON 无效或缺少 message
  /api/me:
    get:
      summary: 查看当前演示会话
      description: 先点击页面右上角的演示登录，再发送请求。浏览器自动发送 HttpOnly Cookie，无需手动填写 Cookie。
      operationId: me
      responses:
        '200':
          description: 已登录
          content:
            application/json:
              schema:
                type: object
                properties:
                  user:
                    type: string
                  message:
                    type: string
        '401':
          description: 未登录
components:
  schemas:
    Message:
      type: object
      required: [message]
      properties:
        message:
          type: string
          minLength: 1
`
