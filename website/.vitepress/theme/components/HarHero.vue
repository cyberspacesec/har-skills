<script setup lang="ts">
import { computed } from 'vue'

// 分屏 Hero：左侧标题与定位句，右侧渲染一段真实 HAR JSON 片段
// 这比"大数字+渐变"更贴题 —— HAR 本质是 JSON 文本，直接展示它

const props = defineProps<{
  eyebrow?: string
  title: string
  highlight?: string
  tagline: string
  primary?: { text: string; link: string }
  secondary?: { text: string; link: string }
}>()

// 高亮标题中的一部分（用 highlight 标记）
const titleParts = computed(() => {
  if (!props.highlight) return [{ text: props.title, hl: false }]
  const idx = props.title.indexOf(props.highlight)
  if (idx < 0) return [{ text: props.title, hl: false }]
  return [
    { text: props.title.slice(0, idx), hl: false },
    { text: props.highlight, hl: true },
    { text: props.title.slice(idx + props.highlight.length), hl: false }
  ]
})
</script>

<template>
  <section class="har-hero">
    <div class="har-hero__grid">
      <div class="har-hero__copy">
        <p class="har-hero__eyebrow" v-if="eyebrow">{{ eyebrow }}</p>
        <h1 class="har-hero__title">
          <span v-for="(p, i) in titleParts" :key="i" :class="{ 'har-hero__title-hl': p.hl }">{{ p.text }}</span>
        </h1>
        <p class="har-hero__tagline">{{ tagline }}</p>
        <div class="har-hero__actions" v-if="primary || secondary">
          <a v-if="primary" class="har-hero__btn har-hero__btn--primary" :href="primary.link">{{ primary.text }}</a>
          <a v-if="secondary" class="har-hero__btn har-hero__btn--ghost" :href="secondary.link">{{ secondary.text }}</a>
        </div>
      </div>
      <div class="har-hero__terminal">
        <div class="har-hero__term-bar">
          <span class="har-hero__dot har-hero__dot--r"></span>
          <span class="har-hero__dot har-hero__dot--y"></span>
          <span class="har-hero__dot har-hero__dot--g"></span>
          <span class="har-hero__term-name">capture.har</span>
        </div>
        <pre class="har-hero__code"><code><span class="t-key">"log"</span>: {
  <span class="t-key">"version"</span>: <span class="t-str">"1.2"</span>,
  <span class="t-key">"creator"</span>: {<span class="t-key">"name"</span>: <span class="t-str">"WebInspector"</span>},
  <span class="t-key">"entries"</span>: [
    {
      <span class="t-key">"startedDateTime"</span>: <span class="t-str">"2024-07-03T10:00:00Z"</span>,
      <span class="t-key">"request"</span>: {<span class="t-key">"method"</span>: <span class="t-str">"GET"</span>, <span class="t-key">"url"</span>: <span class="t-str">"https://api.example.com/users"</span>},
      <span class="t-key">"response"</span>: {<span class="t-key">"status"</span>: <span class="t-num">200</span>, <span class="t-key">"content"</span>: {<span class="t-key">"mimeType"</span>: <span class="t-str">"application/json"</span>}},
      <span class="t-key">"timings"</span>: {<span class="t-key">"dns"</span>: <span class="t-num">12</span>, <span class="t-key">"connect"</span>: <span class="t-num">45</span>, <span class="t-key">"wait"</span>: <span class="t-num">128</span>}
    }
  ]
}</code></pre>
      </div>
    </div>
  </section>
</template>

<style scoped>
.har-hero {
  max-width: 1152px;
  margin: 0 auto;
  padding: 3rem 1.5rem 2rem;
}
.har-hero__grid {
  display: grid;
  grid-template-columns: 1.1fr 1fr;
  gap: 3rem;
  align-items: center;
}
.har-hero__eyebrow {
  font-family: var(--vp-font-family-mono);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  color: var(--har-amber);
  margin: 0 0 1rem;
  font-weight: 500;
}
.har-hero__title {
  font-family: var(--vp-font-family-mono);
  font-size: 3.5rem;
  line-height: 1.05;
  letter-spacing: -0.03em;
  margin: 0 0 1.25rem;
  font-weight: 700;
}
.har-hero__title-hl {
  color: var(--har-amber);
}
.har-hero__tagline {
  font-size: 1.2rem;
  line-height: 1.55;
  color: var(--har-muted);
  max-width: 32ch;
  margin: 0 0 2rem;
}
.har-hero__actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.har-hero__btn {
  display: inline-flex;
  align-items: center;
  padding: 0.6rem 1.1rem;
  border-radius: 6px;
  font-family: var(--vp-font-family-mono);
  font-size: 0.875rem;
  font-weight: 500;
  text-decoration: none;
  transition: transform 0.15s ease, background 0.15s ease;
}
.har-hero__btn--primary {
  background: var(--har-amber);
  color: #fff;
}
.har-hero__btn--primary:hover {
  background: #c2410c;
  transform: translateY(-1px);
}
.har-hero__btn--ghost {
  border: 1px solid var(--har-line);
  color: var(--har-ink);
  background: transparent;
}
.dark .har-hero__btn--ghost {
  color: var(--har-paper);
}
.har-hero__btn--ghost:hover {
  border-color: var(--har-blue);
  color: var(--har-blue);
}
.har-hero__terminal {
  border: 1px solid var(--har-line);
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 20px 50px -20px rgba(11, 18, 32, 0.25);
  background: var(--vp-bg, #fff);
}
.dark .har-hero__terminal {
  background: #0f172a;
}
.har-hero__term-bar {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.6rem 0.9rem;
  background: rgba(0, 0, 0, 0.04);
  border-bottom: 1px solid var(--har-line);
}
.dark .har-hero__term-bar {
  background: rgba(255, 255, 255, 0.03);
}
.har-hero__dot {
  width: 11px;
  height: 11px;
  border-radius: 50%;
}
.har-hero__dot--r { background: #ef4444; }
.har-hero__dot--y { background: #eab308; }
.har-hero__dot--g { background: #22c55e; }
.har-hero__term-name {
  margin-left: 0.6rem;
  font-family: var(--vp-font-family-mono);
  font-size: 0.75rem;
  color: var(--har-muted);
}
.har-hero__code {
  margin: 0;
  padding: 1.1rem 1.25rem;
  overflow-x: auto;
  font-family: var(--vp-font-family-mono);
  font-size: 0.8rem;
  line-height: 1.65;
  color: var(--har-ink);
}
.dark .har-hero__code { color: var(--har-paper); }
.t-key { color: var(--har-blue); }
.dark .t-key { color: #60a5fa; }
.t-str { color: var(--har-green); }
.dark .t-str { color: #4ade80; }
.t-num { color: var(--har-violet); }
.dark .t-num { color: #c084fc; }

@media (max-width: 960px) {
  .har-hero__grid {
    grid-template-columns: 1fr;
    gap: 2rem;
  }
  .har-hero__title { font-size: 2.5rem; }
}
@media (max-width: 480px) {
  .har-hero__title { font-size: 2rem; }
  .har-hero__code { font-size: 0.7rem; }
}
</style>
