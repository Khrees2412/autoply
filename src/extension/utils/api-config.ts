/**
 * Manages the API Base URL for the Autoply extension.
 * Persists the URL to chrome.storage.local.
 */

const STORAGE_KEY = 'autoply_api_base_url';
const DEFAULT_API_BASE = 'http://localhost:8088';

export async function getApiBaseUrl(): Promise<string> {
  const result = await chrome.storage.local.get(STORAGE_KEY);
  return (result[STORAGE_KEY] as string) || DEFAULT_API_BASE;
}

export async function setApiBaseUrl(url: string): Promise<void> {
  await chrome.storage.local.set({ [STORAGE_KEY]: url });
}

export async function __clearApiBaseUrl(): Promise<void> {
  await chrome.storage.local.remove(STORAGE_KEY);
}
