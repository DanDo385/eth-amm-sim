import { chromium } from 'playwright';
import { mkdir, readdir, unlink, rmdir } from 'fs/promises';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);
const OUT = '/Users/openclaw/Code/eth-amm-sim/public/demos/detailed';
const FRAMES_DIR = '/tmp/eth-amm-sim-detailed-frames';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

// Ensure output directories exist
await mkdir(OUT, { recursive: true });
await mkdir(FRAMES_DIR, { recursive: true });

console.log('Launching browser...');
const browser = await chromium.launch({ 
  headless: false,
  executablePath: '/Users/openclaw/Library/Caches/ms-playwright/chromium-1223/chrome-mac-x64/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing'
});
const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });

let frameCount = 0;

async function captureFrame() {
  const framePath = `${FRAMES_DIR}/frame-${String(frameCount).padStart(4, '0')}.png`;
  await page.screenshot({ path: framePath });
  frameCount++;
}

try {
  console.log('Phase 1: Initial dashboard view (15s)');
  await page.goto('http://localhost:3000?backend=http://localhost:8080', { waitUntil: 'networkidle' });
  await page.waitForSelector('h1:has-text("Market Dashboard")', { timeout: 30000 });
  
  // Capture frames at 2 fps for 15 seconds
  for (let i = 0; i < 30; i++) {
    await captureFrame();
    await sleep(500);
  }

  console.log('Phase 2: Starting trading session (5s)');
  const startButton = page.getByRole('button', { name: /start/i });
  await startButton.waitFor({ timeout: 10000 });
  await startButton.click();
  
  // Capture frames at 2 fps for 5 seconds
  for (let i = 0; i < 10; i++) {
    await captureFrame();
    await sleep(500);
  }

  console.log('Phase 3: Observing bot trading activity (30s)');
  // Capture frames at 2 fps for 30 seconds
  for (let i = 0; i < 60; i++) {
    await captureFrame();
    await sleep(500);
    if (i % 10 === 0) {
      console.log(`Phase 3: Captured ${i + 1}/60 frames...`);
    }
  }

  console.log('Phase 4: Executing user trade - buying 1000 APPL (15s)');
  // Find the "Buy APPL with ETH" section and target its input
  const buySection = page.locator('label:has-text("Buy APPL with ETH")').locator('..').locator('input[type="number"]');
  await buySection.scrollIntoViewIfNeeded();
  await sleep(1000);
  
  await buySection.fill('1000');
  await sleep(1000);
  
  // Capture frame before clicking buy
  await captureFrame();
  
  // Click the buy button (it's in the same section)
  const buyButton = buySection.locator('..').locator('button:has-text("Buy")');
  await buyButton.click();
  console.log('User trade executed: Buy 1000 APPL');
  
  // Capture frames at 2 fps for 14 seconds after trade
  for (let i = 0; i < 28; i++) {
    await captureFrame();
    await sleep(500);
  }

  console.log('Phase 5: Observing price movement and impact (30s)');
  // Scroll to price chart to show the impact
  const priceChart = page.locator('text=APPL/ETH Price').first();
  await priceChart.scrollIntoViewIfNeeded();
  
  // Capture frames at 2 fps for 30 seconds
  for (let i = 0; i < 60; i++) {
    await captureFrame();
    await sleep(500);
    if (i % 10 === 0) {
      console.log(`Phase 5: Captured ${i + 1}/60 frames...`);
    }
  }

  console.log('Phase 6: Final metrics view (20s)');
  // Scroll to LP stats section
  const lpStats = page.locator('text=LP Value').first();
  await lpStats.scrollIntoViewIfNeeded();
  
  // Capture frames at 2 fps for 20 seconds
  for (let i = 0; i < 40; i++) {
    await captureFrame();
    await sleep(500);
    if (i % 10 === 0) {
      console.log(`Phase 6: Captured ${i + 1}/40 frames...`);
    }
  }

  console.log(`Total frames captured: ${frameCount}`);
  console.log('Converting frames to video...');
  
  const { stdout, stderr } = await execAsync(
    `ffmpeg -y -framerate 2 -i ${FRAMES_DIR}/frame-%04d.png -c:v libx264 -pix_fmt yuv420p -movflags +faststart ${OUT}/detailed.mp4`
  );
  
  console.log(`Video saved to: ${OUT}/detailed.mp4`);
  console.log(`Recording complete. Total duration: ~115 seconds`);
  
} catch (error) {
  console.error('Error during recording:', error.message);
} finally {
  await browser.close();
  
  // Cleanup frames
  try {
    const files = await readdir(FRAMES_DIR);
    for (const file of files) {
      await unlink(`${FRAMES_DIR}/${file}`);
    }
    await rmdir(FRAMES_DIR);
    console.log('Cleanup complete');
  } catch (cleanupError) {
    console.error('Cleanup error:', cleanupError.message);
  }
}
