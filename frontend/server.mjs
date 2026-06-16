import { createReadStream, existsSync, statSync } from 'node:fs';
import { createServer } from 'node:http';
import { extname, join, normalize } from 'node:path';

const port = Number(process.env.PORT || 4321);
const distDir = join(process.cwd(), 'dist');

const contentTypes = {
  '.css': 'text/css; charset=utf-8',
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.png': 'image/png',
  '.svg': 'image/svg+xml',
  '.ttf': 'font/ttf',
  '.webp': 'image/webp',
  '.ico': 'image/x-icon'
};

function resolveFile(requestUrl) {
  const url = new URL(requestUrl, `http://localhost:${port}`);
  const safePath = normalize(decodeURIComponent(url.pathname)).replace(/^(\.\.[/\\])+/, '');
  let filePath = join(distDir, safePath);

  if (existsSync(filePath) && statSync(filePath).isDirectory()) {
    filePath = join(filePath, 'index.html');
  }

  if (!existsSync(filePath)) {
    filePath = join(distDir, safePath, 'index.html');
  }

  if (!existsSync(filePath)) {
    filePath = join(distDir, 'index.html');
  }

  return filePath;
}

createServer((request, response) => {
  const filePath = resolveFile(request.url || '/');
  const extension = extname(filePath);
  response.setHeader('Content-Type', contentTypes[extension] || 'application/octet-stream');
  createReadStream(filePath)
    .on('error', () => {
      response.statusCode = 404;
      response.end('No encontrado');
    })
    .pipe(response);
}).listen(port, '0.0.0.0', () => {
  console.log(`MendoCultura frontend disponible en http://localhost:${port}`);
});
