# JustLearn 客户端

基于 [MyGo](https://mygo.egoist.dev) 的原生桌面客户端：Go 绘制界面，不需要 webview，也没有网页版。

## 运行

```bash
go tool mygo dev
```

首次运行需要模型配置：应用会打开「设置」，填入 OpenAI 兼容接口的 Base URL、模型名和 API Key。Key 存在系统钥匙串里；开发时也可以直接使用项目根目录 `.env` 里的 `PI_BASE_URL` / `PI_MODEL` / `PI_API_KEY`。

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
- `sidebar.go` / `chat.go` / `route.go` / `settings.go`：界面
- `internal/store`：SQLite 数据层（路线、节点、热力图格子、会话、消息）
- `internal/agent`：模型流式调用、工具循环、教学提示词

## 测试

```bash
go test ./...
```

`TestRender` 会把界面渲染成 `testdata/*.png`；`TestAgentLive` 会用真实模型跑一轮「定路线 → 讲课 → 出题 → 答题 → 更新热力图」，需要 `JUSTLEARN_LIVE=1` 和 API Key，默认跳过。
