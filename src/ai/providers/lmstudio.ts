import type { AIProvider, AIConfig } from '../../types';

interface LMStudioMessage {
  role: 'system' | 'user' | 'assistant';
  content: string;
}

interface LMStudioRequest {
  model: string;
  messages: LMStudioMessage[];
  temperature?: number;
  stream: boolean;
}

interface LMStudioResponse {
  id: string;
  choices: Array<{
    message: {
      content: string;
    };
  }>;
}

export class LMStudioProvider implements AIProvider {
  name = 'lmstudio' as const;
  private baseUrl: string;
  private model: string;
  private temperature: number;

  constructor(config: AIConfig) {
    this.baseUrl = config.baseUrl ?? 'http://localhost:1234';
    this.model = config.model ?? 'local-model';
    this.temperature = config.temperature ?? 0.7;
  }

  async isAvailable(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/v1/models`);
      return response.ok;
    } catch {
      return false;
    }
  }

  async generateText(prompt: string, systemPrompt?: string): Promise<string> {
    const messages: LMStudioMessage[] = [];

    if (systemPrompt) {
      messages.push({ role: 'system', content: systemPrompt });
    }

    messages.push({ role: 'user', content: prompt });

    const request: LMStudioRequest = {
      model: this.model,
      messages,
      temperature: this.temperature,
      stream: false,
    };

    const response = await fetch(`${this.baseUrl}/v1/chat/completions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`LMStudio API error: ${error}`);
    }

    const data = (await response.json()) as LMStudioResponse;
    return data.choices[0]?.message?.content ?? '';
  }

  async listModels(): Promise<string[]> {
    const response = await fetch(`${this.baseUrl}/v1/models`);
    if (!response.ok) {
      throw new Error('Failed to list LMStudio models');
    }
    const data = (await response.json()) as { data: Array<{ id: string }> };
    return data.data.map((m) => m.id);
  }
}
