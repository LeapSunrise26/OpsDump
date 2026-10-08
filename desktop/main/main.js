const { app, BrowserWindow, Menu, dialog, shell } = require('electron');
const fs = require('fs');
const path = require('path');

const paths = require('./paths');
const config = require('./config');
const { FileLogger } = require('./logger');
const sidecar = require('./sidecar');

const EXTERNAL = process.env.SIDECAR_EXTERNAL === '1';
const EXTERNAL_PORT = Number(process.env.SIDECAR_PORT || 28090);

const ENV_TEMPLATE = `# OpsDump desktop: generated on first launch. Sensitive values live only
# here (never in code or logs). File permissions should stay restrictive.
ADMIN_PASSWORD=Opsdump1!
BACKUP_DIR=

# MySQL
MYSQL_ROOT_PASSWORD=
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_DATABASE=

# TDengine
TDENGINE_ROOT_PASSWORD=
TDENGINE_HOST=127.0.0.1
TDENGINE_PORT=6030
TDENGINE_DATABASE=

# MinIO
MINIO_ROOT_USER=
MINIO_ROOT_PASSWORD=
MINIO_ENDPOINT=127.0.0.1:9000
MINIO_BUCKET=
`;

let win = null;
let child = null;
let logger = null;
let backendPort = null;
let stopping = false;
let quitting = false;

function loadingPage(text) {
  const html = `<!doctype html><meta charset="utf-8"><style>
    body{margin:0;height:100vh;display:flex;align-items:center;justify-content:center;
      font:14px/1.6 "Segoe UI",system-ui,sans-serif;background:#f5f7fa;color:#303133}
    .box{text-align:center}
    .spin{width:34px;height:34px;margin:0 auto 14px;border:3px solid #dcdfe6;
      border-top-color:#409eff;border-radius:50%;animation:sp .8s linear infinite}
    @keyframes sp{to{transform:rotate(360deg)}}
    @media(prefers-reduced-motion:reduce){.spin{animation:none}}
  </style><div class="box"><div class="spin"></div><div>${text}</div></div>`;
  return `data:text/html;charset=utf-8,${encodeURIComponent(html)}`;
}

function errorPage(msg, logPath) {
  const esc = (s) => String(s).replace(/[&<>]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' }[c]));
  const html = `<!doctype html><meta charset="utf-8"><style>
    body{margin:0;height:100vh;display:flex;align-items:center;justify-content:center;
      font:14px/1.7 "Segoe UI",system-ui,sans-serif;background:#f5f7fa;color:#303133}
    .card{max-width:560px;background:#fff;border:1px solid #e4e7ed;border-radius:8px;
      padding:24px 28px;box-shadow:0 2px 12px rgba(0,0,0,.06)}
    h1{font-size:16px;margin:0 0 10px;color:#f56c6c}
    code{background:#f0f2f5;padding:2px 6px;border-radius:4px;font-size:12px;word-break:break-all}
  </style><div class="card"><h1>后端启动失败</h1><div>${esc(msg)}</div>
    <p>日志：<code>${esc(logPath)}</code></p></div>`;
  return `data:text/html;charset=utf-8,${encodeURIComponent(html)}`;
}

function createWindow() {
  win = new BrowserWindow({
    width: 1280,
    height: 840,
    minWidth: 960,
    minHeight: 600,
    show: false,
    webPreferences: {
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
    },
  });

  win.once('ready-to-show', () => win.show());

  // No application menu bar (removeMenu also stops Alt from revealing it).
  win.removeMenu();

  // Keep DevTools reachable in development (the removed menu used to own
  // the F12 accelerator); packaged builds expose nothing.
  if (!app.isPackaged) {
    win.webContents.on('before-input-event', (e, input) => {
      if (input.type === 'keyDown' && input.key === 'F12') {
        win.webContents.toggleDevTools();
        e.preventDefault();
      }
    });
  }

  // Only the local backend may open windows / navigate; everything else
  // goes to the system browser or is denied.
  win.webContents.setWindowOpenHandler(({ url }) => {
    try {
      if (new URL(url).hostname === '127.0.0.1') return { action: 'allow' };
    } catch (_) {}
    shell.openExternal(url);
    return { action: 'deny' };
  });
  win.webContents.on('will-navigate', (e, url) => {
    try {
      if (new URL(url).hostname !== '127.0.0.1') e.preventDefault();
    } catch (_) {
      e.preventDefault();
    }
  });

  win.loadURL(loadingPage('正在启动 OpsDump 后端…'));
}

// Resolve the backend: either attach to an externally running instance
// (dev with `go run`) or spawn our own sidecar, retrying once with a new
// port if it dies during startup (port stolen between pick and bind).
async function bootBackend() {
  if (EXTERNAL) {
    await sidecar.ping(EXTERNAL_PORT).catch(() => {
      throw new Error(
        `外部实例不可达 (127.0.0.1:${EXTERNAL_PORT})，请先运行 go run main.go 或去掉 SIDECAR_EXTERNAL`
      );
    });
    return EXTERNAL_PORT;
  }

  const exe = paths.sidecarPath();
  if (!fs.existsSync(exe)) {
    throw new Error(`未找到 sidecar: ${exe}（先执行 npm run build:sidecar）`);
  }

  const dataDir = paths.dataDir();
  fs.mkdirSync(dataDir, { recursive: true });

  const isFirstRun = config.ensureEnv(dataDir, ENV_TEMPLATE);
  if (isFirstRun) {
    setTimeout(
      () =>
        dialog.showMessageBox(win, {
          type: 'info',
          title: 'OpsDump 首次启动',
          message: '已生成初始配置',
          detail:
            '管理员账号：admin\n默认口令：Opsdump1!\n\n请登录后尽快在设置中修改口令，并按需配置 .env 中的备份参数。\n\n数据目录：' +
            dataDir,
          buttons: ['知道了'],
        }),
      1500
    );
  }

  let lastErr = null;
  for (let attempt = 0; attempt < 2; attempt++) {
    const port = await config.pickPort();
    config.writeConfig({ dataDir, port });
    logger.log(`spawn attempt=${attempt + 1} port=${port} exe=${exe}`);

    child = sidecar.spawnSidecar({
      exePath: exe,
      cwd: dataDir,
      onOutput: (d) => logger.write(d),
    });

    try {
      const readyPort = await sidecar.waitForReady(child, port);
      logger.log(`sidecar ready on ${readyPort}`);
      watchChild();
      return readyPort;
    } catch (err) {
      lastErr = err;
      logger.log(`startup failed: ${err.message}`);
      await sidecar.stop(child, port);
      child = null;
    }
  }
  throw lastErr || new Error('sidecar 启动失败');
}

// Runtime crash detection (not fired during our own two-phase stop).
function watchChild() {
  child.once('exit', (code) => {
    if (stopping || quitting) return;
    logger.log(`sidecar crashed unexpectedly (code ${code})`);
    dialog
      .showMessageBox(win, {
        type: 'error',
        title: 'OpsDump',
        message: '后端进程意外退出',
        detail: `退出码：${code}\n\n日志：${logger.filePath}`,
        buttons: ['重启', '退出'],
        defaultId: 0,
        cancelId: 1,
      })
      .then(({ response }) => {
        if (response === 0) restartBackend();
        else app.quit();
      });
  });
}

async function restartBackend() {
  try {
    win.loadURL(loadingPage('正在重启后端…'));
    backendPort = await bootBackend();
    win.loadURL(`http://127.0.0.1:${backendPort}/`);
  } catch (err) {
    win.loadURL(errorPage(err.message, logger.filePath));
  }
}

// Quit path: block default quit until the sidecar is really gone.
function setupQuitHook() {
  app.on('before-quit', (e) => {
    if (quitting) return;
    e.preventDefault();
    quitting = true;
    (async () => {
      if (child) {
        stopping = true;
        logger.log(`stopping sidecar pid=${child.pid} (http graceful, 6s budget)`);
        await sidecar.stop(child, backendPort);
        logger.log(`sidecar stopped, exitCode=${child.exitCode}`);
        child = null;
      }
      if (logger) logger.close();
      app.exit(0);
    })();
  });
}

const gotLock = app.requestSingleInstanceLock();
if (!gotLock) {
  app.quit();
} else {
  app.on('second-instance', () => {
    if (win) {
      if (win.isMinimized()) win.restore();
      win.focus();
    }
  });

  app.whenReady().then(async () => {
    logger = new FileLogger(path.join(paths.logDir(), 'desktop.log'));
    logger.log(`start packaged=${app.isPackaged} external=${EXTERNAL}`);
    createWindow();
    setupQuitHook();

    try {
      backendPort = await bootBackend();
      win.loadURL(`http://127.0.0.1:${backendPort}/`);
    } catch (err) {
      logger.log(`boot failed: ${err.message}`);
      win.loadURL(errorPage(err.message, logger.filePath));
    }
  });

  app.on('window-all-closed', () => {
    app.quit();
  });
}
