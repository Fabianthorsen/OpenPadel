// See https://svelte.dev/docs/kit/types#app.d.ts
declare global {
	namespace App {
		interface User {
			id: string;
			email: string;
			display_name: string;
			avatar_icon: string;
			avatar_color: string;
			self_rating?: number | null;
			created_at: string;
		}

		// Cross-mode profile headline (ADR 0007). Percentages are 0–100; both are
		// zero when games is zero (the UI hides the stat rather than showing 0%/100%).
		interface CareerSummary {
			games: number;
			winrate: number;
			point_win_pct: number;
			// Placement stats from finishing rank per Session, blended across modes
			// (ADR 0007). titles = rank-1 finishes, podiums = rank ≤ 3, best_finish =
			// lowest rank, average_finish = mean rank. All zero at zero games.
			titles: number;
			podiums: number;
			best_finish: number;
			average_finish: number;
		}

		// Per-Game-Mode career aggregate for the Career Stats page (ADR 0007).
		// Point-share only compares like-for-like within one scoring model, so
		// Americano and Mexicano are never blended here. point_win_pct is 0–100;
		// net_points = total_points − points_conceded and may be negative. Both
		// modes are always present; a mode with no games is zero-valued.
		interface ModeStats {
			mode: 'americano' | 'mexicano';
			games: number;
			wins: number;
			draws: number;
			losses: number;
			total_points: number;
			points_conceded: number;
			net_points: number;
			point_win_pct: number;
			tournaments: number;
		}

		// One fully-scored Match of the cross-mode results series behind the Career
		// Stats recent-form curve (ADR 0007). Ordered oldest-first. result is the
		// match outcome from its point differential; date is the session's date
		// (matches carry no timestamp of their own). Per-match (not per-session)
		// because the match is the metric's atomic Point Win % unit — see the ADR.
		// Point Win % per match is derived from points/conceded client-side.
		interface MatchResult {
			match_id: string;
			mode: 'americano' | 'mexicano';
			date: string;
			points: number;
			conceded: number;
			result: 'win' | 'draw' | 'loss';
		}

		type SessionStatus = 'lobby' | 'playing' | 'done';

		interface ValidationError {
			code: string;
			params?: Record<string, string | number | boolean | null | undefined>;
		}

		interface Session {
			id: string;
			admin_token?: string;
			status: SessionStatus;
			name?: string;
			game_mode: 'americano' | 'mexicano';
			courts: number;
			points: number;
			rounds_total?: number;
			current_round?: number;
			creator_player_id?: string;
			club_id?: string;
			club_name?: string;
			is_creator?: boolean;
			can_start?: boolean;
			validation_errors?: ValidationError[];
			scheduled_at?: string;
			court_duration_minutes?: number;
			ends_at?: string;
			players: Player[];
			created_at: string;
			updated_at: string;
		}

		interface Player {
			id: string;
			session_id: string;
			user_id?: string;
			name: string;
			avatar_icon: string;
			avatar_color: string;
			rating: number;
			added_by_admin: boolean;
			active: boolean;
			joined_at: string;
			// Per-player self-removal secret, returned only in the join response
			// (never in the session listing). Stored client-side to self-leave (#241).
			player_token?: string;
		}

		interface Round {
			id: string;
			session_id: string;
			number: number;
			bench: string[];
			matches: Match[];
		}

		interface Match {
			id: string;
			round_id: string;
			court: number;
			team_a: [string, string];
			team_b: [string, string];
			score: { a: number; b: number } | null;
			live?: { a: number; b: number; server?: 'a' | 'b' };
		}

		interface Standing {
			rank: number;
			player_id: string;
			user_id?: string;
			name: string;
			points: number;
			games_played: number;
			wins: number;
			draws: number;
			avatar_icon: string;
			avatar_color: string;
		}

		interface TournamentEntry {
			session_id: string;
			name: string;
			club_name?: string;
			game_mode?: 'americano' | 'mexicano';
			status: string;
			played_at: string;
			rank: number;
			points: number;
			games_played: number;
			ended_early: boolean;
		}

		interface UpcomingEntry {
			session_id: string;
			name: string;
			club_name?: string;
			status: 'lobby' | 'playing';
			game_mode: 'americano' | 'mexicano';
			courts: number;
			player_count: number;
			scheduled_at?: string;
		}

		interface Invite {
			id: string;
			session_id: string;
			session_name: string;
			from_user_id: string;
			from_display_name: string;
			to_user_id: string;
			to_display_name?: string;
			status: 'pending' | 'accepted' | 'declined';
			created_at: string;
		}

		interface ClubInvite {
			id: string;
			club_id: string;
			club_name: string;
			club_avatar_icon: string;
			club_avatar_color: string;
			inviter_id: string;
			inviter_display_name: string;
			invitee_id: string;
			status: 'pending' | 'accepted' | 'declined';
			created_at: string;
		}

		interface Contact {
			user_id: string;
			display_name: string;
			added_at: string;
		}

		interface UserSearchResult {
			id: string;
			display_name: string;
			is_contact: boolean;
			avatar_icon: string;
			avatar_color: string;
		}

		interface Leaderboard {
			session_id: string;
			status: SessionStatus;
			current_round: number | null;
			total_rounds: number | null;
			standings: Standing[];
			updated_at: string;
		}

		interface Club {
			id: string;
			name: string;
			description: string;
			avatar_icon: string;
			avatar_color: string;
			join_code: string;
			created_by: string;
			created_at: string;
		}

		interface ClubMember {
			user_id: string;
			display_name: string;
			role: string;
			avatar_icon: string;
			avatar_color: string;
			joined_at: string;
		}

		interface ClubListItem {
			id: string;
			name: string;
			avatar_icon: string;
			avatar_color: string;
			my_role: string;
			roster_count: number;
		}

		interface ClubJoinPreview {
			name: string;
			avatar_icon: string;
			avatar_color: string;
			member_count: number;
		}

		interface ClubDetail {
			club: Club;
			members: ClubMember[];
			is_admin: boolean;
			my_role: string;
			roster_count: number;
		}
	}
}

export {};
