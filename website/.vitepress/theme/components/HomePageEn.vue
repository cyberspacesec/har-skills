<script setup lang="ts">
import HarHero from './HarHero.vue'
import HarWaterfall from './HarWaterfall.vue'
import FeatureCard from './FeatureCard.vue'
import CapabilityTree from './CapabilityTree.vue'
</script>

<template>
  <HarHero
    eyebrow="AI-Native · Go SDK · CLI · Skill"
    title="A toolkit for every byte of HTTP traffic"
    highlight="HTTP traffic"
    tagline="HAR Skills is an AI-agent-first HAR (HTTP Archive) analysis toolkit. One binary lets an agent parse, search, audit, score, redact, replay, transform, and diff network captures — and export results in machine-readable formats."
    :primary="{ text: 'Get started →', link: '/en/quick-start' }"
    :secondary="{ text: 'CLI reference', link: '/en/cli/global-flags' }"
  />

  <section class="har-section">
    <p class="har-section__eyebrow">Capability tree</p>
    <h2 class="har-section__title">Six domains across the HAR lifecycle</h2>
    <p class="har-section__lede">From ingest, parse, inspect, analyze, operate to compare & export — each domain maps to a set of CLI commands and SDK methods. Color is classification.</p>
    <CapabilityTree />
  </section>

  <section class="har-section">
    <p class="har-section__eyebrow">Access methods</p>
    <h2 class="har-section__title">Four ways in, one capability set</h2>
    <p class="har-section__lede">Whether you install a skill for an AI agent, run commands in a terminal, call the SDK from Go, or wrap it as an MCP service — you reach the same 40 modules and 70+ methods.</p>
    <div class="har-grid har-grid--4">
      <FeatureCard color="#2563eb" tag="Agent Access" title="AI Agent Skill" desc="Via the CLAUDE.md progressive-disclosure doc, an agent bootstraps in one prompt." link="/en/access/skill" />
      <FeatureCard color="#0891b2" tag="CLI" title="Command-line tool" desc="24 Cobra commands, stdin/JSON/CSV/YAML output, pipe-friendly." link="/en/access/cli" />
      <FeatureCard color="#9333ea" tag="SDK" title="Go SDK" desc="40 root-package modules, zero runtime deps, functional + struct dual API." link="/en/access/sdk" />
      <FeatureCard color="#ea580c" tag="MCP" title="MCP wrapper" desc="Wrap as a Model Context Protocol server for Claude and other clients." link="/en/access/mcp" />
    </div>
  </section>

  <section class="har-section">
    <p class="har-section__eyebrow">Signature visualization</p>
    <h2 class="har-section__title">The waterfall, on the home page</h2>
    <p class="har-section__lede">This is the tool's native visual language: each request is a bar, blocked→dns→connect→ssl→send→wait→receive colored by phase, scaled to real duration. Below is what the waterfall command produces.</p>
    <HarWaterfall />
    <p class="har-section__caption">Tip: run <code>har -f capture.har waterfall</code> for the ASCII version; <code>--critical-path</code> for the critical path, <code>--sla</code> to check SLA compliance.</p>
  </section>

  <section class="har-section">
    <p class="har-section__eyebrow">What agents can do</p>
    <h2 class="har-section__title">From one sentence to one command</h2>
    <p class="har-section__lede">Natural-language tasks an agent receives, mapped to the command it runs.</p>
    <div class="har-grid har-grid--3">
      <FeatureCard color="#16a34a" tag="Inspect" title="Find all failed requests" desc="har -f capture.har find --errors --format json" />
      <FeatureCard color="#9333ea" tag="Analyze" title="Score this capture" desc="har -f capture.har performance" />
      <FeatureCard color="#ea580c" tag="Operate" title="Redact before sharing" desc="har -f capture.har redact --redact-ips -o clean.har" />
      <FeatureCard color="#475569" tag="Compare" title="Diff two releases" desc="har diff v1.har v2.har --compare-by-url" />
      <FeatureCard color="#475569" tag="Export" title="Turn into curl" desc="har -f capture.har export curl" />
      <FeatureCard color="#0891b2" tag="Ingest" title="Split a big file by domain" desc="har -f big.har split --by domain" />
    </div>
  </section>

  <section class="har-section har-cta">
    <h2 class="har-section__title">Start</h2>
    <p class="har-section__lede">Three steps: install the binary, run one command, read one page.</p>
    <div class="har-cta__code"><pre><code># 1. Install
go install github.com/cyberspacesec/har-skills/cmd/har@latest

# 2. Run a command
har -f capture.har info

# 3. Read the docs
# EN: /en/quick-start   ZH: /zh/quick-start</code></pre></div>
    <div class="har-cta__actions">
      <a class="har-btn har-btn--primary" href="/en/quick-start">Quick start</a>
      <a class="har-btn har-btn--ghost" href="/en/sdk/data-structures">SDK guide</a>
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
