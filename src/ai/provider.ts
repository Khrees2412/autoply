import type { AIProvider, AIProviderType, AIConfig } from '../types';
import { OllamaProvider } from './providers/ollama';
import { LMStudioProvider } from './providers/lmstudio';
import { OpenAIProvider } from './providers/openai';
import { AnthropicProvider } from './providers/anthropic';
import { configRepository } from '../db/repositories/config';

export function createAIProvider(config?: AIConfig): AIProvider {
  const aiConfig = config ?? configRepository.loadAppConfig().ai;

  switch (aiConfig.provider) {
    case 'ollama':
      return new OllamaProvider(aiConfig);
    case 'lmstudio':
      return new LMStudioProvider(aiConfig);
    case 'openai':
      return new OpenAIProvider(aiConfig);
    case 'anthropic':
      return new AnthropicProvider(aiConfig);
    default:
      throw new Error(`Unknown AI provider: ${aiConfig.provider}`);
  }
}

export function getAvailableProviders(): AIProviderType[] {
  return ['ollama', 'lmstudio', 'openai', 'anthropic'];
}

export async function testProvider(provider: AIProvider): Promise<{ success: boolean; error?: string }> {
  try {
    const available = await provider.isAvailable();
    if (!available) {
      return { success: false, error: 'Provider is not available or not running' };
    }

    // Try a simple generation
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

export { AIProvider };
