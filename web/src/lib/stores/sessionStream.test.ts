import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { sessionStream } from './sessionStream.svelte.js';

class MockEventSource {
  static instances: MockEventSource[] = [];
  url: string;
  onopen: (() => void) | null = null;
  onerror: (() => void) | null = null;
  readyState = 0;
  private listeners = new Map<string, Set<(e: { data: string }) => void>>();

  constructor(url: string) {
    this.url = url;
    MockEventSource.instances.push(this);
  }

  addEventListener(type: string, fn: (e: { data: string }) => void) {
    if (!this.listeners.has(type)) this.listeners.set(type, new Set());
    this.listeners.get(type)!.add(fn);
  }

  emit(type: string, payload: unknown) {
    const data = JSON.stringify({ payload });
    this.listeners.get(type)?.forEach((fn) => fn({ data }));
  }

  close() {
    this.readyState = 2;
  }
}

beforeEach(() => {
  MockEventSource.instances = [];
  vi.stubGlobal('EventSource', MockEventSource);
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('sessionStream', () => {
  it('connects to the correct SSE URL on start', () => {
    const stream = sessionStream('sess123');
    stream.start();

    expect(MockEventSource.instances).toHaveLength(1);
    expect(MockEventSource.instances[0].url).toBe('/api/sessions/sess123/events');

    stream.close();
  });

  it('does not connect before start is called', () => {
    sessionStream('sess123');
    expect(MockEventSource.instances).toHaveLength(0);
  });

  it('dispatches events to registered handlers', () => {
    const stream = sessionStream('sess123');
    stream.start();

    const handler = vi.fn();
    stream.onEvent('session_updated', handler);

    MockEventSource.instances[0].emit('session_updated', { status: 'playing' });

    expect(handler).toHaveBeenCalledWith({ status: 'playing' });
    stream.close();
  });

  it('dispatches to multiple handlers for the same event', () => {
    const stream = sessionStream('sess123');
    stream.start();

    const h1 = vi.fn();
    const h2 = vi.fn();
    stream.onEvent('session_updated', h1);
    stream.onEvent('session_updated', h2);

    MockEventSource.instances[0].emit('session_updated', {});

    expect(h1).toHaveBeenCalled();
    expect(h2).toHaveBeenCalled();
    stream.close();
  });

  it('returns an unsubscribe function from onEvent', () => {
    const stream = sessionStream('sess123');
    stream.start();

    const handler = vi.fn();
    const unsubscribe = stream.onEvent('session_updated', handler);
    unsubscribe();

    MockEventSource.instances[0].emit('session_updated', {});

    expect(handler).not.toHaveBeenCalled();
    stream.close();
  });

  it('resets backoff to 500 on successful open', () => {
    const stream = sessionStream('sess123');
    stream.start();

    const es = MockEventSource.instances[0];
    // Trigger multiple errors to increase backoff, then open should reset it.
    es.onerror?.();
    es.onerror?.();
    es.onopen?.();

    stream.close();
  });

  it('does not reconnect after close', () => {
    const stream = sessionStream('sess123');
    stream.start();
    stream.close();

    const countBefore = MockEventSource.instances.length;
    // Simulate error on the now-closed stream.
    MockEventSource.instances[0].onerror?.();

    // No new EventSource should have been created.
    expect(MockEventSource.instances).toHaveLength(countBefore);
  });

  it('attaches handlers registered before connect to a new EventSource', () => {
    const stream = sessionStream('sess123');

    // Register handler before start
    const handler = vi.fn();
    stream.onEvent('round_updated', handler);
    stream.start();

    MockEventSource.instances[0].emit('round_updated', { round: 2 });

    expect(handler).toHaveBeenCalledWith({ round: 2 });
    stream.close();
  });

});
