// What the server sends for a Devaneio match (api/games/devaneio/view.go).
export interface PlayerView {
	id: string;
	score: number;
	/** Has done their part of the phase (played a card, voted). */
	done: boolean;
}

export interface TableCard {
	card: number;
	mine: boolean;
	/** Only after the vote. */
	owner?: string;
	teller?: boolean;
	voters?: string[];
}

export interface MatchView {
	phase: 'story' | 'play' | 'vote' | 'reveal' | 'over';
	round: number;
	rounds: number;
	teller: string;
	clue?: string;
	players: PlayerView[];
	/** Only the cards of whoever is looking. */
	hand: number[];
	table: TableCard[];
	/** The table position the viewer voted for, or -1. */
	my_vote: number;
	/** Points each person got this round, once the vote is scored. */
	gains: Record<string, number>;
	winners?: string[];
}

export interface Options {
	rounds: number;
}

export const roundChoices: { value: number; label: string }[] = [
	{ value: 0, label: 'Uma volta' },
	{ value: 5, label: '5 rodadas' },
	{ value: 8, label: '8 rodadas' }
];
