import { BaseScraper } from './base';
import type { JobData, CustomQuestion, Platform } from '../types';

export class TeamtailorScraper extends BaseScraper {
  platform: Platform = 'teamtailor';

  protected async waitForContent(): Promise<void> {
    if (!this.page) return;
    await this.page.waitForSelector('.job-ad, .careersite-job, [class*="job-page"]', {
      timeout: 10000,
    }).catch(() => {});
  }

  protected async extractJobData(url: string): Promise<JobData> {
    if (!this.page) throw new Error('Page not initialized');

    // Extract job title
    const title = await this.extractText(
      'h1[class*="title"], .job-header h1, .careersite-job__title'
    );

    // Extract company from URL
    const urlMatch = url.match(/([^.]+)\.teamtailor\.com/);
    const company = urlMatch ? urlMatch[1].replace(/-/g, ' ') : 'Unknown Company';

    // Extract location
    const location = await this.extractText(
      '[class*="location"], .job-header__location, .careersite-job__location'
    );

    // Extract job type
    const jobType = await this.extractText(
      '[class*="employment-type"], .job-header__employment-type'
    );

    // Extract description
    const descriptionParts = await this.extractAllText(
      '.job-ad__content, .careersite-job__content, [class*="job-description"]'
    );
    const description = descriptionParts.join('\n\n');

    // Check remote
    const remote =
      location.toLowerCase().includes('remote') ||
      jobType.toLowerCase().includes('remote');

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
      job_type: jobType.trim() || undefined,
      remote,
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
    const questionContainers = await this.page.$$(
      '.application-form__question, [class*="custom-question"], [class*="form-group"]'
    );

    for (let i = 0; i < questionContainers.length; i++) {
      const container = questionContainers[i];

      const questionText = await container.$eval(
        'label, .question-label',
        (el) => el.textContent?.trim() ?? ''
      ).catch(() => '');

      if (!questionText) continue;

      // Skip common form labels that aren't questions
      const skipLabels = ['name', 'email', 'phone', 'resume', 'cv', 'cover letter'];
      if (skipLabels.some((skip) => questionText.toLowerCase().includes(skip))) {
        continue;
      }

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
        id: `teamtailor_q_${i}`,
        question: questionText,
        type,
        required,
        options,
      });
    }

    return questions;
  }
}
