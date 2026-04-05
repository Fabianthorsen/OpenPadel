// See https://svelte.dev/docs/kit/types#app.d.ts
declare global {
  namespace App {
    interface User {
      id: string;
      email: string;
      display_name: string;
      created_at: string;
    }

    type SessionStatus = 'lobby' | 'active' | 'complete';

    interface Session {
      id: string;
      admin_token?: string;
      status: SessionStatus;
      name?: string;
      courts: number;
      points: number;
      rounds_total?: number;
      current_round?: number;
      creator_player_id?: string;
      players: Player[];
      created_at: string;
      updated_at: string;
    }

    interface Player {
      id: string;
      session_id: string;
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
    }

    interface Standing {
      rank: number;
      player_id: string;
      name: string;
      points: number;
      games_played: number;
      wins: number;
      draws: number;
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
