import { describe, expect, test } from 'bun:test';
import {
  isSupportedExtension,
  validateDocumentPath,
  getSupportedFormatsDescription,
  SUPPORTED_EXTENSIONS,
} from './document-extractor';

describe('document-extractor', () => {
  describe('SUPPORTED_EXTENSIONS', () => {
    test('includes PDF', () => {
      expect(SUPPORTED_EXTENSIONS).toContain('.pdf');
    });

    test('includes markdown formats', () => {
      expect(SUPPORTED_EXTENSIONS).toContain('.md');
      expect(SUPPORTED_EXTENSIONS).toContain('.markdown');
    });

    test('includes plain text', () => {
      expect(SUPPORTED_EXTENSIONS).toContain('.txt');
    });
  });

  describe('isSupportedExtension', () => {
    test('returns true for PDF files', () => {
      expect(isSupportedExtension('resume.pdf')).toBe(true);
      expect(isSupportedExtension('/path/to/Resume.PDF')).toBe(true);
    });

    test('returns true for markdown files', () => {
      expect(isSupportedExtension('resume.md')).toBe(true);
      expect(isSupportedExtension('resume.markdown')).toBe(true);
    });

    test('returns true for text files', () => {
      expect(isSupportedExtension('resume.txt')).toBe(true);
    });

    test('returns false for unsupported extensions', () => {
      expect(isSupportedExtension('resume.docx')).toBe(false);
      expect(isSupportedExtension('resume.doc')).toBe(false);
      expect(isSupportedExtension('resume.rtf')).toBe(false);
      expect(isSupportedExtension('resume.html')).toBe(false);
    });

    test('handles files without extension', () => {
      expect(isSupportedExtension('resume')).toBe(false);
    });
  });

  describe('validateDocumentPath', () => {
    test('returns invalid for empty path', () => {
      const result = validateDocumentPath('');
      expect(result.valid).toBe(false);
      expect(result.error).toContain('required');
    });

    test('returns invalid for whitespace-only path', () => {
      const result = validateDocumentPath('   ');
      expect(result.valid).toBe(false);
    });

    test('returns invalid for non-existent file', () => {
      const result = validateDocumentPath('/nonexistent/path/resume.pdf');
      expect(result.valid).toBe(false);
      expect(result.error).toContain('not found');
    });

    test('returns invalid for unsupported extension', () => {
      // This test uses a file that might exist
      const result = validateDocumentPath('/tmp/test.docx');
      expect(result.valid).toBe(false);
      // Either file not found or unsupported extension
      expect(result.error).toBeTruthy();
    });
  });

  describe('getSupportedFormatsDescription', () => {
    test('returns human-readable description', () => {
      const description = getSupportedFormatsDescription();
      expect(description).toContain('PDF');
      expect(description).toContain('.pdf');
      expect(description).toContain('Markdown');
      expect(description).toContain('.md');
      expect(description).toContain('text');
      expect(description).toContain('.txt');
    });
  });
});
