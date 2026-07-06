<script setup lang="ts">
// 签名元素：HAR entries 的 waterfall 时间轴条。
// 把工具最原生的视觉语言搬到首页 —— 每条请求是一根横条，
// blocked→dns→connect→ssl→send→wait→receive 分段着色，按真实比例排布。
// 这不是装饰，而是工具本身的核心可视化。

interface Entry {
  url: string
  method: string
  status: number
  // 各阶段耗时(ms)，用于计算宽度比例
  blocked: number
  dns: number
  connect: number
  ssl: number
  send: number
  wait: number
  receive: number
}

// 模拟一组真实抓包的 timing（比例接近真实网络场景）
const entries: Entry[] = [
  { url: 'api.example.com/users',     method: 'GET',  status: 200, blocked: 5,  dns: 12,  connect: 45, ssl: 38, send: 2,  wait: 128, receive: 22 },
  { url: 'cdn.example.com/app.js',    method: 'GET',  status: 200, blocked: 3,  dns: 8,   connect: 0,  ssl: 0,  send: 1,  wait: 45,  receive: 180 },
  { url: 'api.example.com/login',     method: 'POST', status: 401, blocked: 4,  dns: 0,   connect: 0,  ssl: 0,  send: 15, wait: 310, receive: 8 },
  { url: 'cdn.example.com/style.css', method: 'GET',  status: 304, blocked: 2,  dns: 0,   connect: 0,  ssl: 0,  send: 1,  wait: 38,  receive: 6 },
  { url: 'api.example.com/profile',   method: 'GET',  status: 500, blocked: 6,  dns: 0,   connect: 0,  ssl: 0,  send: 2,  wait: 520, receive: 14 },
  { url: 'cdn.example.com/logo.png',  method: 'GET',  status: 200, blocked: 1,  dns: 0,   connect: 0,  ssl: 0,  send: 1,  wait: 60,  receive: 240 }
]

const phases = [
  { key: 'blocked',  color: '#94a3b8' },
  { key: 'dns',      color: '#0891b2' },
  { key: 'connect',  color: '#2563eb' },
  { key: 'ssl',      color: '#9333ea' },
  { key: 'send',     color: '#16a34a' },
  { key: 'wait',     color: '#ea580c' },
  { key: 'receive',  color: '#f59e0b' }
] as const

// 计算每条 entry 的总时长，用于归一化宽度
const totalMax = Math.max(...entries.map(e =>
  e.blocked + e.dns + e.connect + e.ssl + e.send + e.wait + e.receive
))

function statusClass(s: number) {
  if (s >= 500) return 'is-5xx'
  if (s >= 400) return 'is-4xx'
  if (s >= 300) return 'is-3xx'
  return 'is-2xx'
}

function widthOf(ms: number) {
  return (ms / totalMax) * 100
}
</script>

<template>
  <div class="har-wf">
    <div class="har-wf__legend">
      <span v-for="p in phases" :key="p.key" class="har-wf__legend-item">
        <span class="har-wf__swatch" :style="{ background: p.color }"></span>{{ p.key }}
      </span>
    </div>
    <div class="har-wf__rows">
      <div v-for="(e, i) in entries" :key="i" class="har-wf__row">
        <div class="har-wf__label">
          <span class="har-wf__method" :class="e.method === 'POST' ? 'm-post' : 'm-get'">{{ e.method }}</span>
          <span class="har-wf__url">{{ e.url }}</span>
          <span class="har-wf__status" :class="statusClass(e.status)">{{ e.status }}</span>
        </div>
        <div class="har-wf__bar" role="img" :aria-label="`${e.method} ${e.url} 状态 ${e.status}`">
          <span
            v-for="p in phases"
            :key="p.key"
            class="har-wf__seg"
            :style="{ background: (e as any)[p.key] > 0 ? p.color : 'transparent', width: widthOf((e as any)[p.key]) + '%' }"
            :title="`${p.key}: ${(e as any)[p.key]}ms`"
          ></span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.har-wf {
  border: 1px solid var(--har-line);
  border-radius: 10px;
  padding: 1.25rem;
  background: var(--vp-bg, #fff);
}
.dark .har-wf { background: #0f172a; }
.har-wf__legend {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-bottom: 1rem;
  padding-bottom: 0.9rem;
  border-bottom: 1px solid var(--har-line);
}
.har-wf__legend-item {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-family: var(--vp-font-family-mono);
  font-size: 0.7rem;
  color: var(--har-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.har-wf__swatch {
  width: 10px;
  height: 10px;
  border-radius: 2px;
}
.har-wf__rows {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.har-wf__row {
  display: grid;
  grid-template-columns: 240px 1fr;
  gap: 1rem;
  align-items: center;
}
.har-wf__label {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-family: var(--vp-font-family-mono);
  font-size: 0.75rem;
  overflow: hidden;
}
.har-wf__method {
  font-weight: 700;
  padding: 1px 5px;
  border-radius: 3px;
  font-size: 0.65rem;
}
.m-get { background: rgba(22, 163, 74, 0.12); color: var(--har-green); }
.m-post { background: rgba(234, 88, 12, 0.12); color: var(--har-amber); }
.har-wf__url {
  color: var(--har-ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}
.dark .har-wf__url { color: var(--har-paper); }
.har-wf__status {
  font-size: 0.65rem;
  padding: 1px 5px;
  border-radius: 3px;
  font-weight: 600;
}
.is-2xx { color: var(--har-green); }
.is-3xx { color: var(--har-blue); }
.is-4xx { color: var(--har-amber); }
.is-5xx { color: #ef4444; }
.har-wf__bar {
  display: flex;
  height: 18px;
  border-radius: 3px;
  overflow: hidden;
  background: rgba(0, 0, 0, 0.04);
}
.dark .har-wf__bar { background: rgba(255, 255, 255, 0.04); }
.har-wf__seg {
  display: inline-block;
  height: 100%;
  min-width: 0;
  transition: opacity 0.15s ease;
}
.har-wf__row:hover .har-wf__seg { opacity: 0.85; }

@media (max-width: 768px) {
  .har-wf__row {
    grid-template-columns: 1fr;
    gap: 0.35rem;
  }
  .har-wf__label { font-size: 0.7rem; }
}
</style>
