import { Command } from 'commander';
import { applicationOrchestrator } from '../../core/application';
import { parseJobUrl, validateUrls, readUrlsFromFile, getSupportedPlatforms } from '../../utils/url-parser';
import { profileRepository } from '../../db/repositories/profile';
import { logger, chalk } from '../../utils/logger';
import { existsSync } from 'fs';

export const applyCommand = new Command('apply')
  .description('Apply to job(s)')
  .argument('[urls...]', 'Job URL(s) to apply to')
  .option('-f, --file <path>', 'Read URLs from file (one per line)')
  .option('-d, --dry-run', 'Generate documents without submitting')
  .action(async (urls: string[], options: { file?: string; dryRun?: boolean }) => {
    // Check for profile
    const profile = profileRepository.findFirst();
    if (!profile) {
      logger.error('No profile found. Run "autoply init" first.');
      process.exit(1);
    }

    // Collect URLs
    let allUrls: string[] = urls || [];

    // Read from file if specified
    if (options.file) {
      if (!existsSync(options.file)) {
        logger.error(`File not found: ${options.file}`);
        process.exit(1);
      }
      const fileUrls = await readUrlsFromFile(options.file);
      allUrls = [...allUrls, ...fileUrls];
    }

    // Check if we have URLs
    if (allUrls.length === 0) {
      logger.error('No URLs provided. Usage: autoply apply <url> or autoply apply --file urls.txt');
      logger.newline();
      logger.info('Supported platforms:');
      for (const platform of getSupportedPlatforms()) {
        console.log(`  - ${platform}`);
      }
      process.exit(1);
    }

    // Validate URLs
    const { valid, invalid } = validateUrls(allUrls);

    if (invalid.length > 0) {
      logger.warning(`${invalid.length} invalid URL(s):`);
      for (const inv of invalid) {
        logger.error(`  ${inv.url}: ${inv.error}`);
      }
      logger.newline();
    }

    if (valid.length === 0) {
      logger.error('No valid URLs to process.');
      process.exit(1);
    }

    logger.info(`Processing ${valid.length} job(s)...`);
    if (options.dryRun) {
      logger.info(chalk.yellow('Dry run mode - applications will not be submitted'));
    }
    logger.newline();

    // Process applications
    const results = await applicationOrchestrator.applyToMultipleJobs(
      valid.map((v) => v.url),
      { dryRun: options.dryRun, profile }
    );

    // Summary
    logger.header('Summary');
    const successful = results.filter((r) => r.success);
    const failed = results.filter((r) => !r.success);

    logger.keyValue('Total', results.length.toString());
    logger.keyValue('Successful', chalk.green(successful.length.toString()));
    logger.keyValue('Failed', failed.length > 0 ? chalk.red(failed.length.toString()) : '0');

    if (successful.length > 0) {
      logger.newline();
      console.log(chalk.bold('Processed:'));
      for (const result of successful) {
        console.log(`  ${chalk.green('✔')} ${result.application?.job_title} at ${result.application?.company}`);
      }
    }

    if (failed.length > 0) {
      logger.newline();
      console.log(chalk.bold('Failed:'));
      for (const result of failed) {
        console.log(`  ${chalk.red('✖')} ${result.error}`);
      }
    }
  });
