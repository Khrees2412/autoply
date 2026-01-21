import { chromium, type Browser, type Page, type BrowserContext } from 'playwright';
import type { JobData, FormField, CustomQuestion, Platform } from '../types';
import { configRepository } from '../db/repositories/config';

export abstract class BaseScraper {
  abstract platform: Platform;
  protected browser: Browser | null = null;
  protected context: BrowserContext | null = null;
  protected page: Page | null = null;

  async initialize(): Promise<void> {
    const config = configRepository.loadAppConfig();
    this.browser = await chromium.launch({
      headless: config.browser.headless,
    });
    this.context = await this.browser.newContext({
      userAgent:
        'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
    });
    this.page = await this.context.newPage();
    this.page.setDefaultTimeout(config.browser.timeout);
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

      await this.page.goto(url, { waitUntil: 'networkidle' });
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
