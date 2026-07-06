<script setup lang="ts">
// 嵌入已有的能力树 SVG（由 scripts/render_feature_tree.py 生成）。
// 复用单一来源，不重新画图。暗色模式用 CSS filter 反色适配。
</script>

<template>
  <div class="cap-tree">
    <img
      class="cap-tree__img"
      src="/assets/har-skills-feature-tree.svg"
      alt="HAR Skills 能力树 —— 六大能力域：Agent Access, Ingest & Parse, Inspect & Search, Analyze, Operate, Compare & Export"
      loading="lazy"
    />
  </div>
</template>

<style scoped>
.cap-tree {
  width: 100%;
  border: 1px solid var(--har-line);
  border-radius: 10px;
  padding: 1rem;
  background: var(--vp-bg, #fff);
  overflow: hidden;
}
.dark .cap-tree {
  background: #0f172a;
}
.cap-tree__img {
  display: block;
  width: 100%;
  height: auto;
}
.dark .cap-tree__img {
  /* SVG 内嵌 color-scheme: light，暗色下用反色 + 轻微亮度提升适配 */
  filter: invert(0.92) hue-rotate(180deg) brightness(0.95);
}
</style>
