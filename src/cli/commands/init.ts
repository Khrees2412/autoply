import { Command } from 'commander';
import { profileRepository } from '../../db/repositories/profile';
import { configRepository } from '../../db/repositories/config';
import { promptForProfile } from '../prompts/profile';
import { logger } from '../../utils/logger';
import { DEFAULT_CONFIG } from '../../types';
import { getDb, ensureAutoplyDir, getAutoplyDir } from '../../db';

export const initCommand = new Command('init')
  .description('Initialize Autoply with your profile')
  .action(async () => {
    try {
      // Ensure directories exist
      ensureAutoplyDir();

      // Initialize database
      getDb();

      // Check if profile already exists
      const existingProfile = profileRepository.findFirst();
      if (existingProfile) {
        logger.warning('A profile already exists. Use "autoply profile edit" to modify it.');
        logger.info(`Current profile: ${existingProfile.name} <${existingProfile.email}>`);
        return;
      }

      // Prompt for profile information
      const profileData = await promptForProfile();

      // Create profile
      const profile = profileRepository.create(profileData);

      // Save default config
      configRepository.saveAppConfig(DEFAULT_CONFIG);

      logger.newline();
      logger.success('Profile created successfully!');
      logger.newline();
      logger.keyValue('Name', profile.name);
      logger.keyValue('Email', profile.email);
      logger.keyValue('Skills', profile.skills.join(', ') || 'None');
      logger.keyValue('Education', `${profile.education.length} entries`);
      logger.newline();
      logger.info(`Data stored in: ${getAutoplyDir()}`);
      logger.newline();
      logger.info('Next steps:');
      logger.info('  1. Configure AI provider: autoply config set ai.provider ollama');
      logger.info('  2. Apply to a job: autoply apply <job-url>');
    } catch (error) {
      if (error instanceof Error && error.message.includes('ExitPromptError')) {
        logger.info('Setup cancelled.');
        return;
      }
      logger.error(`Initialization failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
      process.exit(1);
    }
  });
