import Anthropic from '@anthropic-ai/sdk';
import type { AIProvider, AIConfig } from '../../types';

export class AnthropicProvider implements AIProvider {
  name = 'anthropic' as const;
  private client: Anthropic;
  private model: string;
  private temperature: number;

  constructor(config: AIConfig) {
    const apiKey = process.env.ANTHROPIC_API_KEY;
    if (!apiKey) {
      throw new Error('ANTHROPIC_API_KEY environment variable is not set');
    }

    this.client = new Anthropic({ apiKey });
    this.model = config.model ?? 'claude-3-sonnet-20240229';
    this.temperature = config.temperature ?? 0.7;
  }

  async isAvailable(): Promise<boolean> {
    try {
      // Try a minimal API call to verify connection
      await this.client.messages.create({
        model: this.model,
        max_tokens: 10,
        messages: [{ role: 'user', content: 'Hi' }],
      });
      return true;
    } catch {
      return false;
    }
  }

  async generateText(prompt: string, systemPrompt?: string): Promise<string> {
    const response = await this.client.messages.create({
      model: this.model,
      max_tokens: 4096,
      system: systemPrompt,
      messages: [{ role: 'user', content: prompt }],
    });

    const content = response.content[0];
    if (content.type === 'text') {
      return content.text;
    }

    return '';
  }
}
