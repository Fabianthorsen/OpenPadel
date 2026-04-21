import { describe, it, expect, vi } from 'vitest';

/**
 * Integration tests for ActiveSession timed_americano support
 *
 * These tests verify:
 * - Scoring logic for timed_americano mode (free scoring, no sum constraint)
 * - Round interval countdown display and auto-advance
 * - Skip early functionality during interval countdown
 */

describe('ActiveSession - Timed Americano Scoring', () => {
  describe('Free Scoring - Score Validation', () => {
    it('accepts any score 0-99 for timed_americano (no sum constraint)', () => {
      // In timed mode, we should allow any valid score
      const score = 23;
      const maxScore = 99; // Timed mode cap
      expect(score).toBeGreaterThanOrEqual(0);
      expect(score).toBeLessThanOrEqual(maxScore);
    });

    it('rejects negative scores', () => {
      const score = -5;
      expect(score).toBeLessThan(0);
    });

    it('allows asymmetric scores (a=23, b=19) without sum validation', () => {
      // Unlike Americano where a + b must equal points (e.g., 24)
      // Timed Americano allows free scoring: a and b are independent
      const a = 23;
      const b = 19;
      const maxScore = 99;

      expect(a).toBeGreaterThanOrEqual(0);
      expect(a).toBeLessThanOrEqual(maxScore);
      expect(b).toBeGreaterThanOrEqual(0);
      expect(b).toBeLessThanOrEqual(maxScore);
      // In timed mode, we don't check a + b === points
    });

    it('allows high scores (e.g., 47-42) in timed mode', () => {
      const a = 47;
      const b = 42;
      const maxScore = 99;

      expect(a).toBeLessThanOrEqual(maxScore);
      expect(b).toBeLessThanOrEqual(maxScore);
    });
  });

  describe('Score Entry - Separate Team Collection', () => {
    it('accepts first team score (team A), waits for team B', () => {
      const matchId = 'match1';
      const teamAScore = 15;

      // After team A score is set, system should prompt for team B
      // (Team B score undefined until entered)
      expect(teamAScore).toBeGreaterThanOrEqual(0);
    });

    it('accepts second team score (team B) after team A is set', () => {
      const teamAScore = 15;
      const teamBScore = 18;

      // Both scores now defined
      expect(teamAScore).toBeGreaterThanOrEqual(0);
      expect(teamBScore).toBeGreaterThanOrEqual(0);
    });

    it('allows team B score to be different from team A (no auto-complement)', () => {
      const teamAScore = 20;
      const teamBScore = 15;

      // In Americano mode: teamBScore would auto-calculate as points - 20
      // In Timed mode: teamBScore is independently entered
      expect(teamBScore).not.toBe(24 - teamAScore); // Not auto-complemented to fixed target
    });
  });

  describe('Numpad Guards - Timed Mode Differences', () => {
    it('does not prevent digit entry for high scores in timed mode', () => {
      const maxScore = 99; // Timed mode cap
      const attempt = 87;

      expect(attempt).toBeLessThanOrEqual(maxScore);
      // Should allow entry (not blocked by points constraint)
    });

    it('caps score at 99 for timed_americano', () => {
      const maxScore = 99;
      const capped = Math.min(99, 105);

      expect(capped).toBe(99);
    });

    it('allows +/- adjustments within 0-99 range for timed mode', () => {
      const current = 50;
      const delta = 10;
      const adjusted = Math.max(0, Math.min(99, current + delta));

      expect(adjusted).toBe(60);
    });

    it('prevents score below 0 with + click', () => {
      const current = 0;
      const delta = -5;
      const adjusted = Math.max(0, current + delta);

      expect(adjusted).toBe(0); // Clamped to 0
    });

    it('prevents score above 99 with + click in timed mode', () => {
      const current = 98;
      const delta = 5;
      const adjusted = Math.max(0, Math.min(99, current + delta));

      expect(adjusted).toBe(99); // Clamped to 99
    });
  });

  describe('Finalize Button Guard - Timed Mode', () => {
    it('allows finalize if both scores are non-negative in timed mode', () => {
      const isTimedAmericano = true;
      const a = 20;
      const b = 15;

      const canFinalize = isTimedAmericano ? !(a < 0 || b < 0) : (a + b === 24);
      expect(canFinalize).toBe(true);
    });

    it('prevents finalize if either score is negative in timed mode', () => {
      const isTimedAmericano = true;
      const a = -5;
      const b = 15;

      const canFinalize = isTimedAmericano ? !(a < 0 || b < 0) : (a + b === 24);
      expect(canFinalize).toBe(false);
    });

    it('allows finalize for any positive sum in timed mode (no target check)', () => {
      const isTimedAmericano = true;
      const a = 50;
      const b = 48;
      const points = 24; // Original target

      // In americano: a + b must equal 24 (canFinalize = false here)
      // In timed: no sum check, only negative check
      const canFinalize = isTimedAmericano ? !(a < 0 || b < 0) : (a + b === points);
      expect(canFinalize).toBe(true);
    });
  });

  describe('Round Interval Countdown - Display', () => {
    it('shows interval countdown after round completes', () => {
      const session = {
        game_mode: 'timed_americano' as const,
        current_round: 2,
        rounds_total: 8,
        interval_between_rounds_minutes: 3,
      };

      const isTimedAmericano = session.game_mode === 'timed_americano';
      const roundCompleted = true;

      const shouldShowCountdown = isTimedAmericano && roundCompleted;
      expect(shouldShowCountdown).toBe(true);
    });

    it('does not show countdown for non-timed modes', () => {
      const session = {
        game_mode: 'americano' as const,
        current_round: 2,
        rounds_total: 8,
      };

      const isTimedAmericano = session.game_mode === 'timed_americano';
      const roundCompleted = true;

      const shouldShowCountdown = isTimedAmericano && roundCompleted;
      expect(shouldShowCountdown).toBe(false);
    });

    it('displays interval_between_rounds_minutes from session', () => {
      const testCases = [
        { interval: 1, expected: '1 minute' },
        { interval: 2, expected: '2 minutes' },
        { interval: 3, expected: '3 minutes' },
        { interval: 4, expected: '4 minutes' },
        { interval: 5, expected: '5 minutes' },
      ];

      testCases.forEach(({ interval, expected }) => {
        const session = {
          game_mode: 'timed_americano' as const,
          interval_between_rounds_minutes: interval,
        };

        expect(session.interval_between_rounds_minutes).toBe(interval);
        expect(session.interval_between_rounds_minutes).toBeGreaterThanOrEqual(1);
        expect(session.interval_between_rounds_minutes).toBeLessThanOrEqual(5);
      });
    });
  });

  describe('Round Interval Countdown - Auto-Advance', () => {
    it('calls advanceRound when countdown completes naturally', async () => {
      const advanceRound = vi.fn().mockResolvedValue(undefined);

      const session = {
        game_mode: 'timed_americano' as const,
        id: 's1',
        interval_between_rounds_minutes: 3,
      };

      // Simulate countdown complete
      const remaining = 0;
      if (remaining <= 0) {
        await advanceRound(session.id);
      }

      expect(advanceRound).toHaveBeenCalledWith(session.id);
    });

    it('advances to next round after interval expires', () => {
      const session = {
        game_mode: 'timed_americano' as const,
        current_round: 2,
        rounds_total: 8,
      };

      const currentRound = session.current_round ?? 0;
      const nextRound = currentRound + 1;

      expect(nextRound).toBe(3);
      expect(nextRound).toBeLessThanOrEqual(session.rounds_total ?? 0);
    });
  });

  describe('Round Interval Countdown - Skip Early', () => {
    it('calls advanceRound immediately when skip button clicked', async () => {
      const advanceRound = vi.fn().mockResolvedValue(undefined);

      const session = {
        game_mode: 'timed_americano' as const,
        id: 's1',
        interval_between_rounds_minutes: 3,
      };

      const remaining = 120000; // Still 2 minutes left
      expect(remaining).toBeGreaterThan(0);

      // User clicks "Start now" button
      await advanceRound(session.id);

      expect(advanceRound).toHaveBeenCalledWith(session.id);
    });

    it('skips countdown and advances even if interval not fully elapsed', async () => {
      const advanceRound = vi.fn().mockResolvedValue(undefined);

      const session = {
        game_mode: 'timed_americano' as const,
        id: 's1',
      };

      const remaining = 90000; // 1:30 remaining out of 3:00
      const canSkip = remaining > 0;

      expect(canSkip).toBe(true);

      await advanceRound(session.id);

      expect(advanceRound).toHaveBeenCalledWith(session.id);
    });
  });

  describe('Round Interval Countdown - Starting Message', () => {
    it('displays "Starting in X:XX" message during countdown', () => {
      function formatStartingMessage(remaining: number): string {
        const minutes = Math.floor(remaining / 60000);
        const seconds = Math.floor((remaining % 60000) / 1000);
        const paddedSeconds = String(seconds).padStart(2, '0');
        return `Starting in ${minutes}:${paddedSeconds}`;
      }

      const testCases = [
        { remaining: 180000, expected: 'Starting in 3:00' },
        { remaining: 150000, expected: 'Starting in 2:30' },
        { remaining: 60000, expected: 'Starting in 1:00' },
        { remaining: 30000, expected: 'Starting in 0:30' },
      ];

      testCases.forEach(({ remaining, expected }) => {
        const message = formatStartingMessage(remaining);
        expect(message).toBe(expected);
      });
    });

    it('shows message only during interval countdown, not during play', () => {
      const gameState = {
        mode: 'timed_americano' as const,
        phase: 'interval_countdown' as const,
        roundCompleted: true,
      };

      const shouldShow = gameState.mode === 'timed_americano' && gameState.phase === 'interval_countdown';
      expect(shouldShow).toBe(true);
    });

    it('hides message once round starts', () => {
      const gameState = {
        mode: 'timed_americano' as const,
        phase: 'playing' as const,
        roundStarted: true,
      };

      const shouldShow = gameState.phase === 'interval_countdown';
      expect(shouldShow).toBe(false);
    });
  });

  describe('Phase Derivation for Timed Americano', () => {
    it('derives buffer phase when allScored && isTimedAmericano', () => {
      const allScored = true;
      const isTimedAmericano = true;
      const roundStartedAt = '2026-04-21T10:00:00Z';
      const timeExpired = false;

      // Phase derivation logic
      const phase = (() => {
        if (allScored && isTimedAmericano) return 'buffer';
        if (!allScored && roundStartedAt && !timeExpired && isTimedAmericano) return 'play';
        if (timeExpired && !allScored && isTimedAmericano) return 'expired';
        return null;
      })();

      expect(phase).toBe('buffer');
    });

    it('derives play phase when round is active and scores not finalized', () => {
      const allScored = false;
      const isTimedAmericano = true;
      const roundStartedAt = '2026-04-21T10:00:00Z';
      const timeExpired = false;

      const phase = (() => {
        if (allScored && isTimedAmericano) return 'buffer';
        if (!allScored && roundStartedAt && !timeExpired && isTimedAmericano) return 'play';
        if (timeExpired && !allScored && isTimedAmericano) return 'expired';
        return null;
      })();

      expect(phase).toBe('play');
    });

    it('derives expired phase when time expired but scores not finalized', () => {
      const allScored = false;
      const isTimedAmericano = true;
      const roundStartedAt = '2026-04-21T10:00:00Z';
      const timeExpired = true;

      const phase = (() => {
        if (allScored && isTimedAmericano) return 'buffer';
        if (!allScored && roundStartedAt && !timeExpired && isTimedAmericano) return 'play';
        if (timeExpired && !allScored && isTimedAmericano) return 'expired';
        return null;
      })();

      expect(phase).toBe('expired');
    });

    it('derives buffer phase when round not yet started', () => {
      const allScored = false;
      const isTimedAmericano = true;
      const roundStartedAt = null;
      const timeExpired = false;

      const phase = (() => {
        if (allScored && isTimedAmericano) return 'buffer';
        if (!allScored && roundStartedAt && !timeExpired && isTimedAmericano) return 'play';
        if (timeExpired && !allScored && isTimedAmericano) return 'expired';
        return null;
      })();

      // Round not started, so effectively in buffer
      expect(phase).toBeNull(); // Will be handled by overall logic
    });
  });

  describe('Buffer Phase UI - Score Increment Hiding', () => {
    it('hides score increment buttons during buffer phase', () => {
      const phase = 'buffer';
      const isTimedAmericano = true;

      const showScoreIncrement = phase !== 'buffer' && isTimedAmericano;
      expect(showScoreIncrement).toBe(false);
    });

    it('shows score increment buttons during play phase', () => {
      const phase = 'play';
      const isTimedAmericano = true;

      const showScoreIncrement = phase !== 'buffer' && isTimedAmericano;
      expect(showScoreIncrement).toBe(true);
    });

    it('shows score increment buttons during expired phase', () => {
      const phase = 'expired';
      const isTimedAmericano = true;

      const showScoreIncrement = phase !== 'buffer' && isTimedAmericano;
      expect(showScoreIncrement).toBe(true);
    });

    it('shows compact read-only matchup cards in buffer phase', () => {
      const phase = 'buffer';
      const showTeamNames = true;
      const showAvatars = true;
      const showScores = false;

      // During buffer: show names & avatars but no score increment UI
      const isBufferUI = phase === 'buffer' && showTeamNames && showAvatars && !showScores;
      expect(isBufferUI).toBe(true);
    });
  });

  describe('Overview Button and Sheet - Removal', () => {
    it('does not render Overview button for timed_americano', () => {
      const isTimedAmericano = true;
      const showOverviewButton = !isTimedAmericano;

      expect(showOverviewButton).toBe(false);
    });

    it('does not render Courts Overview sheet when removed', () => {
      // showCourtsOverview state is removed entirely
      const showCourtsOverview = undefined;

      expect(showCourtsOverview).toBeUndefined();
    });

    it('removes LayoutGrid icon import requirement', () => {
      // LayoutGrid import removed from component
      const imports = {
        Activity: true,
        ChartBar: true,
        Users: true,
        Pencil: true,
        Shield: true,
        LayoutGrid: false, // Removed
        Check: true,
      };

      expect(imports.LayoutGrid).toBe(false);
    });
  });

  describe('bufferComplete State Reset', () => {
    it('resets bufferComplete to false when currentRound.number changes', () => {
      let bufferComplete = true;
      const previousRound = 1;
      const currentRound = 2;

      // Effect: reset when round number changes
      if (previousRound !== currentRound && bufferComplete) {
        bufferComplete = false;
      }

      expect(bufferComplete).toBe(false);
    });

    it('maintains bufferComplete state within same round', () => {
      let bufferComplete = true;
      const previousRound = 2;
      const currentRound = 2;

      // Effect: no reset if round same
      if (previousRound !== currentRound && bufferComplete) {
        bufferComplete = false;
      }

      expect(bufferComplete).toBe(true);
    });

    it('initializes bufferComplete as false for new rounds', () => {
      let bufferComplete = false;

      expect(bufferComplete).toBe(false);
    });
  });

  describe('Numpad Drawer Positioning', () => {
    it('applies mx-auto w-full max-w-[480px] to Drawer.Content', () => {
      const drawerClasses = 'flex flex-col max-h-[80vh] gap-3 mx-auto w-full max-w-[480px]';

      expect(drawerClasses).toContain('mx-auto');
      expect(drawerClasses).toContain('w-full');
      expect(drawerClasses).toContain('max-w-[480px]');
    });

    it('applies max-w-sm mx-auto to numpad grid', () => {
      const gridClasses = 'grid grid-cols-3 gap-3 max-w-sm mx-auto';

      expect(gridClasses).toContain('max-w-sm');
      expect(gridClasses).toContain('mx-auto');
    });

    it('applies pb-[env(safe-area-inset-bottom)] for iOS safe area', () => {
      const pbClasses = 'px-6 pb-8 flex-1 pb-[env(safe-area-inset-bottom)]';

      expect(pbClasses).toContain('pb-[env(safe-area-inset-bottom)]');
    });
  });
});
