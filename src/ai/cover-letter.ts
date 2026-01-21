import type { AIProvider } from '../types';
import type { Profile, JobData } from '../types';

const COVER_LETTER_SYSTEM_PROMPT = `You are a cover letter writer who crafts warm, human, and passionate letters. Your goal is to help the candidate stand out by showing who they truly are - not just what they can do.

Writing style guidelines:
- Write like a real person, not a corporate template
- Lead with genuine excitement and curiosity about the role
- Focus on impact and stories, not technical jargon or buzzwords
- Show heart - let the candidate's passion and drive shine through
- Subtly weave in the candidate's unique perspective as someone bringing diverse global experience
- Keep it conversational yet professional
- Be confident but humble, ambitious but grounded
- 3-4 short paragraphs maximum - every sentence should earn its place

Avoid:
- Stiff, formal language ("I am writing to express my interest...")
- Listing skills or technologies - the resume does that
- Generic flattery about the company
- Overused phrases like "passionate about", "excited to", "leverage my skills"

The best cover letters feel like the start of a conversation, not a sales pitch.`;

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

Write a cover letter that:
1. Opens with something genuine - what specifically draws them to this role or company? Make it personal, not generic
2. Tells a brief story or two that shows their impact - focus on the human side, not technical details
3. Connects their journey and perspective to why this opportunity matters to them
4. Closes warmly with a clear next step

Remember: This person brings a unique perspective shaped by their background and experiences. Let that authenticity come through naturally - it's a strength, not something to hide. Write something that could only come from this specific person.`;
}

export async function answerApplicationQuestion(
  provider: AIProvider,
  profile: Profile,
  jobData: JobData,
  question: string
): Promise<string> {
  const systemPrompt = `You help job applicants answer application questions in a warm, authentic voice.

Guidelines:
- Sound like a real person, not a template
- Draw from actual experiences with specific examples
- Be honest and genuine - don't oversell
- Keep answers focused and appropriately brief
- Show personality and enthusiasm without being over the top
- Avoid corporate buzzwords and jargon`;

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
