# 前端扩展契约

本库将公共导航、查看器初始化和业务工具栏分开。登录表单、SSO 链接、环境切换等由接入方扩展提供，业务认证协议不属于库。

## Go 配置

```go
Extension: &oasviewer.Extension{
    ModuleURL: "/custom/login.js",
    StylesheetURLs: []string{"/custom/login.css"},
},
```

Extension 为 nil 时不加载业务扩展。ModuleURL 可为空（仅加载样式）；StylesheetURLs 可为空。模块和样式支持 HTTP(S)、根相对及目录相对 URL。相对路径从当前查看器目录解析；例如 /docs/ 下 ./custom/login.js 为 /docs/custom/login.js。不要把静态文件路径误当作浏览器 URL。

资源由接入方提供独立的 Gin 静态路由或反向代理托管。本库不会读取指定磁盘文件、内嵌扩展资源或代理远程模块。Gin 的 /docs/*path 已占用子路径，推荐 /custom/* 路由。可使用调用方自己的 go:embed + http.FS 托管，见 examples/extension。

## 模块接口与顺序

```js
export async function mount({ toolbar, viewer, specURL }) {
  // toolbar: 导航右侧的 HTMLElement（#viewer-toolbar）
  // viewer: "scalar" | "elements" | "swagger"
  // specURL: 浏览器解析后的绝对 OAS URL
}
```

1. 页面加载内置 CSS/JS，解析 Go 传入的配置。
2. 自定义 CSS 在内置样式之后按配置顺序插入 head；CSS 加载不作为异步屏障。
3. 动态 import ModuleURL，执行模块顶层代码，然后调用并 await mount(context)。
4. mount 成功返回后，初始化当前查看器并开始加载 OAS。

查看器依赖会在 mount 前加载，但查看器实例尚未创建。mount 必须及时返回，不能等用户提交登录表单或一直等待会话事件；注册事件后即返回。必要的会话查询可以 await，但应自行设置超时。未结束的 mount 会阻止查看器初始化；库不擅自中断扩展。

模块加载失败、未导出 mount、mount 抛出异常时，在页面状态区域显示文本错误，仍继续初始化查看器。库不会重试。查看器自身的 OAS 加载错误由各自 UI 显示。模块可在同一页面内复用浏览器模块缓存，但三个 HTML 是完整页面导航，每次重新挂载；不提供 SPA 卸载回调。

每个页面都会获得同一契约；不要访问 Scalar、Elements 或 Swagger 内部 DOM/全局变量。当前扩展不支持替换查看器配置、全局 fetch 拦截器或直接修改 OAS。需要 Bearer Token/签名请求时，先设计明确的各查看器适配接口，不要 monkey-patch window.fetch。

## 登录表单示例

下面是接入骨架，/api/login 和 /api/session 以及响应结构由业务自行实现，不是本库提供的端点。已有 RSA/SSO 协议时直接复用业务认证模块。

```js
export function mount({ toolbar }) {
  const form = document.createElement('form');
  form.className = 'viewer-login';
  form.setAttribute('aria-label', 'API 会话登录');
  // 以下为开发者编写的固定 HTML，不能拼接不可信输入。
  form.innerHTML = `
    <input name="username" aria-label="用户名" autocomplete="username" required>
    <input name="password" type="password" aria-label="口令" autocomplete="current-password" required>
    <button type="submit">登录</button>
    <span role="status" aria-live="polite"></span>`;
  toolbar.append(form);
  const status = form.querySelector('[role=status]');
  const button = form.querySelector('button');
  let busy = false;
  form.addEventListener('submit', async event => {
    event.preventDefault();
    if (busy) return;
    busy = true;
    button.disabled = true;
    let password = form.elements.password.value;
    form.elements.password.value = '';
    status.textContent = '正在登录…';
    try {
      // 根据业务要求替换这部分协议，含 CSRF 和口令加密等要求。
      const response = await fetch('/api/login', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: form.elements.username.value, password }),
        signal: AbortSignal.timeout(20000),
      });
      if (!response.ok) throw new Error('登录失败');
      status.textContent = '已登录';
      // 文档本身受保护时，登录成功后刷新页面让查看器重新加载文档。
    } catch (error) {
      status.textContent = error.message;
    } finally {
      password = '';
      form.elements.password.value = '';
      busy = false;
      button.disabled = false;
    }
  });
  // 刷新/切换页面后可请求真实会话端点确认状态；不要用 localStorage 标记已登录。
}
```

CSS 可自行使用以下稳定入口：

```css
:root { --viewer-nav-height: 48px; }
#viewer-toolbar .viewer-login { display: flex; align-items: center; gap: 8px; }
#viewer-toolbar input { width: 140px; }
```

公共入口：.viewer-nav、#viewer-toolbar、#viewer-status、--viewer-nav-height。窄屏导航横向滚动；表单请保持紧凑。自定义 CSS 能影响整个页面，推荐限制在 #viewer-toolbar 内。

## Cookie 与状态

登录请求由扩展发起，业务服务设置 HttpOnly Cookie。浏览器在符合 Cookie 策略时，向三个查看器的 Try it 请求发送 Cookie。库不获取 Cookie 值、不保存口令/Token、不假设任何会话 API 或登录响应格式。

跨域登录仍需要业务 CORS 和正确 Cookie 属性；加载远程模块还需要正确 JavaScript MIME、CORS 和 CSP。生产应托管可信扩展，因为它以当前页面权限执行。当前没有内置 CSP 响应头；请结合查看器的样式、脚本和远程规格需求设置策略。

## 从 ind2 迁移

1. 保留 ind2 的 auth.js（authenticatedFetch、login 和 RSA 封套协议）在业务项目内，通过业务自己的构建流程产出模块。
2. 将 toolbar.js 中 DOM 创建、会话查询和事件注册移进导出的 mount({ toolbar })，将 nav.append(form) 改为 toolbar.append(form)。不要在模块顶层查找业务表单。
3. 将 viewer-login 相关 CSS 放进业务 login.css，用 StylesheetURLs 引入。
4. 通过独立静态路由提供 login.js、auth.js 及依赖文件；设置 ModuleURL。
5. 三查看器已启用 Cookie 凭证；库不会改写 servers 或提供 /api 代理，业务需提供正确 OAS servers 和 API 路由/代理。
6. 分别验证登录、登录失败、刷新、切换三页面、会话过期和 Try it 的实际 Cookie 行为。文档受保护时还需验证未登录首次加载与登录后重载。

## 验收清单

- 三个入口导航与当前页高亮正常，右侧扩展均挂载一次。
- mount 返回前不发起 OAS 加载；表单挂载后无需登录即可完成 mount。
- 模块不存在或抛错时显示扩展错误，查看器仍能加载。
- CSS 无 404，模块及相对 import 的文件均可访问。
- 登录状态以服务端响应为准，切页后能重新确认；不保存口令和 Token。
- 不依赖当前目录、开发服务器、CDN 或 ind2 仓库中的源码路径。
