// What the server sends for a Língua Presa match (api/games/linguapresa/view.go).
export interface Team {
	players: string[];
	score: number;
}

export interface Card {
	word: string;
	taboo: string[];
}

export interface Played {
	word: string;
	outcome: 'hit' | 'buzz' | 'skip';
}

export interface MatchView {
	phase: 'ready' | 'turn' | 'summary' | 'over';
	turn: number;
	turns: number;
	teams: Team[];
	/** The team of whoever is looking, or -1. */
	my_team: number;
	/** The team playing this turn, -1 when the match is over. */
	active_team: number;
	narrator: string;
	/** narrator: gives clues. mate: guesses. watcher: sees the card and can buzz. */
	role: 'narrator' | 'mate' | 'watcher';
	seconds: number;
	/** Time left on the clock when this snapshot was made. */
	remaining_ms: number;
	/** Only the narrator and the other teams get the card. */
	card?: Card;
	card_seq: number;
	skips_left: number;
	played: Played[];
	winners?: number[];
}

export interface Options {
	mode: 'two' | 'pairs' | 'trios';
	laps: number;
	seconds: number;
}

export const modeChoices: { id: Options['mode']; label: string; hint: string }[] = [
	{ id: 'two', label: 'Dois times', hint: 'Qualquer número de pessoas, de 4 a 10' },
	{ id: 'pairs', label: 'Duplas', hint: 'Número par de pessoas: 4, 6, 8 ou 10' },
	{ id: 'trios', label: 'Trios', hint: '6, 9, 12 ou 15 pessoas' }
];
export const lapChoices = [1, 2];
export const secondChoices = [45, 60, 90];

/** Every team gets a color and a name, by position. */
/** Eight, enough for 16 people in pairs. */
export const teamColors = ['#ff4fa3', '#ffb347', '#5fd38d', '#35d4ff', '#c58bff', '#ff6b4a', '#5b7cff', '#d4f25a'];
export const teamNames = ['Rosa', 'Dourado', 'Verde', 'Ciano', 'Roxo', 'Coral', 'Azul', 'Lima'];
