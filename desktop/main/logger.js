const fs = require('fs');
const path = require('path');

const MAX_BYTES = 10 * 1024 * 1024;

// Append-only file logger with a single-generation size rotation.
// Used for sidecar stdout/stderr and main-process events.
class FileLogger {
  constructor(filePath) {
    this.filePath = filePath;
    fs.mkdirSync(path.dirname(filePath), { recursive: true });
    this.stream = fs.createWriteStream(filePath, { flags: 'a' });
  }

  write(chunk) {
    if (!this.stream) return;
    const line = Buffer.isBuffer(chunk) ? chunk : Buffer.from(String(chunk));
    this.stream.write(line);
    this.rotateIfNeeded();
  }

  log(msg) {
    this.write(`[${new Date().toISOString()}] ${msg}\n`);
  }

  rotateIfNeeded() {
    try {
      const { size } = fs.statSync(this.filePath);
      if (size < MAX_BYTES) return;
      this.stream.end();
      fs.renameSync(this.filePath, `${this.filePath}.1`);
      this.stream = fs.createWriteStream(this.filePath, { flags: 'a' });
    } catch (_) {
      /* rotation is best-effort */
    }
  }

  close() {
    if (this.stream) {
      this.stream.end();
      this.stream = null;
    }
  }
}

module.exports = { FileLogger };
