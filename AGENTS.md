# Agent 工作约定

这是 Go library，不是独立业务服务。先读 README.md，修改扩展前读 docs/extensions.md。

- 公开接口为 URL、Bytes、New，以及 Config、Source、Extension。变更接口必须同步 README、Go doc、示例和相关测试。
- 前端维护源在 frontend/；web/ 为 Vite 构建结果，必须随源码一起分发和提交，不能手改。
- 更新前端：npm ci --ignore-scripts，然后 npm run build。使用 package-lock.json，不混用包管理器。
- 普通库使用者只需要 Go，不需要 Node；发布前核对 web/ 已构建且被 go:embed 包含。
- handler 挂载 GET /docs/*path，通配符名称必须为 path。支持根挂载 /*path。扩展资源用其他路由托管。
- 扩展只负责导航右侧 DOM 和业务行为。不得将具体项目的登录协议、口令、API servers 或代理写进库。
- mount 的 context 契约和执行顺序见 docs/extensions.md；改动时同步文档、示例及验证。
- 不默认下载 OAS、不读取运行时磁盘、不修改 servers、不打包外部 $ref。URL 由浏览器加载。
- 运行 gofmt、go test ./...、go vet ./...。前端变化还须构建并在浏览器检查三个页面和扩展。
- 检查前端依赖许可并维护 THIRD_PARTY_NOTICES.md；不随意移除许可证和版权声明。
- 默认简体中文文档。未经用户授权不要推送、发布或创建远端仓库。
