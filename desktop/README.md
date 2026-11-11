# JustLearn 客户端

基于 [MyGo](https://mygo.egoist.dev) 的原生桌面客户端：Go 绘制界面，不需要 webview，也没有网页版。打开就是一个居中的输入框，发送后整个窗口进入聊天；聊天页左侧是路线预览轨，右侧是默认退后的能力画像，设置收在右上角一个极小入口里。

## 运行

```bash
go tool mygo dev
```

开发版以无窗口方式启动，热重载不会把窗口拉到前台；点 Dock 里的「JustLearn Dev」，或用菜单「文件 → 显示窗口」（`⌘O`）打开。这样改代码时不会打断你正在做的事。

首次运行需要模型配置：应用会打开完整设置页，填入 OpenAI 兼容接口的 Base URL、模型名和 API Key。Key 存在系统钥匙串里；开发时也可以直接使用项目根目录 `.env` 里的 `PI_BASE_URL` / `PI_MODEL` / `PI_API_KEY`。主界面的右上角“设置”是空白占位，完整设置从菜单「文件 → 设置…」（`⌘,`）进入。

## 打包

```bash
go tool mygo build
```

产物在 `build/<os>-<arch>/`：macOS 生成 `.app` 和 `.dmg`，Windows 生成安装包，Linux 生成 `.deb`。

## 数据

SQLite 存在 `mygo.App.Path(PathUserData)/justlearn.db`（macOS 为 `~/Library/Application Support/JustLearn/`）。开发版使用单独的 App 标识，数据与正式版分开。

## 结构

- `main.go`：应用入口、窗口与菜单
- `app.go`：状态、刷新与会话切换
- `toolbar.go`：右上角设置入口
- `panels.go`：左侧路线预览轨与设置空白面板
- `heatmap.go`：右侧能力画像、能力聚合与详情卡
- `chat.go`：首页、聊天、课段、小测、输入框、模型浮层
- `canvas.go`：生成式画布：引用、表格、对比、步骤、代码、记忆卡、柱状图
- `settings.go`：完整设置页（菜单 / 首次运行的兜底）
- `icons.go`：线性图标
- `theme.go`：暖纸 / 深墨调色板、字体、Liquid Glass
- `internal/genui`：画布 block 协议、校验与历史摘要
- `internal/store`：SQLite 数据层（路线、节点、能力、知识格子、会话、消息）
- `internal/agent`：模型流式调用、工具循环、教学提示词、`render_canvas`

## 测试

```bash
go test ./...
```

`TestRender` 会把界面渲染成 `testdata/*.png`；`TestAgentLive` 会用真实模型跑一轮「定路线 → 定义能力 → 讲课 → 出题 → 答题 → 更新能力证据」，需要 `JUSTLEARN_LIVE=1` 和 API Key，默认跳过。
