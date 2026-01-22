import { BaseScraper, type SubmissionOptions, type SubmissionResult } from './base';
import type { JobData, CustomQuestion, Platform } from '../types';
import { FormFiller } from '../core/form-filler';

export class AshbyScraper extends BaseScraper {
  platform: Platform = 'ashby';

  protected async waitForContent(): Promise<void> {
    if (!this.page) return;
    await this.page.waitForSelector('[data-testid="job-post-title"], .ashby-job-posting-heading, h1', {
      timeout: 10000,
    }).catch(() => {});
  }

  // ============ Ashby Form Submission ============

  override async submitApplication(url: string, options: SubmissionOptions): Promise<SubmissionResult> {
    const errors: string[] = [];

    try {
      await this.initialize();
      if (!this.page) throw new Error('Browser not initialized');

      await this.humanDelay();
      await this.page.goto(url, { waitUntil: 'networkidle' });
      await this.humanDelay(true);
      await this.humanScroll();

      // Navigate to application form
      await this.navigateToAshbyApplication();
      await this.waitForAshbyApplicationForm();

      // Fill form
      await this.fillAshbyForm(options, errors);

      // Submit
      const submitted = await this.clickAshbySubmit();
      if (!submitted) {
        return { success: false, message: 'Could not find submit button', errors };
      }

      // Wait for confirmation
      const confirmation = await this.waitForAshbyConfirmation();

      // Screenshot
      const { configRepository } = await import('../db/repositories/config');
      const config = configRepository.loadAppConfig();
      let screenshotPath: string | undefined;
      if (config.application.saveScreenshots) {
        const { getAutoplyDir } = await import('../db');
        const { join } = await import('path');
        screenshotPath = join(getAutoplyDir(), 'screenshots', `ashby_${Date.now()}.png`);
        await this.takeScreenshot(screenshotPath);
      }

      return { success: confirmation.success, message: confirmation.message, screenshotPath, errors };
    } catch (error) {
      errors.push(error instanceof Error ? error.message : 'Unknown error');
      return { success: false, message: 'Ashby submission failed', errors };
    } finally {
      await this.cleanup();
    }
  }

  private async navigateToAshbyApplication(): Promise<void> {
    if (!this.page) return;

    const selectors = [
      '[data-testid="apply-button"]',
      'button:has-text("Apply")',
      'a:has-text("Apply for this job")',
      'a[href*="apply"]',
    ];

    for (const selector of selectors) {
      const button = await this.page.$(selector);
      if (button) {
        await this.humanDelay(true);
        await button.click();
        await this.page.waitForLoadState('networkidle');
        return;
      }
    }
  }

  private async waitForAshbyApplicationForm(): Promise<void> {
    if (!this.page) return;

    await this.page.waitForSelector('form, [data-testid*="application"], .ashby-application-form', {
      timeout: 10000,
    }).catch(() => {});
    await this.humanDelay(true);
  }

  private async fillAshbyForm(options: SubmissionOptions, errors: string[]): Promise<void> {
    if (!this.page) return;

    const { profile } = options;
    const filler = new FormFiller(this.page, profile, options.jobData, {
      resumePath: options.resumePath,
      coverLetterPath: options.coverLetterPath,
      answeredQuestions: options.answeredQuestions,
    });

    // Fill basic fields
    await this.fillInput('input[name*="name"], input[data-testid*="name"]', profile.name);
    await this.fillInput('input[name*="email"], input[type="email"]', profile.email);
    if (profile.phone) {
      await this.fillInput('input[name*="phone"], input[type="tel"]', profile.phone);
    }
    if (profile.linkedin_url) {
      await this.fillInput('input[name*="linkedin"], input[placeholder*="LinkedIn"]', profile.linkedin_url);
    }

    // Upload resume
    if (options.resumePath) {
      const fileInput = await this.page.$('input[type="file"]');
      if (fileInput) {
        await fileInput.setInputFiles(options.resumePath);
        await this.page.waitForTimeout(2000);
      }
    }

    // Custom questions
    if (options.answeredQuestions) {
      const result = await filler.fillCustomQuestions(options.answeredQuestions);
      errors.push(...result.errors);
    }

    await this.humanDelay(true);
  }

  private async fillInput(selector: string, value: string): Promise<boolean> {
    if (!this.page || !value) return false;
    try {
      const input = await this.page.$(selector);
      if (input) {
        await input.fill(value);
        await this.humanDelay(true);
        return true;
      }
      return false;
    } catch {
      return false;
    }
  }

  private async clickAshbySubmit(): Promise<boolean> {
    if (!this.page) return false;

    const selectors = [
      'button[type="submit"]',
      'button:has-text("Submit")',
      '[data-testid="submit-button"]',
    ];

    for (const selector of selectors) {
      const button = await this.page.$(selector);
      if (button) {
        const isEnabled = await button.isEnabled();
        if (isEnabled) {
          await this.humanDelay(true);
          await button.click();
          return true;
        }
      }
    }
    return false;
  }

  private async waitForAshbyConfirmation(): Promise<{ success: boolean; message: string }> {
    if (!this.page) return { success: false, message: 'Page not initialized' };

    try {
      await this.page.waitForTimeout(3000);

      const successElement = await this.page.$('[class*="success"], :has-text("Thank you"), :has-text("Application submitted")');
      if (successElement) {
        return { success: true, message: 'Ashby application submitted' };
      }

      return { success: true, message: 'Submission completed' };
    } catch {
      return { success: false, message: 'Confirmation check failed' };
    }
  }

  protected async extractJobData(url: string): Promise<JobData> {
    if (!this.page) throw new Error('Page not initialized');

    // Extract job title
    const title = await this.extractText(
      '[data-testid="job-post-title"], .ashby-job-posting-heading h1, h1'
    );

    // Extract company name from URL or page
    let company = await this.extractText(
      '[data-testid="company-name"], .ashby-company-name, [class*="companyName"]'
    );
    if (!company) {
      // Try to extract from URL (jobs.ashbyhq.com/companyname)
      const urlMatch = url.match(/jobs\.ashbyhq\.com\/([^/]+)/);
      company = urlMatch ? this.formatCompanyName(urlMatch[1]) : 'Unknown Company';
    }

    // Extract job description
    const description = await this.extractText(
      '[data-testid="job-post-description"], .ashby-job-posting-description, [class*="jobDescription"]'
    );

    // Extract location
    const location = await this.extractText(
      '[data-testid="job-post-location"], .ashby-job-posting-location, [class*="location"]'
    );

    // Extract form fields
    const formFields = await this.extractFormFields();

    // Extract custom questions
    const customQuestions = await this.extractCustomQuestions();

    // Extract requirements and qualifications from description
    const requirements = this.extractRequirements(description);
    const qualifications = this.extractQualifications(description);

    return {
      url,
      platform: this.platform,
      title: title.trim() || 'Unknown Position',
      company: company.trim(),
      description: description.trim(),
      requirements,
      qualifications,
      location: location.trim() || undefined,
      form_fields: formFields,
      custom_questions: customQuestions,
    };
  }

  private formatCompanyName(name: string): string {
    return name
      .replace(/-/g, ' ')
      .replace(/\b\w/g, (c) => c.toUpperCase());
  }

  private async extractCustomQuestions(): Promise<CustomQuestion[]> {
    if (!this.page) return [];

    const questions: CustomQuestion[] = [];

    // Ashby uses specific patterns for custom questions in their application forms
    const customFields = await this.page.$$(
      '[data-testid*="question"], [class*="customQuestion"], .ashby-application-form-field'
    );

    for (let i = 0; i < customFields.length; i++) {
      const field = customFields[i];
      const questionText = await field.$eval(
        'label, [class*="label"], [data-testid*="label"]',
        (el) => el.textContent?.trim() ?? ''
      ).catch(() => '');

      if (!questionText) continue;

      // Determine question type
      const hasTextarea = (await field.$('textarea')) !== null;
      const hasSelect = (await field.$('select')) !== null;
      const hasRadio = (await field.$('input[type="radio"]')) !== null;
      const hasCheckbox = (await field.$('input[type="checkbox"]')) !== null;

      let type: CustomQuestion['type'] = 'text';
      let options: string[] | undefined;

      if (hasTextarea) {
        type = 'textarea';
      } else if (hasSelect) {
        type = 'select';
        options = await field.$$eval('select option', (opts) =>
          opts.map((o) => o.textContent?.trim() ?? '').filter(Boolean)
        ).catch(() => []);
      } else if (hasRadio) {
        type = 'radio';
        options = await field.$$eval('input[type="radio"]', (inputs) =>
          inputs.map((inp) => inp.getAttribute('value') ?? '').filter(Boolean)
        ).catch(() => []);
      } else if (hasCheckbox) {
        type = 'checkbox';
        options = await field.$$eval('input[type="checkbox"]', (inputs) =>
          inputs.map((inp) => inp.getAttribute('value') ?? '').filter(Boolean)
        ).catch(() => []);
      }

      const required = (await field.$('[required], [aria-required="true"]')) !== null;

      questions.push({
        id: `question_${i}`,
        question: questionText,
        type,
        required,
        options,
      });
    }

    return questions;
  }
}
