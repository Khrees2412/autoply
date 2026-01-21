import { getDb, getAutoplyDir } from '../index';
import { join } from 'path';
import { readFileSync, writeFileSync, existsSync } from 'fs';
import type { AppConfig } from '../../types';
import { DEFAULT_CONFIG } from '../../types';

const CONFIG_FILE = join(getAutoplyDir(), 'config.json');

export class ConfigRepository {
  // Database-based config (for key-value pairs)
  get(key: string): string | null {
    const db = getDb();
    const row = db.query<{ value: string }, [string]>('SELECT value FROM config WHERE key = ?').get(key);
    return row?.value ?? null;
  }

  set(key: string, value: string): void {
    const db = getDb();
    db.run(
      'INSERT INTO config (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = ?',
      [key, value, value]
    );
  }

  delete(key: string): boolean {
    const db = getDb();
    const result = db.run('DELETE FROM config WHERE key = ?', [key]);
    return result.changes > 0;
  }

  getAll(): Record<string, string> {
    const db = getDb();
    const rows = db.query<{ key: string; value: string }, []>('SELECT key, value FROM config').all();
    const config: Record<string, string> = {};
    for (const row of rows) {
      config[row.key] = row.value;
    }
    return config;
  }

  // File-based config (for AppConfig object)
  loadAppConfig(): AppConfig {
    if (existsSync(CONFIG_FILE)) {
      try {
        const content = readFileSync(CONFIG_FILE, 'utf-8');
        return { ...DEFAULT_CONFIG, ...JSON.parse(content) };
      } catch {
        return DEFAULT_CONFIG;
      }
    }
    return DEFAULT_CONFIG;
  }

  saveAppConfig(config: AppConfig): void {
    writeFileSync(CONFIG_FILE, JSON.stringify(config, null, 2));
  }

  updateAppConfig(updates: Partial<AppConfig>): AppConfig {
    const current = this.loadAppConfig();
    const updated = {
      ...current,
      ...updates,
      ai: { ...current.ai, ...updates.ai },
      browser: { ...current.browser, ...updates.browser },
      application: { ...current.application, ...updates.application },
    };
    this.saveAppConfig(updated);
    return updated;
  }

  setConfigValue(path: string, value: unknown): AppConfig {
    const config = this.loadAppConfig();
    const parts = path.split('.');

    // Navigate to the nested location
    let current: Record<string, unknown> = config as unknown as Record<string, unknown>;
    for (let i = 0; i < parts.length - 1; i++) {
      if (typeof current[parts[i]] !== 'object' || current[parts[i]] === null) {
        current[parts[i]] = {};
      }
      current = current[parts[i]] as Record<string, unknown>;
    }

    // Set the value
    const lastKey = parts[parts.length - 1];

    // Try to parse as JSON if it's a string
    if (typeof value === 'string') {
      try {
        value = JSON.parse(value);
      } catch {
        // Keep as string
      }
    }

    current[lastKey] = value;
    this.saveAppConfig(config);
    return config;
  }

  getConfigValue(path: string): unknown {
    const config = this.loadAppConfig();
    const parts = path.split('.');

    let current: unknown = config;
    for (const part of parts) {
      if (typeof current !== 'object' || current === null) {
        return undefined;
      }
      current = (current as Record<string, unknown>)[part];
    }

    return current;
  }
}

export const configRepository = new ConfigRepository();
