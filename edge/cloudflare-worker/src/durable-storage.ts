/**
 * Durable Storage - Zero-Drop Edge Storage
 * Uses Durable Objects for persistent, crash-safe storage
 */

import type { Env, EdgeClickEvent } from './index';

// ============================================
// DURABLE OBJECT: CLICK AGGREGATOR
// ============================================

export interface DurableStorageState {
  storage: DurableObjectStorage;
  blockConcurrencyWhile: (callback: () => Promise<void>) => void;
}

export class ClickAggregator {
  private state: DurableStorageState;
  private env: Env;
  private buffer: EdgeClickEvent[] = [];
  private bufferSize: number = 0;
  private maxBufferSize: number = 10000;
  private flushInterval: number = 5000; // 5 seconds
  private lastFlush: number = Date.now();

  constructor(state: DurableStorageState, env: Env) {
    this.state = state;
    this.env = env;

    // Restore buffer from storage on initialization
    this.state.blockConcurrencyWhile(async () => {
      await this.restoreBuffer();
    });
  }

  // ============================================
  // FETCH HANDLER
  // ============================================

  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);
    const path = url.pathname;

    switch (path) {
      case '/queue':
        if (request.method === 'POST') {
          return this.handleQueue(request);
        }
        break;

      case '/flush':
        if (request.method === 'POST') {
          return this.handleFlush();
        }
        break;

      case '/status':
        return this.handleStatus();

      case '/recover':
        if (request.method === 'POST') {
          return this.handleRecover();
        }
        break;
    }

    return new Response('Not Found', { status: 404 });
  }

  // ============================================
  // QUEUE OPERATIONS
  // ============================================

  private async handleQueue(request: Request): Promise<Response> {
    try {
      const event = await request.json() as EdgeClickEvent;

      // Add to buffer
      this.buffer.push(event);
      this.bufferSize++;

      // Persist to durable storage immediately (crash-safe)
      await this.persistEvent(event);

      // Check if we need to flush
      if (this.bufferSize >= this.maxBufferSize || 
          Date.now() - this.lastFlush > this.flushInterval) {
        // Flush in background
        this.flushToBackend();
      }

      return new Response(JSON.stringify({
        success: true,
        queued: true,
        buffer_size: this.bufferSize
      }), {
        headers: { 'Content-Type': 'application/json' }
      });

    } catch (error) {
      return new Response(JSON.stringify({
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error'
      }), {
        status: 500,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  }

  private async handleFlush(): Promise<Response> {
    const result = await this.flushToBackend();
    return new Response(JSON.stringify(result), {
      headers: { 'Content-Type': 'application/json' }
    });
  }

  private handleStatus(): Response {
    return new Response(JSON.stringify({
      buffer_size: this.bufferSize,
      max_buffer_size: this.maxBufferSize,
      last_flush: new Date(this.lastFlush).toISOString(),
      time_since_flush_ms: Date.now() - this.lastFlush
    }), {
      headers: { 'Content-Type': 'application/json' }
    });
  }

  private async handleRecover(): Promise<Response> {
    const recovered = await this.restoreBuffer();
    return new Response(JSON.stringify({
      success: true,
      recovered_events: recovered
    }), {
      headers: { 'Content-Type': 'application/json' }
    });
  }

  // ============================================
  // PERSISTENCE
  // ============================================

  private async persistEvent(event: EdgeClickEvent): Promise<void> {
    const key = `event:${event.tracking_code}:${Date.now()}`;
    await this.state.storage.put(key, JSON.stringify(event));
  }

  private async restoreBuffer(): Promise<number> {
    // List all stored events
    const entries = await this.state.storage.list({ prefix: 'event:' });
    
    let restored = 0;
    for (const [key, value] of entries) {
      try {
        const event = JSON.parse(value as string) as EdgeClickEvent;
        this.buffer.push(event);
        restored++;
      } catch {
        // Skip corrupted entries
      }
    }

    this.bufferSize = this.buffer.length;
    return restored;
  }

  private async clearProcessedEvents(eventKeys: string[]): Promise<void> {
    for (const key of eventKeys) {
      await this.state.storage.delete(key);
    }
  }

  // ============================================
  // FLUSH TO BACKEND
  // ============================================

  private async flushToBackend(): Promise<{
    success: boolean;
    sent: number;
    failed: number;
  }> {
    if (this.buffer.length === 0) {
      return { success: true, sent: 0, failed: 0 };
    }

    const backendUrl = this.env.BACKEND_URL || 'https://api.afftok.com';
    const endpoint = `${backendUrl}/api/internal/edge-click`;

    // Take batch from buffer
    const batchSize = Math.min(100, this.buffer.length);
    const batch = this.buffer.splice(0, batchSize);
    this.bufferSize -= batchSize;

    // Collect keys for cleanup
    const eventKeys: string[] = [];
    for (const event of batch) {
      eventKeys.push(`event:${event.tracking_code}:${Date.now()}`);
    }

    try {
      // Compress with gzip if available
      const body = JSON.stringify({ events: batch });

      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Edge-Batch': 'true',
          'X-Edge-Count': String(batch.length),
          'X-Edge-Source': 'durable-object'
        },
        body
      });

      if (response.ok) {
        // Clear processed events from storage
        await this.clearProcessedEvents(eventKeys);
        this.lastFlush = Date.now();

        return { success: true, sent: batch.length, failed: 0 };
      } else {
        // Re-add to buffer for retry
        this.buffer.unshift(...batch);
        this.bufferSize += batchSize;

        return { success: false, sent: 0, failed: batch.length };
      }

    } catch (error) {
      // Re-add to buffer for retry
      this.buffer.unshift(...batch);
      this.bufferSize += batchSize;

      return { success: false, sent: 0, failed: batch.length };
    }
  }
}

// ============================================
// R2 OVERFLOW STORAGE
// ============================================

export class R2OverflowStorage {
  private env: Env;
  private bucketName: string = 'afftok-edge-overflow';

  constructor(env: Env) {
    this.env = env;
  }

  /**
   * Store events to R2 when buffer exceeds threshold
   */
  async storeOverflow(events: EdgeClickEvent[]): Promise<string> {
    // Generate unique key
    const key = `overflow/${new Date().toISOString().split('T')[0]}/${Date.now()}_${Math.random().toString(36).substring(7)}.jsonl`;

    // Convert to JSONL format
    const content = events.map(e => JSON.stringify(e)).join('\n');

    // Compress with gzip
    const compressed = await this.gzipCompress(content);

    // Store to R2 (if bucket is available)
    // Note: R2 binding would be: this.env.OVERFLOW_BUCKET
    // For now, we'll use KV as fallback
    try {
      await this.env.CLICK_QUEUE.put(`r2_overflow:${key}`, compressed, {
        expirationTtl: 86400 * 7 // 7 days
      });
    } catch {
      // Fallback: store uncompressed
      await this.env.CLICK_QUEUE.put(`r2_overflow:${key}`, content, {
        expirationTtl: 86400 * 7
      });
    }

    return key;
  }

  /**
   * List overflow files
   */
  async listOverflow(): Promise<string[]> {
    const list = await this.env.CLICK_QUEUE.list({ prefix: 'r2_overflow:' });
    return list.keys.map(k => k.name.replace('r2_overflow:', ''));
  }

  /**
   * Restore events from overflow
   */
  async restoreOverflow(key: string): Promise<EdgeClickEvent[]> {
    const data = await this.env.CLICK_QUEUE.get(`r2_overflow:${key}`);
    if (!data) return [];

    // Try to decompress
    let content: string;
    try {
      content = await this.gzipDecompress(data);
    } catch {
      content = data;
    }

    // Parse JSONL
    const events: EdgeClickEvent[] = [];
    for (const line of content.split('\n')) {
      if (line.trim()) {
        try {
          events.push(JSON.parse(line));
        } catch {
          // Skip corrupted lines
        }
      }
    }

    return events;
  }

  /**
   * Delete overflow file after processing
   */
  async deleteOverflow(key: string): Promise<void> {
    await this.env.CLICK_QUEUE.delete(`r2_overflow:${key}`);
  }

  /**
   * Gzip compression
   */
  private async gzipCompress(data: string): Promise<string> {
    const encoder = new TextEncoder();
    const stream = new CompressionStream('gzip');
    const writer = stream.writable.getWriter();
    writer.write(encoder.encode(data));
    writer.close();

    const reader = stream.readable.getReader();
    const chunks: Uint8Array[] = [];
    
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      chunks.push(value);
    }

    // Combine chunks
    const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
    const result = new Uint8Array(totalLength);
    let offset = 0;
    for (const chunk of chunks) {
      result.set(chunk, offset);
      offset += chunk.length;
    }

    // Convert to base64
    return btoa(String.fromCharCode(...result));
  }

  /**
   * Gzip decompression
   */
  private async gzipDecompress(data: string): Promise<string> {
    // Convert from base64
    const binary = atob(data);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i);
    }

    const stream = new DecompressionStream('gzip');
    const writer = stream.writable.getWriter();
    writer.write(bytes);
    writer.close();

    const reader = stream.readable.getReader();
    const chunks: Uint8Array[] = [];
    
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      chunks.push(value);
    }

    // Combine and decode
    const decoder = new TextDecoder();
    return chunks.map(chunk => decoder.decode(chunk)).join('');
  }
}

// ============================================
// OFFLINE QUEUE MANAGER
// ============================================

export class OfflineQueueManager {
  private env: Env;
  private durableStorage: R2OverflowStorage;
  private localBuffer: EdgeClickEvent[] = [];
  private maxLocalBuffer: number = 10000;

  constructor(env: Env) {
    this.env = env;
    this.durableStorage = new R2OverflowStorage(env);
  }

  /**
   * Queue an event (handles overflow automatically)
   */
  async queue(event: EdgeClickEvent): Promise<void> {
    this.localBuffer.push(event);

    // Check if we need to overflow to R2
    if (this.localBuffer.length >= this.maxLocalBuffer) {
      await this.overflowToR2();
    }
  }

  /**
   * Overflow local buffer to R2
   */
  private async overflowToR2(): Promise<void> {
    if (this.localBuffer.length === 0) return;

    // Store to R2
    await this.durableStorage.storeOverflow(this.localBuffer);

    // Clear local buffer
    this.localBuffer = [];
  }

  /**
   * Flush all queued events to backend
   */
  async flushAll(ctx: ExecutionContext): Promise<{
    local_sent: number;
    overflow_sent: number;
    failed: number;
  }> {
    let localSent = 0;
    let overflowSent = 0;
    let failed = 0;

    // Flush local buffer
    const localResult = await this.flushLocalBuffer();
    localSent = localResult.sent;
    failed += localResult.failed;

    // Flush overflow files
    const overflowFiles = await this.durableStorage.listOverflow();
    for (const file of overflowFiles) {
      const events = await this.durableStorage.restoreOverflow(file);
      const result = await this.sendBatch(events);
      
      if (result.success) {
        await this.durableStorage.deleteOverflow(file);
        overflowSent += events.length;
      } else {
        failed += events.length;
      }
    }

    return { local_sent: localSent, overflow_sent: overflowSent, failed };
  }

  /**
   * Flush local buffer to backend
   */
  private async flushLocalBuffer(): Promise<{ sent: number; failed: number }> {
    if (this.localBuffer.length === 0) {
      return { sent: 0, failed: 0 };
    }

    const result = await this.sendBatch(this.localBuffer);
    
    if (result.success) {
      const sent = this.localBuffer.length;
      this.localBuffer = [];
      return { sent, failed: 0 };
    } else {
      return { sent: 0, failed: this.localBuffer.length };
    }
  }

  /**
   * Send batch to backend
   */
  private async sendBatch(events: EdgeClickEvent[]): Promise<{ success: boolean }> {
    const backendUrl = this.env.BACKEND_URL || 'https://api.afftok.com';
    const endpoint = `${backendUrl}/api/internal/edge-click`;

    try {
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Edge-Batch': 'true',
          'X-Edge-Count': String(events.length)
        },
        body: JSON.stringify({ events })
      });

      return { success: response.ok };
    } catch {
      return { success: false };
    }
  }

  /**
   * Get queue status
   */
  async getStatus(): Promise<{
    local_buffer_size: number;
    overflow_files: number;
    total_queued: number;
  }> {
    const overflowFiles = await this.durableStorage.listOverflow();
    
    return {
      local_buffer_size: this.localBuffer.length,
      overflow_files: overflowFiles.length,
      total_queued: this.localBuffer.length + (overflowFiles.length * this.maxLocalBuffer)
    };
  }
}

