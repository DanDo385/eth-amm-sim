import { chromium } from 'playwright-core';

const OUT = '/Users/openclaw/Code/eth-amm-sim/public/screenshots/03-price-movement.png';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const browser = await chromium.launch({ channel: 'chrome', headless: true });
const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });

await page.goto('http://localhost:3000', { waitUntil: 'networkidle' });
await page.waitForSelector('h1:has-text("Market Dashboard")', { timeout: 30000 });
await sleep(3000);

const heading = page.getByText('APPL/ETH Price', { exact: true }).first();
await heading.waitFor({ timeout: 10000 });

const card = heading.locator('xpath=ancestor::div[contains(@class,"rounded-lg") and contains(@class,"border")][1]');

try {
  await card.waitFor({ timeout: 5000 });
  await card.screenshot({ path: OUT });
  console.log('captured via card locator');
} catch (e) {
  console.error(`card locator failed: ${e.message}, falling back to full page`);
  await page.screenshot({ path: OUT, fullPage: true });
  console.log('fallback full-page captured');
}

await browser.close();
console.log('done');
