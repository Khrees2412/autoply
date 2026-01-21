import { BaseScraper } from './base';
import type { JobData, CustomQuestion, Platform } from '../types';

export class WorkdayScraper extends BaseScraper {
  platform: Platform = 'workday';

  protected async waitForContent(): Promise<void> {
    if (!this.page) return;
    await this.page.waitForSelector('[data-automation-id="jobPostingHeader"], .css-1q2dra3, [data-automation-id="jobPostingDescription"]', {
      timeout: 15000,
    }).catch(() => {});
  }

  protected async extractJobData(url: string): Promise<JobData> {
    if (!this.page) throw new Error('Page not initialized');

    // Extract job title
    const title = await this.extractText(
      '[data-automation-id="jobPostingHeader"] h2, [data-automation-id="jobTitle"], h1[data-automation-id]'
    );

    // Extract company name from URL or page
    let company = await this.extractText(
      '[data-automation-id="jobPostingCompanyName"], .css-1q2dra3 [data-automation-id="companyName"]'
    );
    if (!company) {
      // Try to extract from URL (company.myworkdayjobs.com)
      const urlMatch = url.match(/([^.]+)\.myworkdayjobs\.com/);
      company = urlMatch ? this.formatCompanyName(urlMatch[1]) : 'Unknown Company';
    }

    // Extract job description
    const description = await this.extractText(
      '[data-automation-id="jobPostingDescription"], [data-automation-id="jobDescription"], .job-description'
    );

    // Extract location
    const location = await this.extractText(
      '[data-automation-id="locations"], [data-automation-id="jobPostingLocation"], [data-automation-id="location"]'
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

    // Workday uses data-automation-id for form elements
    const customFields = await this.page.$$(
      '[data-automation-id*="question"], [data-automation-id*="formField"], .css-1mog1xl'
    );

    for (let i = 0; i < customFields.length; i++) {
      const field = customFields[i];
      const questionText = await field.$eval(
        'label, [data-automation-id*="label"]',
        (el) => el.textContent?.trim() ?? ''
      ).catch(() => '');

      if (!questionText) continue;

      // Determine question type
      const hasTextarea = (await field.$('textarea')) !== null;
      const hasSelect = (await field.$('select, [data-automation-id="selectWidget"]')) !== null;
      const hasRadio = (await field.$('input[type="radio"]')) !== null;
      const hasCheckbox = (await field.$('input[type="checkbox"]')) !== null;

      let type: CustomQuestion['type'] = 'text';
      let options: string[] | undefined;

      if (hasTextarea) {
        type = 'textarea';
      } else if (hasSelect) {
        type = 'select';
        options = await field.$$eval('[data-automation-id*="option"], option', (opts) =>
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

      const required = (await field.$('[required], [data-automation-id*="required"]')) !== null;

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
