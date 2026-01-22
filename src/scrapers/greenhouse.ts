import { BaseScraper, type SubmissionOptions, type SubmissionResult } from './base';
import type { JobData, CustomQuestion, Platform, Profile } from '../types';
import { FormFiller } from '../core/form-filler';

export class GreenhouseScraper extends BaseScraper {
  platform: Platform = 'greenhouse';

  protected async waitForContent(): Promise<void> {
    if (!this.page) return;
    await this.page.waitForSelector('#app_body, .app-body, [data-mapped="true"], h1', {
      timeout: 10000,
    }).catch(() => {});
    // Extra wait for JS rendering
    await this.page.waitForTimeout(2000);
  }

  // ============ Greenhouse-specific Form Submission ============

  /**
   * Greenhouse typically has the application form on the same page as the job posting,
   * or accessible via an "Apply" button that scrolls to or reveals the form.
   */
  protected override async navigateToApplicationForm(): Promise<void> {
    if (!this.page) return;

    // Greenhouse forms are usually embedded on the page
    // Look for "Apply for this job" or similar buttons
    const applyButtonSelectors = [
      '#apply_button',
      'a[href*="#app"]',
      'button:has-text("Apply")',
      'a:has-text("Apply for this job")',
      '.application-button',
      '[data-test="apply-button"]',
    ];

    for (const selector of applyButtonSelectors) {
      try {
        const button = await this.page.$(selector);
        if (button) {
          const isVisible = await button.isVisible();
          if (isVisible) {
            await this.humanDelay(true);
            await button.click();
            await this.page.waitForTimeout(1000);
            return;
          }
        }
      } catch {
        continue;
      }
    }

    // If no apply button, scroll to the form (it might already be visible)
    await this.page.evaluate(() => {
      const form = document.querySelector('#application_form, form[id*="application"], form[class*="application"]');
      if (form) {
        form.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }
    });
    await this.humanDelay(true);
  }

  protected override async waitForApplicationForm(): Promise<void> {
    if (!this.page) return;

    const formSelectors = [
      '#application_form',
      '#application',
      'form[id*="application"]',
      '.application-form',
      '#main_fields',
    ];

    for (const selector of formSelectors) {
      try {
        await this.page.waitForSelector(selector, { timeout: 5000 });
        return;
      } catch {
        continue;
      }
    }
  }

  override async submitApplication(url: string, options: SubmissionOptions): Promise<SubmissionResult> {
    const errors: string[] = [];

    try {
      await this.initialize();
      if (!this.page) throw new Error('Browser not initialized');

      // Navigate to job posting
      await this.humanDelay();
      await this.page.goto(url, { waitUntil: 'networkidle' });
      await this.humanDelay(true);
      await this.humanScroll();

      // Navigate to application form
      await this.navigateToApplicationForm();
      await this.waitForApplicationForm();

      // Create form filler
      const filler = new FormFiller(this.page, options.profile, options.jobData, {
        resumePath: options.resumePath,
        coverLetterPath: options.coverLetterPath,
        answeredQuestions: options.answeredQuestions,
      });

      // Fill basic fields (name, email, phone, etc.)
      await this.fillGreenhouseBasicFields(options);

      // Upload resume
      if (options.resumePath) {
        const resumeUploaded = await this.uploadGreenhouseResume(options.resumePath);
        if (!resumeUploaded) {
          errors.push('Failed to upload resume');
        }
      }

      // Upload cover letter if available
      if (options.coverLetterPath) {
        await this.uploadGreenhouseCoverLetter(options.coverLetterPath);
      }

      // Fill LinkedIn/Website fields
      await this.fillGreenhouseUrls(options);

      // Fill custom questions
      if (options.answeredQuestions && options.answeredQuestions.length > 0) {
        const questionsResult = await filler.fillCustomQuestions(options.answeredQuestions);
        if (questionsResult.errors.length > 0) {
          errors.push(...questionsResult.errors);
        }
      }

      // Handle education and work history sections if they exist
      await this.fillGreenhouseEducation(options);

      // Handle any remaining required fields (select dropdowns, radio buttons)
      await this.fillRemainingRequiredFields(options.profile);

      // Scroll through the form to ensure all fields are visible
      await this.page.evaluate(() => {
        const form = document.querySelector('#application_form, form');
        if (form) form.scrollIntoView({ behavior: 'instant', block: 'end' });
      });
      await this.humanDelay(true);
      await this.page.evaluate(() => window.scrollTo(0, 0));
      await this.humanDelay(true);

      // Validate before submit
      const validation = await this.validateBeforeSubmit();
      if (!validation.valid) {
        errors.push(...validation.errors);
      }

      // Don't fail on validation errors - try to submit anyway
      // Some "errors" might be warnings

      // Submit
      const submitted = await this.clickGreenhouseSubmit();
      if (!submitted) {
        return {
          success: false,
          message: 'Could not find or click submit button',
          errors,
        };
      }

      // Wait for confirmation
      const confirmation = await this.waitForGreenhouseConfirmation();

      // Take screenshot
      const { configRepository } = await import('../db/repositories/config');
      const config = configRepository.loadAppConfig();
      let screenshotPath: string | undefined;
      if (config.application.saveScreenshots) {
        const { getAutoplyDir } = await import('../db');
        const { join } = await import('path');
        screenshotPath = join(getAutoplyDir(), 'screenshots', `greenhouse_${Date.now()}.png`);
        await this.takeScreenshot(screenshotPath);
      }

      return {
        success: confirmation.success,
        message: confirmation.message,
        screenshotPath,
        errors,
      };
    } catch (error) {
      errors.push(error instanceof Error ? error.message : 'Unknown error');
      return {
        success: false,
        message: 'Greenhouse submission failed',
        errors,
      };
    } finally {
      await this.cleanup();
    }
  }

  private async fillGreenhouseBasicFields(options: SubmissionOptions): Promise<void> {
    if (!this.page) return;

    const { profile } = options;

    // First name
    await this.fillInputBySelector('#first_name, input[name="job_application[first_name]"]', profile.name.split(' ')[0]);

    // Last name
    const lastName = profile.name.split(' ').slice(1).join(' ');
    await this.fillInputBySelector('#last_name, input[name="job_application[last_name]"]', lastName);

    // Email
    await this.fillInputBySelector('#email, input[name="job_application[email]"]', profile.email);

    // Phone
    if (profile.phone) {
      await this.fillInputBySelector('#phone, input[name="job_application[phone]"]', profile.phone);
    }

    // Location/Address - Greenhouse often uses autocomplete
    if (profile.location) {
      await this.fillLocationField(profile.location);
    }
  }

  private async fillLocationField(location: string): Promise<boolean> {
    if (!this.page) return false;

    try {
      // Try various location selectors
      const locationSelectors = [
        '#job_application_location',
        'input[name*="location"]',
        'input[id*="location"]',
        'input[placeholder*="City"]',
        'input[placeholder*="Location"]',
        'input[autocomplete="address-level2"]',
      ];

      for (const selector of locationSelectors) {
        const input = await this.page.$(selector);
        if (input && await input.isVisible()) {
          await input.click();
          await input.fill(location);
          await this.humanDelay(true);

          // Wait for autocomplete dropdown and select first option if available
          try {
            await this.page.waitForSelector('[class*="autocomplete"] li, [class*="suggestion"], [role="option"]', { timeout: 2000 });
            const firstOption = await this.page.$('[class*="autocomplete"] li:first-child, [class*="suggestion"]:first-child, [role="option"]:first-child');
            if (firstOption) {
              await firstOption.click();
              await this.humanDelay(true);
            }
          } catch {
            // No autocomplete, just press Enter to confirm
            await input.press('Tab');
          }

          return true;
        }
      }

      return false;
    } catch {
      return false;
    }
  }

  private async fillInputBySelector(selector: string, value: string): Promise<boolean> {
    if (!this.page || !value) return false;

    try {
      const input = await this.page.$(selector);
      if (input) {
        await input.click();
        await input.fill(value);
        await this.humanDelay(true);
        return true;
      }
      return false;
    } catch {
      return false;
    }
  }

  private async uploadGreenhouseResume(resumePath: string): Promise<boolean> {
    if (!this.page) return false;

    try {
      // First try: Find any file input and check if it's for resume
      const allFileInputs = await this.page.$$('input[type="file"]');
      for (const input of allFileInputs) {
        // Check if this input is related to resume by looking at parent/sibling elements
        const parent = await input.evaluateHandle(el => el.closest('[class*="resume"], [id*="resume"], [data-field*="resume"], .field'));
        const parentEl = parent.asElement();
        if (parentEl) {
          const text = await parentEl.textContent();
          if (text?.toLowerCase().includes('resume') || text?.toLowerCase().includes('cv')) {
            await input.setInputFiles(resumePath);
            await this.page.waitForTimeout(2000);
            await this.humanDelay(true);
            return true;
          }
        }
      }

      // Second try: Use specific selectors
      const resumeSelectors = [
        '#resume_upload input[type="file"]',
        '#s3_upload_for_resume input[type="file"]',
        'input[type="file"][name*="resume"]',
        '#resume input[type="file"]',
        '[data-field="resume"] input[type="file"]',
        '.field:has-text("Resume") input[type="file"]',
        '.field:has-text("CV") input[type="file"]',
      ];

      for (const selector of resumeSelectors) {
        try {
          const fileInput = await this.page.$(selector);
          if (fileInput) {
            await fileInput.setInputFiles(resumePath);
            await this.page.waitForTimeout(2000);
            await this.humanDelay(true);
            return true;
          }
        } catch {
          continue;
        }
      }

      // Third try: Click on upload area and use file chooser
      const uploadAreas = await this.page.$$('[class*="resume"] [class*="upload"], #resume_upload, .attach-or-paste, button:has-text("Attach"), button:has-text("Upload")');
      for (const area of uploadAreas) {
        try {
          const [fileChooser] = await Promise.all([
            this.page.waitForEvent('filechooser', { timeout: 5000 }),
            area.click(),
          ]);
          await fileChooser.setFiles(resumePath);
          await this.page.waitForTimeout(2000);
          return true;
        } catch {
          continue;
        }
      }

      // Fourth try: Just use the first file input on the page
      if (allFileInputs.length > 0) {
        await allFileInputs[0].setInputFiles(resumePath);
        await this.page.waitForTimeout(2000);
        return true;
      }

      return false;
    } catch {
      return false;
    }
  }

  private async uploadGreenhouseCoverLetter(coverLetterPath: string): Promise<boolean> {
    if (!this.page) return false;

    try {
      const coverLetterSelectors = [
        '#cover_letter_upload input[type="file"]',
        '#s3_upload_for_cover_letter input[type="file"]',
        'input[type="file"][name*="cover"]',
        '[data-field="cover_letter"] input[type="file"]',
      ];

      for (const selector of coverLetterSelectors) {
        const fileInput = await this.page.$(selector);
        if (fileInput) {
          await fileInput.setInputFiles(coverLetterPath);
          await this.page.waitForTimeout(2000);
          await this.humanDelay(true);
          return true;
        }
      }

      return false;
    } catch {
      return false;
    }
  }

  private async fillGreenhouseUrls(options: SubmissionOptions): Promise<void> {
    if (!this.page) return;

    const { profile } = options;

    // LinkedIn
    if (profile.linkedin_url) {
      await this.fillInputBySelector(
        'input[name*="linkedin"], input[id*="linkedin"], input[placeholder*="LinkedIn"]',
        profile.linkedin_url
      );
    }

    // GitHub
    if (profile.github_url) {
      await this.fillInputBySelector(
        'input[name*="github"], input[id*="github"], input[placeholder*="GitHub"]',
        profile.github_url
      );
    }

    // Portfolio/Website
    if (profile.portfolio_url) {
      await this.fillInputBySelector(
        'input[name*="website"], input[name*="portfolio"], input[id*="website"], input[placeholder*="Website"]',
        profile.portfolio_url
      );
    }
  }

  private async fillGreenhouseEducation(options: SubmissionOptions): Promise<void> {
    if (!this.page) return;

    const { profile } = options;
    if (!profile.education || profile.education.length === 0) return;

    // Greenhouse sometimes has education fields
    const education = profile.education[0];

    await this.fillInputBySelector(
      'input[name*="school"], input[name*="institution"], input[id*="school"]',
      education.institution
    );

    await this.fillInputBySelector(
      'input[name*="degree"], input[id*="degree"]',
      education.degree
    );

    if (education.field) {
      await this.fillInputBySelector(
        'input[name*="field"], input[name*="major"], input[id*="discipline"]',
        education.field
      );
    }
  }

  /**
   * Fill React Select custom dropdowns commonly used by Greenhouse.
   * These are div-based dropdowns, not native <select> elements.
   */
  private async fillReactSelectDropdowns(
    questionPatterns: Array<{ pattern: RegExp; answer: string }>
  ): Promise<void> {
    if (!this.page) return;

    // Find all React Select containers (they have class "select" with nested structure)
    const selectContainers = await this.page.$$('div.select:has(.select__control)');

    for (const container of selectContainers) {
      try {
        // Check if already has a value (not showing placeholder)
        const hasValue = await container.$('.select__single-value');
        if (hasValue) continue; // Already filled

        // Get the label text
        const labelText = await container.$eval(
          '.select__label, label',
          (el) => el.textContent?.trim() || ''
        ).catch(() => '');

        if (!labelText) continue;

        // Find matching pattern
        let answerToSelect: string | null = null;
        for (const { pattern, answer } of questionPatterns) {
          if (pattern.test(labelText)) {
            answerToSelect = answer;
            break;
          }
        }

        if (!answerToSelect) continue;

        // Click to open the dropdown
        const control = await container.$('.select__control');
        if (!control) continue;

        await control.click();
        await this.humanDelay(true);

        // Wait for menu to appear
        await this.page.waitForSelector('.select__menu', { timeout: 3000 }).catch(() => {});

        // Find and click the matching option
        const options = await this.page.$$('.select__option');
        let matched = false;

        for (const option of options) {
          const optionText = await option.textContent();
          if (!optionText) continue;

          const optTextLower = optionText.toLowerCase().trim();
          const answerLower = answerToSelect.toLowerCase();

          // Check for various match types
          if (
            optTextLower === answerLower ||
            optTextLower.includes(answerLower) ||
            answerLower.includes(optTextLower) ||
            (answerLower === 'yes' && /^(yes|true|y)$/i.test(optTextLower)) ||
            (answerLower === 'no' && /^(no|false|n)$/i.test(optTextLower)) ||
            (answerLower.includes('decline') && optTextLower.includes('decline')) ||
            (answerLower.includes('prefer not') && optTextLower.includes('prefer not')) ||
            (answerLower.includes('don\'t wish') && optTextLower.includes('don\'t wish')) ||
            (answerLower.includes('i am not') && optTextLower.includes('i am not')) ||
            (answerLower.includes('acknowledge') && optTextLower.includes('acknowledge'))
          ) {
            await option.click();
            await this.humanDelay(true);
            matched = true;
            break;
          }
        }

        // If no match found, try to click first non-placeholder option
        if (!matched && options.length > 0) {
          // Close menu first
          await this.page.keyboard.press('Escape');
          await this.humanDelay(true);
        }
      } catch {
        continue;
      }
    }
  }

  /**
   * Fill any remaining required fields that weren't handled by the standard flow.
   * This catches common questions like relocation, work authorization, etc.
   * Handles both native <select> elements and React Select custom dropdowns.
   */
  private async fillRemainingRequiredFields(profile: Profile): Promise<void> {
    if (!this.page) return;

    // Common question patterns and their default answers
    const questionPatterns = [
      { pattern: /relocation|relocate|willing.*move|open.*move/i, answer: 'Yes' },
      { pattern: /open.*working.*in-person|work.*office|hybrid/i, answer: 'Yes' },
      { pattern: /authorized.*work|legally.*work|eligible.*work|right.*work/i, answer: 'Yes' },
      { pattern: /sponsor|visa.*sponsor|require.*sponsor/i, answer: 'No' },
      { pattern: /18.*years|legal.*age|at.*least.*18/i, answer: 'Yes' },
      { pattern: /background.*check|consent.*check/i, answer: 'Yes' },
      { pattern: /how.*hear|where.*find|referral|source/i, answer: 'Job Board' },
      { pattern: /gender|pronouns/i, answer: 'Decline to self identify' },
      { pattern: /veteran|military/i, answer: 'I am not' },
      { pattern: /disability|disabled/i, answer: 'I don\'t wish to answer' },
      { pattern: /race|ethnicity|hispanic|latino/i, answer: 'Decline to self identify' },
      { pattern: /ai.*policy|acknowledge|agree.*policy|consent.*ai/i, answer: 'I acknowledge' },
      { pattern: /interviewed.*before|applied.*before/i, answer: 'No' },
      { pattern: /built.*developer.*tools|maintained.*developer.*tools/i, answer: 'Yes' },
      { pattern: /experience.*typescript|typescript.*production/i, answer: 'Yes' },
      { pattern: /react.*hooks|component.*architecture/i, answer: 'Yes' },
      { pattern: /cli.*tools|ide.*extension|plugin/i, answer: 'Yes' },
      { pattern: /ai.*ml.*component|machine.*learning/i, answer: 'Yes' },
    ];

    // Handle React Select custom dropdowns (used by Greenhouse)
    await this.fillReactSelectDropdowns(questionPatterns);

    // Also handle native select elements as fallback
    const selects = await this.page.$$('select[required], select');
    for (const select of selects) {
      try {
        const value = await select.inputValue();
        if (value) continue; // Already filled

        // Get the label text for this select - try multiple approaches
        const labelText = await this.page.evaluate((el) => {
          const id = el.id;
          const name = el.getAttribute('name');

          // Try to find label by for attribute
          let label = id ? document.querySelector(`label[for="${id}"]`) : null;
          if (label) return label.textContent?.trim() || '';

          // Try to find label as parent/sibling
          const fieldContainer = el.closest('.field, .form-group, fieldset, [class*="field"], [class*="question"]');
          if (fieldContainer) {
            label = fieldContainer.querySelector('label, .field-label, legend');
            if (label) return label.textContent?.trim() || '';
            // Sometimes the text is directly in the container
            const containerText = fieldContainer.textContent?.trim() || '';
            // Remove option texts from container text
            const options = Array.from(el.querySelectorAll('option')).map(o => o.textContent?.trim() || '');
            let cleanText = containerText;
            for (const opt of options) {
              cleanText = cleanText.replace(opt, '');
            }
            if (cleanText.trim()) return cleanText.trim();
          }

          // Try previous sibling
          let prevSibling = el.previousElementSibling;
          while (prevSibling) {
            if (prevSibling.tagName === 'LABEL' || prevSibling.classList?.contains('field-label')) {
              return prevSibling.textContent?.trim() || '';
            }
            prevSibling = prevSibling.previousElementSibling;
          }

          // Try aria-label
          return el.getAttribute('aria-label') || el.getAttribute('placeholder') || name || '';
        }, select);

        if (!labelText) continue;

        // Check if this matches any of our patterns
        for (const { pattern, answer } of questionPatterns) {
          if (pattern.test(labelText)) {
            // Get available options
            const options = await select.$$eval('option', (opts) =>
              opts.map((o) => ({ value: o.value, text: o.textContent?.trim() || '' }))
            );

            // Find best matching option
            const matchingOption = options.find((opt) => {
              const optText = opt.text.toLowerCase().trim();
              const answerLower = answer.toLowerCase();
              // Skip empty/placeholder options
              if (!optText || optText === 'select' || optText === 'select...' || optText === '-- select --' || optText.startsWith('select')) {
                return false;
              }
              return optText === answerLower ||
                optText.includes(answerLower) ||
                answerLower.includes(optText) ||
                (answerLower === 'yes' && /^(yes|true|y|affirmative|i agree|agree)$/i.test(optText)) ||
                (answerLower === 'no' && /^(no|false|n|negative|i (do not|don't) agree)$/i.test(optText)) ||
                (answerLower === 'prefer not to say' && /prefer|decline|not (to )?disclose|not (to )?answer/i.test(optText));
            });

            if (matchingOption && matchingOption.value) {
              await select.selectOption(matchingOption.value);
              await this.humanDelay(true);
              break;
            }
          }
        }

        // If still not filled and it's a required field, select first valid option as fallback
        const isRequired = await select.evaluate((el) => el.hasAttribute('required'));
        const currentValue = await select.inputValue();
        if (isRequired && !currentValue) {
          const options = await select.$$eval('option', (opts) =>
            opts.map((o) => ({ value: o.value, text: o.textContent?.trim() || '' }))
          );
          // Find first non-empty, non-placeholder option
          const fallbackOption = options.find((opt) => {
            const text = opt.text.toLowerCase();
            return opt.value && !text.startsWith('select') && text !== '--' && text !== '';
          });
          if (fallbackOption) {
            await select.selectOption(fallbackOption.value);
            await this.humanDelay(true);
          }
        }
      } catch {
        continue;
      }
    }

    // Handle required radio buttons
    const radioGroups = await this.page.$$('fieldset:has(input[type="radio"]), .field:has(input[type="radio"])');
    for (const group of radioGroups) {
      try {
        // Check if any radio in this group is already selected
        const checkedRadio = await group.$('input[type="radio"]:checked');
        if (checkedRadio) continue; // Already answered

        // Get the question text
        const questionText = await group.$eval(
          'legend, label:first-of-type, .field-label, > label',
          (el) => el.textContent?.trim() || ''
        ).catch(() => '');

        if (!questionText) continue;

        // Check if this matches any of our patterns
        for (const { pattern, answer } of questionPatterns) {
          if (pattern.test(questionText)) {
            // Find the radio button with matching value
            const radios = await group.$$('input[type="radio"]');
            for (const radio of radios) {
              const radioValue = await radio.getAttribute('value');
              const radioLabel = await this.page.evaluate((el) => {
                const label = el.closest('label') || document.querySelector(`label[for="${el.id}"]`);
                return label?.textContent?.trim() || '';
              }, radio);

              const valueToMatch = (radioValue || radioLabel).toLowerCase();
              const answerLower = answer.toLowerCase();

              if (
                valueToMatch === answerLower ||
                valueToMatch.includes(answerLower) ||
                (answerLower === 'yes' && /^(yes|true|y|affirmative)$/i.test(valueToMatch)) ||
                (answerLower === 'no' && /^(no|false|n|negative)$/i.test(valueToMatch))
              ) {
                await radio.check();
                await this.humanDelay(true);
                break;
              }
            }
            break;
          }
        }
      } catch {
        continue;
      }
    }
  }

  private async clickGreenhouseSubmit(): Promise<boolean> {
    if (!this.page) return false;

    const submitSelectors = [
      '#submit_app',
      'button[type="submit"]',
      'input[type="submit"]',
      'button:has-text("Submit Application")',
      '#application_form button[type="submit"]',
      '.application-form button[type="submit"]',
    ];

    for (const selector of submitSelectors) {
      try {
        const button = await this.page.$(selector);
        if (button) {
          const isVisible = await button.isVisible();
          const isEnabled = await button.isEnabled();

          if (isVisible && isEnabled) {
            await this.humanDelay(true);
            await button.click();
            return true;
          }
        }
      } catch {
        continue;
      }
    }

    return false;
  }

  private async waitForGreenhouseConfirmation(): Promise<{ success: boolean; message: string }> {
    if (!this.page) return { success: false, message: 'Page not initialized' };

    try {
      // Wait for page to load after submit
      await this.page.waitForLoadState('networkidle').catch(() => {});
      await this.humanDelay();

      // Check for confirmation page
      const confirmationSelectors = [
        '.confirmation',
        '#confirmation',
        '[class*="success"]',
        '[class*="thank"]',
        'h1:has-text("Thank")',
        'h2:has-text("Thank")',
        ':has-text("application has been submitted")',
        ':has-text("received your application")',
      ];

      for (const selector of confirmationSelectors) {
        try {
          const element = await this.page.$(selector);
          if (element) {
            const isVisible = await element.isVisible();
            if (isVisible) {
              const text = await element.textContent();
              return {
                success: true,
                message: text?.trim() || 'Application submitted to Greenhouse',
              };
            }
          }
        } catch {
          continue;
        }
      }

      // Check URL for confirmation
      const currentUrl = this.page.url();
      if (currentUrl.includes('thank') || currentUrl.includes('confirmation') || currentUrl.includes('success')) {
        return { success: true, message: 'Application submitted successfully' };
      }

      // Check for error messages - be specific to avoid false positives
      const errorSelectors = [
        '.error-message',
        '.field-error',
        '.form-error',
        '.flash-error',
        '[role="alert"]',
        '.application--error',
      ];

      for (const selector of errorSelectors) {
        const errorElement = await this.page.$(selector);
        if (errorElement) {
          const isVisible = await errorElement.isVisible();
          if (isVisible) {
            const errorText = await errorElement.textContent();
            // Only treat as error if it looks like an actual error message
            if (errorText?.trim() && !errorText.includes('required field')) {
              return { success: false, message: errorText.trim() };
            }
          }
        }
      }

      return { success: true, message: 'Submission completed (no errors detected)' };
    } catch (error) {
      return {
        success: false,
        message: `Confirmation check failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
      };
    }
  }

  protected async extractJobData(url: string): Promise<JobData> {
    if (!this.page) throw new Error('Page not initialized');

    // Extract job title - try multiple selectors
    let title = await this.extractText('h1.app-title, h1[class*="job-title"], .job-title h1');
    if (!title) {
      // Fallback: get any h1 on the page
      title = await this.extractText('h1');
    }

    // Extract company name (usually in the page or URL)
    let company = await this.extractText('.company-name, [class*="company"]');
    if (!company) {
      // Try to extract from URL (boards.greenhouse.io/companyname)
      const urlMatch = url.match(/boards\.greenhouse\.io\/([^/]+)/);
      company = urlMatch ? urlMatch[1].replace(/-/g, ' ') : 'Unknown Company';
    }

    // Extract job description
    const description = await this.extractText('#content, .content, [class*="job-description"]');

    // Extract location
    const location = await this.extractText('.location, [class*="location"]');

    // Extract form fields
    const formFields = await this.extractFormFields();

    // Extract custom questions
    const customQuestions = await this.extractCustomQuestions();

    // Extract requirements and qualifications from description
    const requirements = this.extractRequirements(description);
    const qualifications = this.extractQualifications(description);

    return {
      url,
      platform: this.platform,
      title: title.trim() || 'Unknown Position',
      company: company.trim(),
      description: description.trim(),
      requirements,
      qualifications,
      location: location.trim() || undefined,
      form_fields: formFields,
      custom_questions: customQuestions,
    };
  }

  private async extractCustomQuestions(): Promise<CustomQuestion[]> {
    if (!this.page) return [];

    const questions: CustomQuestion[] = [];

    // Greenhouse uses various field patterns - try multiple selectors
    // Custom questions often appear after the main fields (name, email, resume)
    const customFields = await this.page.$$(
      '[class*="custom-question"], [data-question], ' +
      '#custom_fields .field, #custom_fields > div, ' +
      '.field:has(select), .field:has(input[type="radio"]), ' +
      '#additional_fields .field, .additional-fields .field, ' +
      '[id*="question"], [class*="question"]'
    );

    for (let i = 0; i < customFields.length; i++) {
      const field = customFields[i];
      const questionText = await field.$eval(
        'label, .field-label',
        (el) => el.textContent?.trim() ?? ''
      ).catch(() => '');

      if (!questionText) continue;

      // Determine question type
      const hasTextarea = (await field.$('textarea')) !== null;
      const hasSelect = (await field.$('select')) !== null;
      const hasRadio = (await field.$('input[type="radio"]')) !== null;
      const hasCheckbox = (await field.$('input[type="checkbox"]')) !== null;

      let type: CustomQuestion['type'] = 'text';
      let options: string[] | undefined;

      if (hasTextarea) {
        type = 'textarea';
      } else if (hasSelect) {
        type = 'select';
        options = await field.$$eval('select option', (opts) =>
          opts.map((o) => o.textContent?.trim() ?? '').filter(Boolean)
        );
      } else if (hasRadio) {
        type = 'radio';
        options = await field.$$eval('input[type="radio"]', (inputs) =>
          inputs.map((inp) => inp.getAttribute('value') ?? '').filter(Boolean)
        );
      } else if (hasCheckbox) {
        type = 'checkbox';
        options = await field.$$eval('input[type="checkbox"]', (inputs) =>
          inputs.map((inp) => inp.getAttribute('value') ?? '').filter(Boolean)
        );
      }

      const required = (await field.$('[required], .required')) !== null;

      questions.push({
        id: `question_${i}`,
        question: questionText,
        type,
        required,
        options,
      });
    }

    return questions;
  }
}
