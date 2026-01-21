import type { QueueItem } from '../types';
import { randomUUID } from 'crypto';

export class ApplicationQueue {
  private items: Map<string, QueueItem> = new Map();
  private processing = false;

  add(url: string): QueueItem {
    const item: QueueItem = {
      id: randomUUID(),
      url,
      status: 'pending',
    };
    this.items.set(item.id, item);
    return item;
  }

  addMany(urls: string[]): QueueItem[] {
    return urls.map((url) => this.add(url));
  }

  get(id: string): QueueItem | undefined {
    return this.items.get(id);
  }

  getAll(): QueueItem[] {
    return Array.from(this.items.values());
  }

  getPending(): QueueItem[] {
    return this.getAll().filter((item) => item.status === 'pending');
  }

  getProcessing(): QueueItem | undefined {
    return this.getAll().find((item) => item.status === 'processing');
  }

  getCompleted(): QueueItem[] {
    return this.getAll().filter((item) => item.status === 'completed');
  }

  getFailed(): QueueItem[] {
    return this.getAll().filter((item) => item.status === 'failed');
  }

  updateStatus(id: string, status: QueueItem['status'], error?: string): void {
    const item = this.items.get(id);
    if (item) {
      item.status = status;
      if (error) item.error = error;
    }
  }

  setResult(id: string, result: QueueItem['result']): void {
    const item = this.items.get(id);
    if (item) {
      item.result = result;
    }
  }

  remove(id: string): boolean {
    return this.items.delete(id);
  }

  clear(): void {
    this.items.clear();
  }

  size(): number {
    return this.items.size;
  }

  isEmpty(): boolean {
    return this.items.size === 0;
  }

  hasNext(): boolean {
    return this.getPending().length > 0;
  }

  getNext(): QueueItem | undefined {
    return this.getPending()[0];
  }

  isProcessing(): boolean {
    return this.processing;
  }

  setProcessing(value: boolean): void {
    this.processing = value;
  }

  getStats(): {
    total: number;
    pending: number;
    processing: number;
    completed: number;
    failed: number;
  } {
    const all = this.getAll();
    return {
      total: all.length,
      pending: all.filter((i) => i.status === 'pending').length,
      processing: all.filter((i) => i.status === 'processing').length,
      completed: all.filter((i) => i.status === 'completed').length,
      failed: all.filter((i) => i.status === 'failed').length,
    };
  }
}

export const applicationQueue = new ApplicationQueue();
