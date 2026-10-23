# go-oas-viewer

Go library，提供 Gin handler，内嵌 Scalar、Stoplight Elements、Swagger UI 的 HTML、JS、CSS 和字体。三个页面顶部有切换链接，右侧可加载业务扩展。运行时不需要 Node、CDN 或公共 CORS 代理。

Module：`github.com/iseki0/go-oas-viewer`，包名 `oasviewer`。仓库：https://github.com/iseki0/go-oas-viewer （当前为私有仓库，需要相应访问权限）。

## Viewer 版本与构建信息

| Viewer | npm package | 当前锁定版本 |
| --- | --- | --- |
| Scalar | @scalar/api-reference | 1.75.0 |
| Stoplight Elements | @stoplight/elements | 9.0.29 |
| Swagger UI | swagger-ui-dist | 5.34.0 |

每个查看器页面初始化时输出两条 console.log：

- `[go-oas-viewer] viewer versions`：当前 viewer，以及三个 viewer 的包名和版本。构建脚本从 node_modules 中实际安装的 package.json 读取版本并打入前端资源，不手写前端版本常量。
- `[go-oas-viewer] Go build info`：Go handler 构造时通过 runtime/debug.ReadBuildInfo() 获取的信息，包含 Go 版本、主 module、依赖、替换关系及构建设置；无法获取时输出 unavailable。

BuildInfo 描述运行本库的宿主程序，不是独立的前端版本或本库自己的 Git 版本。VCS 信息属于宿主主 module；库版本可在依赖及 Replace 字段中查看，go run/本地替换可能显示 (devel) 或没有版本。此信息随查看器 HTML 提供给浏览器。demo server 启动时也将完整 BuildInfo.String() 写入服务端日志。

升级 viewer 后须重建 web/，同步本表、THIRD_PARTY_NOTICES.md 和锁文件。

## 从 GitHub 接入

```powershell
go env -w GOPRIVATE=github.com/iseki0/go-oas-viewer
go get github.com/iseki0/go-oas-viewer@latest
```

私有仓库需要先配置 GitHub Git 访问凭证。若已有 GOPRIVATE，请将上述路径追加到现有配置，不要覆盖其他私有 module 配置。当前未发布版本标签，@latest 使用默认分支的伪版本。import 路径与 module 均为 github.com/iseki0/go-oas-viewer，包名为 oasviewer。

## 本地接入

要求 Go 1.24 或更高版本。另一个项目的 go.mod 添加：

```go
require github.com/iseki0/go-oas-viewer v0.0.0
replace github.com/iseki0/go-oas-viewer => C:/Users/iseki/code/go-oas-viewer
```

然后 `go mod tidy`。本库直接运行示例：

```powershell
cd C:\Users\iseki\code\go-oas-viewer
go run ./examples/basic
# 浏览器打开 http://127.0.0.1:8080/docs/
```

## 完整 Demo（手动验收）

```powershell
go run ./examples/demo
# 打开 http://127.0.0.1:18080/
# 自定义地址：go run ./examples/demo -addr 127.0.0.1:19090
```

所有 demo 页面、扩展 JS/CSS 和 OAS 都在 examples/demo/main.go，无额外运行时文件。首页列出 Bytes 与相对 URL 两种来源的三个查看器入口，以及手动验收步骤。可用 Try it 测试 GET /api/ping、POST /api/echo 和 GET /api/me。右上角“演示登录”建立 HttpOnly Cookie 会话，切页共享，退出后 /api/me 返回 401。演示登录不验证账号口令，服务重启后会话失效，只用于本地验收。

## 内存文档与 go:embed

```go
package main

import (
    _ "embed"
    "log"
    "github.com/gin-gonic/gin"
    oasviewer "github.com/iseki0/go-oas-viewer"
)

//go:embed openapi.yaml
var spec []byte

func main() {
    r := gin.Default()
    handler, err := oasviewer.New(oasviewer.Config{
        BasePath: "/docs",
        Spec: oasviewer.Bytes(spec),
        Title: "我的 API",
    })
    if err != nil { log.Fatal(err) }
    r.GET("/docs/*path", handler)
    // 可选：Gin 不会自动将 HEAD 转到 GET handler。
    r.HEAD("/docs/*path", handler)
    log.Fatal(r.Run(":8080"))
}
```

New 复制 bytes 并预渲染三个 HTML，可并发复用 handler。调用 New 期间不得并发改动输入。空文档报错；库不校验 OAS 语法。JSON bytes 也支持，仍使用 openapi.yaml 路径，响应类型为 application/json。

## URL 文档

只需将 Spec 改为以下之一：

```go
oasviewer.URL("https://example.com/openapi.yaml")
oasviewer.URL("/api/openapi.yaml")
oasviewer.URL("./spec.yaml")
```

URL 在浏览器端请求，Go 服务不下载、不做代理。绝对 URL 支持 http/https，根相对 URL 从当前域名根目录解析，目录相对 URL 从查看器目录解析。例如 /docs/ 和三个 HTML 中的 ./spec.yaml 都指向 /docs/spec.yaml，../spec.yaml 指向 /spec.yaml。片段和查询参数按浏览器 URL 规则处理。配置可来自调用方环境变量，库本身不读取环境变量。

调用方负责托管文档；查看器 handler 只提供内嵌资源和 Bytes 文档，不能自动读取磁盘上的 spec.yaml。不要将自定义规格路由放进已经占用的 Gin /docs/*path，推荐单独用 /api/openapi.yaml。

远程文档、扩展和 API 请求须满足浏览器 CORS、Cookie、CSP 和混合内容规则。三个查看器的 Try it 配置为发送浏览器凭证（include/withCredentials），库不会保存 Token。Cookie 仍按自身 Domain、Path、SameSite、Secure 限制发送。OAS 的 servers 保持原样，不强制使用 /api。

相对 $ref 以 OAS URL 为基准。Bytes 模式以 /docs/openapi.yaml 为基准，相关文件由调用方托管。各查看器的外部引用支持可能不同，推荐提供已打包为单文件的 OAS；库不执行 bundle。

## 路径与 HTTP 行为

| 路径 | 内容 |
| --- | --- |
| /docs | Gin 默认重定向到 /docs/（请保留 RedirectTrailingSlash） |
| /docs/、/docs/index.html | Scalar |
| /docs/elements.html | Stoplight Elements |
| /docs/swagger.html | Swagger UI |
| /docs/assets/* | 内嵌且带 hash 的静态资源 |
| /docs/openapi.yaml | 仅 Bytes 模式提供；URL 模式为 404 |

必须使用名为 path 的通配符；不要改成 *any。BasePath 必填，允许 / 或规范的绝对 URL 路径，可带一个末尾斜线。根路径挂载使用 BasePath: "/" 与 GET /*path。未知资源、目录和路径穿越返回 404，不返回默认 HTML。

HTML 和内存文档使用 Cache-Control: no-cache；带 hash 的资源缓存一年。支持 GET/HEAD（HEAD 需显式注册）；其他方法不会由 GET 路由匹配。New 遇到缺失来源、空 bytes、不合法 URL/BasePath 返回错误，不 panic。

反向代理：BasePath 填浏览器可见的路径；Gin 路由填写代理剥离前缀后的内部路径。页面资源、导航和相对规格 URL 均为目录相对 URL。确保代理保留末尾斜线，并把全部子路径转发到该 handler。

## 登录框与其他扩展

```go
Extension: &oasviewer.Extension{
    ModuleURL: "/custom/login.js",
    StylesheetURLs: []string{"/custom/login.css"},
},
```

模块导出 `mount({ toolbar, viewer, specURL })`；toolbar 是导航右侧 HTMLElement，viewer 是 scalar/elements/swagger，specURL 是解析后的绝对 OAS URL。完整执行顺序、错误处理、登录表单示例和 ind2 迁移步骤见 [扩展接入文档](docs/extensions.md)。

运行 `go run ./examples/extension` 可查看可运行的登录链接扩展示例。两个示例使用同一端口，请分别运行；示例不伪造登录成功。

## 维护与构建

普通使用者无需 npm；web/ 已包含构建资源，go:embed 在 Go 构建时内嵌它们。维护前端才需要 Node（Vite 8 所支持的 Node 版本）和 npm：

```powershell
npm ci --ignore-scripts
npm run build
gofmt -w .
go test ./...
go vet ./...
go build -trimpath -ldflags "-s -w -buildid=" ./examples/basic
```

frontend/ 是维护源，web/ 是构建产物，必须一起提交。package-lock.json 锁定前端依赖，go.sum 锁定 Go 依赖。构建脚本会在 node_modules 创建独立 go.mod 边界，避免 go test ./... 扫描 npm 依赖树。Vite 大 chunk 提示是资源体积提示，不表示构建失败。第三方许可见 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) 和 web/assets 下的许可文件。

首次接入或更新依赖后，在浏览器打开三个页面，确认标题、导航高亮、规格内容、Try it、扩展和网络资源均正确。Go HTTP 测试不代替浏览器渲染检查。Swagger OAuth2 重定向登录不是本库的业务登录扩展；当前不提供专用 OAuth2 回调页，需额外接入后再启用该流程。
