import { chromium, type Browser, type Page, type BrowserContext } from 'playwright';
import type { JobData, FormField, CustomQuestion, Platform } from '../types';
import { configRepository } from '../db/repositories/config';

// Random delay to mimic human behavior
function randomDelay(min: number, max: number): Promise<void> {
  const delay = Math.floor(Math.random() * (max - min + 1)) + min;
  return new Promise((resolve) => setTimeout(resolve, delay));
}

export abstract class BaseScraper {
  abstract platform: Platform;
  protected browser: Browser | null = null;
  protected context: BrowserContext | null = null;
  protected page: Page | null = null;

  async initialize(): Promise<void> {
    const config = configRepository.loadAppConfig();
    this.browser = await chromium.launch({
      headless: config.browser.headless,
      args: [
        '--disable-blink-features=AutomationControlled',
        '--disable-features=IsolateOrigins,site-per-process',
      ],
    });
    this.context = await this.browser.newContext({
      userAgent:
        'Mozilla/5.0 (Macintosh; Apple Silicon Mac OS X 14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
      storageState: config.browser.storageState,
      viewport: { width: 1920, height: 1080 },
      locale: 'en-NG',
      timezoneId: 'Africa/Lagos',
    });

    // Mask automation indicators
    await this.context.addInitScript(() => {
      // Remove webdriver flag
      Object.defineProperty(navigator, 'webdriver', { get: () => undefined });

      // Mock plugins (real browsers have these)
      Object.defineProperty(navigator, 'plugins', {
        get: () => [
          { name: 'Chrome PDF Plugin', filename: 'internal-pdf-viewer' },
          { name: 'Chrome PDF Viewer', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai' },
          { name: 'Native Client', filename: 'internal-nacl-plugin' },
        ],
      });

      // Mock languages
      Object.defineProperty(navigator, 'languages', {
        get: () => ['en-NG', 'en-GB', 'en'],
      });

      // Hide automation-related Chrome properties
      const originalQuery = window.navigator.permissions.query;
      window.navigator.permissions.query = (parameters: PermissionDescriptor) => {
        if (parameters.name === 'notifications') {
          return Promise.resolve({ state: 'prompt', onchange: null } as PermissionStatus);
        }
        return originalQuery(parameters);
      };

      // Mask Chrome property
      (window as unknown as { chrome: unknown }).chrome = { runtime: {} };
    });

    this.page = await this.context.newPage();
    this.page.setDefaultTimeout(config.browser.timeout);
  }

  // Add human-like delay between actions
  protected async humanDelay(short = false): Promise<void> {
    if (short) {
      await randomDelay(300, 800);
    } else {
      await randomDelay(1000, 3000);
    }
  }

  // Simulate human-like scrolling
  protected async humanScroll(): Promise<void> {
    if (!this.page) return;

    const scrolls = Math.floor(Math.random() * 3) + 2; // 2-4 scrolls
    for (let i = 0; i < scrolls; i++) {
      const scrollAmount = Math.floor(Math.random() * 300) + 100;
      await this.page.mouse.wheel(0, scrollAmount);
      await randomDelay(500, 1500);
    }
  }

  async cleanup(): Promise<void> {
    if (this.context) {
      await this.context.close();
      this.context = null;
    }
    if (this.browser) {
      await this.browser.close();
      this.browser = null;
    }
    this.page = null;
  }

  async scrape(url: string): Promise<JobData> {
    try {
      await this.initialize();
      if (!this.page) throw new Error('Browser not initialized');

      // Random delay before navigation
      await this.humanDelay();

      await this.page.goto(url, { waitUntil: 'networkidle' });

      // Simulate human behavior: mouse movement and scrolling
      await this.humanDelay(true);
      await this.page.mouse.move(
        Math.random() * 500 + 100,
        Math.random() * 300 + 100
      );
      await this.humanScroll();

      await this.waitForContent();

      const jobData = await this.extractJobData(url);
      return jobData;
    } finally {
      await this.cleanup();
    }
  }

  protected abstract waitForContent(): Promise<void>;
  protected abstract extractJobData(url: string): Promise<JobData>;

  protected async extractText(selector: string): Promise<string> {
    if (!this.page) return '';
    try {
      const element = await this.page.$(selector);
      if (!element) return '';
      return (await element.textContent()) ?? '';
    } catch {
      return '';
    }
  }

  protected async extractAllText(selector: string): Promise<string[]> {
    if (!this.page) return [];
    try {
      const elements = await this.page.$$(selector);
      const texts: string[] = [];
      for (const element of elements) {
        const text = await element.textContent();
        if (text) texts.push(text.trim());
      }
      return texts;
    } catch {
      return [];
    }
  }

  protected async extractFormFields(): Promise<FormField[]> {
    if (!this.page) return [];

    const fields: FormField[] = [];

    // Extract input fields
    const inputs = await this.page.$$('input:not([type="hidden"]):not([type="submit"])');
    for (const input of inputs) {
      const name = (await input.getAttribute('name')) ?? '';
      const type = ((await input.getAttribute('type')) ?? 'text') as FormField['type'];
      const label = await this.findLabelForInput(input);
      const required = (await input.getAttribute('required')) !== null;

      if (name || label) {
        fields.push({ name, type, label, required });
      }
    }

    // Extract textareas
    const textareas = await this.page.$$('textarea');
    for (const textarea of textareas) {
      const name = (await textarea.getAttribute('name')) ?? '';
      const label = await this.findLabelForInput(textarea);
      const required = (await textarea.getAttribute('required')) !== null;

      if (name || label) {
        fields.push({ name, type: 'textarea', label, required });
      }
    }

    // Extract selects
    const selects = await this.page.$$('select');
    for (const select of selects) {
      const name = (await select.getAttribute('name')) ?? '';
      const label = await this.findLabelForInput(select);
      const required = (await select.getAttribute('required')) !== null;
      const options = await select.$$eval('option', (opts) =>
        opts.map((o) => o.textContent?.trim() ?? '').filter(Boolean)
      );

      if (name || label) {
        fields.push({ name, type: 'select', label, required, options });
      }
    }

    return fields;
  }

  protected async findLabelForInput(input: unknown): Promise<string> {
    if (!this.page) return '';

    try {
      // Try to find associated label by id
      const id = await (input as { getAttribute: (attr: string) => Promise<string | null> }).getAttribute('id');
      if (id) {
        const label = await this.page.$(`label[for="${id}"]`);
        if (label) {
          const text = await label.textContent();
          if (text) return text.trim();
        }
      }

      // Try to find parent label
      const parentLabel = await this.page.evaluate((el) => {
        const parent = (el as HTMLElement).closest('label');
        return parent?.textContent?.trim() ?? '';
      }, input);

      if (parentLabel) return parentLabel;

      // Try aria-label
      const ariaLabel = await (input as { getAttribute: (attr: string) => Promise<string | null> }).getAttribute('aria-label');
      if (ariaLabel) return ariaLabel;

      // Try placeholder
      const placeholder = await (input as { getAttribute: (attr: string) => Promise<string | null> }).getAttribute('placeholder');
      if (placeholder) return placeholder;

      return '';
    } catch {
      return '';
    }
  }

  protected extractRequirements(description: string): string[] {
    const lines = description.split('\n');
    const requirements: string[] = [];
    let inRequirements = false;

    for (const line of lines) {
      const trimmed = line.trim();
      const lower = trimmed.toLowerCase();

      if (
        lower.includes('requirement') ||
        lower.includes('must have') ||
        lower.includes('you will need')
      ) {
        inRequirements = true;
        continue;
      }

      if (
        inRequirements &&
        (lower.includes('nice to have') ||
          lower.includes('preferred') ||
          lower.includes('bonus') ||
          lower.includes('what we offer'))
      ) {
        inRequirements = false;
      }

      if (inRequirements && (trimmed.startsWith('-') || trimmed.startsWith('•') || trimmed.startsWith('*'))) {
        requirements.push(trimmed.replace(/^[-•*]\s*/, ''));
      }
    }

    return requirements;
  }

  protected extractQualifications(description: string): string[] {
    const lines = description.split('\n');
    const qualifications: string[] = [];
    let inQualifications = false;

    for (const line of lines) {
      const trimmed = line.trim();
      const lower = trimmed.toLowerCase();

      if (
        lower.includes('qualification') ||
        lower.includes('nice to have') ||
        lower.includes('preferred')
      ) {
        inQualifications = true;
        continue;
      }

      if (
        inQualifications &&
        (lower.includes('responsibilit') || lower.includes('what we offer') || lower.includes('benefit'))
      ) {
        inQualifications = false;
      }

      if (
        inQualifications &&
        (trimmed.startsWith('-') || trimmed.startsWith('•') || trimmed.startsWith('*'))
      ) {
        qualifications.push(trimmed.replace(/^[-•*]\s*/, ''));
      }
    }

    return qualifications;
  }

  async takeScreenshot(path: string): Promise<void> {
    if (this.page) {
      await this.page.screenshot({ path, fullPage: true });
    }
  }
}

export interface ScraperConstructor {
  new (): BaseScraper;
}
