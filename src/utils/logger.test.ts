import { describe, test, expect, spyOn, beforeEach } from 'bun:test';
import { setVerbose, logger } from './logger';

describe('logger verbose mode', () => {
  let consoleSpy: ReturnType<typeof spyOn>;

  beforeEach(() => {
    consoleSpy = spyOn(console, 'log').mockImplementation(() => {});
    setVerbose(false);
    delete process.env.DEBUG;
  });

  test('debug does not log when verbose is off', () => {
    logger.debug('test message');
    expect(consoleSpy).not.toHaveBeenCalled();
  });

  test('debug logs when verbose is on', () => {
    setVerbose(true);
    logger.debug('test message');
    expect(consoleSpy).toHaveBeenCalled();
  });

  test('debug logs when DEBUG env is set', () => {
    process.env.DEBUG = '1';
    logger.debug('test message');
    expect(consoleSpy).toHaveBeenCalled();
  });
});
