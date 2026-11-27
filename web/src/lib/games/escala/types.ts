// What the server sends for an Escala match (api/games/escala/view.go).
export interface Theme {
	title: string;
	low: string;
	high: string;
}

export interface PlayerView {
	id: string;
	clue?: string;
	has_clue: boolean;
	placed: boolean;
}

export interface Card {
	player: string;
	/** Only after the reveal. */
	number?: number;
	wrong?: boolean;
}

export interface Result {
	theme: Theme;
	ok: boolean;
}

export interface MatchView {
	phase: 'clues' | 'order' | 'reveal' | 'over';
	round: number;
	rounds: number;
	/** Rounds sorted right so far. */
	score: number;
	theme: Theme;
	players: PlayerView[];
	line: Card[];
	results: Result[];
	/** An open vote to change the theme. Votes are public. */
	swap?: { proposer: string; votes: Record<string, boolean> };
	swaps_left: number;
	/** Counts closed votes, so every phone can announce one once. */
	swap_seq: number;
	swap_ok: boolean;
	me: { number: number };
}

export interface Options {
	rounds: number;
}

export const roundChoices = [3, 5, 7];
