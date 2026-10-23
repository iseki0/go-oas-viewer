// 必须导出 mount；每个页面加载时调用一次。
export function mount({ toolbar, viewer, specURL }) {
  const link = document.createElement('a');
  link.className = 'login-link';
  link.textContent = '登录';
  link.href = '/login?returnTo=' + encodeURIComponent(window.location.pathname);
  toolbar.append(link);
  toolbar.dataset.viewer = viewer;
  // 可在这里挂载登录表单、查询真实会话状态等。
  // 不等待用户完成登录：mount 返回后查看器才开始加载 OAS。
  console.debug('OAS URL:', specURL);
}
