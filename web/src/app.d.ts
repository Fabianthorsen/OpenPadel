// See https://svelte.dev/docs/kit/types#app.d.ts
declare global {
  namespace App {
    interface User {
      id: string;
      email: string;
      display_name: string;
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

    interface TennisCareerStats {
      tournaments: number;
      wins: number;
      losses: number;
    }

    type SessionStatus = 'lobby' | 'active' | 'complete';

    interface Session {
      id: string;
      admin_token?: string;
      status: SessionStatus;
      name?: string;
      game_mode: 'americano' | 'mexicano' | 'tennis';
      sets_to_win: number;
      games_per_set: number;
      courts: number;
      points: number;
      rounds_total?: number;
      current_round?: number;
      creator_player_id?: string;
      scheduled_at?: string;
      players: Player[];
      created_at: string;
      updated_at: string;
    }

    interface TennisTeam {
      player_id: string;
      name: string;
      team: 'a' | 'b';
    }

    interface TennisState {
      sets: [number, number][];
      games_a: number;
      games_b: number;
      points_a: number;
      points_b: number;
      in_tiebreak: boolean;
      tiebreak_a: number;
      tiebreak_b: number;
      server: 'a' | 'b' | '';
      winner: 'a' | 'b' | '';
    }

    interface TennisMatch {
      id: string;
      session_id: string;
      state: TennisState;
      teams: { a: TennisTeam[]; b: TennisTeam[] };
      created_at: string;
      updated_at: string;
    }

    interface Player {
      id: string;
      session_id: string;
      user_id?: string;
      name: string;
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
      game_mode: 'americano' | 'mexicano' | 'tennis';
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
