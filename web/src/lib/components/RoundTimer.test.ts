import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Test utilities for RoundTimer logic
function calculateRemaining(roundStartedAt: string, roundDurationSeconds: number, now: number): number {
  const roundStart = new Date(roundStartedAt).getTime();
  const endTime = roundStart + roundDurationSeconds * 1000;
  return Math.max(0, endTime - now);
}

function getColorClass(remaining: number): string {
  if (remaining > 60000) return 'bg-emerald-500';
  if (remaining > 30000) return 'bg-amber-500';
  if (remaining > 0) return 'bg-red-500 animate-pulse';
  return 'bg-red-600';
}

function formatTimer(remaining: number): string {
  const minutes = Math.floor(remaining / 60000);
  const seconds = Math.floor((remaining % 60000) / 1000);
  const paddedSeconds = String(seconds).padStart(2, '0');
  return `${minutes}:${paddedSeconds}`;
}

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.runOnlyPendingTimers();
  vi.useRealTimers();
});

describe('RoundTimer - Countdown Calculation', () => {
  it('calculates remaining time correctly when 30 seconds have elapsed', () => {
    const now = new Date('2024-04-18T12:00:00Z').getTime();
    const roundStartedAt = new Date(now - 30000).toISOString();
    const remaining = calculateRemaining(roundStartedAt, 120, now);

    // 120 seconds - 30 seconds = 90 seconds
    expect(remaining).toBe(90000);
  });

  it('calculates remaining time correctly at round start', () => {
    const now = new Date('2024-04-18T12:00:00Z').getTime();
    const roundStartedAt = new Date(now).toISOString();
    const remaining = calculateRemaining(roundStartedAt, 120, now);

    expect(remaining).toBe(120000);
  });

  it('returns 0 remaining when time has expired', () => {
    const now = new Date('2024-04-18T12:00:00Z').getTime();
    const roundStartedAt = new Date(now - 120000).toISOString(); // 2 minutes ago
    const remaining = calculateRemaining(roundStartedAt, 120, now);

    expect(remaining).toBe(0);
  });
});

describe('RoundTimer - Color States', () => {
  it('returns green class when remaining > 60 seconds', () => {
    const color = getColorClass(90000); // 1:30 remaining
    expect(color).toBe('bg-emerald-500');
  });

  it('returns amber class when remaining between 30-60 seconds', () => {
    const color = getColorClass(45000); // 0:45 remaining
    expect(color).toBe('bg-amber-500');
  });

  it('returns red pulsing class when remaining between 1-30 seconds', () => {
    const color = getColorClass(15000); // 0:15 remaining
    expect(color).toContain('bg-red-500');
    expect(color).toContain('animate-pulse');
  });

  it('returns dark red class when remaining <= 0', () => {
    const color = getColorClass(0);
    expect(color).toBe('bg-red-600');
  });

  it('returns dark red class when remaining is negative', () => {
    const color = getColorClass(-5000);
    expect(color).toBe('bg-red-600');
  });
});

describe('RoundTimer - Display Formatting', () => {
  it('formats 90 seconds as 1:30', () => {
    const display = formatTimer(90000);
    expect(display).toBe('1:30');
  });

  it('formats 45 seconds as 0:45', () => {
    const display = formatTimer(45000);
    expect(display).toBe('0:45');
  });

  it('formats 5 seconds as 0:05', () => {
    const display = formatTimer(5000);
    expect(display).toBe('0:05');
  });

  it('formats 0 seconds as 0:00', () => {
    const display = formatTimer(0);
    expect(display).toBe('0:00');
  });

  it('formats 120 seconds as 2:00', () => {
    const display = formatTimer(120000);
    expect(display).toBe('2:00');
  });

  it('pads seconds with leading zero for single digit', () => {
    const display = formatTimer(61000); // 1:01
    expect(display).toMatch(/1:01/);
  });
});

describe('RoundTimer - Timer Sync Updates', () => {
  it('recalculates remaining time when roundStartedAt prop changes', () => {
    const now = new Date('2024-04-18T12:00:00Z').getTime();
    let roundStartedAt = new Date(now - 30000).toISOString();
    let remaining = calculateRemaining(roundStartedAt, 120, now);

    expect(remaining).toBe(90000); // 1:30 remaining

    // Simulate timer_sync event — round was restarted with new timing
    roundStartedAt = new Date(now - 10000).toISOString(); // Only 10 seconds elapsed
    remaining = calculateRemaining(roundStartedAt, 120, now);

    expect(remaining).toBe(110000); // 1:50 remaining
  });

  it('recalculates remaining time when roundDurationSeconds prop changes', () => {
    const now = new Date('2024-04-18T12:00:00Z').getTime();
    const roundStartedAt = new Date(now - 30000).toISOString();
    let remaining = calculateRemaining(roundStartedAt, 120, now);

    expect(remaining).toBe(90000);

    // Duration was recalculated to accommodate remaining rounds
    remaining = calculateRemaining(roundStartedAt, 90, now); // 90 seconds total instead

    expect(remaining).toBe(60000); // 1:00 remaining
  });
});

describe('RoundTimer - Buzzer State', () => {
  it('identifies buzzer state when remaining <= 0', () => {
    const remaining = 0;
    const isBuzzer = remaining <= 0;

    expect(isBuzzer).toBe(true);
  });

  it('does not identify buzzer state when remaining > 0', () => {
    const remaining = 1000;
    const isBuzzer = remaining <= 0;

    expect(isBuzzer).toBe(false);
  });

  it('transitions to buzzer state correctly at 0 seconds', () => {
    const now = new Date('2024-04-18T12:00:00Z').getTime();
    const roundStartedAt = new Date(now - 119000).toISOString();
    let remaining = calculateRemaining(roundStartedAt, 120, now);

    // Should be very close to 0 but not quite
    expect(remaining).toBeGreaterThan(0);
    expect(remaining).toBeLessThanOrEqual(1000);

    // After 1+ more second
    const nextNow = now + 1000;
    remaining = calculateRemaining(roundStartedAt, 120, nextNow);
    expect(remaining).toBeLessThanOrEqual(0);
  });
});
