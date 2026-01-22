import { generateText } from 'ai';
import { createOpenAI } from '@ai-sdk/openai';
import { createAnthropic } from '@ai-sdk/anthropic';
import { createGoogleGenerativeAI } from '@ai-sdk/google';
import type { AIProvider, AIProviderType, AIConfig } from '../types';
import { configRepository } from '../db/repositories/config';

// Model mappings for each provider
const MODEL_DEFAULTS: Record<AIProviderType, string> = {
  openai: 'gpt-4o',
  anthropic: 'claude-3-5-sonnet-latest',
  google: 'gemini-1.5-pro',
  ollama: 'llama3.2',
  lmstudio: 'local-model',
};

function createModel(config: AIConfig) {
  const modelId = config.model || MODEL_DEFAULTS[config.provider];

  switch (config.provider) {
    case 'openai': {
      const openai = createOpenAI({
        apiKey: process.env.OPENAI_API_KEY,
      });
      return openai(modelId);
    }
    case 'anthropic': {
      const anthropic = createAnthropic({
        apiKey: process.env.ANTHROPIC_API_KEY,
      });
      return anthropic(modelId);
    }
    case 'google': {
      const google = createGoogleGenerativeAI({
        apiKey: process.env.GOOGLE_API_KEY,
      });
      return google(modelId);
    }
    case 'ollama': {
      // Ollama uses OpenAI-compatible API
      let baseUrl = config.baseUrl ?? 'http://localhost:11434';
      if (!baseUrl.endsWith('/v1')) {
        baseUrl = baseUrl.replace(/\/$/, '') + '/v1';
      }
      const ollama = createOpenAI({
        baseURL: baseUrl,
        apiKey: 'ollama', // Ollama doesn't require an API key
      });
      return ollama(modelId);
    }
    case 'lmstudio': {
      // LMStudio uses OpenAI-compatible API
      let lmBaseUrl = config.baseUrl ?? 'http://localhost:1234';
      if (!lmBaseUrl.endsWith('/v1')) {
        lmBaseUrl = lmBaseUrl.replace(/\/$/, '') + '/v1';
      }
      const lmstudio = createOpenAI({
        baseURL: lmBaseUrl,
        apiKey: 'lmstudio', // LMStudio doesn't require an API key
      });
      return lmstudio(modelId);
    }
    default:
      throw new Error(`Unknown AI provider: ${config.provider}`);
  }
}

class UnifiedAIProvider implements AIProvider {
  name: AIProviderType;
  private config: AIConfig;

  constructor(config: AIConfig) {
    this.name = config.provider;
    this.config = config;
  }

  async isAvailable(): Promise<boolean> {
    try {
      const model = createModel(this.config);
      await generateText({
        model,
        prompt: 'Hi',
        maxTokens: 50,
      });
      return true;
    } catch {
      return false;
    }
  }

  async generateText(prompt: string, systemPrompt?: string): Promise<string> {
    const model = createModel(this.config);

    const result = await generateText({
      model,
      system: systemPrompt,
      prompt,
      temperature: this.config.temperature ?? 0.7,
    });

    return result.text;
  }
}

export function createAIProvider(config?: AIConfig): AIProvider {
  const aiConfig = config ?? configRepository.loadAppConfig().ai;
  return new UnifiedAIProvider(aiConfig);
}

export function getAvailableProviders(): AIProviderType[] {
  return ['openai', 'anthropic', 'google', 'ollama', 'lmstudio'];
}

export async function testProvider(provider: AIProvider): Promise<{ success: boolean; error?: string }> {
  try {
    const available = await provider.isAvailable();
    if (!available) {
      return { success: false, error: 'Provider is not available or not running' };
    }

    const response = await provider.generateText('Say "hello" and nothing else.');
    if (!response || response.length === 0) {
      return { success: false, error: 'Provider returned empty response' };
    }

    return { success: true };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Unknown error',
    };
  }
}

export type { AIProvider };
