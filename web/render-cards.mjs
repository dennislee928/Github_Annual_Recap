import fs from "node:fs";
import path from "node:path";
import url from "node:url";
import { chromium } from "playwright";
import { spawn } from "node:child_process";

const __dirname = path.dirname(url.fileURLToPath(import.meta.url));
const outDir = path.join(__dirname, "out");
const cardDir = path.join(outDir, "cards");

fs.mkdirSync(cardDir, { recursive: true });

// start local static server
const port = 4173;
const server = spawn(process.execPath, [path.join(__dirname, "serve.mjs")], {
  env: { ...process.env, PORT: String(port) },
  stdio: ["ignore", "pipe", "pipe"]
});

const waitForServer = async () => {
  return new Promise((resolve, reject) => {
    let ok = false;
    const onData = (d) => {
      const s = d.toString();
      if (s.includes("Static server running")) {
        ok = true;
        cleanup();
        resolve(true);
      }
    };
    const onErr = (d) => {
      // keep stderr for debugging
    };
    const cleanup = () => {
      server.stdout.off("data", onData);
      server.stderr.off("data", onErr);
      clearTimeout(t);
    };
    const t = setTimeout(() => {
      cleanup();
      if (!ok) reject(new Error("Server did not start in time"));
    }, 10000);
    server.stdout.on("data", onData);
    server.stderr.on("data", onErr);
  });
};

try {
  await waitForServer();

  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport: { width: 1080, height: 1080 } });

  const reportUrl = `http://127.0.0.1:${port}/report.html?data=recap_2025.json`;
  await page.goto(reportUrl, { waitUntil: "networkidle" });

  // Ensure JS rendered
  await page.waitForSelector(".card", { timeout: 15000 });

  // Screenshot each card
  const cards = await page.$$(".card");
  for (let i = 0; i < cards.length; i++) {
    const n = String(i + 1).padStart(2, "0");
    const p = path.join(cardDir, `card-${n}.png`);
    await cards[i].screenshot({ path: p });
    console.log(`Wrote ${p}`);
  }

  // Full report HTML copy (for sharing)
  const reportHtml = await page.content();
  fs.writeFileSync(path.join(outDir, "report.html"), reportHtml, "utf-8");

  await browser.close();
} finally {
  server.kill("SIGTERM");
}
