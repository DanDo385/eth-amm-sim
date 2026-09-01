import { chromium } from 'playwright';
import { mkdir, readdir, rename } from 'fs/promises';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);
const OUT = '/Users/openclaw/Code/eth-amm-sim/public/demos/short';
const FRAMES_DIR = '/tmp/eth-amm-sim-short-frames';
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
  console.log('Navigating to dashboard...');
  await page.goto('http://localhost:3000?backend=http://localhost:8080', { waitUntil: 'networkidle' });
  await page.waitForSelector('h1:has-text("Market Dashboard")', { timeout: 30000 });
  
  console.log('Taking initial screenshot of idle state...');
  await captureFrame();
  await sleep(2000);

  console.log('Clicking Start button...');
  const startButton = page.getByRole('button', { name: /start/i });
  await startButton.waitFor({ timeout: 10000 });
  await startButton.click();
  
  console.log('Recording trading session for 60 seconds (2 fps)...');
  for (let i = 0; i < 120; i++) {
    await sleep(500);
    await captureFrame();
    if (i % 10 === 0) {
      console.log(`Captured ${i + 1}/120 frames...`);
    }
  }

  console.log('Converting frames to video...');
  const { stdout, stderr } = await execAsync(
    `ffmpeg -y -framerate 2 -i ${FRAMES_DIR}/frame-%04d.png -c:v libx264 -pix_fmt yuv420p -movflags +faststart ${OUT}/short.mp4`
  );
  
  console.log(`Video saved to: ${OUT}/short.mp4`);
  console.log(`Total frames: ${frameCount}`);
  
} catch (error) {
  console.error('Error during recording:', error.message);
} finally {
  await browser.close();
  
  // Cleanup frames
  const files = await readdir(FRAMES_DIR);
  for (const file of files) {
    await import('fs/promises').then(fs => fs.unlink(`${FRAMES_DIR}/${file}`));
  }
  await import('fs/promises').then(fs => fs.rmdir(FRAMES_DIR));
}
