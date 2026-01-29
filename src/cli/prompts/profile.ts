import { input, confirm, editor } from '@inquirer/prompts';
import type { Profile, Education, Preferences } from '../../types';

export async function promptForProfile(): Promise<Omit<Profile, 'id' | 'created_at' | 'updated_at'>> {
  console.log('\n📝 Let\'s set up your profile\n');

  const name = await input({
    message: 'Full name:',
    validate: (value) => (value.length > 0 ? true : 'Name is required'),
  });

  const email = await input({
    message: 'Email address:',
    validate: (value) => {
      if (!value.includes('@')) return 'Please enter a valid email';
      return true;
    },
  });

  const phone = await input({
    message: 'Phone number (optional):',
  });

  const location = await input({
    message: 'Location (City, Country):',
  });

  const linkedin_url = await input({
    message: 'LinkedIn URL (optional):',
  });

  const github_url = await input({
    message: 'GitHub URL (optional):',
  });

  const portfolio_url = await input({
    message: 'Portfolio URL (optional):',
  });

  // Skills
  const skillsInput = await input({
    message: 'Skills (comma-separated):',
  });
  const skills = skillsInput
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);

  // Education
  const education: Education[] = [];
  const addEducation = await confirm({
    message: 'Add education?',
    default: true,
  });

  if (addEducation) {
    let addMore = true;
    while (addMore) {
      const edu = await promptForEducation();
      education.push(edu);
      addMore = await confirm({
        message: 'Add another education entry?',
        default: false,
      });
    }
  }

  // Preferences
  const preferences = await promptForPreferences();

  // Base resume (required)
  const base_resume = await editor({
    message: 'Enter your base resume (markdown format):',
    validate: (value) =>
      value && value.trim().length > 0 ? true : 'Base resume is required',
  });

  // Base cover letter (required)
  const base_cover_letter = await editor({
    message: 'Enter your base cover letter template:',
    validate: (value) =>
      value && value.trim().length > 0
        ? true
        : 'Base cover letter template is required',
  });

  return {
    name,
    email,
    phone: phone || undefined,
    location: location || undefined,
    linkedin_url: linkedin_url || undefined,
    github_url: github_url || undefined,
    portfolio_url: portfolio_url || undefined,
    base_resume,
    base_cover_letter,
    preferences,
    skills,
    experience: [],
    education,
  };
}

async function promptForEducation(): Promise<Education> {
  const institution = await input({
    message: 'Institution name:',
    validate: (v) => (v.length > 0 ? true : 'Required'),
  });

  const degree = await input({
    message: 'Degree (e.g., Bachelor\'s, Master\'s):',
    validate: (v) => (v.length > 0 ? true : 'Required'),
  });

  const field = await input({
    message: 'Field of study (optional):',
  });

  const start_date = await input({
    message: 'Start date (optional):',
  });

  const end_date = await input({
    message: 'End date or expected graduation:',
  });

  const gpa = await input({
    message: 'GPA (optional):',
  });

  return {
    institution,
    degree,
    field: field || undefined,
    start_date: start_date || undefined,
    end_date: end_date || undefined,
    gpa: gpa || undefined,
  };
}

async function promptForPreferences(): Promise<Preferences> {
  const remote_only = await confirm({
    message: 'Only interested in remote jobs?',
    default: false,
  });

  const minSalaryInput = await input({
    message: 'Minimum salary (optional, numbers only):',
  });
  const min_salary = minSalaryInput ? parseInt(minSalaryInput, 10) : undefined;

  const locationsInput = await input({
    message: 'Preferred locations (comma-separated, optional):',
  });
  const preferred_locations = locationsInput
    .split(',')
    .map((l) => l.trim())
    .filter(Boolean);

  const excludedInput = await input({
    message: 'Companies to exclude (comma-separated, optional):',
  });
  const excluded_companies = excludedInput
    .split(',')
    .map((c) => c.trim())
    .filter(Boolean);

  const jobTypesInput = await input({
    message: 'Job types (comma-separated, e.g., full-time, contract):',
    default: 'full-time',
  });
  const job_types = jobTypesInput
    .split(',')
    .map((t) => t.trim())
    .filter(Boolean);

  return {
    remote_only,
    min_salary,
    preferred_locations,
    excluded_companies,
    job_types,
  };
}

export async function promptForProfileUpdate(
  current: Profile
): Promise<Partial<Profile>> {
  console.log('\n📝 Update your profile (press Enter to keep current value)\n');

  const name = await input({
    message: 'Full name:',
    default: current.name,
  });

  const email = await input({
    message: 'Email:',
    default: current.email,
  });

  const phone = await input({
    message: 'Phone:',
    default: current.phone ?? '',
  });

  const location = await input({
    message: 'Location:',
    default: current.location ?? '',
  });

  const skillsInput = await input({
    message: 'Skills (comma-separated):',
    default: current.skills.join(', '),
  });

  const skills = skillsInput
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);

  return {
    name,
    email,
    phone: phone || undefined,
    location: location || undefined,
    skills,
  };
}
