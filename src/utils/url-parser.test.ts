import { describe, expect, test } from 'bun:test';
import {
  parseJobUrl,
  detectPlatform,
  isValidJobUrl,
  getSupportedPlatforms,
  getPlatformExamples,
  validateUrls,
} from './url-parser';
import type { Platform } from '../types';

describe('url-parser', () => {
  describe('parseJobUrl', () => {
    test('parses valid Greenhouse URL', () => {
      const result = parseJobUrl('https://boards.greenhouse.io/company/jobs/12345');
      expect(result.isValid).toBe(true);
      expect(result.platform).toBe('greenhouse');
      expect(result.error).toBeUndefined();
    });

    test('parses valid LinkedIn URL', () => {
      const result = parseJobUrl('https://linkedin.com/jobs/view/12345');
      expect(result.isValid).toBe(true);
      expect(result.platform).toBe('linkedin');
    });

    test('parses valid Lever URL', () => {
      const result = parseJobUrl('https://jobs.lever.co/company/job-id');
      expect(result.isValid).toBe(true);
      expect(result.platform).toBe('lever');
    });

    test('parses valid Jobvite URL', () => {
      const result = parseJobUrl('https://jobs.jobvite.com/company/job/12345');
      expect(result.isValid).toBe(true);
      expect(result.platform).toBe('jobvite');
    });

    test('parses valid SmartRecruiters URL', () => {
      const result = parseJobUrl('https://jobs.smartrecruiters.com/Company/12345');
      expect(result.isValid).toBe(true);
      expect(result.platform).toBe('smartrecruiters');
    });

    test('parses valid Pinpoint URL', () => {
      const result = parseJobUrl('https://company.pinpointhq.com/jobs/12345');
      expect(result.isValid).toBe(true);
      expect(result.platform).toBe('pinpoint');
    });

    test('parses valid Teamtailor URL', () => {
      const result = parseJobUrl('https://company.teamtailor.com/jobs/12345');
      expect(result.isValid).toBe(true);
      expect(result.platform).toBe('teamtailor');
    });

    test('parses valid Workday URL (myworkdayjobs)', () => {
      const result = parseJobUrl('https://company.myworkdayjobs.com/en-US/External/job/12345');
      expect(result.isValid).toBe(true);
      expect(result.platform).toBe('workday');
    });

    test('parses valid Workday URL (workday.com)', () => {
      const result = parseJobUrl('https://company.workday.com/en-US/job/12345');
      expect(result.isValid).toBe(true);
      expect(result.platform).toBe('workday');
    });

    test('parses valid Ashby URL', () => {
      const result = parseJobUrl('https://jobs.ashbyhq.com/company/job-id');
      expect(result.isValid).toBe(true);
      expect(result.platform).toBe('ashby');
    });

    test('rejects invalid URL format', () => {
      const result = parseJobUrl('not-a-url');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Invalid URL format');
    });

    test('rejects non-HTTP/HTTPS protocols', () => {
      const result = parseJobUrl('ftp://boards.greenhouse.io/company/jobs/12345');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('URL must use HTTP or HTTPS protocol');
    });

    test('falls back to generic for unknown platform', () => {
      const result = parseJobUrl('https://example.com/jobs/12345');
      expect(result.isValid).toBe(true);
      expect(result.platform).toBe('generic');
    });

    test('handles URLs with query parameters', () => {
      const result = parseJobUrl('https://boards.greenhouse.io/company/jobs/12345?gh_jid=123&source=linkedin');
      expect(result.isValid).toBe(true);
      expect(result.platform).toBe('greenhouse');
    });

    test('handles URLs with fragments', () => {
      const result = parseJobUrl('https://boards.greenhouse.io/company/jobs/12345#apply');
      expect(result.isValid).toBe(true);
      expect(result.platform).toBe('greenhouse');
    });

    test('handles HTTP URLs (not just HTTPS)', () => {
      const result = parseJobUrl('http://boards.greenhouse.io/company/jobs/12345');
      expect(result.isValid).toBe(true);
      expect(result.platform).toBe('greenhouse');
    });

    test('handles www prefix in LinkedIn URL', () => {
      const result = parseJobUrl('https://www.linkedin.com/jobs/view/12345');
      expect(result.isValid).toBe(true);
      expect(result.platform).toBe('linkedin');
    });

    test('returns URL in result', () => {
      const url = 'https://boards.greenhouse.io/company/jobs/12345';
      const result = parseJobUrl(url);
      expect(result.url).toBe(url);
    });
  });

  describe('detectPlatform', () => {
    test('detects Greenhouse', () => {
      expect(detectPlatform('https://boards.greenhouse.io/test')).toBe('greenhouse');
    });

    test('detects LinkedIn', () => {
      expect(detectPlatform('https://linkedin.com/jobs/view/123')).toBe('linkedin');
    });

    test('detects Lever', () => {
      expect(detectPlatform('https://jobs.lever.co/company')).toBe('lever');
    });

    test('detects Jobvite', () => {
      expect(detectPlatform('https://jobs.jobvite.com/company')).toBe('jobvite');
    });

    test('detects SmartRecruiters', () => {
      expect(detectPlatform('https://jobs.smartrecruiters.com/Company')).toBe('smartrecruiters');
    });

    test('detects Pinpoint', () => {
      expect(detectPlatform('https://acme.pinpointhq.com/jobs')).toBe('pinpoint');
    });

    test('detects Teamtailor', () => {
      expect(detectPlatform('https://acme.teamtailor.com/jobs')).toBe('teamtailor');
    });

    test('detects Workday (myworkdayjobs)', () => {
      expect(detectPlatform('https://acme.myworkdayjobs.com/jobs')).toBe('workday');
    });

    test('detects Workday (workday.com/job)', () => {
      expect(detectPlatform('https://acme.workday.com/en-US/job/123')).toBe('workday');
    });

    test('detects Ashby', () => {
      expect(detectPlatform('https://jobs.ashbyhq.com/company')).toBe('ashby');
    });

    test('falls back to generic for unknown platform', () => {
      expect(detectPlatform('https://example.com/jobs')).toBe('generic');
    });

    test('falls back to generic for empty string', () => {
      expect(detectPlatform('')).toBe('generic');
    });
  });

  describe('isValidJobUrl', () => {
    test('returns true for valid URL', () => {
      expect(isValidJobUrl('https://boards.greenhouse.io/company/jobs/12345')).toBe(true);
    });

    test('returns false for invalid URL', () => {
      expect(isValidJobUrl('not-a-url')).toBe(false);
    });

    test('returns true for unknown platform (generic fallback)', () => {
      expect(isValidJobUrl('https://example.com/jobs')).toBe(true);
    });
  });

  describe('getSupportedPlatforms', () => {
    test('returns all supported platforms', () => {
      const platforms = getSupportedPlatforms();
      expect(platforms).toContain('greenhouse');
      expect(platforms).toContain('linkedin');
      expect(platforms).toContain('lever');
      expect(platforms).toContain('jobvite');
      expect(platforms).toContain('smartrecruiters');
      expect(platforms).toContain('pinpoint');
      expect(platforms).toContain('teamtailor');
      expect(platforms).toContain('workday');
      expect(platforms).toContain('ashby');
      expect(platforms).toContain('generic');
      expect(platforms).toHaveLength(10);
    });
  });

  describe('getPlatformExamples', () => {
    test('returns examples for all platforms', () => {
      const examples = getPlatformExamples();
      expect(examples.greenhouse).toContain('greenhouse.io');
      expect(examples.linkedin).toContain('linkedin.com');
      expect(examples.lever).toContain('lever.co');
      expect(examples.jobvite).toContain('jobvite.com');
      expect(examples.smartrecruiters).toContain('smartrecruiters.com');
      expect(examples.pinpoint).toContain('pinpointhq.com');
      expect(examples.teamtailor).toContain('teamtailor.com');
      expect(examples.workday).toContain('myworkdayjobs.com');
      expect(examples.ashby).toContain('ashbyhq.com');
    });

    test('all examples are valid URLs', () => {
      const examples = getPlatformExamples();
      for (const [platform, url] of Object.entries(examples)) {
        const result = parseJobUrl(url);
        expect(result.isValid).toBe(true);
        expect(result.platform).toBe(platform as Platform);
      }
    });
  });

  describe('validateUrls', () => {
    test('separates valid and invalid URLs', () => {
      const urls = [
        'https://boards.greenhouse.io/company/jobs/12345',
        'not-a-url',
        'https://jobs.lever.co/company/job-id',
        'https://example.com/jobs',
      ];

      const result = validateUrls(urls);

      expect(result.valid).toHaveLength(3);
      expect(result.invalid).toHaveLength(1);

      expect(result.valid[0].platform).toBe('greenhouse');
      expect(result.valid[1].platform).toBe('lever');
      expect(result.valid[2].platform).toBe('generic');

      expect(result.invalid[0].error).toBe('Invalid URL format');
    });

    test('handles empty array', () => {
      const result = validateUrls([]);
      expect(result.valid).toHaveLength(0);
      expect(result.invalid).toHaveLength(0);
    });

    test('handles all valid URLs', () => {
      const urls = [
        'https://boards.greenhouse.io/company/jobs/12345',
        'https://jobs.lever.co/company/job-id',
      ];

      const result = validateUrls(urls);
      expect(result.valid).toHaveLength(2);
      expect(result.invalid).toHaveLength(0);
    });

    test('handles mix of invalid and generic URLs', () => {
      const urls = ['not-a-url', 'https://example.com/jobs'];

      const result = validateUrls(urls);
      expect(result.valid).toHaveLength(1);
      expect(result.valid[0].platform).toBe('generic');
      expect(result.invalid).toHaveLength(1);
    });
  });
});
