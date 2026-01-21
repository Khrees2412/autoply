import chalk from 'chalk';
import ora, { type Ora } from 'ora';

export const logger = {
  info: (message: string) => console.log(chalk.blue('ℹ'), message),
  success: (message: string) => console.log(chalk.green('✔'), message),
  warning: (message: string) => console.log(chalk.yellow('⚠'), message),
  error: (message: string) => console.log(chalk.red('✖'), message),
  debug: (message: string) => {
    if (process.env.DEBUG) {
      console.log(chalk.gray('⚙'), message);
    }
  },

  // Styled text helpers
  bold: (text: string) => chalk.bold(text),
  dim: (text: string) => chalk.dim(text),
  cyan: (text: string) => chalk.cyan(text),
  green: (text: string) => chalk.green(text),
  yellow: (text: string) => chalk.yellow(text),
  red: (text: string) => chalk.red(text),

  // Table-like output
  keyValue: (key: string, value: string) => {
    console.log(`  ${chalk.gray(key + ':')} ${value}`);
  },

  // Newline
  newline: () => console.log(),

  // Header
  header: (text: string) => {
    console.log();
    console.log(chalk.bold.underline(text));
    console.log();
  },
};

export function createSpinner(text: string): Ora {
  return ora({ text, color: 'cyan' });
}

export { chalk };
