<script setup lang="ts">
import HarHero from './HarHero.vue'
import HarWaterfall from './HarWaterfall.vue'
import FeatureCard from './FeatureCard.vue'
import CapabilityTree from './CapabilityTree.vue'
</script>

<template>
  <HarHero
    eyebrow="AI-Native · Go SDK · CLI · Skill"
    title="读懂每一条HTTP流量的工具箱"
    highlight="HTTP流量"
    tagline="HAR Skills 是一个面向 AI Agent 优先的 HAR（HTTP Archive）分析工具箱。装一个二进制，Agent 就能解析、搜索、审计、评分、脱敏、重放、转换、对比网络抓包，并以机器可读格式导出结果。"
    :primary="{ text: '快速开始 →', link: '/zh/quick-start' }"
    :secondary="{ text: 'CLI 命令参考', link: '/zh/cli/global-flags' }"
  />

  <section class="har-section">
    <p class="har-section__eyebrow">能力树</p>
    <h2 class="har-section__title">六大能力域，覆盖 HAR 全生命周期</h2>
    <p class="har-section__lede">从接入、解析、检视、分析、操作到对比导出 —— 每个能力域对应一组 CLI 命令和 SDK 方法，配色即分类。</p>
    <CapabilityTree />
  </section>

  <section class="har-section">
    <p class="har-section__eyebrow">接入方式</p>
    <h2 class="har-section__title">四种方式，同一套能力</h2>
    <p class="har-section__lede">无论你是给 AI Agent 装技能、在终端跑命令、在 Go 程序里调 SDK，还是封装 MCP 服务，触达的是同一套 40 个模块、70+ 方法。</p>
    <div class="har-grid har-grid--4">
      <FeatureCard color="#2563eb" tag="Agent Access" title="AI Agent Skill" desc="通过 CLAUDE.md 渐进式披露文档，Agent 一键引导即可调用全部能力。" link="/zh/access/skill" />
      <FeatureCard color="#0891b2" tag="CLI" title="命令行工具" desc="24 个 Cobra 命令，支持 stdin/JSON/CSV/YAML 多格式输出，可管道组合。" link="/zh/access/cli" />
      <FeatureCard color="#9333ea" tag="SDK" title="Go SDK" desc="根包 40 模块，零运行时依赖，函数式选项 + 结构体双风格 API。" link="/zh/access/sdk" />
      <FeatureCard color="#ea580c" tag="MCP" title="MCP 封装" desc="可包装为 Model Context Protocol 服务，供 Claude 等客户端调用。" link="/zh/access/mcp" />
    </div>
  </section>

  <section class="har-section">
    <p class="har-section__eyebrow">签名可视化</p>
    <h2 class="har-section__title">把 waterfall 搬到首页</h2>
    <p class="har-section__lede">这是工具最原生的视觉语言：每条请求是一根横条，blocked→dns→connect→ssl→send→wait→receive 分段着色，按真实耗时比例排布。下面是 waterfall 命令的产物。</p>
    <HarWaterfall />
    <p class="har-section__caption">提示：在终端运行 <code>har -f capture.har waterfall</code> 即可生成 ASCII 版瀑布流；<code>--critical-path</code> 看关键路径，<code>--sla</code> 校验 SLA。</p>
  </section>

  <section class="har-section">
    <p class="har-section__eyebrow">Agent 能做什么</p>
    <h2 class="har-section__title">一句话到一条命令</h2>
    <p class="har-section__lede">下面是 Agent 接到自然语言任务后，会映射到的典型命令。</p>
    <div class="har-grid har-grid--3">
      <FeatureCard color="#16a34a" tag="Inspect" title="找出所有失败的请求" desc="har -f capture.har find --errors --format json" />
      <FeatureCard color="#9333ea" tag="Analyze" title="给这次抓包打性能分" desc="har -f capture.har performance" />
      <FeatureCard color="#ea580c" tag="Operate" title="脱敏后再分享" desc="har -f capture.har redact --redact-ips -o clean.har" />
      <FeatureCard color="#475569" tag="Compare" title="对比两次发布" desc="har diff v1.har v2.har --compare-by-url" />
      <FeatureCard color="#475569" tag="Export" title="转成 curl 重放" desc="har -f capture.har export curl" />
      <FeatureCard color="#0891b2" tag="Ingest" title="按域名拆分大文件" desc="har -f big.har split --by domain" />
    </div>
  </section>

  <section class="har-section har-cta">
    <h2 class="har-section__title">开始用</h2>
    <p class="har-section__lede">三步上手：安装二进制、跑一条命令、读一页文档。</p>
    <div class="har-cta__code"><pre><code># 1. 安装
go install github.com/cyberspacesec/har-skills/cmd/har@latest

# 2. 跑一条命令
har -f capture.har info

# 3. 读文档
# 中文：/zh/quick-start   英文：/en/quick-start</code></pre></div>
    <div class="har-cta__actions">
      <a class="har-btn har-btn--primary" href="/zh/quick-start">快速开始</a>
      <a class="har-btn har-btn--ghost" href="/zh/sdk/data-structures">SDK 指南</a>
    </div>
  </section>
</template>

<style scoped>
.har-grid {
  display: grid;
  gap: 1rem;
  margin-bottom: 1rem;
}
.har-grid--4 { grid-template-columns: repeat(4, 1fr); }
.har-grid--3 { grid-template-columns: repeat(3, 1fr); }
.har-section__caption {
  margin-top: 1rem;
  font-size: 0.85rem;
  color: var(--har-muted);
}
.har-section__caption code {
  font-family: var(--vp-font-family-mono);
  font-size: 0.85em;
  background: rgba(37, 99, 235, 0.08);
  color: var(--vp-c-brand-1);
  border-radius: 4px;
  padding: 0.1em 0.35em;
}
.har-cta__code pre {
  border: 1px solid var(--har-line);
  border-radius: 8px;
  padding: 1rem 1.25rem;
  background: var(--vp-bg, #fff);
  font-family: var(--vp-font-family-mono);
  font-size: 0.8rem;
  line-height: 1.7;
  overflow-x: auto;
  margin: 0 0 1.5rem;
}
:deep(.dark) .har-cta__code pre,
.dark .har-cta__code pre { background: #0f172a; }
.har-cta__actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.har-btn {
  display: inline-flex;
  align-items: center;
  padding: 0.6rem 1.1rem;
  border-radius: 6px;
  font-family: var(--vp-font-family-mono);
  font-size: 0.875rem;
  font-weight: 500;
  text-decoration: none;
}
.har-btn--primary { background: var(--har-amber); color: #fff; }
.har-btn--primary:hover { background: #c2410c; }
.har-btn--ghost { border: 1px solid var(--har-line); color: inherit; }
.har-btn--ghost:hover { border-color: var(--har-blue); color: var(--har-blue); }

@media (max-width: 960px) {
  .har-grid--4 { grid-template-columns: repeat(2, 1fr); }
  .har-grid--3 { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 600px) {
  .har-grid--4, .har-grid--3 { grid-template-columns: 1fr; }
}
</style>
