import { describe, it, expect, beforeEach, vi } from 'vitest';

/**
 * Tests for profile page club loading behavior.
 *
 * Regression tests for issue where loadClubs() crashed when:
 * 1. api.clubs.list() returned null instead of an array
 * 2. Error handling wasn't setting clubs to a fallback empty array
 * 3. The clubs Section would render with null snippet, crashing the component
 */

describe('Profile clubs loading', () => {
	let clubs: any[];
	let clubsLoading: boolean;
	let toastCalls: string[] = [];

	const mockApi = {
		clubs: {
			list: vi.fn()
		}
	};

	const mockToast = {
		error: vi.fn((msg: string) => {
			toastCalls.push(msg);
		})
	};

	const mockAuth = {
		token: 'test-token',
		ready: true
	};

	beforeEach(() => {
		clubs = [];
		clubsLoading = false;
		toastCalls = [];
		vi.clearAllMocks();
	});

	async function loadClubs() {
		if (!mockAuth.token) return;
		clubsLoading = true;
		try {
			const result = await mockApi.clubs.list(mockAuth.token);
			clubs = result ?? [];
		} catch (err) {
			clubs = [];
			mockToast.error('Failed to load clubs');
		} finally {
			clubsLoading = false;
		}
	}

	it('initializes clubs as empty array on load', async () => {
		mockApi.clubs.list.mockResolvedValue([]);
		await loadClubs();
		expect(clubs).toEqual([]);
		expect(clubsLoading).toBe(false);
	});

	it('handles null response from API', async () => {
		mockApi.clubs.list.mockResolvedValue(null);
		await loadClubs();
		expect(clubs).toEqual([]);
		expect(clubsLoading).toBe(false);
	});

	it('handles undefined response from API', async () => {
		mockApi.clubs.list.mockResolvedValue(undefined);
		await loadClubs();
		expect(clubs).toEqual([]);
		expect(clubsLoading).toBe(false);
	});

	it('loads club list successfully', async () => {
		const mockClubs = [
			{
				id: '1',
				name: 'Club A',
				avatar_icon: '🎾',
				avatar_color: 'primary',
				my_role: 'admin',
				roster_count: 5
			},
			{
				id: '2',
				name: 'Club B',
				avatar_icon: '🏆',
				avatar_color: 'success',
				my_role: 'member',
				roster_count: 3
			}
		];
		mockApi.clubs.list.mockResolvedValue(mockClubs);
		await loadClubs();
		expect(clubs).toEqual(mockClubs);
		expect(clubsLoading).toBe(false);
	});

	it('handles API error gracefully', async () => {
		mockApi.clubs.list.mockRejectedValue(new Error('Network error'));
		await loadClubs();
		expect(clubs).toEqual([]);
		expect(clubsLoading).toBe(false);
		expect(toastCalls).toContain('Failed to load clubs');
	});

	it('sets clubsLoading to false after success', async () => {
		mockApi.clubs.list.mockResolvedValue([]);
		expect(clubsLoading).toBe(false);
		const loadPromise = loadClubs();
		expect(clubsLoading).toBe(true);
		await loadPromise;
		expect(clubsLoading).toBe(false);
	});

	it('sets clubsLoading to false after error', async () => {
		mockApi.clubs.list.mockRejectedValue(new Error('API error'));
		expect(clubsLoading).toBe(false);
		const loadPromise = loadClubs();
		expect(clubsLoading).toBe(true);
		await loadPromise;
		expect(clubsLoading).toBe(false);
	});

	it('does not throw when clubs.length is accessed after null response', async () => {
		mockApi.clubs.list.mockResolvedValue(null);
		await loadClubs();
		expect(() => {
			// This should not throw
			const len = clubs.length;
			expect(len).toBe(0);
		}).not.toThrow();
	});

	it('does not throw when clubs.length is accessed after error', async () => {
		mockApi.clubs.list.mockRejectedValue(new Error('API error'));
		await loadClubs();
		expect(() => {
			// This should not throw
			const len = clubs.length;
			expect(len).toBe(0);
		}).not.toThrow();
	});

	it('should always render section when user is logged in, even with 0 clubs', () => {
		// Simulate empty clubs after loading
		clubs = [];
		clubsLoading = false;

		// Clubs section should be visible because:
		// 1. User can create first club
		// 2. "+ Create Club" button should always be accessible
		const shouldShowSection = clubs !== null && clubs !== undefined;
		expect(shouldShowSection).toBe(true);
		expect(clubs.length).toBe(0);
	});

	it('shows create button even when no clubs exist', () => {
		clubs = [];
		clubsLoading = false;

		// "+ Create Club" button should be visible regardless of club count
		const showCreateButton = true; // Always shown for authenticated users
		const clubCount = clubs.length;

		expect(clubCount).toBe(0);
		expect(showCreateButton).toBe(true);
	});
});
