package oasviewer

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"path"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:web
var embedded embed.FS

// Source 表示由 URL 或 Bytes 构造的 OAS 文档来源；零值无效。
type Source struct {
	url  string
	data []byte
	kind uint8
}

// URL 指定浏览器加载的 OAS URL，支持 HTTP(S)、根相对和目录相对 URL。
// 相对 URL 以查看器目录为基准；Go 服务不下载此 URL。
func URL(url string) Source { return Source{url: url, kind: 1} }

// Bytes 指定 YAML 或 JSON 文档。New 会复制内容，调用前不能并发修改 data。
// 库不验证或转换文档；相对 $ref 由浏览器相对于 openapi.yaml 解析。
func Bytes(data []byte) Source { return Source{data: data, kind: 2} }

// Extension 指定接入方提供的前端扩展资源，均由接入方自行托管。
type Extension struct {
	// ModuleURL 是 ES 模块 URL，模块必须导出异步或同步 mount(context) 函数。
	// context 包含 toolbar (HTMLElement)、viewer (scalar/elements/swagger)、specURL (绝对 URL)。
	// 三个页面每次加载分别调用一次；mount 完成后才初始化查看器。
	ModuleURL string
	// StylesheetURLs 按顺序插入页面 head，位于查看器内置样式之后。
	// 相对 URL 以查看器目录为基准，CSS 和模块均受浏览器 CORS/CSP 规则约束。
	StylesheetURLs []string
}

// Config 配置查看器页面及 OAS 来源，New 不保存调用方的可变切片。
type Config struct {
	// BasePath 是公开挂载路径，例如 /docs；必填，根路径使用 /。
	// 实际 Gin 路由可以在反向代理剥离前缀后使用不同的内部路径。
	BasePath string
	Spec     Source
	// Title 为空时使用 API 文档，内容按纯文本转义。
	Title     string
	Extension *Extension
}

type pageData struct {
	Title       string
	Config      string
	Stylesheets []string
}

type clientConfig struct {
	BuildInfo *debug.BuildInfo `json:"buildInfo,omitempty"`
	SpecURL   string           `json:"specURL"`
	ModuleURL string           `json:"moduleURL,omitempty"`
}

// New 校验配置，预渲染页面，返回可并发使用的 Gin handler。
// 挂载到 GET /docs/*path（通配符名称必须为 path），可额外挂载 HEAD 同一路由。
// 配置错误返回 error；请求未知资源返回 404，不读取磁盘、不列出目录。
func New(config Config) (gin.HandlerFunc, error) {
	base := config.BasePath
	if base == "" || !strings.HasPrefix(base, "/") || strings.HasPrefix(base, "//") || strings.ContainsAny(base, "?#\\%\r\n\t ") || path.Clean(base) != strings.TrimSuffix(base, "/") && base != "/" {
		return nil, fmt.Errorf("oasviewer: BasePath must be a clean absolute URL path")
	}
	base = strings.TrimSuffix(base, "/")
	var spec []byte
	buildInfo, _ := debug.ReadBuildInfo()
	client := clientConfig{BuildInfo: buildInfo}
	switch config.Spec.kind {
	case 1:
		if err := validateURL(config.Spec.url); err != nil {
			return nil, fmt.Errorf("oasviewer: Spec: %w", err)
		}
		client.SpecURL = config.Spec.url
	case 2:
		if len(config.Spec.data) == 0 {
			return nil, fmt.Errorf("oasviewer: empty Spec bytes")
		}
		spec = bytes.Clone(config.Spec.data)
		client.SpecURL = "./openapi.yaml"
	default:
		return nil, fmt.Errorf("oasviewer: Spec must be constructed with URL or Bytes")
	}
	data := pageData{Title: config.Title}
	if data.Title == "" {
		data.Title = "API 文档"
	}
	if ext := config.Extension; ext != nil {
		if ext.ModuleURL != "" {
			if err := validateURL(ext.ModuleURL); err != nil {
				return nil, fmt.Errorf("oasviewer: ModuleURL: %w", err)
			}
			client.ModuleURL = ext.ModuleURL
		}
		data.Stylesheets = append([]string(nil), ext.StylesheetURLs...)
		for _, value := range data.Stylesheets {
			if err := validateURL(value); err != nil {
				return nil, fmt.Errorf("oasviewer: StylesheetURLs: %w", err)
			}
		}
	}
	encoded, err := json.Marshal(client)
	if err != nil {
		return nil, fmt.Errorf("oasviewer: config: %w", err)
	}
	data.Config = string(encoded)
	pages := make(map[string][]byte, 3)
	for _, name := range []string{"index.html", "elements.html", "swagger.html"} {
		source, err := embedded.ReadFile("web/" + name)
		if err != nil {
			return nil, err
		}
		t, err := template.New(name).Parse(strings.Replace(string(source), "</head>", `{{range .Stylesheets}}<link rel="stylesheet" href="{{.}}">{{end}}</head>`, 1))
		if err != nil {
			return nil, err
		}
		var output bytes.Buffer
		if err := t.Execute(&output, data); err != nil {
			return nil, err
		}
		pages[name] = output.Bytes()
	}
	assets, err := fs.Sub(embedded, "web")
	if err != nil {
		return nil, err
	}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Header("Allow", "GET, HEAD")
			c.Status(http.StatusMethodNotAllowed)
			return
		}
		resource := c.Param("path")
		if resource == "" {
			c.Redirect(http.StatusTemporaryRedirect, base+"/")
			return
		}
		name := strings.TrimPrefix(resource, "/")
		if name == "" {
			name = "index.html"
		}
		if body, ok := pages[name]; ok {
			c.Header("Cache-Control", "no-cache")
			write(c, "text/html; charset=utf-8", body)
			return
		}
		if name == "openapi.yaml" && spec != nil {
			c.Header("Cache-Control", "no-cache")
			contentType := "application/yaml; charset=utf-8"
			if json.Valid(spec) {
				contentType = "application/json; charset=utf-8"
			}
			write(c, contentType, spec)
			return
		}
		if !strings.HasPrefix(name, "assets/") || !fs.ValidPath(name) || strings.Contains(name, "\\") {
			c.Status(http.StatusNotFound)
			return
		}
		body, err := fs.ReadFile(assets, name)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		if strings.HasSuffix(name, ".txt") {
			c.Header("Cache-Control", "no-cache")
		} else {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}
		contentType := mime.TypeByExtension(path.Ext(name))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		write(c, contentType, body)
	}, nil
}

func write(c *gin.Context, contentType string, body []byte) {
	c.Header("X-Content-Type-Options", "nosniff")
	if c.Request.Method == http.MethodHead {
		c.Header("Content-Type", contentType)
		c.Header("Content-Length", fmt.Sprint(len(body)))
		c.Status(http.StatusOK)
		return
	}
	c.Data(http.StatusOK, contentType, body)
}

func validateURL(value string) error {
	u, err := url.Parse(value)
	if err != nil || strings.TrimSpace(value) != value || value == "" || strings.ContainsAny(value, "\r\n\\") {
		return fmt.Errorf("invalid URL %q", value)
	}
	if u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https" || u.Scheme != "" && u.Host == "" || u.User != nil || strings.HasPrefix(value, "//") {
		return fmt.Errorf("only HTTP(S) or relative URLs are supported: %q", value)
	}
	return nil
}
