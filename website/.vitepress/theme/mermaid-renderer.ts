// Mermaid 客户端渲染器
// 接管 .mermaid-host 容器：解码 data-mermaid 源码 → mermaid.render → 替换骨架屏为真 SVG
// 与 vitepress-plugin-mermaid 的 Mermaid.vue 方案不同，本渲染器：
//   1. 不依赖 Suspense，首屏由骨架屏 .mermaid-skeleton 占位（构建时已内联进 HTML）
//   2. 按 documentElement.classList.contains('dark') 选 light/dark 主题
//   3. 路由切换后（VitePress SPA）重新扫描未渲染的 host
//
// mermaid 包动态 import：避免 VitePress SSR 阶段加载它（mermaid 顶层访问 document），
// 同时让 mermaid 进入独立 chunk，按需加载，不拖慢首屏 JS。
let _mermaid = null
async function getMermaid() {
  if (_mermaid) return _mermaid
  const mod = await import('mermaid')
  _mermaid = mod.default
  return _mermaid
}

let booted = false
let renderSeq = 0

async function ensureBoot(isDark) {
  if (booted) return
  const mermaid = await getMermaid()
  mermaid.initialize({
    startOnLoad: false,
    theme: isDark ? 'dark' : 'default',
    securityLevel: 'loose',
    flowchart: { curve: 'basis', useMaxWidth: true },
    sequence: { useMaxWidth: true },
    gantt: { useMaxWidth: true }
  })
  booted = true
}

async function renderOne(host) {
  if (host.dataset.rendered === '1') return
  const code = decodeURIComponent(host.dataset.mermaid || '')
  if (!code) return
  const isDark = document.documentElement.classList.contains('dark')
  await ensureBoot(isDark)
  const mermaid = await getMermaid()
  const id = `mmd-${renderSeq++}`
  try {
    const { svg } = await mermaid.render(id, code)
    const real = host.querySelector('.mermaid-real')
    if (real) {
      real.innerHTML = svg
      host.dataset.rendered = '1'
      host.classList.add('mermaid-rendered')
      // 隐藏骨架
      const skel = host.querySelector('.mermaid-skeleton')
      if (skel) skel.style.display = 'none'
    }
  } catch (e) {
    // 渲染失败：保留骨架，标记错误，把源码+错误信息作为 fallback 展示
    const errMsg = (e && (e.message || String(e))) || 'unknown error'
    host.classList.add('mermaid-error')
    host.dataset.error = errMsg.slice(0, 500)
    const real = host.querySelector('.mermaid-real')
    if (real) {
      real.innerHTML =
        `<pre class="mermaid-fallback"><code>${escapeHtml(code)}</code>` +
        `<span class="mermaid-fallback-err">⚠ ${escapeHtml(errMsg).slice(0,300)}</span></pre>`
    }
    console.warn('[mermaid] render failed:', e)
  }
}

function escapeHtml(s) {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
}

async function scan(root = document) {
  const hosts = root.querySelectorAll('.mermaid-host:not([data-rendered="1"])')
  if (!hosts.length) return
  scanning = true
  try {
    // 主题可能变化，重新 initialize
    const isDark = document.documentElement.classList.contains('dark')
    if (booted) {
      const mermaid = await getMermaid()
      mermaid.initialize({
        startOnLoad: false,
        theme: isDark ? 'dark' : 'default',
        securityLevel: 'loose'
      })
    }
    await Promise.all([...hosts].map(renderOne))
  } finally {
    scanning = false
  }
}

let darkObserver = null
let scanning = false // 防止 scan 重入：渲染会触发 MutationObserver，避免循环

export function setupMermaidRenderer() {
  if (typeof document === 'undefined') return // SSR 跳过

  // 路由切换后扫描（VitePress SPA，每次路由切换重挂内容）
  let routeTimer = null
  const onRouteChange = () => {
    clearTimeout(routeTimer)
    routeTimer = setTimeout(async () => {
      if (scanning) return
      await scan()
    }, 80)
  }
  window.addEventListener('hashchange', onRouteChange)
  // 兜底：MutationObserver 观察 main 内容区新增节点（路由切换重挂内容）
  // 注意：renderOne 也会改 DOM 触发本 observer，靠 scanning flag + :not([rendered]) 选择器收敛
  const mo = new MutationObserver((muts) => {
    // 只在有新增节点时触发，过滤纯属性变化（避免 dark 切换重复触发）
    const hasAdded = muts.some(m => m.addedNodes.length > 0)
    if (hasAdded) onRouteChange()
  })
  mo.observe(document.body, { childList: true, subtree: true })

  // 深浅色切换时重新渲染（dark class 变化）
  darkObserver = new MutationObserver(() => {
    document
      .querySelectorAll('.mermaid-host[data-rendered="1"]')
      .forEach((host) => {
        host.dataset.rendered = '0'
        const skel = host.querySelector('.mermaid-skeleton')
        if (skel) skel.style.display = ''
        const real = host.querySelector('.mermaid-real')
        if (real) real.innerHTML = ''
        host.classList.remove('mermaid-rendered')
      })
    onRouteChange()
  })
  darkObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class']
  })

  // 首次扫描
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => scan())
  } else {
    scan()
  }
}
