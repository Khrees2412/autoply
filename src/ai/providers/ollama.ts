import type { AIProvider, AIConfig } from '../../types';

interface OllamaResponse {
  model: string;
  response: string;
  done: boolean;
}

interface OllamaGenerateRequest {
  model: string;
  prompt: string;
  system?: string;
  stream: boolean;
  options?: {
    temperature?: number;
  };
}

export class OllamaProvider implements AIProvider {
  name = 'ollama' as const;
  private baseUrl: string;
  private model: string;
  private temperature: number;

  constructor(config: AIConfig) {
    this.baseUrl = config.baseUrl ?? 'http://localhost:11434';
    this.model = config.model ?? 'llama3.2';
    this.temperature = config.temperature ?? 0.7;
  }

  async isAvailable(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/api/tags`);
      return response.ok;
    } catch {
      return false;
    }
  }

  async generateText(prompt: string, systemPrompt?: string): Promise<string> {
    const request: OllamaGenerateRequest = {
      model: this.model,
      prompt,
      stream: false,
      options: {
        temperature: this.temperature,
      },
    };

    if (systemPrompt) {
      request.system = systemPrompt;
    }

    const response = await fetch(`${this.baseUrl}/api/generate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Ollama API error: ${error}`);
    }

    const data = (await response.json()) as OllamaResponse;
    return data.response;
  }

  async listModels(): Promise<string[]> {
    const response = await fetch(`${this.baseUrl}/api/tags`);
    if (!response.ok) {
      throw new Error('Failed to list Ollama models');
    }
    const data = (await response.json()) as { models: Array<{ name: string }> };
    return data.models.map((m) => m.name);
  }
}
