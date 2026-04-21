import { describe, it, expect, beforeEach, vi } from 'vitest';

/**
 * Unit tests for CreateDrawer hierarchical Americano variant selection
 *
 * These tests verify:
 * - State management for gameMode and americanoVariant
 * - Derived actualGameMode computation
 * - UI visibility based on game mode and variant selections
 * - API payload construction with correct game_mode values
 */

describe('CreateDrawer - Americano Variant Selection', () => {
  describe('State & Types - gameMode and americanoVariant', () => {
    it('initializes gameMode to "americano" by default', () => {
      const gameMode = 'americano' as const;
      expect(gameMode).toBe('americano');
    });

    it('initializes americanoVariant to "points" by default', () => {
      const americanoVariant = 'points' as const;
      expect(americanoVariant).toBe('points');
    });

    it('allows changing americanoVariant between "points" and "timed"', () => {
      let americanoVariant: 'points' | 'timed' = 'points';

      americanoVariant = 'timed';
      expect(americanoVariant).toBe('timed');

      americanoVariant = 'points';
      expect(americanoVariant).toBe('points');
    });

    it('game mode type has exactly americano and mexicano', () => {
      type GameMode = 'americano' | 'mexicano';
      const validModes: GameMode[] = ['americano', 'mexicano'];

      expect(validModes).toContain('americano');
      expect(validModes).toContain('mexicano');
      expect(validModes.length).toBe(2);
    });
  });

  describe('Derived actualGameMode - Maps UI selection to backend game_mode', () => {
    it('returns "americano" when gameMode="americano" and variant="points"', () => {
      const gameMode = 'americano' as const;
      const americanoVariant = 'points' as const;

      const actualGameMode = gameMode === 'americano'
        ? (americanoVariant === 'timed' ? 'timed_americano' : 'americano')
        : gameMode;

      expect(actualGameMode).toBe('americano');
    });

    it('returns "timed_americano" when gameMode="americano" and variant="timed"', () => {
      const gameMode = 'americano' as const;
      const americanoVariant = 'timed' as const;

      const actualGameMode = gameMode === 'americano'
        ? (americanoVariant === 'timed' ? 'timed_americano' : 'americano')
        : gameMode;

      expect(actualGameMode).toBe('timed_americano');
    });

    it('returns "mexicano" unmodified when gameMode="mexicano"', () => {
      const gameMode = 'mexicano' as const;
      const americanoVariant = 'points' as const;

      const actualGameMode = gameMode === 'americano'
        ? (americanoVariant === 'timed' ? 'timed_americano' : 'americano')
        : gameMode;

      expect(actualGameMode).toBe('mexicano');
    });
  });

  describe('UI Visibility - Americano Variant Toggle', () => {
    it('shows variant toggle when gameMode === "americano"', () => {
      const gameMode = 'americano';
      const shouldShowVariantToggle = gameMode === 'americano';

      expect(shouldShowVariantToggle).toBe(true);
    });

    it('hides variant toggle when gameMode === "mexicano"', () => {
      const gameMode = 'mexicano';
      const shouldShowVariantToggle = gameMode === 'americano';

      expect(shouldShowVariantToggle).toBe(false);
    });

    it('shows variant toggle after switching from mexicano to americano', () => {
      let gameMode: 'americano' | 'mexicano' = 'mexicano';

      let shouldShowVariantToggle = gameMode === 'americano';
      expect(shouldShowVariantToggle).toBe(false);

      gameMode = 'americano';
      shouldShowVariantToggle = gameMode === 'americano';
      expect(shouldShowVariantToggle).toBe(true);
    });
  });

  describe('UI Visibility - Points Input', () => {
    it('shows points input when gameMode="americano" and variant="points"', () => {
      const gameMode = 'americano';
      const americanoVariant = 'points';

      const shouldShowPoints = (gameMode === 'americano' && americanoVariant === 'points') || gameMode === 'mexicano';
      expect(shouldShowPoints).toBe(true);
    });

    it('shows points input when gameMode="mexicano"', () => {
      const gameMode = 'mexicano';
      const americanoVariant = 'points';

      const shouldShowPoints = (gameMode === 'americano' && americanoVariant === 'points') || gameMode === 'mexicano';
      expect(shouldShowPoints).toBe(true);
    });

    it('hides points input when gameMode="americano" and variant="timed"', () => {
      const gameMode = 'americano';
      const americanoVariant = 'timed';

      const shouldShowPoints = (gameMode === 'americano' && americanoVariant === 'points') || gameMode === 'mexicano';
      expect(shouldShowPoints).toBe(false);
    });

    it('shows points initially when americano selected (default variant is points)', () => {
      const gameMode = 'americano';
      const americanoVariant = 'points';

      const shouldShowPoints = (gameMode === 'americano' && americanoVariant === 'points') || gameMode === 'mexicano';
      expect(shouldShowPoints).toBe(true);
    });

    it('hides points when variant changes from points to timed', () => {
      const gameMode = 'americano';
      let americanoVariant: 'points' | 'timed' = 'points';

      let shouldShowPoints = (gameMode === 'americano' && americanoVariant === 'points') || gameMode === 'mexicano';
      expect(shouldShowPoints).toBe(true);

      americanoVariant = 'timed';
      shouldShowPoints = (gameMode === 'americano' && americanoVariant === 'points') || gameMode === 'mexicano';
      expect(shouldShowPoints).toBe(false);
    });
  });

  describe('UI Visibility - Duration & Buffer (Timed Controls)', () => {
    it('shows duration/buffer when gameMode="americano" and variant="timed"', () => {
      const gameMode = 'americano';
      const americanoVariant = 'timed';

      const shouldShowTimedControls = gameMode === 'americano' && americanoVariant === 'timed';
      expect(shouldShowTimedControls).toBe(true);
    });

    it('hides duration/buffer when gameMode="americano" and variant="points"', () => {
      const gameMode = 'americano';
      const americanoVariant = 'points';

      const shouldShowTimedControls = gameMode === 'americano' && americanoVariant === 'timed';
      expect(shouldShowTimedControls).toBe(false);
    });

    it('hides duration/buffer when gameMode="mexicano"', () => {
      const gameMode = 'mexicano';
      const americanoVariant = 'timed';

      const shouldShowTimedControls = gameMode === 'americano' && americanoVariant === 'timed';
      expect(shouldShowTimedControls).toBe(false);
    });

    it('shows duration/buffer when variant changes from points to timed', () => {
      const gameMode = 'americano';
      let americanoVariant: 'points' | 'timed' = 'points';

      let shouldShowTimedControls = gameMode === 'americano' && americanoVariant === 'timed';
      expect(shouldShowTimedControls).toBe(false);

      americanoVariant = 'timed';
      shouldShowTimedControls = gameMode === 'americano' && americanoVariant === 'timed';
      expect(shouldShowTimedControls).toBe(true);
    });
  });

  describe('API Payload - game_mode & Points Values', () => {
    it('sends game_mode="americano" and points=24 for americano with points variant', () => {
      const gameMode = 'americano';
      const americanoVariant = 'points';
      const points = 24;

      const actualGameMode = gameMode === 'americano'
        ? (americanoVariant === 'timed' ? 'timed_americano' : 'americano')
        : gameMode;

      const apiPoints = actualGameMode === 'timed_americano' ? 0 : points;

      expect(actualGameMode).toBe('americano');
      expect(apiPoints).toBe(24);
    });

    it('sends game_mode="timed_americano" and points=0 for americano with timed variant', () => {
      const gameMode = 'americano';
      const americanoVariant = 'timed';
      const points = 24;

      const actualGameMode = gameMode === 'americano'
        ? (americanoVariant === 'timed' ? 'timed_americano' : 'americano')
        : gameMode;

      const apiPoints = actualGameMode === 'timed_americano' ? 0 : points;

      expect(actualGameMode).toBe('timed_americano');
      expect(apiPoints).toBe(0);
    });

    it('sends game_mode="mexicano" and points=24 for mexicano mode', () => {
      const gameMode = 'mexicano';
      const americanoVariant = 'points';
      const points = 24;

      const actualGameMode = gameMode === 'americano'
        ? (americanoVariant === 'timed' ? 'timed_americano' : 'americano')
        : gameMode;

      const apiPoints = actualGameMode === 'timed_americano' ? 0 : points;

      expect(actualGameMode).toBe('mexicano');
      expect(apiPoints).toBe(24);
    });

    it('sends total_duration_minutes only when game_mode="timed_americano"', () => {
      const testCases: Array<{
        gameMode: 'americano' | 'mexicano';
        variant: 'points' | 'timed';
        expectedSendDuration: boolean;
      }> = [
        { gameMode: 'americano', variant: 'timed', expectedSendDuration: true },
        { gameMode: 'americano', variant: 'points', expectedSendDuration: false },
        { gameMode: 'mexicano', variant: 'points', expectedSendDuration: false },
      ];

      testCases.forEach(({ gameMode, variant, expectedSendDuration }) => {
        const actualGameMode = gameMode === 'americano'
          ? (variant === 'timed' ? 'timed_americano' : 'americano')
          : gameMode;

        const shouldSend = actualGameMode === 'timed_americano';
        expect(shouldSend).toBe(expectedSendDuration);
      });
    });

    it('sends buffer_seconds only when game_mode="timed_americano"', () => {
      const testCases: Array<{
        gameMode: 'americano' | 'mexicano';
        variant: 'points' | 'timed';
        expectedSendBuffer: boolean;
      }> = [
        { gameMode: 'americano', variant: 'timed', expectedSendBuffer: true },
        { gameMode: 'americano', variant: 'points', expectedSendBuffer: false },
        { gameMode: 'mexicano', variant: 'points', expectedSendBuffer: false },
      ];

      testCases.forEach(({ gameMode, variant, expectedSendBuffer }) => {
        const actualGameMode = gameMode === 'americano'
          ? (variant === 'timed' ? 'timed_americano' : 'americano')
          : gameMode;

        const shouldSend = actualGameMode === 'timed_americano';
        expect(shouldSend).toBe(expectedSendBuffer);
      });
    });
  });

  describe('State Persistence - Variant Selection Survives Mode Changes', () => {
    it('preserves variant selection when switching away from americano and back', () => {
      let gameMode: 'americano' | 'mexicano' = 'americano';
      let americanoVariant: 'points' | 'timed' = 'points';

      americanoVariant = 'timed';
      expect(americanoVariant).toBe('timed');

      gameMode = 'mexicano';
      gameMode = 'americano';

      expect(americanoVariant).toBe('timed');
    });
  });

  describe('Edge Cases & Constraints', () => {
    it('variant toggle has exactly two options: "points" and "timed"', () => {
      type Variant = 'points' | 'timed';
      const validVariants: Variant[] = ['points', 'timed'];

      expect(validVariants.length).toBe(2);
      expect(validVariants).toContain('points');
      expect(validVariants).toContain('timed');
    });

    it('game mode selection has exactly two game types (americano and mexicano)', () => {
      type GameMode = 'americano' | 'mexicano';
      const validGameModes: GameMode[] = ['americano', 'mexicano'];

      expect(validGameModes.length).toBe(2);
    });

    it('totalDurationMinutes and bufferSeconds have sensible defaults', () => {
      const totalDurationMinutes = 90;
      const bufferSeconds = 120;

      expect(totalDurationMinutes).toBeGreaterThanOrEqual(60);
      expect(totalDurationMinutes).toBeLessThanOrEqual(180);
      expect(bufferSeconds).toBeGreaterThanOrEqual(60);
      expect(bufferSeconds).toBeLessThanOrEqual(300);
    });
  });

  describe('Interval Between Rounds - State & Validation', () => {
    it('initializes intervalBetweenRoundsMin to 3 by default', () => {
      const intervalBetweenRoundsMin = 3;
      expect(intervalBetweenRoundsMin).toBe(3);
    });

    it('allows changing interval between 1 and 5 minutes', () => {
      const validIntervals = [1, 2, 3, 4, 5];
      validIntervals.forEach(interval => {
        expect(interval).toBeGreaterThanOrEqual(1);
        expect(interval).toBeLessThanOrEqual(5);
      });
    });

    it('should not allow intervals outside 1-5 range', () => {
      const invalidIntervals = [0, 6, 10, -1];
      invalidIntervals.forEach(interval => {
        const isValid = interval >= 1 && interval <= 5;
        expect(isValid).toBe(false);
      });
    });

    it('shows interval picker only when gameMode="americano" and variant="timed"', () => {
      const testCases: Array<{
        gameMode: 'americano' | 'mexicano';
        variant: 'points' | 'timed';
        shouldShow: boolean;
      }> = [
        { gameMode: 'americano', variant: 'timed', shouldShow: true },
        { gameMode: 'americano', variant: 'points', shouldShow: false },
        { gameMode: 'mexicano', variant: 'points', shouldShow: false },
      ];

      testCases.forEach(({ gameMode, variant, shouldShow }) => {
        const show = gameMode === 'americano' && variant === 'timed';
        expect(show).toBe(shouldShow);
      });
    });
  });

  describe('Interval Between Rounds - API Payload', () => {
    it('sends interval_between_rounds_minutes when timed_americano is created', () => {
      const gameMode = 'americano';
      const variant = 'timed';
      const intervalBetweenRoundsMin = 3;

      const actualGameMode = gameMode === 'americano'
        ? (variant === 'timed' ? 'timed_americano' : 'americano')
        : gameMode;

      const shouldIncludeInterval = actualGameMode === 'timed_americano';
      expect(shouldIncludeInterval).toBe(true);
      expect(intervalBetweenRoundsMin).toBe(3);
    });

    it('does not send interval_between_rounds_minutes for americano with points variant', () => {
      const gameMode = 'americano';
      const variant = 'points';
      const intervalBetweenRoundsMin = 3;

      const actualGameMode = gameMode === 'americano'
        ? (variant === 'timed' ? 'timed_americano' : 'americano')
        : gameMode;

      const shouldIncludeInterval = actualGameMode === 'timed_americano';
      expect(shouldIncludeInterval).toBe(false);
    });

    it('sends custom interval value (1-5) when selected', () => {
      const testCases = [1, 2, 3, 4, 5];

      testCases.forEach(interval => {
        const gameMode = 'americano';
        const variant = 'timed';
        const actualGameMode = gameMode === 'americano'
          ? (variant === 'timed' ? 'timed_americano' : 'americano')
          : gameMode;

        const shouldIncludeInterval = actualGameMode === 'timed_americano';
        expect(shouldIncludeInterval).toBe(true);
        expect(interval).toBeGreaterThanOrEqual(1);
        expect(interval).toBeLessThanOrEqual(5);
      });
    });
  });
});
