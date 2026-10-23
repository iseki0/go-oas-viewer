import { readFile, writeFile } from 'node:fs/promises';
import { build } from 'vite';

// 阻止 go test ./... 扫描庞大的 npm 依赖树及其中的第三方 Go 示例。
await writeFile(new URL('../node_modules/go.mod', import.meta.url), 'module go-oas-viewer-frontend-dependencies\n\ngo 1.24.0\n');
// 从实际安装的包读取版本，确保控制台反映本次打包的资源。
const packages = {
  scalar: '@scalar/api-reference',
  elements: '@stoplight/elements',
  swagger: 'swagger-ui-dist',
};
const versions = {};
for (const [viewer, name] of Object.entries(packages)) {
  const pkg = JSON.parse(await readFile(new URL('../node_modules/' + name + '/package.json', import.meta.url), 'utf8'));
  versions[viewer] = { package: name, version: pkg.version };
}
await build({ define: { __VIEWER_VERSIONS__: JSON.stringify(versions) } });
