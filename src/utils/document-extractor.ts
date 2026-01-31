/**
 * Document text extraction utilities
 * Supports PDF, Markdown, and plain text files
 */

import { existsSync } from 'fs';
import { readFile } from 'fs/promises';
import { extname, resolve } from 'path';

/**
 * Supported file extensions for document import
 */
export const SUPPORTED_EXTENSIONS = ['.pdf', '.md', '.markdown', '.txt'] as const;

export type SupportedExtension = (typeof SUPPORTED_EXTENSIONS)[number];

/**
 * Result of document extraction
 */
export interface ExtractionResult {
  success: boolean;
  content?: string;
  filePath?: string;
  fileType?: string;
  error?: string;
}

/**
 * Check if a file extension is supported
 */
export function isSupportedExtension(filePath: string): boolean {
  const ext = extname(filePath).toLowerCase();
  return SUPPORTED_EXTENSIONS.includes(ext as SupportedExtension);
}

/**
 * Extract text content from a file (PDF, MD, or TXT)
 *
 * @param filePath - Path to the file
 * @returns Extraction result with content or error
 */
export async function extractTextFromFile(filePath: string): Promise<ExtractionResult> {
  const absolutePath = resolve(filePath);

  // Check file exists
  if (!existsSync(absolutePath)) {
    return {
      success: false,
      error: `File not found: ${filePath}`,
    };
  }

  const ext = extname(absolutePath).toLowerCase();

  // Check supported extension
  if (!isSupportedExtension(absolutePath)) {
    return {
      success: false,
      error: `Unsupported file type: ${ext}. Supported: ${SUPPORTED_EXTENSIONS.join(', ')}`,
    };
  }

  try {
    let content: string;

    if (ext === '.pdf') {
      content = await extractTextFromPdf(absolutePath);
    } else {
      // MD, TXT - read as text
      content = await readFile(absolutePath, 'utf-8');
    }

    // Validate content
    if (!content || content.trim().length === 0) {
      return {
        success: false,
        error: 'File is empty or could not extract text',
      };
    }

    return {
      success: true,
      content: content.trim(),
      filePath: absolutePath,
      fileType: ext.replace('.', ''),
    };
  } catch (error) {
    return {
      success: false,
      error: `Failed to extract text: ${error instanceof Error ? error.message : 'Unknown error'}`,
    };
  }
}

async function extractTextFromPdf(filePath: string): Promise<string> {
  // Import from lib path directly to avoid pdf-parse's test file loading on import
  const pdf = (await import('pdf-parse/lib/pdf-parse.js')).default;
  const fileBuffer = await readFile(filePath);
  const result = await pdf(fileBuffer);
  return result.text;
}

/**
 * Validate that a file path is a valid document
 */
export function validateDocumentPath(filePath: string): { valid: boolean; error?: string } {
  if (!filePath || filePath.trim().length === 0) {
    return { valid: false, error: 'File path is required' };
  }

  const absolutePath = resolve(filePath);

  if (!existsSync(absolutePath)) {
    return { valid: false, error: `File not found: ${filePath}` };
  }

  if (!isSupportedExtension(absolutePath)) {
    const ext = extname(absolutePath).toLowerCase();
    return {
      valid: false,
      error: `Unsupported file type: ${ext}. Supported: ${SUPPORTED_EXTENSIONS.join(', ')}`,
    };
  }

  return { valid: true };
}

/**
 * Get a user-friendly description of supported formats
 */
export function getSupportedFormatsDescription(): string {
  return 'PDF (.pdf), Markdown (.md), or plain text (.txt)';
}
