import { chromium } from 'playwright-core';

const OUT = '/Users/openclaw/Code/eth-amm-sim/public/screenshots';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const browser = await chromium.launch({ channel: 'chrome', headless: true });
const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });

await page.goto('http://localhost:3000', { waitUntil: 'networkidle' });
await page.waitForSelector('h1:has-text("Market Dashboard")', { timeout: 30000 });
await sleep(4000);

async function shotAround(anchor, file, { exact = false } = {}) {
  const loc = page.getByText(anchor, { exact }).first();
  await loc.waitFor({ timeout: 15000 });
  await loc.scrollIntoViewIfNeeded();
  await sleep(1200);
  const path = `${OUT}/${file}`;
  await page.screenshot({ path });
  console.log(`captured ${file}`);
}

// 1. Whale bot activity - Trade Blotter (whale rows highlighted) + Key Events
await shotAround('Trade Blotter', '04-whale-trade.png');

// 2. LP metrics panel - impermanent loss, fees earned, net PnL
await shotAround('Fees Earned', '05-metrics-evolving.png');

// 3. Session timer / elapsed / trade count
await shotAround('Elapsed', '06-session-progress.png');

await browser.close();
console.log('done');
