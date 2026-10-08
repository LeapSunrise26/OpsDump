const { spawn, execFile } = require('child_process');
const http = require('http');

const READY_TIMEOUT_MS = 15000;
const PING_INTERVAL_MS = 200;
// GoFrame's graceful shutdown budget is 5s; wait a bit beyond it.
const GRACEFUL_WAIT_MS = 6000;
const FORCE_WAIT_MS = 2000;

function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms));
}

// GET /ping on the backend; resolves with body when healthy.
function ping(port) {
  return new Promise((resolve, reject) => {
    const req = http.get(
      { host: '127.0.0.1', port, path: '/ping', timeout: 1000 },
      (res) => {
        let body = '';
        res.on('data', (c) => (body += c));
        res.on('end', () => resolve(body));
      }
    );
    req.on('timeout', () => req.destroy(new Error('ping timeout')));
    req.on('error', reject);
  });
}

// Poll /ping until pong, child exits, or timeout.
function waitForReady(child, port, timeoutMs = READY_TIMEOUT_MS) {
  return new Promise((resolve, reject) => {
    let done = false;
    const finish = (fn, arg) => {
      if (done) return;
      done = true;
      clearInterval(timer);
      clearTimeout(timer2);
      child.removeListener('exit', onExit);
      fn(arg);
    };
    const onExit = (code) =>
      finish(reject, new Error(`sidecar exited early (code ${code})`));
    const timer = setInterval(() => {
      ping(port)
        .then((body) => {
          if (body.includes('pong')) finish(resolve, port);
        })
        .catch(() => {});
    }, PING_INTERVAL_MS);
    const timer2 = setTimeout(
      () => finish(reject, new Error(`sidecar not ready within ${timeoutMs}ms`)),
      timeoutMs
    );
    child.once('exit', onExit);
  });
}

// taskkill without /F cannot deliver a control event to a Node-spawned
// child (it owns no console taskkill can close — verified empirically),
// so the graceful phase goes through the backend's HTTP shutdown endpoint
// instead. /F (TerminateProcess) is the last-resort fallback.
function taskkill(pid) {
  return new Promise((resolve) => {
    execFile('taskkill', ['/PID', String(pid), '/F'], () => resolve());
  });
}

// POST /api/system/shutdown with the control header the backend requires
// (blocks browser CSRF: custom headers force a CORS preflight, which the
// server never answers).
function httpShutdown(port) {
  return new Promise((resolve, reject) => {
    const req = http.request(
      {
        host: '127.0.0.1',
        port,
        path: '/api/system/shutdown',
        method: 'POST',
        headers: { 'X-OpsDump-Control': '1' },
        timeout: 2000,
      },
      (res) => {
        res.resume();
        res.on('end', resolve);
      }
    );
    req.on('timeout', () => req.destroy(new Error('shutdown request timeout')));
    req.on('error', reject);
    req.end();
  });
}

// Two-phase stop: HTTP graceful first, forced kill fallback. Resolves when
// the child is really gone (or after best-effort).
async function stop(child, port) {
  if (!child || child.exitCode !== null) return;
  const pid = child.pid;
  const exited = new Promise((r) => child.once('exit', r));

  if (port) {
    try {
      await httpShutdown(port);
    } catch (_) {
      /* backend already down; fall through to kill */
    }
  }
  await Promise.race([exited, sleep(GRACEFUL_WAIT_MS)]);

  if (child.exitCode === null) {
    await taskkill(pid);
    await Promise.race([exited, sleep(FORCE_WAIT_MS)]);
  }
}

// Spawn the sidecar with cwd = data dir; all relative paths in the desktop
// config (.env / logs) resolve there. stdio piped to onOutput callbacks.
// (No detached/console tricks: graceful stop goes over HTTP, see stop().)
function spawnSidecar({ exePath, cwd, onOutput }) {
  const child = spawn(exePath, [], {
    cwd,
    windowsHide: true,
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  child.stdout.on('data', (d) => onOutput(d));
  child.stderr.on('data', (d) => onOutput(d));
  return child;
}

module.exports = { ping, waitForReady, stop, spawnSidecar };
