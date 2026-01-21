import { BaseScraper } from './base';
import type { JobData, CustomQuestion, Platform } from '../types';

export class LinkedInScraper extends BaseScraper {
  platform: Platform = 'linkedin';

  protected async waitForContent(): Promise<void> {
    if (!this.page) return;
    await this.page.waitForSelector('.job-view-layout, .jobs-unified-top-card', {
      timeout: 15000,
    }).catch(() => {});
  }

  protected async extractJobData(url: string): Promise<JobData> {
    if (!this.page) throw new Error('Page not initialized');

    // Extract job title
    const title = await this.extractText(
      '.job-details-jobs-unified-top-card__job-title, .jobs-unified-top-card__job-title, h1.t-24'
    );

    // Extract company name
    const company = await this.extractText(
      '.job-details-jobs-unified-top-card__company-name, .jobs-unified-top-card__company-name, a.ember-view.t-black.t-normal'
    );

    // Extract location
    const location = await this.extractText(
      '.job-details-jobs-unified-top-card__primary-description-container, .jobs-unified-top-card__bullet'
    );

    // Extract job description
    const description = await this.extractText(
      '.jobs-description-content__text, .jobs-box__html-content, .description__text'
    );

    // Extract job type
    const jobType = await this.extractText(
      '.job-details-jobs-unified-top-card__job-insight, .jobs-unified-top-card__workplace-type'
    );

    // Check if remote
    const remote = jobType.toLowerCase().includes('remote') || location.toLowerCase().includes('remote');

    // Form fields for LinkedIn are typically handled through their Easy Apply flow
    const formFields = await this.extractFormFields();

    // Custom questions in Easy Apply
    const customQuestions = await this.extractCustomQuestions();

    // Extract requirements and qualifications
    const requirements = this.extractRequirements(description);
    const qualifications = this.extractQualifications(description);

    return {
      url,
      platform: this.platform,
      title: title.trim() || 'Unknown Position',
      company: company.trim() || 'Unknown Company',
      description: description.trim(),
      requirements,
      qualifications,
      location: location.trim() || undefined,
      job_type: jobType.trim() || undefined,
      remote,
      form_fields: formFields,
      custom_questions: customQuestions,
    };
  }

  private async extractCustomQuestions(): Promise<CustomQuestion[]> {
    if (!this.page) return [];

    const questions: CustomQuestion[] = [];

    // LinkedIn Easy Apply questions
    const questionContainers = await this.page.$$(
      '.jobs-easy-apply-form-section__grouping, [class*="fb-form-element"]'
    );

    for (let i = 0; i < questionContainers.length; i++) {
      const container = questionContainers[i];

      const questionText = await container.$eval(
        'label, .fb-form-element-label, [class*="artdeco-text-input--label"]',
        (el) => el.textContent?.trim() ?? ''
      ).catch(() => '');

      if (!questionText) continue;

      const hasTextarea = (await container.$('textarea')) !== null;
      const hasSelect = (await container.$('select')) !== null;
      const hasRadio = (await container.$('input[type="radio"]')) !== null;

      let type: CustomQuestion['type'] = 'text';
      let options: string[] | undefined;

      if (hasTextarea) {
        type = 'textarea';
      } else if (hasSelect) {
        type = 'select';
        options = await container.$$eval('select option', (opts) =>
          opts.map((o) => o.textContent?.trim() ?? '').filter(Boolean)
        );
      } else if (hasRadio) {
        type = 'radio';
        options = await container.$$eval('[class*="fb-radio-button"] label', (labels) =>
          labels.map((l) => l.textContent?.trim() ?? '').filter(Boolean)
        );
      }

      const required = (await container.$('[required], [aria-required="true"]')) !== null;

      questions.push({
        id: `linkedin_q_${i}`,
        question: questionText,
        type,
        required,
        options,
      });
    }

    return questions;
  }
}
