import type { AIProvider } from '../types';
import type { Profile, JobData } from '../types';

const RESUME_SYSTEM_PROMPT = `You are an expert resume writer. Your task is to tailor a resume to match a specific job posting while maintaining authenticity and accuracy. Follow these guidelines:

1. Highlight relevant skills and experiences that match the job requirements
2. Use keywords from the job description naturally
3. Quantify achievements where possible
4. Keep the format clean and professional
5. Do not fabricate or exaggerate experiences
6. Output the resume in clean markdown format

Be concise and impactful. Focus on achievements over responsibilities.`;

export async function tailorResume(
  provider: AIProvider,
  profile: Profile,
  jobData: JobData
): Promise<string> {
  const prompt = buildResumePrompt(profile, jobData);
  return provider.generateText(prompt, RESUME_SYSTEM_PROMPT);
}

function buildResumePrompt(profile: Profile, jobData: JobData): string {
  return `Please tailor the following resume for the job posting below.

## Candidate Profile

**Name:** ${profile.name}
**Email:** ${profile.email}
${profile.phone ? `**Phone:** ${profile.phone}` : ''}
${profile.location ? `**Location:** ${profile.location}` : ''}
${profile.linkedin_url ? `**LinkedIn:** ${profile.linkedin_url}` : ''}
${profile.github_url ? `**GitHub:** ${profile.github_url}` : ''}
${profile.portfolio_url ? `**Portfolio:** ${profile.portfolio_url}` : ''}

### Skills
${profile.skills.join(', ')}

### Experience
${profile.experience
  .map(
    (exp) => `
**${exp.title}** at ${exp.company}
${exp.location ? `${exp.location} | ` : ''}${exp.start_date} - ${exp.end_date ?? 'Present'}
${exp.description ?? ''}
${exp.highlights.length > 0 ? exp.highlights.map((h) => `- ${h}`).join('\n') : ''}
`
  )
  .join('\n')}

### Education
${profile.education
  .map(
    (edu) => `
**${edu.degree}**${edu.field ? ` in ${edu.field}` : ''} - ${edu.institution}
${edu.start_date ?? ''} - ${edu.end_date ?? ''}
${edu.gpa ? `GPA: ${edu.gpa}` : ''}
`
  )
  .join('\n')}

${profile.base_resume ? `### Additional Information from Base Resume\n${profile.base_resume}` : ''}

---

## Job Posting

**Position:** ${jobData.title}
**Company:** ${jobData.company}
${jobData.location ? `**Location:** ${jobData.location}` : ''}
${jobData.job_type ? `**Type:** ${jobData.job_type}` : ''}

### Description
${jobData.description}

### Requirements
${jobData.requirements.map((r) => `- ${r}`).join('\n')}

### Qualifications
${jobData.qualifications.map((q) => `- ${q}`).join('\n')}

---

Please generate a tailored resume in markdown format that highlights the most relevant qualifications for this specific position.`;
}

export async function generateResumeForMultipleJobs(
  provider: AIProvider,
  profile: Profile,
  jobs: JobData[]
): Promise<Map<string, string>> {
  const results = new Map<string, string>();

  for (const job of jobs) {
    const resume = await tailorResume(provider, profile, job);
    results.set(job.url, resume);
  }

  return results;
}
