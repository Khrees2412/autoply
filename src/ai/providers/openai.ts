import OpenAI from 'openai';
import type { AIProvider, AIConfig } from '../../types';

export class OpenAIProvider implements AIProvider {
  name = 'openai' as const;
  private client: OpenAI;
  private model: string;
  private temperature: number;

  constructor(config: AIConfig) {
    const apiKey = process.env.OPENAI_API_KEY;
    if (!apiKey) {
      throw new Error('OPENAI_API_KEY environment variable is not set');
    }

    this.client = new OpenAI({ apiKey });
    this.model = config.model ?? 'gpt-4-turbo-preview';
    this.temperature = config.temperature ?? 0.7;
  }

  async isAvailable(): Promise<boolean> {
    try {
      await this.client.models.list();
      return true;
    } catch {
      return false;
    }
  }

  async generateText(prompt: string, systemPrompt?: string): Promise<string> {
    const messages: OpenAI.ChatCompletionMessageParam[] = [];

    if (systemPrompt) {
      messages.push({ role: 'system', content: systemPrompt });
    }

    messages.push({ role: 'user', content: prompt });

    const response = await this.client.chat.completions.create({
      model: this.model,
      messages,
      temperature: this.temperature,
    });

    return response.choices[0]?.message?.content ?? '';
  }

  async listModels(): Promise<string[]> {
    const response = await this.client.models.list();
    return response.data
      .filter((m) => m.id.startsWith('gpt'))
      .map((m) => m.id);
  }
}
