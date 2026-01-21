import type { AIProvider } from '../types';
import type { Profile, JobData } from '../types';

const COVER_LETTER_SYSTEM_PROMPT = `You are an expert cover letter writer. Your task is to create compelling, personalized cover letters that:

1. Show genuine enthusiasm for the specific role and company
2. Connect the candidate's experience to the job requirements
3. Demonstrate knowledge of the company (when available)
4. Be concise (3-4 paragraphs maximum)
5. Use a professional but personable tone
6. Avoid generic phrases and clichés
7. Include a clear call to action

Output the cover letter in clean text format, ready to be used.`;

export async function generateCoverLetter(
  provider: AIProvider,
  profile: Profile,
  jobData: JobData
): Promise<string> {
  const prompt = buildCoverLetterPrompt(profile, jobData);
  return provider.generateText(prompt, COVER_LETTER_SYSTEM_PROMPT);
}

function buildCoverLetterPrompt(profile: Profile, jobData: JobData): string {
  return `Please write a cover letter for the following job application.

## Candidate Profile

**Name:** ${profile.name}
**Email:** ${profile.email}
${profile.location ? `**Location:** ${profile.location}` : ''}

### Summary of Qualifications
${profile.skills.slice(0, 10).join(', ')}

### Recent Experience
${profile.experience
  .slice(0, 3)
  .map(
    (exp) => `
**${exp.title}** at ${exp.company} (${exp.start_date} - ${exp.end_date ?? 'Present'})
${exp.description ?? ''}
Key achievements: ${exp.highlights.slice(0, 3).join('; ')}
`
  )
  .join('\n')}

${profile.base_cover_letter ? `### Cover Letter Template/Notes\n${profile.base_cover_letter}` : ''}

---

## Job Posting

**Position:** ${jobData.title}
**Company:** ${jobData.company}
${jobData.location ? `**Location:** ${jobData.location}` : ''}

### Description
${jobData.description}

### Key Requirements
${jobData.requirements.slice(0, 5).map((r) => `- ${r}`).join('\n')}

---

Please write a personalized cover letter that:
1. Opens with a compelling hook about interest in the ${jobData.title} role at ${jobData.company}
2. Highlights 2-3 specific experiences that align with the job requirements
3. Shows enthusiasm for the company and role
4. Ends with a professional call to action

Keep it to 3-4 paragraphs.`;
}

export async function answerApplicationQuestion(
  provider: AIProvider,
  profile: Profile,
  jobData: JobData,
  question: string
): Promise<string> {
  const systemPrompt = `You are helping a job applicant answer application questions.
Provide concise, professional answers that:
1. Are relevant to the specific role and company
2. Draw from the candidate's actual experience
3. Are honest and authentic
4. Are appropriately brief unless a detailed answer is clearly needed`;

  const prompt = `Based on the following candidate profile and job posting, please answer this application question:

## Question
"${question}"

## Candidate Profile
Name: ${profile.name}
Skills: ${profile.skills.join(', ')}
Recent Experience:
${profile.experience
  .slice(0, 2)
  .map((exp) => `- ${exp.title} at ${exp.company}: ${exp.description ?? exp.highlights.join(', ')}`)
  .join('\n')}

## Job
${jobData.title} at ${jobData.company}
${jobData.description.slice(0, 500)}...

Please provide a concise, relevant answer to the question.`;

  return provider.generateText(prompt, systemPrompt);
}

export async function answerMultipleQuestions(
  provider: AIProvider,
  profile: Profile,
  jobData: JobData,
  questions: string[]
): Promise<Map<string, string>> {
  const results = new Map<string, string>();

  for (const question of questions) {
    const answer = await answerApplicationQuestion(provider, profile, jobData, question);
    results.set(question, answer);
  }

  return results;
}
