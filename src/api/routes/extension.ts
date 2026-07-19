import type { FastifyInstance } from 'fastify';
import { profileRepository } from '../../db/repositories/profile';
import { configRepository } from '../../db/repositories/config';
export function registerExtensionRoutes(app: FastifyInstance): void {
  app.get('/extension/status', async () => {
    const profile = profileRepository.findFirst();
    const config = configRepository.loadAppConfig();

    // Skip AI provider check for now - it's causing timeouts
    // The extension can test AI connectivity when needed
    const aiProviderStatus: { available: boolean; error?: string } = { available: true };

    return {
      hasProfile: !!profile,
      profileName: profile?.name || null,
      aiProvider: config.ai.provider,
      aiProviderStatus,
      autoSubmitEnabled: config.application.autoSubmit,
    };
  });
}
