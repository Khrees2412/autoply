import { BaseScraper } from './base';
import type { JobData, CustomQuestion, Platform } from '../types';

export class LeverScraper extends BaseScraper {
  platform: Platform = 'lever';

  protected async waitForContent(): Promise<void> {
    if (!this.page) return;
    await this.page.waitForSelector('.posting-headline, .content', {
      timeout: 10000,
    }).catch(() => {});
  }

  protected async extractJobData(url: string): Promise<JobData> {
    if (!this.page) throw new Error('Page not initialized');

    // Extract job title
    const title = await this.extractText('.posting-headline h2, h1.posting-title');

    // Extract company name
    let company = await this.extractText('.posting-headline .company, .main-header-content h1');
    if (!company) {
      // Extract from URL: jobs.lever.co/companyname
      const urlMatch = url.match(/jobs\.lever\.co\/([^/]+)/);
      company = urlMatch ? urlMatch[1].replace(/-/g, ' ') : 'Unknown Company';
    }

    // Extract location
    const location = await this.extractText('.posting-categories .location, .sort-by-commitment');

    // Extract job description
    const descriptionSections = await this.extractAllText('.posting-description, .section-wrapper');
    const description = descriptionSections.join('\n\n');

    // Extract form fields
    const formFields = await this.extractFormFields();

    // Extract custom questions from application page
    const customQuestions = await this.extractCustomQuestions();

    // Parse requirements and qualifications
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

  private async extractCustomQuestions(): Promise<CustomQuestion[]> {
    if (!this.page) return [];

    const questions: CustomQuestion[] = [];

    // Look for custom question containers
    const questionContainers = await this.page.$$(
      '.custom-question, .application-question, [class*="custom-field"]'
    );

    for (let i = 0; i < questionContainers.length; i++) {
      const container = questionContainers[i];

      const questionText = await container.$eval(
        'label, .question-label',
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
        options = await container.$$eval('label', (labels) =>
          labels
            .filter((l) => l.querySelector('input[type="radio"]'))
            .map((l) => l.textContent?.trim() ?? '')
            .filter(Boolean)
        );
      }

      const required = (await container.$('[required], .required, [aria-required="true"]')) !== null;

      questions.push({
        id: `lever_q_${i}`,
        question: questionText,
        type,
        required,
        options,
      });
    }

    return questions;
  }
}
