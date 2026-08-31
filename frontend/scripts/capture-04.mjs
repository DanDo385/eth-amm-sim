import { chromium } from 'playwright-core';
const OUT = '/Users/openclaw/Code/eth-amm-sim/public/screenshots';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const browser = await chromium.launch({ channel: 'chrome', headless: true });
try {
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
  await page.goto('http://localhost:3000', { waitUntil: 'networkidle' });
  await page.waitForSelector('h1:has-text("Market Dashboard")', { timeout: 30000 });
  await sleep(4000);
  const loc = page.getByText('Key Events', { exact: false }).first();
  await loc.waitFor({ timeout: 15000 });
  await loc.scrollIntoViewIfNeeded();
  await page.mouse.wheel(0, 200);
  await sleep(1500);
  await page.screenshot({ path: `${OUT}/04-whale-trade.png` });
  console.log('captured 04-whale-trade.png');
} finally {
  await browser.close();
}
