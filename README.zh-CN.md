<p align="center">
  <img src="docs/icon.png" alt="" width="96">
</p>

<h1 align="center">Mirra 镜衣</h1>

<p align="center">
  <a href="README.md">English</a> |
  <strong>简体中文</strong>
</p>

<p align="center">
  从三个参考图库里挑一张脸、一套服装和一个姿势，交给图像模型画出它们描述的那个人。用 Go 和 <a href="https://github.com/nsgo-dev/nsgo">NSGo</a> 写成，界面由各平台自己的原生控件绘制；同一份代码构建 macOS、iOS、Android 和 Windows 版本。
</p>

<p align="center">
  <img src="docs/screenshots/zh-CN/phone-home.png" alt="手机上的工作室：今日造型和三个要素" width="260">
  <img src="docs/screenshots/zh-CN/phone-me.png" alt="我的：App 信息和 AI 服务" width="260">
</p>

<p align="center">
  <img src="docs/screenshots/zh-CN/desktop.png" alt="桌面工作室：素材库、画布和本次造型" width="800">
</p>

<p align="center">
  <img src="docs/screenshots/zh-CN/desktop-settings.png" alt="桌面上的 AI 服务设置" width="800">
</p>

## 功能

- **人脸、服装、姿势。** 选一张脸（AI 生成，CC0）、一套服装（没有人穿的衣服）和一个姿势（无脸的灰色人偶），App 让图像模型生成一张具备这些的成年人照片。只有姿势是必选的。
- **两种布局。** 窗口窄于 900pt 时是手机布局：底部有工作室、衣橱、作品、我的四个标签，造型分三步选好。更宽时是桌面工作室：左边是素材库，中间是可缩放的画布（本次造型或生成结果），右边是汇总和「生成造型」，AI 服务设置在弹窗里。
- **贴合图库的筛选。** 人脸按发长，服装按季节和场合，姿势按机位。
- **结果由你决定去留。** 新生成的图片要点「保存到作品」才会入库。可以再生成，也可以一键给同一张脸、同一个姿势换一套衣服；可以收藏、导出、删除。
- **衣橱。** 用 ♡ 收藏喜欢的服装，上传自己的衣服照片（平铺或挂拍），随时看到当前正在穿的那套。
- **造型册。** 保存的作品带编号，可以收藏，再次打开时会带回生成它的那套造型。
- **云端或本地。** 支持任何 OpenAI 兼容的图片编辑接口（`POST /images/edits`）：带 Key 的云端 API，或你自己部署的服务（可以不填 Key）。生成前可以先测试连接；Key 存在系统钥匙串里。
- **补充要求。** 写一句光线、背景之类的要求，随 prompt 一起发送。
- **不用人像照片。** App 不读取任何人的照片：姿势库和服装库的条款禁止用它们的图片表现真实的人。上传的只能是衣服照片。
- 界面有英文和简体中文。

## 素材来源

| 槽位 | 图库 | 条款 |
| --- | --- | --- |
| 姿势 | [POSE-DB-01](https://ai-pose-gallery.pages.dev) | 免费使用，包括商用；禁止再分发，禁止批量或用工具自动下载 |
| 服装 | [COSTUME-DB-01](https://ai-costume-gallery.pages.dev) | 同上 |
| 人脸 | [FACE REFERENCE](https://ai-portrait-gallery.pages.dev) | CC0 |

图片归各图库作者所有。本仓库不包含任何一张，App 也遵守图库的条款：

- 每次都在线读取图库的目录。浏览时缩略图从图库加载，和用浏览器看一样；除了系统的图片缓存，App 自己不保存副本。
- 只有按下「生成」时才下载原图，而且只下载你选中的那几张，每张一次。
- 不提供姿势库里的双人和露骨分类，也不提供服装库里的内衣和「セクシー」分类。

姿势库和服装库要求不要用它们的图片表现真实的人或未成年人，也不要用于违反公序良俗的用途；人脸库要求不要用来冒充真实的人。用这个 App 做出来的东西同样受这些条款约束，详见 [NOTICE.md](NOTICE.md)。

## 工作原理

```
图库目录（HTML / JS）  ──►  条目：id、分类、标题、缩略图、原图
                                │
    选人脸、服装（或你自己的衣服）和姿势
                                │
     原图（各下载一次）+ prompt
                                │
            POST {接口地址}/images/edits（multipart，image[] 按顺序）
                                │
       b64_json 或 url  ──►  先展示；点「保存到作品」才入库
```

图片按姿势、服装、人脸的顺序发送，prompt 按这个顺序给它们编号，并说明每张的用途：人偶的姿势和机位、要穿上的服装、要用的脸和发型。

| 文件 | 内容 |
| --- | --- |
| [`gallery.go`](gallery.go) | 读取三个图库的目录，以及过滤掉哪些内容 |
| [`generate.go`](generate.go) | prompt 和 `/images/edits` 请求 |
| [`works.go`](works.go) | 造型册：带编号的作品，造型和收藏状态存在旁边的 JSON 里 |
| [`service.go`](service.go) | 云端 / 本地服务和连接测试 |
| [`wardrobe.go`](wardrobe.go) | 收藏的服装、你上传的衣服、当前正在穿的那套 |
| [`app.go`](app.go) | App 的状态、布局选择和生成流程 |
| [`library.go`](library.go) | 两种布局共用的部分：素材库的筛选和网格、本次造型的条目、预览 |
| [`phone.go`](phone.go) | 手机布局和三步编辑页 |
| [`desktop.go`](desktop.go) | 桌面工作室的顶栏和三栏 |
| [`pages.go`](pages.go) | 结果、作品和 AI 服务设置 |
| [`theme.yaml`](theme.yaml) | App 固定使用的深色琥珀主题 |
| [`locales/`](locales) | 英文和中文文案 |
| [`tools/appicon`](tools/appicon) | 绘制 `assets/Icon.png` |

## 运行

需要 Go 1.27.4 或更高版本。在 Mac 上（也能构建 iOS 和 Android 版本）：

```sh
git clone https://github.com/houko/nsgo-pose
cd nsgo-pose
go tool nsgo dev                     # 本机
go tool nsgo dev -lang zh-Hans       # 中文界面
go tool nsgo dev -target ios_simulator
go tool nsgo dev -target android     # 已连接的设备或模拟器
go tool nsgo build                   # 打包 nsgo.yaml 里列出的所有平台，输出到 dist/
```

`go tool nsgo doctor` 会告诉你这台机器能构建、运行哪些平台。

### 配置生图接口

手机上打开「我的」，桌面上点右上角的设置按钮，选「云端 API」或「本地部署」，填好后点「测试连接」（会请求服务的 `GET /models`）。

| 字段 | 默认值 | 说明 |
| --- | --- | --- |
| 接口地址 | `https://api.openai.com/v1`（云端）、`http://127.0.0.1:8080/v1`（本地） | 任何 OpenAI 兼容的接口、网关或本地服务；App 会在后面拼上 `/images/edits` |
| API Key | | 存在 macOS / iOS 钥匙串、Android Keystore 或 Windows 凭据管理器里；本地部署可以留空 |
| 模型 | `gpt-image-2` | 需要支持多张参考图的模型 |
| 尺寸 | `1024x1536` | 作为 `size` 参数传入 |

本地服务同样必须是 OpenAI 兼容的接口，比如 ComfyUI 自己的协议就不支持。每按一次「生成」就是一次调用，云端 API 会按次计费。目前这个请求只在测试里的本地模拟接口上验证过，还没有对真实的生图接口测试过；欢迎反馈各个模型或网关的实际表现。

## 平台

| 平台 | 状态 |
| --- | --- |
| macOS | 可以运行；在真实窗口里测试过 |
| iOS、Android、Windows | `nsgo build` 能打包（iOS 模拟器 App、用调试密钥签名的 APK、Windows 安装程序）；还没有在设备或模拟器上运行过 |
| Linux | 还不在构建目标里；NSGo 支持，这里没测过 |
| 网页 | 不支持：浏览器里没有安全存放 API Key 的地方（`cx.Secrets()` 在网页上不可用） |

## 隐私

- 只有按下「生成」或「测试连接」时，App 才会访问你配置的服务，发送的只有你选的图片（图库图片或你的衣服照片）和 prompt。
- 你上传的衣服照片只存在 App 的数据目录里。
- 图库会看到你打开的缩略图和原图的请求，和用浏览器访问一样。
- 新生成的图片在保存前单独存放；保存到作品后一直留在 App 的数据目录里，直到你删除。

## 开发

```sh
go tool nsgo test .                                    # headless 测试，图库和接口都用本地模拟
go tool nsgo test -target macos .                      # 同一套测试在真实窗口里跑
go tool nsgo check -target macos,ios,android,windows   # 对每个平台做 vet
go run ./tools/appicon                                 # 重新绘制图标
```

测试不访问网络：`fixtures_test.go` 按图库的格式提供虚构的页面数据，以及一个模拟的 `/images/edits`。

如果哪个图库改了页面格式，选择页会显示「图库加载失败」和解析器的报错；解析代码在 `gallery.go`。

## 许可

代码使用 [MIT](LICENSE) 许可。图库的图片不属于本仓库，仍受其所有者的条款约束。
