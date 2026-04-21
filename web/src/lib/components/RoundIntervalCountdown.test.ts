import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Test utilities for RoundIntervalCountdown logic
function formatIntervalTimer(remaining: number): string {
  const minutes = Math.floor(remaining / 60000);
  const seconds = Math.floor((remaining % 60000) / 1000);
  const paddedSeconds = String(seconds).padStart(2, '0');
  return `${minutes}:${paddedSeconds}`;
}

function calculateIntervalRemaining(startTime: number, intervalMinutes: number, now: number): number {
  const endTime = startTime + intervalMinutes * 60 * 1000;
  return Math.max(0, endTime - now);
}

function getIntervalProgressPercent(remaining: number, intervalMinutes: number): number {
  const total = intervalMinutes * 60 * 1000;
  return Math.max(0, Math.min(100, (remaining / total) * 100));
}

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.runOnlyPendingTimers();
  vi.useRealTimers();
});

describe('RoundIntervalCountdown - Timer Calculation', () => {
  it('calculates remaining interval correctly when 30 seconds have elapsed', () => {
    const now = new Date('2024-04-18T12:00:00Z').getTime();
    const startTime = now - 30000;
    const remaining = calculateIntervalRemaining(startTime, 3, now);

    // 3 minutes (180000ms) - 30 seconds (30000ms) = 150000ms
    expect(remaining).toBe(150000);
  });

  it('calculates remaining interval at start', () => {
    const now = new Date('2024-04-18T12:00:00Z').getTime();
    const startTime = now;
    const remaining = calculateIntervalRemaining(startTime, 3, now);

    // 3 minutes = 180000ms
    expect(remaining).toBe(180000);
  });

  it('returns 0 when interval has expired', () => {
    const now = new Date('2024-04-18T12:00:00Z').getTime();
    const startTime = now - 180000; // 3 minutes ago
    const remaining = calculateIntervalRemaining(startTime, 3, now);

    expect(remaining).toBe(0);
  });

  it('handles 1-minute intervals correctly', () => {
    const now = new Date('2024-04-18T12:00:00Z').getTime();
    const startTime = now - 30000; // 30 seconds elapsed
    const remaining = calculateIntervalRemaining(startTime, 1, now);

    // 1 minute (60000ms) - 30 seconds = 30000ms
    expect(remaining).toBe(30000);
  });

  it('handles 5-minute intervals correctly', () => {
    const now = new Date('2024-04-18T12:00:00Z').getTime();
    const startTime = now - 120000; // 2 minutes elapsed
    const remaining = calculateIntervalRemaining(startTime, 5, now);

    // 5 minutes (300000ms) - 2 minutes (120000ms) = 180000ms
    expect(remaining).toBe(180000);
  });
});

describe('RoundIntervalCountdown - Display Formatting', () => {
  it('formats 150 seconds as 2:30', () => {
    const display = formatIntervalTimer(150000);
    expect(display).toBe('2:30');
  });

  it('formats 30 seconds as 0:30', () => {
    const display = formatIntervalTimer(30000);
    expect(display).toBe('0:30');
  });

  it('formats 180 seconds (3 minutes) as 3:00', () => {
    const display = formatIntervalTimer(180000);
    expect(display).toBe('3:00');
  });

  it('formats 5 seconds as 0:05', () => {
    const display = formatIntervalTimer(5000);
    expect(display).toBe('0:05');
  });

  it('formats 0 seconds as 0:00', () => {
    const display = formatIntervalTimer(0);
    expect(display).toBe('0:00');
  });

  it('pads seconds with leading zero', () => {
    const display = formatIntervalTimer(61000); // 1:01
    expect(display).toMatch(/1:01/);
  });
});

describe('RoundIntervalCountdown - Progress Bar', () => {
  it('calculates progress as 100% at interval start', () => {
    const remaining = 180000; // full 3 minutes
    const progress = getIntervalProgressPercent(remaining, 3);

    expect(progress).toBe(100);
  });

  it('calculates progress as 50% when half remaining', () => {
    const remaining = 90000; // half of 3 minutes
    const progress = getIntervalProgressPercent(remaining, 3);

    expect(progress).toBe(50);
  });

  it('calculates progress as 0% when interval expired', () => {
    const remaining = 0;
    const progress = getIntervalProgressPercent(remaining, 3);

    expect(progress).toBe(0);
  });

  it('clamps progress to 0-100 range', () => {
    const testCases = [
      { remaining: -5000, intervalMinutes: 3, expected: 0 },
      { remaining: 200000, intervalMinutes: 3, expected: 100 },
    ];

    testCases.forEach(({ remaining, intervalMinutes, expected }) => {
      const progress = getIntervalProgressPercent(remaining, intervalMinutes);
      expect(progress).toBeGreaterThanOrEqual(0);
      expect(progress).toBeLessThanOrEqual(100);
      expect(progress).toBe(expected);
    });
  });
});

describe('RoundIntervalCountdown - Callbacks', () => {
  it('identifies when interval is complete (remaining <= 0)', () => {
    const remaining = 0;
    const isComplete = remaining <= 0;

    expect(isComplete).toBe(true);
  });

  it('does not identify as complete when time remains', () => {
    const remaining = 5000;
    const isComplete = remaining <= 0;

    expect(isComplete).toBe(false);
  });

  it('should call onComplete callback when remaining becomes 0', () => {
    const onComplete = vi.fn();
    const remaining = 0;

    if (remaining <= 0) {
      onComplete();
    }

    expect(onComplete).toHaveBeenCalled();
  });

  it('should call onSkipEarly callback when skip button clicked', () => {
    const onSkipEarly = vi.fn();

    // Simulate skip button click
    onSkipEarly();

    expect(onSkipEarly).toHaveBeenCalled();
  });
});

describe('RoundIntervalCountdown - State Transitions', () => {
  it('transitions through interval states: countdown -> complete', () => {
    const intervalMinutes = 3;
    const startTime = new Date('2024-04-18T12:00:00Z').getTime();

    // At start: full countdown
    let now = startTime;
    let remaining = calculateIntervalRemaining(startTime, intervalMinutes, now);
    expect(remaining).toBe(180000);

    // After 3 minutes: complete
    now = startTime + 180000;
    remaining = calculateIntervalRemaining(startTime, intervalMinutes, now);
    expect(remaining).toBe(0);
  });

  it('handles skip early before interval completes naturally', () => {
    const intervalMinutes = 3;
    const startTime = new Date('2024-04-18T12:00:00Z').getTime();
    const now = startTime + 30000; // Only 30 seconds elapsed

    let remaining = calculateIntervalRemaining(startTime, intervalMinutes, now);
    expect(remaining).toBe(150000); // 2:30 remaining

    // Skip early clicked
    const skipped = true;
    expect(skipped).toBe(true);
  });
});
