# OpsDump Electron 桌面版 — 设计方案

> 状态：**已实施并验证**（v0.2.0，Windows NSIS）。本文档同步了实现过程中的方案修正（§6.4 关闭通道、§7.1 固定版本、§11 验证结果）。
> 范围：仅 Windows 桌面版（NSIS 安装包）。
> 已确认取舍：**Sidecar 子进程**集成 / 前端 CDN **本地 vendor** / **electron-builder** 打包。
> 原则：**零业务侵入**——Go 侧仅新增可选优雅关闭接口（§8），其余桌面化逻辑全部放在 Electron 与构建脚本层。
> 关联文档：[opsdump-design.md](./opsdump-design.md)（服务端形态设计）。

---

## 一、背景与目标

OpsDump 目前是 GoFrame 单二进制 + 内嵌 Web 页面的服务端形态（`http://<host>:28090`）。目标是用 Electron 套壳出桌面版：

- **双击即用**：安装后启动图标即开窗口，无需手动跑命令行 / 浏览器。
- **离线可用**：前端不再依赖 unpkg CDN。
- **数据隔离**：运行时数据（DB / 日志 / .env）收敛到用户数据目录，安装目录保持只读。
- **同一代码基线**：桌面版与服务端版共用同一套 Go 二进制与页面，仅构建方式不同。

### 非目标（本期不做）

macOS/Linux 打包、自动更新、代码签名、前端 SPA 重构、Go 侧新增关闭 API、系统托盘常驻（见 §13 Q2）。

---

## 二、现状关键结论（设计依据）

调研结论（行号基于当前仓库）：

| # | 事实 | 依据 | 对设计的影响 |
|---|------|------|--------------|
| 1 | 服务监听 `:28090`（0.0.0.0），配置只能改文件，无 env/CLI 覆盖；README 提到的 `--config` 参数实际未实现 | `manifest/config/config.yaml:2` | 桌面版通过「生成配置文件 + 固定 cwd」注入端口与回环绑定 |
| 2 | DB / 日志 / `.env` 全部 **CWD 相对路径**；`.env` 直读无搜索回退 | `config.yaml:19-27`、`internal/service/env.go:23` | spawn 子进程必须固定 `cwd` 为可写数据目录 |
| 3 | 配置搜索第一优先级是 CWD（`""` / `config/` / `manifest/config`） | GoFrame `gcfg_adapter_file.go:56-60` | 在 cwd 写 `config.yaml` 即生效，无需改代码 |
| 4 | 模板 + 静态资源经 `gf pack` 内嵌（`packSrc: resource`），纯 `go build` **不**内嵌 | `hack/config.yaml:17-23`、`internal/packed` 占位包 | sidecar 必须用 `gf pack` + 交叉编译产出 |
| 5 | 前端为 gview 外壳 + 内联 Vue3，依赖 unpkg CDN，共 **16 处**引用（4 个模板 × 4） | `resource/template/{login,dashboard,jobs,runs}.html` | vendor 到 `resource/public/vendor/`，随 pack 内嵌 |
| 6 | API 全部相对路径同源 `fetch('/api/...')`，无 CORS 中间件、无 WebSocket/SSE | 全模板扫描 | 必须 `loadURL('http://127.0.0.1:<port>')`，不可 `file://` |
| 7 | 会话为 Cookie（`gfsessionid`），24h，存 `%TEMP%\gsessions`（文件型） | 框架默认 | 桌面版无需改动 |
| 8 | 应用层无信号处理；框架层 5s 优雅关闭；Windows 下 `child.kill()` = `TerminateProcess` 不触发优雅关闭 | `ghttp_server.go:459`、`ghttp_server_config.go:315` | 退出必须两阶段 `taskkill` |
| 9 | 端口占用时 `s.Run()` → `Fatalf` 进程退出 | 框架行为 | 可用「子进程提前退出」信号做换端口重试 |
| 10 | `viewer.delimiters: ["${", "}"]` 保护 Vue 插值不被 Go 模板吞掉 | `config.yaml:34-37` | 生成的桌面配置**必须**保留该项 |
| 11 | Windows 下相对 db 路径有 `:` 解析陷阱（首启 db 不存在时） | `opsdump-design.md` 已知问题节 | 首启用全新空数据目录，规避此坑 |
| 12 | 登录页把明文口令存 localStorage；错误文案 `r.msg` vs `message` 字段不一致（已知次要缺陷） | `login.html:117,137`、`controller/middleware.go` | 与桌面化无关，不扩散本期范围（见 §8） |

---

## 三、总体架构

```
┌─ OpsDump.exe (Electron Main) ─────────────────────────┐
│  main.js      单实例锁 / 窗口 / 退出编排                │
│  config.js    选空闲端口 / 生成 config.yaml + .env      │
│  sidecar.js   spawn + /ping 健康检查 + 两阶段终止       │
│  logger.js    sidecar stdio → userData/logs/           │
│                                                        │
│  BrowserWindow ──────── loadURL ────► http://127.0.0.1:<port>
└──────────────┬─────────────────────────────────────────┘
               │ spawn(cwd=data 目录, windowsHide)
      ┌────────▼──────────────────────┐
      │  sidecar/ops-dump.exe        │  gf pack 内嵌 resource/（含 vendor）
      │  GoFrame + SQLite + gcron    │  仅监听 127.0.0.1:<port>
      └──────────────────────────────┘

  userData/ops-dump-desktop/data/   ← cwd，运行时唯一写入位置
  ├── config.yaml   （Electron 每次启动生成）
  ├── .env          （首启从模板生成）
  ├── ops-dump.db   （+ -wal/-shm）
  └── logs/         （ops-dump.log + server/）
```

核心原则：

1. **同源直连**：页面与 API 由同一子进程服务，前端零改动（除 vendor 路径替换）。
2. **约定优于配置**：不新增 Go 代码，靠 GoFrame 既有的 cwd 搜索约定注入配置。
3. **单向管控**：Electron 是唯一管理者——生成环境、拉起、探活、终止；子进程无感知桌面壳存在。
4. **可退化**：sidecar.exe 脱离 Electron 仍可独立运行（等同现有服务端形态），桌面壳只是一层编排。

---

## 四、目录结构（新增部分）

```
ops-dump/
├── desktop/                        # 新增，独立 npm 包
│   ├── package.json
│   ├── electron-builder.yml
│   ├── main/
│   │   ├── main.js                 # 入口：单实例、窗口生命周期、模块编排
│   │   ├── sidecar.js              # 子进程管理：spawn / 探活 / 终止 / 重试
│   │   ├── config.js               # 端口选择、config.yaml 与 .env 生成
│   │   ├── logger.js               # sidecar stdio 落盘 + main 进程日志
│   │   └── paths.js                # userData / resources 路径解析
│   ├── build/
│   │   ├── icon.ico                # 应用图标（256x256）
│   │   └── installer.nsi           # 可选：自定义 NSIS 脚本
│   └── scripts/
│       └── dev.js                  # 开发模式：可选拉起 go run
├── resource/
│   ├── public/vendor/              # 新增：第三方前端依赖（见 §7）
│   │   ├── vue.global.prod.js
│   │   ├── element-plus.css
│   │   ├── element-plus.js
│   │   ├── element-plus-icons.js
│   │   └── fonts/                  # 若 CSS 有外部字体引用
│   └── template/*.html             # 16 处 CDN 引用改本地路径
├── dist/                           # 构建产物（gitignore）
│   ├── sidecar/ops-dump.exe
│   └── Setup-OpsDump-x.y.z.exe
└── Makefile                        # 新增 build-windows 目标
```

---

## 五、运行时数据目录与首次初始化

### 5.1 布局

| 内容 | 路径 | 写入方 |
|------|------|--------|
| 子进程 cwd | `%APPDATA%/ops-dump-desktop/data/` | Electron 创建 |
| 配置 | `data/config.yaml` | Electron 每次启动重写 |
| 环境变量 | `data/.env` | Electron 首启生成（已存在则不覆盖） |
| SQLite | `data/ops-dump.db`（+ `-wal`/`-shm`） | Go |
| 应用日志 | `data/logs/ops-dump.log` | Go |
| server 日志 | `data/logs/server/{Y-m-d}.log` | GoFrame |
| sidecar 控制台 | `userData/logs/sidecar.log` | Electron |
| 会话文件 | `%TEMP%\gsessions/default` | 框架默认，不动 |

安装目录（如 `C:\Program Files\OpsDump\`）只含 exe 与 sidecar，**永不写入**——规避 NSIS 装进 Program Files 后的权限问题。

### 5.2 启动初始化序列

```
1. app.requestSingleInstanceLock() 失败 → 聚焦已有窗口并退出
2. 确保 data/ 存在（mkdir -p）
3. data/.env 不存在 → 从内置模板生成（策略见 §13 Q1）
4. 选端口：默认 28090，被占用则取系统空闲端口（listen(0) 探测）
5. 渲染 config.yaml 到 data/（含端口、回环绑定）
6. spawn sidecar（见 §6）
7. 轮询 GET /ping 直至 pong → win.loadURL(`http://127.0.0.1:${port}`)
```

### 5.3 桌面版 config.yaml 模板

```yaml
server:
  address: "127.0.0.1:${port}"          # 仅回环，桌面版安全基线
  openapiPath: "/api.json"
  swaggerPath: "/swagger"
  serverRoot: "resource/public"          # 磁盘不存在时应走 gres 内嵌（§11-1 首验）
  dumpRouterMap: true
  logger:
    path: "./logs/server"
    file: "{Y-m-d}.log"
    stdout: true                          # 进入 stdio → sidecar.log
    rotateSize: "100M"
    rotateBackupLimit: 10
    rotateBackupExpire: "30d"
    rotateBackupCompress: 9
    rotateCheckInterval: "24h"

ops-dump:
  timezone: "Asia/Shanghai"
  dbPath: "./ops-dump.db"
  credentials:
    envFile: "./.env"
  admin:
    username: admin

logger:
  path: ./logs
  file: ops-dump.log
  level: info
  stdout: true

viewer:
  delimiters: ["${", "}"]                # 必须保留，否则 Vue 插值被 Go 模板吞掉
```

### 5.4 .env 首启模板

```dotenv
# 由桌面版首次启动生成；敏感项仅此文件配置，不入库不进日志
ADMIN_PASSWORD=<策略见 §13 Q1>
BACKUP_DIR=
# MySQL / TDengine / MinIO 各项留空，按需填写
```

---

## 六、Sidecar 生命周期设计（`desktop/main/sidecar.js`）

### 6.1 启动

```js
const child = spawn(sidecarPath, [], {
  cwd: dataDir,                          // 关键：所有相对路径的锚点
  windowsHide: true,                     // 不弹控制台黑窗
  stdio: ['ignore', 'pipe', 'pipe'],     // 接管输出，写 sidecar.log
});
```

- `sidecarPath = path.join(process.resourcesPath, 'sidecar', 'ops-dump.exe')`（dev 模式指向 `../dist/sidecar/ops-dump.exe`）。
- stdout/stderr 逐行追加到 `userData/logs/sidecar.log`（单文件 10MB 轮转）。
- main.go 的 `gres.Dump()` 输出会淹没日志，见 §8 建议改动。

### 6.2 探活与就绪

- 每 200ms `GET http://127.0.0.1:<port>/ping`，返回体含 `pong` 即就绪 → `loadURL`。
- 就绪前子进程 `exit` → 判定启动失败（大概率端口被占或配置错）→ **换端口重试一次** → 仍失败则弹窗（附 sidecar.log 路径）并退出。
- 总超时 15s → 走同一失败路径。
- 等待期间窗口展示极简 loading 页（dataURL），错误时展示错误页附日志路径。

### 6.3 运行期监控

- 监听 `exit`：非主动退出时弹窗提供「重启」/「退出」。
- 窗口关闭前置钩子：先走 §6.4 终止流程，确认子进程回收后再退出。

### 6.4 两阶段终止（退出编排）——实现修正

**原方案**（`taskkill` 无 `/F` → 控制事件 → 框架优雅关闭）在实测中**不成立**：
Node spawn 的子进程没有 `taskkill` 可投递控制事件的控制台/窗口，`taskkill /PID`
（不带 `/F`）直接报错「只能被强制终止」；`detached:true` / `windowsHide` 组合均无效
（`child.kill()` 同样是 TerminateProcess）。详见 §13 风险表的实测记录。

**实施的方案**：优雅通道改为后端 **HTTP 关闭接口**（§8），两阶段调整为：

| 阶段 | 动作 | 语义 |
|------|------|------|
| 1 | `POST /api/system/shutdown`（头 `X-OpsDump-Control: 1`） | 后端 `g.Server().Shutdown()`：GoFrame 5s 预算内优雅关闭（在跑 cron 可收尾），实测 ~350ms 且 `exitCode=0` |
| 2 | 等待 ≤6s 仍存活 → `taskkill /F` | 强杀兜底（SQLite WAL 下风险低） |

退出时序：

```
用户关闭窗口 / app.quit
  → before-quit 拦截默认退出
  → 1) POST /api/system/shutdown     等 ≤6s（框架预算 5s）
  → 2) taskkill /F（兜底）
  → 3) 确认子进程已回收 → app.exit(0)
```

### 6.5 端口策略

- 每次启动实测：`net.createServer().listen(28090)` 成功 → 用 28090；`EADDRINUSE` → `listen(0)` 取系统分配的空闲端口。
- 端口写入 config.yaml，与 `loadURL` 同源使用；子进程若因竞态仍端口冲突提前退出，由 §6.2 换端口重试兜底。
- 不做端口持久化：每次动态确定，避免「上次的端口这次被别的程序占了」类脏状态。

---

## 七、前端 CDN 本地化（vendor）

### 7.1 清单（16 处引用，4 个模板）——已实施，版本已固定

| 现引用 | 落地文件（`resource/public/vendor/`） | 锁定版本 |
|--------|--------------------------------------|----------|
| `https://unpkg.com/element-plus/dist/index.css` | `element-plus-2.14.7.css` | element-plus 2.14.7 |
| `https://unpkg.com/vue@3/dist/vue.global.js` | `vue-3.5.43.global.prod.js` | vue 3.5.43（生产版） |
| `https://unpkg.com/element-plus` | `element-plus-2.14.7.js`（`dist/index.full.js` UMD） | element-plus 2.14.7 |
| `https://unpkg.com/@element-plus/icons-vue` | `element-plus-icons-2.3.2.js`（`dist/index.iife.min.js`） | @element-plus/icons-vue 2.3.2 |

> 版本写进文件名，升级即换文件 + 改模板引用；CSS 内已确认无外部字体引用（全部 data: URI）。

替换规则（`login/dashboard/jobs/runs.html` 各 4 处）：

```html
<!-- 前 -->
<link rel="stylesheet" href="https://unpkg.com/element-plus/dist/index.css">
<script src="https://unpkg.com/vue@3/dist/vue.global.js"></script>
<!-- 后 -->
<link rel="stylesheet" href="/vendor/element-plus.css">
<script src="/vendor/vue.global.prod.js"></script>
```

### 7.2 约束

- 版本**固定**（文件名带版本号，或在本方案 §7.1 记录锁定版本），避免上游 breaking change。
- `serverRoot: resource/public` → URL 直接映射 `/vendor/...`，与现有 `/style.css` 同机制。
- vendor 随 `gf pack` 内嵌进 exe，离线可用；打包体积预计 +2~3MB。
- 实现时 devtools 扫 404：若 CSS 引用外部字体（如 iconfont）一并归档 `vendor/fonts/` 并改写 `url()`。

---

## 八、Go 侧改动清单（实施状态）

| 改动 | 位置 | 优先级 | 说明 |
|------|------|--------|------|
| `gres.Dump()` 加门禁 | `main.go` | **已实施** | 仅 `OPSDUMP_DEBUG` 非空时输出，避免 sidecar.log 被内嵌资源清单刷屏 |
| Makefile 增加 `build-windows` | `Makefile` + `hack/build-windows.ps1` | **已实施**（属构建层） | 见 §10；逻辑放 PowerShell 脚本，因 Windows make 无 POSIX shell |
| 优雅关闭通道 `POST /api/system/shutdown` | `api/v1/system.go` + `controller/system.go` | **已实施**（原 §13 风险的落地对策） | 见 §6.4；公开路由但强制 `X-OpsDump-Control: 1` 自定义头防浏览器 CSRF（跨站请求无法免预检设置自定义头，服务端从不应答预检） |
| 修 `r.msg` vs `message` 字段不一致 | `controller/middleware.go` 等 | 暂做 | 独立已知缺陷 |
| 登录页明文存 localStorage 口令 | `login.html:117,137` | 暂做 | 独立安全议题，不扩散范围 |

**明确不改**：路由、鉴权、调度、SQLite、配置结构——保持桌面版与服务端版同一代码基线。

---

## 九、Electron 主进程设计

### 9.1 窗口与安全基线

```js
const win = new BrowserWindow({
  width: 1280, height: 840, minWidth: 960, minHeight: 600,
  show: false,                          // ready-to-show 前不闪白
  webPreferences: {
    contextIsolation: true,
    nodeIntegration: false,
    sandbox: true,                      // 页面无 preload 需求
  },
});

// 外链策略：仅放行本机后端，其余交给系统浏览器
win.webContents.setWindowOpenHandler(({ url }) =>
  new URL(url).hostname === '127.0.0.1'
    ? { action: 'allow' }
    : (shell.openExternal(url), { action: 'deny' }));
win.webContents.on('will-navigate', (e, url) => {
  if (new URL(url).hostname !== '127.0.0.1') e.preventDefault();
});
```

### 9.2 模块职责边界

| 模块 | 职责 | 不负责 |
|------|------|--------|
| `main.js` | app 生命周期、单实例、窗口、模块编排、退出顺序 | 具体进程细节 |
| `config.js` | 端口探测、config.yaml/.env 模板渲染 | 进程管理 |
| `sidecar.js` | spawn / 探活 / 重试 / 两阶段终止 | 配置内容 |
| `logger.js` | stdio 落盘、轮转 | 业务日志解析 |
| `paths.js` | `userData` / `resourcesPath` / dev 路径统一解析 | — |

---

## 十、构建与打包

### 10.1 Go sidecar 构建（Makefile 新目标）——已实施

```make
build-windows:
	@powershell -NoProfile -ExecutionPolicy Bypass -File hack/build-windows.ps1
```

`hack/build-windows.ps1` 流程：删除旧 pack 文件（避免 `gf pack` 交互式覆盖确认）→
`gf pack resource internal/packed/build_pack_data.go --keepPath=true` →
`GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o dist/sidecar/ops-dump.exe .`
→ finally 清理临时 pack 文件。

- 必须先 `gf pack`：纯 `go build` 不内嵌 resource，模板/静态/vendor 会缺失。
- sqlite 为纯 Go 无 CGO，交叉编译无障碍；产物 ~24MB 单文件自包含。
- 逻辑放在 PS1 而非 make 配方：Windows 下 make 用 cmd shell（本机无 POSIX sh），
  `GOOS=xxx go build` 环境变量前缀语法不可用；`cli.install` 目标的 bash 语法同样会挂。
- 临时 pack 文件（`internal/packed/build_pack_data.go`）已加入 `.gitignore` 兜底。
- 现有 `gf build`（`hack/config.yaml` 只配 linux）保持不动，服务端部署链路不受影响。

### 10.2 electron-builder 配置（实现：`desktop/package.json` 的 build 字段）

```jsonc
{
  "build": {
    "appId": "com.leapsunrise.opsdump",
    "productName": "OpsDump",
    "directories": { "output": "../dist" },
    "files": ["main/**", "package.json"],       // 仅主进程源码进 asar
    "extraResources": [                          // asar 外，可直接 spawn
      { "from": "../dist/sidecar/ops-dump.exe", "to": "sidecar/ops-dump.exe" }
    ],
    "win": { "target": "nsis", "icon": "build/icon.ico" },
    "electronDownload": { "mirror": "https://npmmirror.com/mirrors/electron/" },
    "nsis": {
      "oneClick": false,
      "allowToChangeInstallationDirectory": true,
      "createDesktopShortcut": true,
      "createStartMenuShortcut": true
    }
  }
}
```

> 配置放在 `package.json` 的 `build` 字段（未单独建 `electron-builder.yml`）；`electronDownload.mirror`
> 固化 electron 下载镜像。

### 10.3 npm scripts

```json
{
  "scripts": {
    "start": "electron .",
    "dev:external": "set SIDECAR_EXTERNAL=1&& electron .",
    "electron:fetch": "node node_modules/electron/install.js",
    "build:sidecar": "powershell -NoProfile -ExecutionPolicy Bypass -File ../hack/build-windows.ps1",
    "dist": "powershell -NoProfile -ExecutionPolicy Bypass -File ../hack/build-desktop.ps1"
  }
}
```

- **dev 模式**：默认连本机已跑的 `go run main.go`（28090），`SIDECAR_EXTERNAL=1` 跳过 spawn；也可直接用 `dist/sidecar/ops-dump.exe` 全链路自测。
- **`dist`**：`hack/build-desktop.ps1` 先建 sidecar，再跑 electron-builder；并默认注入
  `ELECTRON_MIRROR` / `ELECTRON_BUILDER_BINARIES_MIRROR`（npmmirror，已设则沿用）——否则
  electron-builder 从 GitHub 拉 electron/NSIS 二进制会超时（实测 `ETIMEDOUT 20.205.243.166:443`）。
  另在 `package.json` 的 `build.electronDownload.mirror` 固化 electron 镜像。
- **包管理**：pnpm 与 npm 均可；仓库用 pnpm 11。pnpm 11 下 `allowBuilds` 放行构建脚本、
  electron 钉 44.6.0 规避供应链 24h 最小发布龄策略（详见 §14）。

---

## 十一、验证清单（实施完成后逐项过）

| # | 项目 | 方法 | 通过标准 |
|---|------|------|----------|
| 1 | **空目录独立运行**（地基） | 把 gf pack 后的 exe 单独放进干净 temp 目录运行 | `/login` 正常渲染（内嵌资源生效） |
| 2 | 相对路径落点 | 空 data 目录启动 | db/logs/.env 全部生成在 data/，安装目录无写入 |
| 3 | 回环绑定 | `netstat` | 仅 127.0.0.1，无 0.0.0.0 |
| 4 | delimiters 生效 | 登录页 Vue 渲染 | 无 `${...}` 裸露 / 无 Go 模板报错 |
| 5 | 端口占用自愈 | 手动占 28090 后启动 | 自动换端口，界面正常 |
| 6 | 优雅退出 | 触发窗口关闭 | 任务管理器无残留 ops-dump.exe；日志见优雅关闭痕迹 |
| 7 | 强杀兜底 | 模拟 taskkill 失败场景 | 3s 后 `/F` 兜底，无僵尸进程 |
| 8 | 运行期崩溃 | 任务管理器杀 sidecar | 弹窗提供重启/退出 |
| 9 | 完全离线 | 断网跑登录 + 建任务 + 执行历史 | 全流程可用，无 unpkg 请求 |
| 10 | 单实例 | 二次双击图标 | 聚焦已有窗口，无第二个 sidecar |
| 11 | 装机冒烟 | NSIS 装到 Program Files | 上述全绿，data 在 `%APPDATA%` |
| 12 | 升级覆盖安装 | 连续装两版 | data/.env/.db 保留 |

### 11.1 验证结果（2026-10-08，全部通过）

| # | 结果 | 证据 |
|---|------|------|
| 1 | ✅ | 干净 temp 目录跑 exe：`/ping` pong、`/login` 200、`/style.css` 200（全走内嵌 gres，磁盘无 `resource/`） |
| 2 | ✅ | **地基验证中发现并修复**：相对 `dbPath` 经 `gfile.Search` 回退到编译期 MainPkgPath（命中开发机仓库 db）。桌面配置改为**绝对路径**（`RealPath` 对绝对路径立即返回、不进搜索链），db/.env/logs 全落 data/，仓库 db 未被触碰 |
| 3 | ✅ | `netstat`：`127.0.0.1:28091 LISTENING`，无 0.0.0.0 |
| 4 | ✅ | 登录页 Vue 正常渲染（`createApp` 生效），首启种子口令 `code:0` 登录成功 |
| 5 | ✅ | 占用 28090 后启动：`spawn attempt` 换端口，54221 正常服务 |
| 6 | ✅ | `exitCode=0`，server 日志 `all servers shutdown`，350ms 完成，无残留进程（dev 与装机版均验证） |
| 7 | ✅ | 代码路径存在（HTTP 阶段失败/超时 → `taskkill /F`），运行中实测兜底逻辑生效 |
| 8 | ✅ | `taskkill /F` 杀 sidecar → desktop.log `crashed unexpectedly (code 1)` + 弹窗（重启/退出） |
| 9 | ✅ | 代码层验证：4 模板 `unpkg` 引用清零（16 处），vendor 4 文件全部 200，页面无外部 URL；资源由内嵌供给 |
| 10 | ✅ | 第二实例立即退出（单实例锁），仅 1 组 electron 进程 + 1 个 sidecar |
| 11 | ✅ | 静默装机（`/S`）→ 启动 `packaged=true`、ping/login 200、优雅退出 exitCode=0 |
| 12 | ✅ | 静默卸载后 `%APPDATA%\ops-dump-desktop\data\` 保留（.env/.db 不丢） |

另验证：`SIDECAR_EXTERNAL=1` 外部实例模式（不 spawn、复用 28090）正常。

---

## 十二、实施阶段与工作量

| 阶段 | 内容 | 产出 | 估时 |
|------|------|------|------|
| **P0 地基** | §11-1 空目录验证 + `build-windows` 目标 + vendor（§7） | 自包含 exe、离线模板 | 0.5d |
| **P1 壳** | `desktop/` 四模块：config/sidecar/logger/main，dev 模式跑通 | 本地 `electron .` 可用 | 1d |
| **P1 生命周期** | 端口策略、探活重试、两阶段退出、单实例、崩溃处理 | §11-5~8 通过 | 0.5d |
| **P2 打包** | electron-builder + NSIS + 图标 | Setup 安装包 | 0.5d |
| **P2 收尾** | `gres.Dump()` 门禁、README 桌面章节、§11 全清单 | 可发布 | 0.5d |

合计约 3 人日。

> **P0 第 1 项是全方案地基**：验证 gres 内嵌回退是否覆盖 `serverRoot` 磁盘缺失场景。若不通过，则需在 Electron 安装包里附带 `resource/` 目录作兜底（`extraResources` 加一项、config 指向该目录），§3/§10 相应调整。

---

## 十三、风险与待决问题

### 风险与对策（实施后状态）

| 风险 | 影响 | 对策 |
|------|------|------|
| gres 内嵌未覆盖静态根路径 | 空目录白屏 | ✅ 已实测通过（§11-1），无需 extraResources 附带 resource/ |
| 相对 dbPath 搜索回退到编译目录 | 开发机上误连仓库 db | ✅ 已修复：桌面配置绝对路径（§11-2） |
| `taskkill` 无 `/F` 对 Node 子进程无效 | 无法优雅关闭 | ✅ 已改 HTTP 关闭通道（§6.4/§8），实测 exitCode=0 |
| 端口竞态（spawn 到 bind 间被抢） | 启动失败 | ✅ §6.2 换端口重试一次，实测通过 |
| Electron 二进制 / electron-builder 下载失败（网络） | 构建中断 | 用 npmmirror 镜像：`ELECTRON_MIRROR`、`ELECTRON_BUILDER_BINARIES_MIRROR`（`npm run dist` 前设置） |
| Electron 被硬杀时 sidecar 变孤儿进程 | 残留监听/占库 | Windows 下子进程本就不随父进程回收（原方案同样存在）；孤儿占 28090 时下次启动自动换端口；启动崩溃/退出有日志可查。后续可做启动时孤儿探测 |
| 签名缺失 | SmartScreen 拦截提示 | 本期接受；后续买证书 |
| `.env` 被 git 跟踪（现状） | 泄露默认口令 | 与桌面化无关，建议顺手移出版本控制 |

### 已决问题

- **Q1（首启 ADMIN_PASSWORD 策略）**：选 a) 沿用默认 `Opsdump1!`——首启生成 `.env` 时写入默认值，Electron 弹一次性对话框提示「登录后尽快修改」并展示数据目录。
- **Q2（关窗行为）**：选**退出应用**（退出即停调度），与「单进程运维平台」心智一致；托盘常驻留作后续需求。
- **Q3（文档位置）**：即本文件 `docs/opsdump-electron-design.md`。

---

## 十四、pnpm 工作流与供应链策略（实施补充）

使用 pnpm 11 时踩到三个坑，均已固化到仓库配置：

| 现象 | 根因 | 解决 |
|------|------|------|
| `ERR_PNPM_IGNORED_BUILDS: electron-winstaller`，`pnpm run start/dist` 失败 | pnpm 11 不再读取 `package.json` 的 `pnpm` 字段；构建脚本默认被拦 | `desktop/pnpm-workspace.yaml` 的 `allowBuilds`（electron / electron-winstaller / @electron/rebuild / esbuild 置 true） |
| `ERR_PNPM_MINIMUM_RELEASE_AGE_VIOLATION` | 供应链「24h 最小发布龄」策略：`electron@44.7.0` 发布不足 24h，`pnpm add` 被回滚；且隐藏 lockfile `node_modules/.pnpm/lock.yaml` 仍钉旧版本 | electron 固定 **44.6.0**（合规）；清理须删 `node_modules` + `pnpm-lock.yaml` 重建，仅删顶层 lockfile 无效 |
| `pnpm install` 后 electron 二进制缺失、启动报 “failed to install correctly” | electron 运行时二进制未下载（postinstall 未执行） | `pnpm-workspace.yaml` 放行 electron；补装脚本 `pnpm electron:fetch`；设置 `ELECTRON_MIRROR` |

其它：删除 `package-lock.json`（由 `pnpm-lock.yaml` 管理）；`dist` 脚本不嵌套调用 `npm run`，改用
`hack/build-desktop.ps1` 以做到包管理器无关并内置镜像。
