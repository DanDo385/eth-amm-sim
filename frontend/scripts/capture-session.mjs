import { chromium } from 'playwright-core';

const OUT = '/Users/openclaw/Code/eth-amm-sim/public/screenshots';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const browser = await chromium.launch({ channel: 'chrome', headless: true });
const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });

await page.goto('http://localhost:3000', { waitUntil: 'networkidle' });
await page.waitForSelector('h1:has-text("Market Dashboard")', { timeout: 30000 });
// let charts / data render
await sleep(3000);

async function cardFor(text) {
  const loc = page.getByText(text, { exact: false }).first();
  await loc.waitFor({ timeout: 10000 });
  const card = loc.locator('xpath=ancestor::div[contains(@class,"rounded")][1]');
  return (await card.count()) ? card : loc;
}

const steps = [
  { file: '01-dashboard-active.png', shot: async () => page.screenshot({ path: `${OUT}/01-dashboard-active.png`, fullPage: true }) },
  { file: '02-bots-trading.png', shot: async () => (await cardFor('Trade Blotter')).screenshot({ path: `${OUT}/02-bots-trading.png` }) },
  { file: '03-price-movement.png', shot: async () => (await cardFor('Price')).screenshot({ path: `${OUT}/03-price-movement.png` }) },
  { file: '04-whale-trade.png', shot: async () => (await cardFor('Key Events')).screenshot({ path: `${OUT}/04-whale-trade.png` }) },
  { file: '05-metrics-evolving.png', shot: async () => (await cardFor('LP')).screenshot({ path: `${OUT}/05-metrics-evolving.png` }) },
  { file: '06-session-progress.png', shot: async () => (await cardFor('Elapsed')).screenshot({ path: `${OUT}/06-session-progress.png` }) },
];

for (let i = 0; i < steps.length; i++) {
  if (i > 0) await sleep(10000);
  try {
    await steps[i].shot();
    console.log(`captured ${steps[i].file}`);
  } catch (e) {
    console.error(`FAILED ${steps[i].file}: ${e.message}`);
    // fallback to viewport screenshot
    await page.screenshot({ path: `${OUT}/${steps[i].file}` });
    console.log(`fallback viewport ${steps[i].file}`);
  }
}

await browser.close();
console.log('done');
