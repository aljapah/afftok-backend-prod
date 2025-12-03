/**
 * Click Queue - Async click event batching and delivery
 */

import type { Env, EdgeClickEvent } from './index';

export interface QueuedClick {
  event: EdgeClickEvent;
  attempts: number;
  queued_at: string;
  last_attempt?: string;
}

export interface BatchResult {
  success: boolean;
  sent: number;
  failed: number;
  errors?: string[];
}

export class ClickQueue {
  private env: Env;
  private batchSize: number;
  private inMemoryQueue: QueuedClick[] = [];

  constructor(env: Env) {
    this.env = env;
    this.batchSize = parseInt(env.BATCH_SIZE || '50', 10);
  }

  /**
   * Queue a click event for async processing
   */
  async queueClick(event: EdgeClickEvent): Promise<void> {
    const queuedClick: QueuedClick = {
      event,
      attempts: 0,
      queued_at: new Date().toISOString()
    };

    // Add to in-memory queue
    this.inMemoryQueue.push(queuedClick);

    // Also persist to KV for durability
    try {
      const key = `click:${event.tracking_code}:${Date.now()}`;
      await this.env.CLICK_QUEUE.put(key, JSON.stringify(queuedClick), {
        expirationTtl: 3600 // 1 hour
      });
    } catch (error) {
      console.error('Failed to persist click to KV:', error);
    }

    // Flush if batch size reached
    if (this.inMemoryQueue.length >= this.batchSize) {
      await this.flushInMemory();
    }
  }

  /**
   * Flush in-memory queue to backend
   */
  private async flushInMemory(): Promise<BatchResult> {
    if (this.inMemoryQueue.length === 0) {
      return { success: true, sent: 0, failed: 0 };
    }

    const batch = this.inMemoryQueue.splice(0, this.batchSize);
    return this.sendBatch(batch);
  }

  /**
   * Flush all queued clicks to backend (called by scheduler)
   */
  async flushToBackend(ctx: ExecutionContext): Promise<Response> {
    const results: BatchResult[] = [];

    // 1. Flush in-memory queue
    if (this.inMemoryQueue.length > 0) {
      const memoryResult = await this.flushInMemory();
      results.push(memoryResult);
    }

    // 2. Flush KV queue
    const kvResult = await this.flushKVQueue();
    results.push(kvResult);

    const totalSent = results.reduce((sum, r) => sum + r.sent, 0);
    const totalFailed = results.reduce((sum, r) => sum + r.failed, 0);

    return new Response(JSON.stringify({
      success: true,
      total_sent: totalSent,
      total_failed: totalFailed,
      batches: results.length,
      timestamp: new Date().toISOString()
    }), {
      headers: { 'Content-Type': 'application/json' }
    });
  }

  /**
   * Flush KV queue
   */
  private async flushKVQueue(): Promise<BatchResult> {
    let sent = 0;
    let failed = 0;
    const errors: string[] = [];

    try {
      // List all queued clicks
      const list = await this.env.CLICK_QUEUE.list({ prefix: 'click:', limit: 100 });
      
      if (list.keys.length === 0) {
        return { success: true, sent: 0, failed: 0 };
      }

      // Fetch all clicks
      const clicks: QueuedClick[] = [];
      for (const key of list.keys) {
        try {
          const data = await this.env.CLICK_QUEUE.get(key.name);
          if (data) {
            clicks.push(JSON.parse(data));
          }
        } catch {
          // Skip invalid entries
        }
      }

      // Send batch
      if (clicks.length > 0) {
        const result = await this.sendBatch(clicks);
        sent = result.sent;
        failed = result.failed;
        if (result.errors) {
          errors.push(...result.errors);
        }

        // Delete sent clicks from KV
        if (result.success) {
          for (const key of list.keys) {
            try {
              await this.env.CLICK_QUEUE.delete(key.name);
            } catch {
              // Ignore delete errors
            }
          }
        }
      }

    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Unknown error';
      errors.push(errorMsg);
      failed++;
    }

    return { success: failed === 0, sent, failed, errors };
  }

  /**
   * Send batch to backend
   */
  private async sendBatch(clicks: QueuedClick[]): Promise<BatchResult> {
    const backendUrl = this.env.BACKEND_URL || 'https://api.afftok.com';
    const endpoint = `${backendUrl}/api/internal/edge-click`;

    try {
      const events = clicks.map(c => c.event);
      
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Edge-Batch': 'true',
          'X-Edge-Count': String(events.length),
          'X-Edge-Location': 'cloudflare'
        },
        body: JSON.stringify({ events })
      });

      if (response.ok) {
        return { success: true, sent: events.length, failed: 0 };
      } else {
        const errorText = await response.text();
        return {
          success: false,
          sent: 0,
          failed: events.length,
          errors: [`Backend error: ${response.status} - ${errorText}`]
        };
      }

    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Unknown error';
      return {
        success: false,
        sent: 0,
        failed: clicks.length,
        errors: [errorMsg]
      };
    }
  }

  /**
   * Get queue status
   */
  async getQueueStatus(): Promise<{
    in_memory: number;
    in_kv: number;
    oldest_event?: string;
  }> {
    let kvCount = 0;
    let oldestEvent: string | undefined;

    try {
      const list = await this.env.CLICK_QUEUE.list({ prefix: 'click:', limit: 1000 });
      kvCount = list.keys.length;

      if (list.keys.length > 0) {
        // Get oldest event
        const oldestKey = list.keys[0];
        const data = await this.env.CLICK_QUEUE.get(oldestKey.name);
        if (data) {
          const click = JSON.parse(data) as QueuedClick;
          oldestEvent = click.queued_at;
        }
      }
    } catch {
      // Ignore errors
    }

    return {
      in_memory: this.inMemoryQueue.length,
      in_kv: kvCount,
      oldest_event: oldestEvent
    };
  }

  /**
   * Retry failed clicks with exponential backoff
   */
  async retryFailed(ctx: ExecutionContext): Promise<void> {
    try {
      const list = await this.env.CLICK_QUEUE.list({ prefix: 'failed:', limit: 50 });

      for (const key of list.keys) {
        const data = await this.env.CLICK_QUEUE.get(key.name);
        if (!data) continue;

        const click = JSON.parse(data) as QueuedClick;
        
        // Check retry limit
        if (click.attempts >= 5) {
          // Move to dead letter
          await this.moveToDeadLetter(click);
          await this.env.CLICK_QUEUE.delete(key.name);
          continue;
        }

        // Calculate backoff
        const backoffMs = Math.pow(2, click.attempts) * 1000;
        const lastAttempt = click.last_attempt ? new Date(click.last_attempt).getTime() : 0;
        
        if (Date.now() - lastAttempt < backoffMs) {
          continue; // Not ready for retry
        }

        // Retry
        click.attempts++;
        click.last_attempt = new Date().toISOString();

        const result = await this.sendBatch([click]);
        
        if (result.success) {
          await this.env.CLICK_QUEUE.delete(key.name);
        } else {
          await this.env.CLICK_QUEUE.put(key.name, JSON.stringify(click), {
            expirationTtl: 3600
          });
        }
      }
    } catch (error) {
      console.error('Retry failed:', error);
    }
  }

  /**
   * Move to dead letter queue
   */
  private async moveToDeadLetter(click: QueuedClick): Promise<void> {
    try {
      const key = `dlq:${click.event.tracking_code}:${Date.now()}`;
      await this.env.CLICK_QUEUE.put(key, JSON.stringify(click), {
        expirationTtl: 604800 // 7 days
      });
    } catch {
      // Ignore errors
    }
  }

  /**
   * Get dead letter queue items
   */
  async getDeadLetterQueue(): Promise<QueuedClick[]> {
    const items: QueuedClick[] = [];

    try {
      const list = await this.env.CLICK_QUEUE.list({ prefix: 'dlq:', limit: 100 });

      for (const key of list.keys) {
        const data = await this.env.CLICK_QUEUE.get(key.name);
        if (data) {
          items.push(JSON.parse(data));
        }
      }
    } catch {
      // Ignore errors
    }

    return items;
  }
}

