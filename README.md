# JustLearn

JustLearn 是一个原生桌面的 AI 私人教师：在聊天里学、在聊天里测，每个学习节点有一张知识热力图。路线、格子、阈值和掌握度都由模型在对话中定义并更新。

用 Go + [MyGo](https://mygo.egoist.dev) 绘制原生界面，不使用 webview，也没有网页版。数据保存在本机 SQLite，模型通过任意 OpenAI 兼容接口调用。

![JustLearn](docs/screenshot.png)

## 功能

- **聊天式教学**：一次只问一个问题，讲一小段、问一个检查问题，答完再继续；不会一次倒完整节课。路线放在左侧，聊天里只给一句摘要。
- **知识热力图**：模型在对话里为每个节点维护 3–9 个知识格子与 0–3 级掌握度；当前界面暂时收起右侧热力图面板，数据和工具仍然保留。
- **要点卡可选**：只有用户通过几轮问答弄懂一个侧面后，才生成一张 200–300 字的要点卡，不把整节课写成文章。
- **深色工作台**：Xcode 式全局工具栏、源码列表式侧栏和系统蓝强调；课文与聊天统一无衬线，课文配 2px 蓝色页边线。
- **小测**：一次一题，答题后行内给出正确答案与解析，并把真实回答写回热力图证据；不会一口气发整份测验。
- **路线与节点**：左侧列出学习路线和全部节点；节点状态用绿点（学习中）、对号（掌握）和留空（未学）表示。
- **搜索与管理**：侧栏搜索节点；新建、切换、删除路线；删除路线会连同节点、格子、相关会话和消息一起删除，删除前有确认。
- **模型选择**：工具栏右侧显示当前模型，菜单可切换；`/models` 可用时自动列出服务端模型。
- **导入文本**：输入框左侧 `+` 可导入 `.txt` / `.md`，把内容放进草稿继续编辑。
- **应用内更新**：从 GitHub Release 检查签名更新，下载后重启进入新版本；菜单「JustLearn → 检查更新…」可手动检查。
- **原生桌面**：固定深色主题；标题栏与交通灯融合；SQLite 本地数据；Key 存系统钥匙串。

![Composer](docs/composer.png)

## 运行

要求 Go 1.27+（MyGo 使用 Go 1.27.3 工具链），macOS 12+ / Windows 10+ / Linux（WebKitGTK 4.1）。

```bash
git clone https://github.com/justdoswift/justlearn.git
cd justlearn
cp .env.example .env        # 填入 PI_BASE_URL / PI_MODEL / PI_API_KEY
cd desktop
go tool mygo dev
```

首次运行也可以在应用内「设置」里填 Base URL、模型和 API Key，Key 会存进系统钥匙串；开发时用根目录 `.env` 更方便。

## 打包

```bash
cd desktop
go tool mygo build
```

产物在 `desktop/build/<os>-<arch>/`。macOS 会生成 `.app` 和 `.dmg`。

## 发布与更新

正式包用 Developer ID 签名、公证，并从 GitHub Release 分发；安装后的副本通过 MyGo 的签名更新机制自动检查新版本。

首次发布前，确认 `desktop/mygo.json` 里的 `macos.signingIdentity` 与 `macos.notarize.keychainProfile` 对应本机的 Developer ID 和公证凭据。更新签名密钥单独生成，私钥不要进仓库：

```bash
cd desktop
go tool mygo keygen -o "$HOME/Library/Application Support/mygo/update-keys/justlearn"
```

把输出的 `publicKey` 填进 `desktop/mygo.json` 的 `updates`，私钥保存在本机。发布时先提升 `version` 并写 `CHANGELOG.md` 里对应的 `## <version>` 段，提交后从仓库根目录运行：

```bash
scripts/release.sh
```

脚本会构建 `darwin/universal`、签名、公证、生成更新归档与 manifest，并用 `-upload` 上传到该版本的 GitHub **draft** release。检查草稿后发布：

```bash
gh release edit v0.0.1 --draft=false
```

也可以 `PUBLISH=1 scripts/release.sh` 在构建上传后直接发布。更新私钥默认从 `~/Library/Application Support/mygo/update-keys/justlearn/mygo-update.key` 读取，可用 `MYGO_UPDATER_PRIVATE_KEY_FILE` 或 `MYGO_UPDATER_PRIVATE_KEY` 覆盖。

## 模型

任何 OpenAI 兼容的 `/chat/completions` 服务都可以。默认读取：

- `PI_BASE_URL`
- `PI_MODEL`
- `PI_API_KEY`

模型列表接口是可选的；服务不提供 `/models` 时，工具栏的模型菜单会保留当前模型和「在设置中配置…」。应用内的模型选择会写入本地设置，并优先于环境变量。

## 结构

```
desktop/
  main.go            应用入口、窗口、菜单
  app.go             状态、刷新、模型与文本导入
  toolbar.go         全局工具栏、模型菜单
  sidebar.go         搜索、路线与节点
  chat.go            消息、课段、小测、输入条
  route.go           热力图面板（当前暂不渲染）
  settings.go        模型设置对话框
  internal/store     SQLite 数据层
  internal/agent     流式模型客户端、工具循环、教学提示词
scripts/release.sh   签名、公证并上传 GitHub Release
CHANGELOG.md         应用内更新显示的版本说明
```

设计规范见 [DESIGN.md](DESIGN.md)。

## 测试

```bash
cd desktop
go test ./...
```

渲染测试会把界面输出到 `testdata/`；`TestAgentLive` 会跑一轮真实模型闭环（定路线 → 讲课 → 出题 → 答题 → 更新热力图），需要 `JUSTLEARN_LIVE=1` 和 API Key，默认跳过。

## License

MIT
