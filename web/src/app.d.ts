// See https://svelte.dev/docs/kit/types#app.d.ts
declare global {
  namespace App {
    interface User {
      id: string;
      email: string;
      display_name: string;
      avatar_icon: string;
      avatar_color: string;
      created_at: string;
    }

    interface AmericanoCareerStats {
      games_played: number;
      wins: number;
      draws: number;
      losses: number;
      total_points: number;
      tournaments: number;
    }

    type SessionStatus = 'lobby' | 'active' | 'complete';

    interface Session {
      id: string;
      admin_token?: string;
      status: SessionStatus;
      name?: string;
      game_mode: 'americano' | 'mexicano' | 'timed_americano';
      courts: number;
      points: number;
      rounds_total?: number;
      current_round?: number;
      creator_player_id?: string;
      is_creator?: boolean;
      scheduled_at?: string;
      court_duration_minutes?: number;
      ends_at?: string;
      players: Player[];
      created_at: string;
      updated_at: string;
      total_duration_minutes?: number;
      buffer_seconds?: number;
      round_duration_seconds?: number;
      round_started_at?: string;
      interval_between_rounds_minutes?: number;
    }

    interface Player {
      id: string;
      session_id: string;
      user_id?: string;
      name: string;
      avatar_icon: string;
      avatar_color: string;
      active: boolean;
      joined_at: string;
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
      status: 'lobby' | 'active';
      game_mode: 'americano' | 'mexicano' | 'timed_americano';
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
  }
}

export {};
