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

async function getSessionStatus() {
  try {
    const statusElement = await page.locator('text=/Idle|Running|Completed|Stopped/i').first();
    const status = await statusElement.textContent();
    return status?.trim() || 'Unknown';
  } catch {
    return 'Unknown';
  }
}

try {
  console.log('Phase 1: Navigate to dashboard (5s)');
  await page.goto('http://localhost:3000?backend=http://localhost:8080', { waitUntil: 'networkidle' });
  await page.waitForSelector('h1:has-text("Market Dashboard")', { timeout: 30000 });
  
  // Capture initial state
  for (let i = 0; i < 10; i++) {
    await captureFrame();
    await sleep(500);
  }

  console.log('Phase 2: Stop current session if running (10s)');
  const initialStatus = await getSessionStatus();
  console.log(`Initial session status: ${initialStatus}`);
  
  if (initialStatus === 'Running') {
    const stopButton = page.getByRole('button', { name: /stop/i });
    await stopButton.waitFor({ timeout: 5000 });
    await stopButton.click();
    console.log('Clicked Stop button');
    await sleep(2000);
    
    // Capture stop action
    for (let i = 0; i < 6; i++) {
      await captureFrame();
      await sleep(500);
    }
  } else {
    // Capture current state
    for (let i = 0; i < 6; i++) {
      await captureFrame();
      await sleep(500);
    }
  }

  console.log('Phase 3: Reseed session (10s)');
  const reseedButton = page.getByRole('button', { name: /reseed/i });
  await reseedButton.waitFor({ timeout: 5000 });
  await reseedButton.click();
  console.log('Clicked Reseed button');
  await sleep(3000);
  
  // Capture reseed action
  for (let i = 0; i < 14; i++) {
    await captureFrame();
    await sleep(500);
  }

  console.log('Phase 4: Start new session (5s)');
  const startButton = page.getByRole('button', { name: /start/i });
  await startButton.waitFor({ timeout: 10000 });
  await startButton.click();
  console.log('Clicked Start button');
  
  // Capture session start
  for (let i = 0; i < 10; i++) {
    await captureFrame();
    await sleep(500);
  }

  console.log('Phase 5: Observe bot trading (20s)');
  for (let i = 0; i < 40; i++) {
    await captureFrame();
    await sleep(500);
    if (i % 10 === 0) {
      console.log(`Phase 5: Captured ${i + 1}/40 frames...`);
    }
  }

  console.log('Phase 6: Execute user trade - buy 500 APPL (15s)');
  // Scroll to User Trading section
  const userTradingSection = page.locator('text=User Trading').first();
  await userTradingSection.scrollIntoViewIfNeeded();
  await sleep(1000);
  
  // Find the "Buy APPL with ETH" input
  const buyInput = page.locator('label:has-text("Buy APPL with ETH")').locator('..').locator('input[type="number"]');
  await buyInput.scrollIntoViewIfNeeded();
  await sleep(500);
  
  // Clear and fill with 500
  await buyInput.fill('500');
  console.log('Filled buy input with 500');
  await sleep(1000);
  
  // Capture before clicking buy
  await captureFrame();
  
  // Click the Buy button
  const buyButton = page.locator('label:has-text("Buy APPL with ETH")').locator('..').locator('button:has-text("Buy")');
  await buyButton.click();
  console.log('Clicked Buy button for 500 APPL');
  
  // Capture the trade execution
  for (let i = 0; i < 20; i++) {
    await captureFrame();
    await sleep(500);
  }

  console.log('Phase 7: Observe price impact (20s)');
  const priceChart = page.locator('text=APPL/ETH Price').first();
  await priceChart.scrollIntoViewIfNeeded();
  
  for (let i = 0; i < 40; i++) {
    await captureFrame();
    await sleep(500);
    if (i % 10 === 0) {
      console.log(`Phase 7: Captured ${i + 1}/40 frames...`);
    }
  }

  console.log('Phase 8: Show final metrics (15s)');
  const lpStats = page.locator('text=LP Value').first();
  await lpStats.scrollIntoViewIfNeeded();
  
  for (let i = 0; i < 30; i++) {
    await captureFrame();
    await sleep(500);
    if (i % 10 === 0) {
      console.log(`Phase 8: Captured ${i + 1}/30 frames...`);
    }
  }

  console.log(`Total frames captured: ${frameCount}`);
  console.log('Converting frames to video...');
  
  const { stdout, stderr } = await execAsync(
    `ffmpeg -y -framerate 2 -i ${FRAMES_DIR}/frame-%04d.png -c:v libx264 -pix_fmt yuv420p -movflags +faststart ${OUT}/detailed.mp4`
  );
  
  console.log(`Video saved to: ${OUT}/detailed.mp4`);
  console.log(`Recording complete. Total duration: ~100 seconds`);
  
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
