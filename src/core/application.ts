import type { Profile, JobData, Application, GeneratedDocuments } from '../types';
import { parseJobUrl } from '../utils/url-parser';
import { scrapeJob, createScraper } from '../scrapers';
import { createAIProvider } from '../ai/provider';
import { tailorResume } from '../ai/resume';
import { generateCoverLetter, answerApplicationQuestion } from '../ai/cover-letter';
import { profileRepository } from '../db/repositories/profile';
import { applicationRepository } from '../db/repositories/application';
import { configRepository } from '../db/repositories/config';
import { ApplicationQueue } from './queue';
import { generateResumePdf, generateCoverLetterPdf } from './document';
import { logger, createSpinner } from '../utils/logger';
import { join } from 'path';
import { mkdir } from 'fs/promises';
import { getAutoplyDir, ensureAutoplyDir } from '../db';

export interface ApplicationResult {
  success: boolean;
  application?: Application;
  error?: string;
  documents?: GeneratedDocuments;
}

export interface ApplyOptions {
  dryRun?: boolean;
  profile?: Profile;
  generateOnly?: boolean;
}

export class ApplicationOrchestrator {
  private queue: ApplicationQueue;

  constructor() {
    this.queue = new ApplicationQueue();
  }

  async applyToJob(url: string, options: ApplyOptions = {}): Promise<ApplicationResult> {
    const { dryRun = false, generateOnly = false } = options;

    // Validate URL
    const parsedUrl = parseJobUrl(url);
    if (!parsedUrl.isValid) {
      return { success: false, error: parsedUrl.error };
    }

    // Get profile
    const profile = options.profile ?? profileRepository.findFirst();
    if (!profile) {
      return { success: false, error: 'No profile found. Run "autoply init" to create one.' };
    }

    const spinner = createSpinner(`Scraping job from ${parsedUrl.platform}...`);
    spinner.start();

    let jobData: JobData;
    try {
      jobData = await scrapeJob(url, parsedUrl.platform);
      spinner.succeed(`Scraped: ${jobData.title} at ${jobData.company}`);
    } catch (error) {
      spinner.fail('Failed to scrape job posting');
      return {
        success: false,
        error: `Scraping failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
      };
    }

    // Don't submit applications with unknown job titles
    if (jobData.title === 'Unknown Position' && !dryRun && !generateOnly) {
      return {
        success: false,
        error: 'Cannot submit application: job title could not be scraped. Try with --dry-run to generate documents only.',
      };
    }

    // Generate documents
    spinner.start('Generating tailored resume...');
    let documents: GeneratedDocuments;
    try {
      const provider = createAIProvider();
      const isAvailable = await provider.isAvailable();
      if (!isAvailable) {
        spinner.fail('AI provider not available');
        return { success: false, error: 'AI provider is not running or configured' };
      }

      const resume = await tailorResume(provider, profile, jobData);
      spinner.succeed('Resume generated');

      spinner.start('Generating cover letter...');
      const coverLetter = await generateCoverLetter(provider, profile, jobData);
      spinner.succeed('Cover letter generated');

      documents = { resume, coverLetter };
    } catch (error) {
      spinner.fail('Document generation failed');
      return {
        success: false,
        error: `AI generation failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
      };
    }

    // If generate only or dry run, save and return
    if (generateOnly || dryRun) {
      if (dryRun) {
        logger.info('Dry run mode - not submitting application');
        logger.newline();
        logger.header('Generated Resume Preview');
        console.log(documents.resume.slice(0, 500) + '...');
        logger.newline();
        logger.header('Generated Cover Letter Preview');
        console.log(documents.coverLetter.slice(0, 500) + '...');
      }

      // Create application record
      const application = applicationRepository.create({
        profile_id: profile.id!,
        url,
        platform: parsedUrl.platform,
        company: jobData.company,
        job_title: jobData.title,
        status: dryRun ? 'pending' : 'submitted',
        generated_resume: documents.resume,
        generated_cover_letter: documents.coverLetter,
      });

      return { success: true, application, documents };
    }

    // Answer custom questions
    if (jobData.custom_questions.length > 0) {
      spinner.start(`Answering ${jobData.custom_questions.length} custom questions...`);
      try {
        const provider = createAIProvider();
        for (const question of jobData.custom_questions) {
          if (!question.answer) {
            question.answer = await answerApplicationQuestion(
              provider,
              profile,
              jobData,
              question.question
            );
          }
        }
        spinner.succeed('Custom questions answered');
      } catch (error) {
        spinner.warn('Some questions could not be auto-answered');
      }
    }

    // Create application record
    const application = applicationRepository.create({
      profile_id: profile.id!,
      url,
      platform: parsedUrl.platform,
      company: jobData.company,
      job_title: jobData.title,
      status: 'pending',
      generated_resume: documents.resume,
      generated_cover_letter: documents.coverLetter,
      form_data: {
        fields: jobData.form_fields,
        questions: jobData.custom_questions,
      },
    });

    // Check if auto-submit is enabled
    const config = configRepository.loadAppConfig();
    if (config.application.autoSubmit) {
      spinner.start('Submitting application...');
      try {
        await this.submitApplication(application, jobData, profile, documents);
        applicationRepository.update(application.id!, {
          status: 'submitted',
          applied_at: new Date().toISOString(),
        });
        spinner.succeed('Application submitted!');
      } catch (error) {
        applicationRepository.update(application.id!, {
          status: 'failed',
          error_message: error instanceof Error ? error.message : 'Unknown error',
        });
        spinner.fail('Application submission failed');
        return {
          success: false,
          application,
          error: `Submission failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
          documents,
        };
      }
    } else {
      logger.info('Auto-submit disabled. Application prepared but not submitted.');
      logger.info('Set autoSubmit to true in config to enable automatic submission.');
    }

    return { success: true, application, documents };
  }

  private async submitApplication(
    application: Application,
    jobData: JobData,
    profile: Profile,
    documents: GeneratedDocuments
  ): Promise<void> {
    const config = configRepository.loadAppConfig();

    // Ensure directories exist
    ensureAutoplyDir();
    const docsDir = join(getAutoplyDir(), 'documents');
    const screenshotsDir = join(getAutoplyDir(), 'screenshots');

    await mkdir(docsDir, { recursive: true });
    await mkdir(screenshotsDir, { recursive: true });

    // Save documents (markdown and PDF)
    const resumeMdPath = join(docsDir, `${application.id}_resume.md`);
    const coverLetterMdPath = join(docsDir, `${application.id}_cover_letter.md`);
    const resumePdfPath = join(docsDir, `${application.id}_resume.pdf`);
    const coverLetterPdfPath = join(docsDir, `${application.id}_cover_letter.pdf`);

    // Save markdown versions
    await Bun.write(resumeMdPath, documents.resume);
    await Bun.write(coverLetterMdPath, documents.coverLetter);

    // Generate PDFs for uploading
    await generateResumePdf(documents.resume, resumePdfPath, profile.name);
    await generateCoverLetterPdf(documents.coverLetter, coverLetterPdfPath, profile.name);

    // Create scraper for this platform
    const scraper = createScraper(application.platform);

    // Prepare answered questions
    const answeredQuestions = jobData.custom_questions;

    // Submit the application using platform-specific scraper
    const result = await scraper.submitApplication(application.url, {
      profile,
      jobData,
      documents,
      resumePath: resumePdfPath,
      coverLetterPath: coverLetterPdfPath,
      answeredQuestions,
    });

    if (!result.success) {
      const errorMsg = result.errors.length > 0
        ? `${result.message}: ${result.errors.join(', ')}`
        : result.message;
      throw new Error(errorMsg);
    }

    // Log success details
    if (result.screenshotPath) {
      logger.info(`Screenshot saved to: ${result.screenshotPath}`);
    }
  }

  async applyToMultipleJobs(urls: string[], options: ApplyOptions = {}): Promise<ApplicationResult[]> {
    const results: ApplicationResult[] = [];

    for (const url of urls) {
      logger.header(`Processing: ${url}`);
      const result = await this.applyToJob(url, options);
      results.push(result);

      if (result.success) {
        logger.success(`Completed: ${result.application?.job_title} at ${result.application?.company}`);
      } else {
        logger.error(`Failed: ${result.error}`);
      }

      logger.newline();
    }

    return results;
  }

  async generateDocuments(
    url: string,
    outputDir: string,
    type: 'resume' | 'cover-letter' | 'both' = 'both'
  ): Promise<{ resumePath?: string; coverLetterPath?: string }> {
    const parsedUrl = parseJobUrl(url);
    if (!parsedUrl.isValid) {
      throw new Error(parsedUrl.error);
    }

    const profile = profileRepository.findFirst();
    if (!profile) {
      throw new Error('No profile found. Run "autoply init" first.');
    }

    const spinner = createSpinner('Scraping job...');
    spinner.start();

    const jobData = await scrapeJob(url, parsedUrl.platform);
    spinner.succeed(`Scraped: ${jobData.title} at ${jobData.company}`);

    const provider = createAIProvider();
    const result: { resumePath?: string; coverLetterPath?: string } = {};

    if (type === 'resume' || type === 'both') {
      spinner.start('Generating tailored resume...');
      const resume = await tailorResume(provider, profile, jobData);
      const resumePath = join(outputDir, `resume_${jobData.company.replace(/\s+/g, '_')}.pdf`);
      await generateResumePdf(resume, resumePath, profile.name);
      result.resumePath = resumePath;
      spinner.succeed(`Resume saved to: ${resumePath}`);
    }

    if (type === 'cover-letter' || type === 'both') {
      spinner.start('Generating cover letter...');
      const coverLetter = await generateCoverLetter(provider, profile, jobData);
      const coverPath = join(outputDir, `cover_letter_${jobData.company.replace(/\s+/g, '_')}.pdf`);
      await generateCoverLetterPdf(coverLetter, coverPath, profile.name);
      result.coverLetterPath = coverPath;
      spinner.succeed(`Cover letter saved to: ${coverPath}`);
    }

    return result;
  }
}

export const applicationOrchestrator = new ApplicationOrchestrator();
