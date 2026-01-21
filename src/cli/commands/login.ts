import { Command } from 'commander';
import { chromium } from 'playwright';
import { join } from 'path';
import { getAutoplyDir } from '../../db';
import { configRepository } from '../../db/repositories/config';

const STORAGE_STATE_PATH = join(getAutoplyDir(), 'browser-state.json');

export const loginCommand = new Command('login')
  .description('Login to job platforms and save browser session')
  .argument('[platform]', 'Platform to login to (linkedin, etc.)', 'linkedin')
  .action(async (platform: string) => {
    const urls: Record<string, string> = {
      linkedin: 'https://www.linkedin.com/login',
      greenhouse: 'https://www.greenhouse.io',
      lever: 'https://www.lever.co',
    };

    const loginUrl = urls[platform];
    if (!loginUrl) {
      console.error(`Unknown platform: ${platform}`);
      console.log(`Supported platforms: ${Object.keys(urls).join(', ')}`);
      process.exit(1);
    }

    console.log(`Opening ${platform} login page...`);
    console.log('Please login manually in the browser window.');
    console.log('The browser will close automatically after you login.\n');

    const browser = await chromium.launch({
      headless: false,
      args: [
        '--disable-blink-features=AutomationControlled',
        '--disable-features=IsolateOrigins,site-per-process',
      ],
    });
    const context = await browser.newContext({
      userAgent:
        'Mozilla/5.0 (Macintosh; Apple Silicon Mac OS X 14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
      viewport: { width: 1920, height: 1080 },
      locale: 'en-NG',
      timezoneId: 'Africa/Lagos',
    });

    // Remove webdriver detection flag
    await context.addInitScript(() => {
      Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
    });

    const page = await context.newPage();

    await page.goto(loginUrl);

    // Wait for user to login - detect by URL change or specific elements
    if (platform === 'linkedin') {
      console.log('Waiting for LinkedIn login...');
      await page.waitForURL('**/feed/**', { timeout: 300000 }); // 5 min timeout
    } else {
      // Generic wait - user closes browser or timeout
      console.log('Press Enter in this terminal when you have finished logging in...');
      await new Promise<void>((resolve) => {
        process.stdin.once('data', () => resolve());
      });
    }

    // Save storage state
    await context.storageState({ path: STORAGE_STATE_PATH });
    console.log(`\nSession saved to: ${STORAGE_STATE_PATH}`);

    // Update config to use the storage state
    configRepository.updateAppConfig({
      browser: {
        ...configRepository.loadAppConfig().browser,
        storageState: STORAGE_STATE_PATH,
      },
    });
    console.log('Config updated to use saved session.');

    await browser.close();
    console.log('\nLogin complete! Your session will be reused for future scraping.');
  });
