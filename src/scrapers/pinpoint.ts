import { BaseScraper } from './base';
import type { JobData, CustomQuestion, Platform } from '../types';

export class PinpointScraper extends BaseScraper {
  platform: Platform = 'pinpoint';

  protected async waitForContent(): Promise<void> {
    if (!this.page) return;
    await this.page.waitForSelector('.job-page, .job-content, [class*="vacancy"]', {
      timeout: 10000,
    }).catch(() => {});
  }

  protected async extractJobData(url: string): Promise<JobData> {
    if (!this.page) throw new Error('Page not initialized');

    // Extract job title
    const title = await this.extractText('h1.job-title, h1[class*="title"], .vacancy-title');

    // Extract company from URL
    const urlMatch = url.match(/([^.]+)\.pinpointhq\.com/);
    const company = urlMatch ? urlMatch[1].replace(/-/g, ' ') : 'Unknown Company';

    // Extract location
    const location = await this.extractText('.job-location, [class*="location"], .vacancy-location');

    // Extract description
    const description = await this.extractText(
      '.job-description, .job-content, .vacancy-description, [class*="description"]'
    );

    // Extract form fields
    const formFields = await this.extractFormFields();

    // Extract custom questions
    const customQuestions = await this.extractCustomQuestions();

    // Parse requirements
    const requirements = this.extractRequirements(description);
    const qualifications = this.extractQualifications(description);

    return {
      url,
      platform: this.platform,
      title: title.trim() || 'Unknown Position',
      company: this.capitalizeWords(company),
      description: description.trim(),
      requirements,
      qualifications,
      location: location.trim() || undefined,
      form_fields: formFields,
      custom_questions: customQuestions,
    };
  }

  private capitalizeWords(str: string): string {
    return str
      .split(' ')
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join(' ');
  }

  private async extractCustomQuestions(): Promise<CustomQuestion[]> {
    if (!this.page) return [];

    const questions: CustomQuestion[] = [];
    const questionContainers = await this.page.$$('[class*="question"], [class*="custom-field"]');

    for (let i = 0; i < questionContainers.length; i++) {
      const container = questionContainers[i];

      const questionText = await container.$eval(
        'label, .question-text, [class*="label"]',
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
      }

      const required = (await container.$('[required]')) !== null;

      questions.push({
        id: `pinpoint_q_${i}`,
        question: questionText,
        type,
        required,
        options,
      });
    }

    return questions;
  }
}
