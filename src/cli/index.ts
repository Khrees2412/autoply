#!/usr/bin/env bun
import { Command } from 'commander';
import { initCommand } from './commands/init';
import { profileCommand } from './commands/profile';
import { configCommand } from './commands/config';
import { applyCommand } from './commands/apply';
import { generateCommand } from './commands/generate';
import { historyCommand } from './commands/history';
import { closeDb } from '../db';

const program = new Command();

program
  .name('autoply')
  .description('Automated job application CLI - Apply to jobs with AI-generated resumes')
  .version('1.0.0');

// Register commands
program.addCommand(initCommand);
program.addCommand(profileCommand);
program.addCommand(configCommand);
program.addCommand(applyCommand);
program.addCommand(generateCommand);
program.addCommand(historyCommand);

// Cleanup on exit
process.on('exit', () => {
  closeDb();
});

process.on('SIGINT', () => {
  closeDb();
  process.exit(0);
});

process.on('SIGTERM', () => {
  closeDb();
  process.exit(0);
});

// Parse and execute
program.parse(process.argv);

// Show help if no command provided
if (!process.argv.slice(2).length) {
  program.outputHelp();
}
