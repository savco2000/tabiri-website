#!/usr/bin/env node
const { spawn } = require('child_process');
const net = require('net');
const path = require('path');

const rootDir = path.resolve(__dirname, '..');
const preferredPort = Number.parseInt(process.env.PORT || '8000', 10);

function findAvailablePort(startPort, maxAttempts = 10) {
  return new Promise((resolve, reject) => {
    const tryPort = (port, attempt) => {
      const server = net.createServer();
      server.once('error', (error) => {
        if (error.code === 'EADDRINUSE' && attempt < maxAttempts) {
          tryPort(port + 1, attempt + 1);
          return;
        }
        reject(error);
      });
      server.once('listening', () => {
        server.close(() => resolve(port));
      });
      server.listen(port, '0.0.0.0');
    };

    tryPort(startPort, 1);
  });
}

async function main() {
  const port = await findAvailablePort(preferredPort);
  console.log(`Starting static server at http://localhost:${port}`);

  const command = process.platform === 'win32' ? 'npx.cmd' : 'npx';
  const child = spawn(command, ['http-server', '.', '-p', String(port), '-c-1'], {
    cwd: rootDir,
    stdio: 'inherit',
    env: { ...process.env, PORT: String(port) },
  });

  child.on('exit', (code, signal) => {
    process.exit(code ?? (signal ? 1 : 0));
  });
}

main().catch((error) => {
  console.error(error.message);
  process.exit(1);
});
