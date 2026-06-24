#!/usr/bin/env python3
"""Render the HAR Skills README feature tree as a static SVG.

The script intentionally has no third-party dependencies so the diagram can be
regenerated in any checkout with a standard Python installation.
"""

from __future__ import annotations

import argparse
import html
import textwrap
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable


WIDTH = 1440
HEIGHT = 940


@dataclass(frozen=True)
class Branch:
    title: str
    summary: str
    color: str
    items: tuple[str, ...]
    x: int
    y: int


BRANCHES: tuple[Branch, ...] = (
    Branch(
        title="Agent Access",
        summary="AI-first entry points",
        color="#2563eb",
        items=(
            "CLAUDE.md skill guide",
            "24 CLI tool calls",
            "Go SDK embedding",
            "MCP wrapper ready",
        ),
        x=380,
        y=150,
    ),
    Branch(
        title="Ingest & Parse",
        summary="Turn captures into structured data",
        color="#0891b2",
        items=(
            "file, bytes, reader, stdin",
            "gzip auto-detect",
            "standard and optimized parsers",
            "lazy and streaming workflows",
        ),
        x=910,
        y=150,
    ),
    Branch(
        title="Inspect & Search",
        summary="Find the requests that matter",
        color="#16a34a",
        items=(
            "info, list, find",
            "URL, regex, method, status",
            "headers, cookies, domains",
            "timing, waterfall, content",
        ),
        x=380,
        y=375,
    ),
    Branch(
        title="Analyze",
        summary="Explain risk and performance",
        color="#9333ea",
        items=(
            "security audit",
            "performance score",
            "cache analysis",
            "connection reuse",
            "duplicate detection",
        ),
        x=910,
        y=375,
    ),
    Branch(
        title="Operate",
        summary="Clean, transform, and replay",
        color="#ea580c",
        items=(
            "validate HAR spec",
            "redact secrets and PII",
            "rewrite URLs and headers",
            "replay HTTP requests",
            "extract response bodies",
        ),
        x=380,
        y=600,
    ),
    Branch(
        title="Compare & Export",
        summary="Package traffic for teams and tools",
        color="#475569",
        items=(
            "diff, merge, split",
            "build indexes",
            "curl, wget, Python",
            "Postman, CSV, Markdown, HTML",
            "JSON, JSONL, YAML, XML",
        ),
        x=910,
        y=600,
    ),
)


def esc(value: str) -> str:
    return html.escape(value, quote=True)


def wrap_lines(text: str, width: int) -> list[str]:
    return textwrap.wrap(text, width=width, break_long_words=False)


def svg_text(
    x: int,
    y: int,
    text: str,
    *,
    css_class: str,
    max_chars: int | None = None,
    line_height: int = 22,
) -> str:
    lines = wrap_lines(text, max_chars) if max_chars else [text]
    tspans = []
    for index, line in enumerate(lines):
        dy = 0 if index == 0 else line_height
        tspans.append(f'<tspan x="{x}" dy="{dy}">{esc(line)}</tspan>')
    return f'<text class="{css_class}" x="{x}" y="{y}">{"".join(tspans)}</text>'


def connector(start_x: int, start_y: int, end_x: int, end_y: int, color: str) -> str:
    mid = start_x + (end_x - start_x) * 0.58
    return (
        f'<path class="connector" d="M {start_x} {start_y} '
        f'C {mid:.0f} {start_y}, {mid:.0f} {end_y}, {end_x} {end_y}" '
        f'stroke="{color}" />'
    )


def branch_card(branch: Branch) -> str:
    card_w = 450
    card_h = 204
    item_x = branch.x + 28
    title_x = branch.x + 28
    parts = [
        f'<g class="branch" id="{esc(branch.title.lower().replace(" ", "-").replace("&", "and"))}">',
        f'<rect class="card" x="{branch.x}" y="{branch.y}" width="{card_w}" height="{card_h}" rx="8" />',
        f'<rect x="{branch.x}" y="{branch.y}" width="8" height="{card_h}" rx="4" fill="{branch.color}" />',
        f'<circle cx="{branch.x + 35}" cy="{branch.y + 36}" r="10" fill="{branch.color}" opacity="0.18" />',
        f'<circle cx="{branch.x + 35}" cy="{branch.y + 36}" r="5" fill="{branch.color}" />',
        svg_text(title_x + 24, branch.y + 35, branch.title, css_class="branch-title"),
        svg_text(title_x, branch.y + 66, branch.summary, css_class="branch-summary"),
    ]

    y = branch.y + 98
    for item in branch.items:
        parts.extend(
            [
                f'<circle cx="{item_x}" cy="{y - 5}" r="3.5" fill="{branch.color}" />',
                svg_text(item_x + 16, y, item, css_class="item", max_chars=43, line_height=18),
            ]
        )
        y += 23

    parts.append("</g>")
    return "\n".join(parts)


def render(branches: Iterable[Branch] = BRANCHES) -> str:
    root_x = 62
    root_y = 386
    root_w = 248
    root_h = 120
    root_anchor_x = root_x + root_w
    root_anchor_y = root_y + root_h // 2

    branch_list = list(branches)
    connectors = [
        connector(root_anchor_x, root_anchor_y, branch.x, branch.y + 88, branch.color)
        for branch in branch_list
    ]
    cards = [branch_card(branch) for branch in branch_list]

    return f'''<svg xmlns="http://www.w3.org/2000/svg" width="{WIDTH}" height="{HEIGHT}" viewBox="0 0 {WIDTH} {HEIGHT}" role="img" aria-labelledby="title desc">
  <title id="title">HAR Skills capability tree</title>
  <desc id="desc">A tree-style mind map showing HAR Skills agent access, parsing, inspection, analysis, operations, comparison, and export capabilities.</desc>
  <style>
    :root {{
      color-scheme: light;
    }}
    .bg {{
      fill: #f8fafc;
    }}
    .grid {{
      stroke: #e2e8f0;
      stroke-width: 1;
      opacity: 0.55;
    }}
    .title {{
      fill: #0f172a;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      font-size: 34px;
      font-weight: 780;
      letter-spacing: 0;
    }}
    .subtitle {{
      fill: #475569;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      font-size: 17px;
      font-weight: 450;
      letter-spacing: 0;
    }}
    .root-card {{
      fill: #0f172a;
      filter: drop-shadow(0 16px 34px rgba(15, 23, 42, 0.2));
    }}
    .root-title {{
      fill: #ffffff;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      font-size: 23px;
      font-weight: 760;
      letter-spacing: 0;
    }}
    .root-subtitle {{
      fill: #cbd5e1;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      font-size: 15px;
      font-weight: 500;
      letter-spacing: 0;
    }}
    .tag {{
      fill: #d9f99d;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      font-size: 13px;
      font-weight: 720;
      letter-spacing: 0;
    }}
    .connector {{
      fill: none;
      stroke-width: 3;
      stroke-linecap: round;
      opacity: 0.55;
    }}
    .card {{
      fill: #ffffff;
      stroke: #dbe4ee;
      stroke-width: 1;
      filter: drop-shadow(0 10px 22px rgba(15, 23, 42, 0.08));
    }}
    .branch-title {{
      fill: #111827;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      font-size: 21px;
      font-weight: 760;
      letter-spacing: 0;
    }}
    .branch-summary {{
      fill: #64748b;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      font-size: 14px;
      font-weight: 520;
      letter-spacing: 0;
    }}
    .item {{
      fill: #334155;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      font-size: 14px;
      font-weight: 470;
      letter-spacing: 0;
    }}
    .footer {{
      fill: #64748b;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      font-size: 13px;
      font-weight: 500;
      letter-spacing: 0;
    }}
  </style>

  <rect class="bg" width="{WIDTH}" height="{HEIGHT}" />
  <path class="grid" d="M 0 96 H {WIDTH} M 0 824 H {WIDTH} M 340 116 V 812 M 870 116 V 812" />

  {svg_text(58, 64, "HAR Skills Capability Tree", css_class="title")}
  {svg_text(58, 95, "AI-native HAR analysis for agents, CLIs, SDKs, and automation pipelines", css_class="subtitle")}

  <g id="root">
    <rect class="root-card" x="{root_x}" y="{root_y}" width="{root_w}" height="{root_h}" rx="8" />
    <rect x="{root_x + 22}" y="{root_y + 22}" width="94" height="24" rx="6" fill="#1e293b" />
    {svg_text(root_x + 36, root_y + 39, "AI-native", css_class="tag")}
    {svg_text(root_x + 24, root_y + 76, "HAR Skills", css_class="root-title")}
    {svg_text(root_x + 24, root_y + 102, "HTTP Archive toolkit", css_class="root-subtitle")}
  </g>

  <g id="connectors">
    {chr(10).join(connectors)}
  </g>

  <g id="branches">
    {chr(10).join(cards)}
  </g>

  {svg_text(58, 875, "Generated by scripts/render_feature_tree.py", css_class="footer")}
</svg>
'''


def main() -> None:
    repo_root = Path(__file__).resolve().parents[1]
    default_output = repo_root / "docs" / "assets" / "har-skills-feature-tree.svg"

    parser = argparse.ArgumentParser(description="Render the HAR Skills feature tree SVG.")
    parser.add_argument(
        "-o",
        "--output",
        type=Path,
        default=default_output,
        help=f"output SVG path (default: {default_output})",
    )
    args = parser.parse_args()

    output = args.output
    if not output.is_absolute():
        output = Path.cwd() / output
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(render(), encoding="utf-8")
    print(output)


if __name__ == "__main__":
    main()
