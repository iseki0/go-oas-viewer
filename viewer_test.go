package oasviewer

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"html"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func routerFor(t *testing.T, config Config) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	handler, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	route := strings.TrimSuffix(config.BasePath, "/") + "/*path"
	r.GET(route, handler)
	r.HEAD(route, handler)
	return r
}

func request(r http.Handler, method, target string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(method, target, nil))
	return w
}

func TestPagesAndEmbeddedAssets(t *testing.T) {
	r := routerFor(t, Config{BasePath: "/docs", Spec: URL("./spec.yaml"), Title: `<script>alert("x")</script>`, Extension: &Extension{ModuleURL: "/custom/login.js", StylesheetURLs: []string{"/custom/login.css"}}})
	for _, page := range []string{"index.html", "elements.html", "swagger.html"} {
		w := request(r, "GET", "/docs/"+page)
		body := w.Body.String()
		if w.Code != 200 || strings.Contains(body, `<script>alert`) || !strings.Contains(body, "&lt;script&gt;") {
			t.Fatalf("page %s: %d %s", page, w.Code, body)
		}
		if strings.Count(body, `aria-current="page"`) != 1 || !strings.Contains(body, `href="./`+page+`" aria-current="page"`) {
			t.Fatalf("bad navigation: %s", body)
		}
		if !strings.Contains(body, `href="/custom/login.css"`) || !strings.Contains(body, `id="viewer-toolbar"`) {
			t.Fatal("missing extension hooks")
		}
		match := regexp.MustCompile(`data-config="([^"]+)"`).FindStringSubmatch(body)
		if len(match) != 2 {
			t.Fatal("missing config")
		}
		var cfg clientConfig
		if err := json.Unmarshal([]byte(html.UnescapeString(match[1])), &cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.SpecURL != "./spec.yaml" || cfg.ModuleURL != "/custom/login.js" {
			t.Fatalf("config: %+v", cfg)
		}
		assets := regexp.MustCompile(`(?:src|href)="(\./assets/[^"]+)"`).FindAllStringSubmatch(body, -1)
		if len(assets) == 0 {
			t.Fatal("no built assets")
		}
		for _, asset := range assets {
			result := request(r, "GET", "/docs/"+strings.TrimPrefix(asset[1], "./"))
			if result.Code != 200 || result.Body.Len() == 0 {
				t.Fatalf("asset %s: %d", asset[1], result.Code)
			}
			if !strings.Contains(result.Header().Get("Cache-Control"), "immutable") {
				t.Fatal("asset not cached")
			}
		}
	}
	if w := request(r, "GET", "/docs"); w.Code != 301 || w.Header().Get("Location") != "/docs/" {
		t.Fatalf("redirect: %d %s", w.Code, w.Header())
	}
	if w := request(r, "GET", "/docs/"); w.Code != 200 {
		t.Fatal(w.Code)
	}
	for _, name := range []string{"missing.html", "openapi.yaml", "assets/", "assets/../index.html", "../viewer.go"} {
		if w := request(r, "GET", "/docs/"+name); w.Code != 404 {
			t.Fatalf("%s: %d", name, w.Code)
		}
	}
}

func TestBytesSnapshotHEADAndRoot(t *testing.T) {
	source := []byte(`{"openapi":"3.0.3"}`)
	r := routerFor(t, Config{BasePath: "/", Spec: Bytes(source)})
	source[0] = 'x'
	w := request(r, "GET", "/openapi.yaml")
	if w.Code != 200 || !json.Valid(w.Body.Bytes()) || !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("snapshot: %d %s", w.Code, w.Body.String())
	}
	head := request(r, "HEAD", "/openapi.yaml")
	if head.Code != 200 || head.Body.Len() != 0 || head.Header().Get("Content-Length") == "" {
		t.Fatal("HEAD failed")
	}
	for _, page := range []string{"/", "/index.html", "/elements.html", "/swagger.html"} {
		if result := request(r, "GET", page); result.Code != 200 {
			t.Fatalf("%s: %d", page, result.Code)
		}
	}
}

func TestInvalidConfig(t *testing.T) {
	for _, config := range []Config{
		{BasePath: "/docs"}, {BasePath: "docs", Spec: URL("/s")},
		{BasePath: "/docs/../x", Spec: URL("/s")}, {BasePath: "//host", Spec: URL("/s")},
		{BasePath: "/docs", Spec: Bytes(nil)}, {BasePath: "/docs", Spec: URL("")},
		{BasePath: "/docs", Spec: URL("javascript:alert(1)")},
		{BasePath: "/docs", Spec: URL("//host/spec")},
		{BasePath: "/docs", Spec: URL("/s"), Extension: &Extension{ModuleURL: "data:text/javascript,test"}},
		{BasePath: "/docs", Spec: URL("/s"), Extension: &Extension{StylesheetURLs: []string{"javascript:alert(1)"}}},
	} {
		if _, err := New(config); err == nil {
			t.Fatalf("expected error: %+v", config)
		}
	}
}

func TestEscapedConfig(t *testing.T) {
	value := `/spec.yaml?x="/><script>alert(1)</script>&y=1`
	r := routerFor(t, Config{BasePath: "/docs", Spec: URL(value)})
	body := request(r, "GET", "/docs/").Body.String()
	if strings.Contains(body, "<script>alert(1)</script>") {
		t.Fatal("unsafe config")
	}
	match := regexp.MustCompile(`data-config="([^"]+)"`).FindStringSubmatch(body)
	var cfg clientConfig
	if err := json.Unmarshal([]byte(html.UnescapeString(match[1])), &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.BuildInfo == nil || cfg.BuildInfo.GoVersion == "" {
		t.Fatal("missing Go build info in rendered page config")
	}
	if cfg.SpecURL != value {
		t.Fatalf("changed URL: %s", cfg.SpecURL)
	}
}
