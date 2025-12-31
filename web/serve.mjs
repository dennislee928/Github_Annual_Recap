import http from "node:http";
import fs from "node:fs";
import path from "node:path";
import url from "node:url";

const __dirname = path.dirname(url.fileURLToPath(import.meta.url));
const root = __dirname;
const port = process.env.PORT ? Number(process.env.PORT) : 4173;

const mime = {
  ".html": "text/html; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".svg": "image/svg+xml",
  ".png": "image/png"
};

function safeJoin(rootDir, reqPath) {
  const p = path.normalize(path.join(rootDir, reqPath));
  if (!p.startsWith(rootDir)) return null;
  return p;
}

const server = http.createServer((req, res) => {
  const u = new URL(req.url, `http://127.0.0.1:${port}`);
  const pathname = u.pathname === "/" ? "/report.html" : u.pathname;
  const filePath = safeJoin(root, pathname);
  if (!filePath) {
    res.writeHead(403); res.end("Forbidden"); return;
  }

  fs.stat(filePath, (err, st) => {
    if (err || !st.isFile()) {
      res.writeHead(404); res.end("Not found"); return;
    }
    const ext = path.extname(filePath).toLowerCase();
    res.writeHead(200, { "Content-Type": mime[ext] || "application/octet-stream" });
    fs.createReadStream(filePath).pipe(res);
  });
});

server.listen(port, "127.0.0.1", () => {
  console.log(`Static server running at http://127.0.0.1:${port}`);
});
