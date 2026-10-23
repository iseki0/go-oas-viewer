import './viewer.css';

export async function setup(viewer) {
  const config = JSON.parse(document.querySelector('#viewer-config').dataset.config);
  console.log('[go-oas-viewer] viewer versions', { current: viewer, viewers: __VIEWER_VERSIONS__ });
  console.log('[go-oas-viewer] Go build info', config.buildInfo ?? 'unavailable');
  const specURL = new URL(config.specURL, document.baseURI).href;
  if (config.moduleURL) {
    try {
      const extension = await import(/* @vite-ignore */ new URL(config.moduleURL, document.baseURI).href);
      if (typeof extension.mount !== 'function') throw new Error('扩展模块必须导出 mount 函数');
      await extension.mount({ toolbar: document.querySelector('#viewer-toolbar'), viewer, specURL });
    } catch (error) {
      showError(new Error(`扩展加载失败：${error instanceof Error ? error.message : String(error)}`));
    }
  }
  return specURL;
}

export function showError(error) {
  const status = document.querySelector('#viewer-status');
  status.hidden = false;
  status.textContent = error instanceof Error ? error.message : String(error);
}
