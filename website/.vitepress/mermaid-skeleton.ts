// Mermaid 骨架屏插件
// 作用：vitepress-plugin-mermaid 默认把 mermaid 代码块渲染成空 <div class="mermaid">
// （外加一个写死 "Loading..." 的 Suspense fallback），首屏看到一片空白，
// 等 JS 加载完客户端渲染才出图——这是文档站"看着还是干巴/没图"的根因。
//
// 本插件在 markdown-it 层接管 mermaid fence，输出一个带骨架屏占位的容器：
//   <div class="mermaid-host" data-mermaid="<源码>">
//     <div class="mermaid-skeleton" aria-hidden="true">…骨架占位…</div>
//   </div>
// 骨架屏本身有视觉结构（脉冲框 + 图标 + 类别标签），首屏不再是空白；
// 客户端 mermaid 渲染完成后，把骨架替换为真 SVG，无感知升级。
//
// 与 vitepress-plugin-mermaid 共存：本插件先注册，后者的 fence 拦截因
// 我们已 return 而不再触发。我们仍依赖 mermaid 包做客户端渲染，只是占位更好看。

const CATEGORY_LABELS = {
  flowchart: '流程图',
  'sequenceDiagram': '时序图',
  'stateDiagram': '状态图',
  'classDiagram': '类图',
  'erDiagram': 'ER 图',
  gantt: '甘特图',
  pie: '饼图',
  mindmap: '思维导图',
  journey: '用户旅程',
  gitGraph: 'Git 图'
}

function detectCategory(code) {
  const c = code.trim()
  for (const key of Object.keys(CATEGORY_LABELS)) {
    if (c.startsWith(key) || c.includes('\n' + key)) return key
  }
  // flowchart / graph 也算流程图
  if (/^(flowchart|graph)\s/.test(c)) return 'flowchart'
  return null
}

export function mermaidSkeletonPlugin(md) {
  const original = md.renderer.rules.fence
  md.renderer.rules.fence = (tokens, idx, options, env, self) => {
    const token = tokens[idx]
    if (token.info.trim() === 'mermaid') {
      const code = token.content
      const cat = detectCategory(code)
      const label = cat ? CATEGORY_LABELS[cat] : '示意图'
      // 把源码做 URL 编码塞进 data 属性，客户端脚本解码后喂给 mermaid.render
      const encoded = encodeURIComponent(code)
      return (
        `<div class="mermaid-host" data-mermaid="${encoded}" data-cat="${cat || ''}">` +
          `<div class="mermaid-skeleton" aria-hidden="true">` +
            `<div class="mermaid-skeleton-bar"></div>` +
            `<div class="mermaid-skeleton-body">` +
              `<span class="mermaid-skeleton-icon">◷</span>` +
              `<span class="mermaid-skeleton-label">${label}</span>` +
            `</div>` +
            `<div class="mermaid-skeleton-shimmer"></div>` +
          `</div>` +
          `<div class="mermaid-real" role="img"></div>` +
        `</div>`
      )
    }
    return original.call(this, tokens, idx, options, env, self)
  }
}
