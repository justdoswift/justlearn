# JustLearn

JustLearn 是一个原生桌面的 AI 私人教师：在聊天里学、在聊天里测，每个学习节点有一张知识热力图。路线、格子、阈值和掌握度都由模型在对话中定义并更新。

用 Go + [MyGo](https://mygo.egoist.dev) 绘制原生界面，不使用 webview，也没有网页版。数据保存在本机 SQLite，模型通过任意 OpenAI 兼容接口调用。

![JustLearn](docs/screenshot.png)

## 功能

- **聊天式教学**：一次只问一个问题，讲一小段、问一个检查问题，答完再继续；不会一次倒完整节课。路线放在右侧，聊天里只给一句摘要。
- **知识热力图**：每个节点 3–9 个知识格子，按 0–3 级掌握度着色；所有格子达到阈值后前进到下一节。
- **要点卡可选**：只有用户通过几轮问答弄懂一个侧面后，才生成一张 200–300 字的要点卡，不把整节课写成文章。
- **小测**：一次一题，答题后行内给出正确答案与解析，并把真实回答写回热力图证据；不会一口气发整份测验。
- **多条路线**：左侧列出所有学习路线，点击切换；右侧只展示当前节点的知识热力图。
- **搜索与管理**：侧栏搜索路线；路线右侧的 `×` 可删除（会清理节点、格子和相关消息）。
- **路线管理**：新建、切换、删除路线；删除路线会连同节点、格子、相关会话和消息一起删除，删除前有确认。
- **模型选择**：输入框右侧显示当前模型，菜单可切换；`/models` 可用时自动列出服务端模型。
- **导入文本**：输入框左侧 `+` 可导入 `.txt` / `.md`，把内容放进草稿继续编辑。
- **原生桌面**：墨色/浅色跟随系统；标题栏与交通灯融合；SQLite 本地数据；Key 存系统钥匙串。

![Composer states](docs/composer.png)

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

## 模型

任何 OpenAI 兼容的 `/chat/completions` 服务都可以。默认读取：

- `PI_BASE_URL`
- `PI_MODEL`
- `PI_API_KEY`

模型列表接口是可选的；服务不提供 `/models` 时，输入框的模型菜单会保留当前模型和「在设置中配置…」。应用内的模型选择会写入本地设置，并优先于环境变量。

## 结构

```
desktop/
  main.go            应用入口、窗口、菜单
  app.go             状态、刷新、模型与文本导入
  sidebar.go         会话与目标
  chat.go            顶栏、消息、课段、小测、胶囊输入框
  route.go           路线、热力图、节点列表
  settings.go        模型设置浮层
  internal/store     SQLite 数据层
  internal/agent     流式模型客户端、工具循环、教学提示词
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
