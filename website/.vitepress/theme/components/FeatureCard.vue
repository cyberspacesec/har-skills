<script setup lang="ts">
// 特性卡片 —— 用能力树六色作为分类语义色，每张卡片绑定一个能力域。
// 左侧色条编码分类，hover 时色条延伸。

defineProps<{
  color?: string // hex
  tag?: string   // 分类标签
  title: string
  desc: string
  link?: string
}>()
</script>

<template>
  <a class="feat-card" :href="link" :style="{ '--card-color': color || 'var(--har-blue)' }">
    <span class="feat-card__tag" v-if="tag">{{ tag }}</span>
    <h3 class="feat-card__title">{{ title }}</h3>
    <p class="feat-card__desc">{{ desc }}</p>
    <span class="feat-card__more" v-if="link">→</span>
  </a>
</template>

<style scoped>
.feat-card {
  position: relative;
  display: block;
  padding: 1.25rem 1.25rem 1.25rem 1.5rem;
  border: 1px solid var(--har-line);
  border-radius: 8px;
  text-decoration: none;
  color: inherit;
  background: var(--vp-bg, #fff);
  transition: transform 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease;
  overflow: hidden;
}
.feat-card::before {
  content: '';
  position: absolute;
  left: 0; top: 0; bottom: 0;
  width: 4px;
  background: var(--card-color);
  transition: width 0.18s ease;
}
.feat-card:hover {
  transform: translateY(-2px);
  border-color: var(--card-color);
  box-shadow: 0 12px 30px -12px rgba(11, 18, 32, 0.18);
}
.feat-card:hover::before { width: 6px; }
.feat-card__tag {
  font-family: var(--vp-font-family-mono);
  font-size: 0.65rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--card-color);
  font-weight: 600;
}
.feat-card__title {
  font-family: var(--vp-font-family-mono);
  font-size: 1.05rem;
  margin: 0.4rem 0 0.5rem;
  font-weight: 600;
  letter-spacing: -0.01em;
}
.feat-card__desc {
  font-size: 0.85rem;
  line-height: 1.55;
  color: var(--har-muted);
  margin: 0;
}
.feat-card__more {
  display: inline-block;
  margin-top: 0.75rem;
  font-family: var(--vp-font-family-mono);
  font-size: 0.8rem;
  color: var(--card-color);
  font-weight: 600;
}
</style>
