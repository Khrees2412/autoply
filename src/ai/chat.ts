import type { AIProvider } from '../types';
import type { Profile } from '../types';
import { fetchProfileLinksContext } from '../utils/link-context-fetcher';

export interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
}

const CHAT_SYSTEM_PROMPT = `You are a career coach and interview prep assistant. Your goal is to help the user answer interview questions, behavioral questions, and career-related questions by drawing ONLY from their actual experience and background.

STRICT RESPONSE RULES:
1. BREVITY & CONCISENESS: Be extremely direct and concise. NO fluff, preamble, greetings, filler text, or conversational intros/outros (e.g. "Sure!", "Here is an answer", "Good luck!"). Jump straight to the point unless the user explicitly asks for a long/detailed breakdown.
2. ZERO HALLUCINATION: Rely ONLY on facts explicitly stated in the profile and link context below. Do NOT invent, assume, or hallucinate any skills, experience, company names, metrics, or credentials not present in the context. If information is missing, state briefly that it is not in the profile.
3. STORYTELLING & STRUCTURE: When drafting interview answers, use the STAR method naturally (Situation, Task, Action, Result) with real metrics and examples from their profile.
4. CLEAR FORMATTING: Format responses with clean Markdown (lists, bolding, headings) for readability.`;

export async function answerQuestionFromProfile(
  provider: AIProvider,
  profile: Profile,
  question: string,
  history: ChatMessage[] = []
): Promise<string> {
  const linkContext = await fetchProfileLinksContext(profile);
  const prompt = buildChatPrompt(profile, question, history, linkContext);
  return provider.generateText(prompt, CHAT_SYSTEM_PROMPT);
}

function formatHistory(history: ChatMessage[]): string {
  if (history.length === 0) return '';
  const lines = history.map((m) => `${m.role === 'user' ? 'User' : 'Assistant'}: ${m.content}`);
  return `\n## Conversation So Far\n${lines.join('\n\n')}\n`;
}

function buildChatPrompt(
  profile: Profile,
  question: string,
  history: ChatMessage[],
  linkContext: string = ''
): string {
  return `The user "${profile.name}" is preparing for job interviews and has a question. Use ONLY the information from their profile, resume, and external links below to craft your answer. Do not make up any experiences or details that aren't listed.
${formatHistory(history)}
## Current Question
${question}

---

## ${profile.name}'s Profile

${profile.email ? `Email: ${profile.email}` : ''}
${profile.phone ? `Phone: ${profile.phone}` : ''}
${profile.location ? `Location: ${profile.location}` : ''}
${profile.linkedin_url ? `LinkedIn: ${profile.linkedin_url}` : ''}
${profile.github_url ? `GitHub: ${profile.github_url}` : ''}
${profile.portfolio_url ? `Portfolio: ${profile.portfolio_url}` : ''}

### Skills
${profile.skills.length > 0 ? profile.skills.join(', ') : 'None listed.'}

### Work Experience
${
  profile.experience.length > 0
    ? profile.experience
        .map(
          (exp) => `
**${exp.title}** at ${exp.company}
${exp.start_date} - ${exp.end_date ?? 'Present'}
${exp.location ? `Location: ${exp.location}` : ''}
${exp.description ? `Description: ${exp.description}` : ''}
${exp.highlights.length > 0 ? `Key highlights: ${exp.highlights.join('; ')}` : ''}
`
        )
        .join('\n')
    : 'No work experience listed in profile.'
}

### Education
${
  profile.education.length > 0
    ? profile.education
        .map(
          (edu) => `
**${edu.degree}**${edu.field ? ` in ${edu.field}` : ''}
${edu.institution}
${edu.start_date ?? ''} - ${edu.end_date ?? ''}
${edu.gpa ? `GPA: ${edu.gpa}` : ''}
`
        )
        .join('\n')
    : 'No education listed in profile.'
}

${profile.base_resume ? `### Base Resume\n${profile.base_resume}\n` : ''}
${profile.base_cover_letter ? `### Base Cover Letter\n${profile.base_cover_letter}\n` : ''}
${linkContext ? `### External Links & Portfolio Context\n${linkContext}\n` : ''}

---

Based ONLY on the profile and link information above, craft a compelling answer to the current question. ${history.length > 0 ? 'The conversation history above provides context — reference it where relevant.' : ''} Use specific examples, metrics, and experiences from their background. Structure your response using storytelling - don't just list facts, but show how their experience prepared them for this type of question.

If the profile doesn't contain relevant information for this question, be honest about that but still provide helpful guidance on how they might approach answering it.`;
}
