const fs = require('fs');
const net = require('net');
const path = require('path');

const PREFERRED_PORT = 28090;

// Find a free port: try 28090 first, fall back to an ephemeral one.
// Port is decided per launch (no persistence) so a stale port can never
// brick startup.
function pickPort(preferred = PREFERRED_PORT) {
  return tryPort(preferred).catch(() => tryPort(0));
}

function tryPort(port) {
  return new Promise((resolve, reject) => {
    const srv = net.createServer();
    srv.once('error', reject);
    srv.listen(port, '127.0.0.1', () => {
      const chosen = srv.address().port;
      srv.close(() => resolve(chosen));
    });
  });
}

// Desktop config template. Differences vs manifest/config/config.yaml:
//  - address pinned to loopback with the discovered port
//  - dbPath / envFile are ABSOLUTE paths. Critical: with a relative dbPath,
//    gf's gfile.Search falls back to the compile-time package dir when the
//    file is missing from cwd, silently opening the developer's repo db.
//  - YAML single quotes (backslashes in double-quoted YAML are escapes)
function renderConfig({ port, dbPath, envFile }) {
  const abs = (p) => `'${p.replace(/\\/g, '/')}'`;
  return `server:
  address: "127.0.0.1:${port}"
  openapiPath: "/api.json"
  swaggerPath: "/swagger"
  serverRoot: "resource/public"
  dumpRouterMap: true
  logger:
    path: "./logs/server"
    file: "{Y-m-d}.log"
    stdout: true
    rotateSize: "100M"
    rotateBackupLimit: 10
    rotateBackupExpire: "30d"
    rotateBackupCompress: 9
    rotateCheckInterval: "24h"

ops-dump:
  timezone: "Asia/Shanghai"
  dbPath: ${abs(dbPath)}
  credentials:
    envFile: ${abs(envFile)}
  admin:
    username: admin

logger:
  path: ./logs
  file: ops-dump.log
  level: info
  stdout: true

viewer:
  delimiters:
    - "\${"
    - "}"
`;
}

// First-run .env. Never overwritten afterwards (user secrets live there).
// Returns true when the file was created now (i.e. this is first run).
function ensureEnv(dataDir, envTemplate) {
  const envPath = path.join(dataDir, '.env');
  if (fs.existsSync(envPath)) return false;
  fs.writeFileSync(envPath, envTemplate, 'utf8');
  return true;
}

function writeConfig({ dataDir, port }) {
  const dbPath = path.join(dataDir, 'ops-dump.db');
  const envFile = path.join(dataDir, '.env');
  const cfgPath = path.join(dataDir, 'config.yaml');
  fs.writeFileSync(cfgPath, renderConfig({ port, dbPath, envFile }), 'utf8');
  return cfgPath;
}

module.exports = { PREFERRED_PORT, pickPort, ensureEnv, writeConfig, renderConfig };
