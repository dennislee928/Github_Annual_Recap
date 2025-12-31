import fs from "node:fs";
import path from "node:path";
import url from "node:url";
import { chromium } from "playwright";
import { spawn } from "node:child_process";

const __dirname = path.dirname(url.fileURLToPath(import.meta.url));
const outDir = path.join(__dirname, "out");
const instagramDir = path.join(outDir, "instagram");

fs.mkdirSync(instagramDir, { recursive: true });

// start local static server
const port = 4174;
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
  const page = await browser.newPage({ viewport: { width: 1080, height: 1920 } });

  const reportUrl = `http://127.0.0.1:${port}/report-story.html?data=recap_2025.json`;
  await page.goto(reportUrl, { waitUntil: "networkidle" });

  // Wait for cards to be attached to DOM (not necessarily visible)
  await page.waitForSelector(".story-card", { state: "attached", timeout: 15000 });
  
  // Wait for JavaScript to populate data - check if values are filled
  await page.waitForFunction(() => {
    const commits = document.getElementById("t_commits");
    return commits && commits.textContent !== "—" && commits.textContent !== "";
  }, { timeout: 10000 }).catch(() => {
    console.warn("Data might not be fully loaded, continuing anyway...");
  });
  
  await page.waitForTimeout(1000);

  // Get all story cards
  const cardCount = await page.evaluate(() => {
    return document.querySelectorAll(".story-card").length;
  });
  console.log(`Found ${cardCount} story cards`);

  // Show each card one by one and screenshot
  for (let i = 0; i < cardCount; i++) {
    // Hide all cards and show only the current one
    await page.evaluate((index) => {
      const cards = document.querySelectorAll(".story-card");
      cards.forEach((card, idx) => {
        if (idx === index) {
          card.style.display = "flex";
          card.style.visibility = "visible";
        } else {
          card.style.display = "none";
          card.style.visibility = "hidden";
        }
      });
    }, i);

    // Wait for the card to be visible and rendered
    const cardId = `#card-${String(i + 1).padStart(2, "0")}`;
    try {
      await page.waitForSelector(`${cardId}.story-card`, { state: "visible", timeout: 5000 });
    } catch (e) {
      console.warn(`Card ${i + 1} might not be visible, continuing anyway...`);
    }
    
    // Wait a bit more for any animations or rendering
    await page.waitForTimeout(1500);

    // Screenshot the entire viewport (which contains the single visible card)
    const n = String(i + 1).padStart(2, "0");
    const p = path.join(instagramDir, `story-${n}.png`);
    await page.screenshot({ path: p, fullPage: false });
    console.log(`Wrote ${p}`);
  }

  await browser.close();
} finally {
  server.kill("SIGTERM");
}
