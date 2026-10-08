const path = require('path');
const { app } = require('electron');

// Writable data directory: everything the sidecar writes relative to its cwd
// (db / logs / .env / config.yaml) lands here. Install dir stays read-only.
function dataDir() {
  return path.join(app.getPath('userData'), 'data');
}

// Electron-side logs (sidecar stdio capture, main-process log).
function logDir() {
  return path.join(app.getPath('userData'), 'logs');
}

// sidecar exe: inside app resources when packaged, repo dist/ in dev.
function sidecarPath() {
  if (app.isPackaged) {
    return path.join(process.resourcesPath, 'sidecar', 'ops-dump.exe');
  }
  return path.join(__dirname, '..', '..', 'dist', 'sidecar', 'ops-dump.exe');
}

module.exports = { dataDir, logDir, sidecarPath };
