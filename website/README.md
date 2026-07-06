# HAR Skills 文档站 / Documentation Site

本目录是 HAR Skills 项目的 VitePress 文档站源码。

This directory contains the VitePress documentation site for the HAR Skills project.

## 本地开发 / Local Development

前置条件 / Prerequisites：

- Node.js 22+（推荐）/ Node.js 22+ (recommended)
- npm 10+

```bash
# 1. 安装依赖 / Install dependencies
cd website
npm install

# 2. 启动开发服务器（热更新）/ Start dev server with HMR
npm run docs:dev
# 访问 http://localhost:5173

# 3. 构建生产产物 / Build for production
npm run docs:build
# 产物输出到 website/.vitepress/dist

# 4. 本地预览生产构建 / Preview the production build
npm run docs:preview
# 访问 http://localhost:4173
```

## 目录结构 / Directory Structure

```
website/
├── package.json              # 依赖与脚本 / dependencies & scripts
├── .vitepress/
│   ├── config.ts             # VitePress 站点配置 / site config
│   ├── theme/                # 自定义主题 / custom theme
│   └── dist/                 # 构建产物（gitignore）/ build output (gitignored)
├── public/                   # 静态资源（原样拷贝）/ static assets copied as-is
│   └── assets/
│       └── har-skills-feature-tree.svg   # 能力树图（CI 自动生成）/ capability tree SVG (CI-generated)
├── en/                       # 英文文档 / English docs
└── zh/                       # 中文文档 / Chinese docs
```

## 能力树图 / Capability Tree SVG

`website/public/assets/har-skills-feature-tree.svg` 在 CI 构建时由仓库根目录的 `scripts/render_feature_tree.py` 自动重新生成，无需手动维护。

The capability tree SVG is regenerated automatically during CI by `scripts/render_feature_tree.py` at the repository root — no manual maintenance needed.

如需本地重新生成 / To regenerate locally：

```bash
# 从仓库根目录运行 / Run from repository root
python scripts/render_feature_tree.py
cp docs/assets/har-skills-feature-tree.svg website/public/assets/har-skills-feature-tree.svg
```

## 添加新页面 / Adding New Pages

1. 在 `website/en/` 或 `website/zh/` 下新建 `.md` 文件 / Create a new `.md` file under `website/en/` or `website/zh/`.
2. 文件顶部使用 YAML frontmatter 声明标题等元数据 / Use YAML frontmatter at the top for metadata:

   ```markdown
   ---
   title: 页面标题 / Page Title
   ---
   ```

3. 在对应语言的 `index.md` 或侧边栏配置中引用新页面，使其出现在导航中 / Reference the new page in the corresponding sidebar/index config so it appears in navigation.
4. 中英文内容需保持同步 / Keep English and Chinese content in sync.

## 部署 / Deployment

推送到 `main` 分支且 `website/**` 有变更时，GitHub Actions 工作流 `.github/workflows/docs.yml` 会自动构建并部署到 GitHub Pages。

Pushing to `main` with changes under `website/**` triggers the `.github/workflows/docs.yml` GitHub Actions workflow, which builds and deploys to GitHub Pages automatically.
