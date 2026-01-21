import { BaseScraper } from './base';
import type { JobData, CustomQuestion, Platform } from '../types';

export class AshbyScraper extends BaseScraper {
  platform: Platform = 'ashby';

  protected async waitForContent(): Promise<void> {
    if (!this.page) return;
    await this.page.waitForSelector('[data-testid="job-post-title"], .ashby-job-posting-heading, h1', {
      timeout: 10000,
    }).catch(() => {});
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
