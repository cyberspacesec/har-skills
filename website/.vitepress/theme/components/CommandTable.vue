<script setup lang="ts">
// CLI 命令速查表组件 —— 用于命令参考页统一渲染命令清单。
// 数据驱动，避免在多处 Markdown 重复维护表格。

defineProps<{
  commands: Array<{
    name: string
    desc: string
    level?: string
  }>
}>()
</script>

<template>
  <div class="cmd-table">
    <table>
      <thead>
        <tr>
          <th>命令</th>
          <th>用途</th>
          <th>分级</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="c in commands" :key="c.name">
          <td><code>{{ c.name }}</code></td>
          <td>{{ c.desc }}</td>
          <td><span class="cmd-table__level" v-if="c.level">{{ c.level }}</span></td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.cmd-table {
  width: 100%;
  overflow-x: auto;
}
.cmd-table table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}
.cmd-table th,
.cmd-table td {
  text-align: left;
  padding: 0.6rem 0.75rem;
  border-bottom: 1px solid var(--har-line);
}
.cmd-table th {
  font-family: var(--vp-font-family-mono);
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--har-muted);
  font-weight: 600;
}
.cmd-table code {
  font-family: var(--vp-font-family-mono);
  font-size: 0.85em;
  background: rgba(37, 99, 235, 0.08);
  color: var(--vp-c-brand-1);
  border-radius: 4px;
  padding: 0.1em 0.35em;
}
.cmd-table__level {
  font-family: var(--vp-font-family-mono);
  font-size: 0.65rem;
  color: var(--har-muted);
  border: 1px solid var(--har-line);
  border-radius: 3px;
  padding: 1px 5px;
}
</style>
