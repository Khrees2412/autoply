import { describe, expect, test } from 'bun:test';
import { createScraper } from './index';
import { GreenhouseScraper } from './greenhouse';
import { LeverScraper } from './lever';
import { LinkedInScraper } from './linkedin';
import { JobviteScraper } from './jobvite';
import { SmartRecruitersScraper } from './smartrecruiters';
import { PinpointScraper } from './pinpoint';
import { TeamtailorScraper } from './teamtailor';
import { WorkdayScraper } from './workday';
import { AshbyScraper } from './ashby';
import type { Platform } from '../types';

describe('createScraper', () => {
  test('creates GreenhouseScraper for greenhouse platform', () => {
    const scraper = createScraper('greenhouse');
    expect(scraper).toBeInstanceOf(GreenhouseScraper);
    expect(scraper.platform).toBe('greenhouse');
  });

  test('creates LeverScraper for lever platform', () => {
    const scraper = createScraper('lever');
    expect(scraper).toBeInstanceOf(LeverScraper);
    expect(scraper.platform).toBe('lever');
  });

  test('creates LinkedInScraper for linkedin platform', () => {
    const scraper = createScraper('linkedin');
    expect(scraper).toBeInstanceOf(LinkedInScraper);
    expect(scraper.platform).toBe('linkedin');
  });

  test('creates JobviteScraper for jobvite platform', () => {
    const scraper = createScraper('jobvite');
    expect(scraper).toBeInstanceOf(JobviteScraper);
    expect(scraper.platform).toBe('jobvite');
  });

  test('creates SmartRecruitersScraper for smartrecruiters platform', () => {
    const scraper = createScraper('smartrecruiters');
    expect(scraper).toBeInstanceOf(SmartRecruitersScraper);
    expect(scraper.platform).toBe('smartrecruiters');
  });

  test('creates PinpointScraper for pinpoint platform', () => {
    const scraper = createScraper('pinpoint');
    expect(scraper).toBeInstanceOf(PinpointScraper);
    expect(scraper.platform).toBe('pinpoint');
  });

  test('creates TeamtailorScraper for teamtailor platform', () => {
    const scraper = createScraper('teamtailor');
    expect(scraper).toBeInstanceOf(TeamtailorScraper);
    expect(scraper.platform).toBe('teamtailor');
  });

  test('creates WorkdayScraper for workday platform', () => {
    const scraper = createScraper('workday');
    expect(scraper).toBeInstanceOf(WorkdayScraper);
    expect(scraper.platform).toBe('workday');
  });

  test('creates AshbyScraper for ashby platform', () => {
    const scraper = createScraper('ashby');
    expect(scraper).toBeInstanceOf(AshbyScraper);
    expect(scraper.platform).toBe('ashby');
  });

  test('throws error for unsupported platform', () => {
    expect(() => createScraper('unsupported' as Platform)).toThrow(
      'No scraper available for platform: unsupported'
    );
  });

  test('all platforms have corresponding scrapers', () => {
    const platforms: Platform[] = [
      'greenhouse',
      'lever',
      'linkedin',
      'jobvite',
      'smartrecruiters',
      'pinpoint',
      'teamtailor',
      'workday',
      'ashby',
    ];

    for (const platform of platforms) {
      const scraper = createScraper(platform);
      expect(scraper).toBeDefined();
      expect(scraper.platform).toBe(platform);
    }
  });

  test('each call creates a new scraper instance', () => {
    const scraper1 = createScraper('greenhouse');
    const scraper2 = createScraper('greenhouse');
    expect(scraper1).not.toBe(scraper2);
  });
});
