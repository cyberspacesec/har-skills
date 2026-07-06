import { h } from 'vue'
import DefaultTheme from 'vitepress/theme'
import type { Theme } from 'vitepress'
import './styles/theme.css'
import { setupMermaidRenderer } from './mermaid-renderer'
import HarHero from './components/HarHero.vue'
import HarWaterfall from './components/HarWaterfall.vue'
import FeatureCard from './components/FeatureCard.vue'
import CapabilityTree from './components/CapabilityTree.vue'
import CommandTable from './components/CommandTable.vue'

export default {
  extends: DefaultTheme,
  Layout: () => {
    // 使用默认布局，仅通过全局样式覆盖外观
    return h(DefaultTheme.Layout, null, {})
  },
  enhanceApp({ app }) {
    app.component('HarHero', HarHero)
    app.component('HarWaterfall', HarWaterfall)
    app.component('FeatureCard', FeatureCard)
    app.component('CapabilityTree', CapabilityTree)
    app.component('CommandTable', CommandTable)
  },
  setup() {
    // 客户端启动后接管 mermaid 骨架屏容器：解码源码 → mermaid.render → 替换骨架
    setupMermaidRenderer()
  }
} satisfies Theme
