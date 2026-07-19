import { browserManager } from '../core/browser-manager';
import { logger } from '../utils/logger';

export interface LinkedInExperienceItem {
  title: string;
  company: string;
  location?: string;
  startDate?: string;
  endDate?: string;
  description?: string;
}

export interface LinkedInEducationItem {
  institution: string;
  degree?: string;
  field?: string;
  startDate?: string;
  endDate?: string;
}

export interface ExtractedLinkedInProfile {
  name?: string;
  headline?: string;
  location?: string;
  about?: string;
  experience: LinkedInExperienceItem[];
  education: LinkedInEducationItem[];
  skills: string[];
  rawText?: string;
}

/**
 * Normalizes a LinkedIn profile URL
 */
export function normalizeLinkedInUrl(url: string): string {
  try {
    const parsed = new URL(url);
    if (!parsed.hostname.includes('linkedin.com')) return url;
    parsed.search = '';
    parsed.hash = '';
    return parsed.toString().replace(/\/+$/, '');
  } catch {
    return url;
  }
}

/**
 * Scrapes detailed LinkedIn profile information (headline, about, experience, education, skills).
 * Uses Playwright browser session if available, with HTTP fallback.
 */
export async function scrapeLinkedInProfile(
  url: string
): Promise<ExtractedLinkedInProfile> {
  const cleanUrl = normalizeLinkedInUrl(url);

  // 1. Try Playwright browser scraping
  try {
    const session = await browserManager.createSession('linkedin', cleanUrl);
    try {
      const { page } = session;
      await page.goto(cleanUrl, { waitUntil: 'domcontentloaded', timeout: 10000 });
      await page.waitForTimeout(1500);

      // Check if hit login redirect or authwall
      const currentUrl = page.url();
      if (!currentUrl.includes('/authwall') && !currentUrl.includes('/login')) {
        // Scroll down to load Experience & Education lazy sections
        await page.evaluate(() => window.scrollBy(0, 1200));
        await page.waitForTimeout(1000);
        await page.evaluate(() => window.scrollBy(0, 1500));
        await page.waitForTimeout(1000);

        const extracted = await page.evaluate(() => {
          // Name
          const nameEl = document.querySelector('h1, .text-heading-xlarge');
          const name = nameEl?.textContent?.trim();

          // Headline
          const headlineEl = document.querySelector('.text-body-medium, [data-generated-suggestion-target]');
          const headline = headlineEl?.textContent?.trim();

          // Location
          const locationEl = document.querySelector('.text-body-small.inline, .top-card__subline-item');
          const location = locationEl?.textContent?.trim();

          // About / Summary
          const aboutSection = document.querySelector('section:has(#about), .pv-about-section');
          const aboutText = aboutSection?.textContent?.replace(/^About\s*/i, '').trim();

          // Experience items
          const expItems: Array<{
            title: string;
            company: string;
            location?: string;
            startDate?: string;
            endDate?: string;
            description?: string;
          }> = [];

          const expNodes = document.querySelectorAll(
            'section:has(#experience) li.artdeco-list__item, section[id*="experience"] li'
          );

          expNodes.forEach((node) => {
            const lines = (node.textContent || '')
              .split('\n')
              .map((s) => s.trim())
              .filter(Boolean);

            if (lines.length >= 2) {
              const title = lines[0];
              const company = lines[1];
              let dateStr: string | undefined;
              let desc: string | undefined;

              for (let i = 2; i < lines.length; i++) {
                if (/\d{4}|present|jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec/i.test(lines[i])) {
                  dateStr = lines[i];
                  break;
                }
              }

              if (lines.length > 3) {
                desc = lines.slice(3).join(' ');
              }

              let startDate: string | undefined;
              let endDate: string | undefined;
              if (dateStr) {
                const parts = dateStr.split(/–|-|to/i);
                if (parts[0]) startDate = parts[0].trim();
                if (parts[1]) endDate = parts[1].trim();
              }

              expItems.push({
                title,
                company,
                startDate,
                endDate,
                description: desc,
              });
            }
          });

          // Education items
          const eduItems: Array<{
            institution: string;
            degree?: string;
            field?: string;
            startDate?: string;
            endDate?: string;
          }> = [];

          const eduNodes = document.querySelectorAll(
            'section:has(#education) li.artdeco-list__item, section[id*="education"] li'
          );

          eduNodes.forEach((node) => {
            const lines = (node.textContent || '')
              .split('\n')
              .map((s) => s.trim())
              .filter(Boolean);

            if (lines.length >= 1) {
              const institution = lines[0];
              const degree = lines[1];

              eduItems.push({
                institution,
                degree,
              });
            }
          });

          // Skills
          const skillNodes = document.querySelectorAll(
            'section:has(#skills) li span[aria-hidden="true"]'
          );
          const skills = Array.from(skillNodes)
            .map((s) => s.textContent?.trim() || '')
            .filter((s) => s && s.length < 50);

          const rawText = document.body.innerText.slice(0, 4000);

          return {
            name,
            headline,
            location,
            about: aboutText,
            experience: expItems,
            education: eduItems,
            skills,
            rawText,
          };
        });

        if (extracted.name || extracted.experience.length > 0 || extracted.headline) {
          logger.info(`Successfully scraped LinkedIn profile via browser for ${cleanUrl}`, {}, 'api');
          return extracted;
        }
      }
    } finally {
      await session.release();
    }
  } catch (err) {
    logger.debug(`Playwright browser scraping skipped for ${cleanUrl}: ${err}`, {}, 'api');
  }

  // 2. HTTP Fetch fallback with JSON-LD & meta tag parsing
  return await fetchLinkedInPublicProfile(cleanUrl);
}

/**
 * Fallback HTTP parser for public LinkedIn profiles
 */
async function fetchLinkedInPublicProfile(url: string): Promise<ExtractedLinkedInProfile> {
  try {
    const res = await fetch(url, {
      headers: {
        'User-Agent':
          'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
        Accept: 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
        'Accept-Language': 'en-US,en;q=0.9',
      },
      signal: AbortSignal.timeout(5000),
    });

    if (!res.ok) {
      return { experience: [], education: [], skills: [] };
    }

    const html = await res.text();

    let name: string | undefined;
    let headline: string | undefined;
    let location: string | undefined;
    const experience: LinkedInExperienceItem[] = [];
    const education: LinkedInEducationItem[] = [];

    // Parse JSON-LD script tags
    const jsonLdMatches = html.match(/<script[^>]*type=["']application\/ld\+json["'][^>]*>([\s\S]*?)<\/script>/gi);
    if (jsonLdMatches) {
      for (const match of jsonLdMatches) {
        try {
          const jsonText = match.replace(/<script[^>]*>/i, '').replace(/<\/script>/i, '').trim();
          const data = JSON.parse(jsonText);

          if (data['@type'] === 'Person') {
            name = name || data.name;
            headline = headline || data.jobTitle;
            if (data.address && typeof data.address === 'object') {
              location = location || `${data.address.addressLocality || ''} ${data.address.addressCountry || ''}`.trim();
            }

            if (Array.isArray(data.worksFor)) {
              data.worksFor.forEach((org: { name?: string; jobTitle?: string }) => {
                if (org.name) {
                  experience.push({
                    company: org.name,
                    title: org.jobTitle || headline || 'Role',
                  });
                }
              });
            }

            if (Array.isArray(data.alumniOf)) {
              data.alumniOf.forEach((org: { name?: string }) => {
                if (org.name) {
                  education.push({ institution: org.name });
                }
              });
            }
          }
        } catch {
          // Ignore invalid JSON-LD snippet
        }
      }
    }

    // Parse OpenGraph & Meta Tags
    if (!name || !headline) {
      const ogTitle = html.match(/<meta[^>]*property=["']og:title["'][^>]*content=["']([\s\S]*?)["']/i)?.[1];
      if (ogTitle) {
        const parts = ogTitle.split('-').map((s) => s.trim());
        if (parts[0]) name = name || parts[0];
        if (parts[1]) headline = headline || parts[1];
      }

      const ogDesc = html.match(/<meta[^>]*property=["']og:description["'][^>]*content=["']([\s\S]*?)["']/i)?.[1] ||
        html.match(/<meta[^>]*name=["']description["'][^>]*content=["']([\s\S]*?)["']/i)?.[1];

      if (ogDesc && !headline) {
        headline = ogDesc.slice(0, 200);
      }
    }

    // Clean html to body text for snippet fallback
    const bodyText = html
      .replace(/<script[\s\S]*?<\/script>/gi, '')
      .replace(/<style[\s\S]*?<\/style>/gi, '')
      .replace(/<[^>]+>/g, ' ')
      .replace(/\s+/g, ' ')
      .trim()
      .slice(0, 2000);

    return {
      name,
      headline,
      location,
      experience,
      education,
      skills: [],
      rawText: bodyText,
    };
  } catch (err) {
    logger.warn(`Public HTTP fetch failed for LinkedIn ${url}: ${err}`, {}, 'api');
    return { experience: [], education: [], skills: [] };
  }
}

/**
 * Formats an ExtractedLinkedInProfile into clean Markdown text for prompt context
 */
export function formatLinkedInProfileContext(profile: ExtractedLinkedInProfile, url: string): string {
  const parts: string[] = [`LinkedIn Profile (${url}):`];

  if (profile.name) parts.push(`- Full Name: ${profile.name}`);
  if (profile.headline) parts.push(`- Headline / Current Role: ${profile.headline}`);
  if (profile.location) parts.push(`- Location: ${profile.location}`);
  if (profile.about) parts.push(`- About / Summary: ${profile.about}`);

  if (profile.experience.length > 0) {
    parts.push('- Work Experience:');
    profile.experience.forEach((exp) => {
      const dates = exp.startDate ? ` (${exp.startDate}${exp.endDate ? ` - ${exp.endDate}` : ''})` : '';
      const loc = exp.location ? ` - ${exp.location}` : '';
      const desc = exp.description ? `\n  Details: ${exp.description}` : '';
      parts.push(`  * ${exp.title} at ${exp.company}${dates}${loc}${desc}`);
    });
  }

  if (profile.education.length > 0) {
    parts.push('- Education:');
    profile.education.forEach((edu) => {
      const degree = edu.degree ? ` (${edu.degree}${edu.field ? ` in ${edu.field}` : ''})` : '';
      parts.push(`  * ${edu.institution}${degree}`);
    });
  }

  if (profile.skills.length > 0) {
    parts.push(`- Skills: ${profile.skills.join(', ')}`);
  }

  if (parts.length === 1 && profile.rawText) {
    parts.push(`- Snippet: ${profile.rawText.slice(0, 1000)}`);
  }

  return parts.join('\n');
}
