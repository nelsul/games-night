// Mirrors api/games/capybaras/view.go: what one player may see of a match.

export interface Card {
	id: number;
	kind: string;
}

export interface PlayerView {
	id: string;
	cards: number;
	out: boolean;
	turns: number;
}

export interface PendingView {
	player: string;
	kinds: string[];
	effect: 'nap' | 'current' | 'peek' | 'swirl' | 'cuddle' | 'pair' | 'trio';
	target?: string;
	ask?: string;
	nopes: number;
	deadline: string;
}

export interface GameEvent {
	seq: number;
	type: string;
	player?: string;
	target?: string;
	kinds?: string[];
	card?: { id?: number; kind: string };
	ok?: boolean;
}

export interface MatchView {
	phase: 'play' | 'defuse' | 'favor' | 'over';
	players: PlayerView[];
	turn: string;
	turns_left: number;
	deck: number;
	jaguars: number;
	discard: Card[];
	discard_len: number;
	pending?: PendingView;
	favor?: { from: string; to: string };
	hiding: boolean;
	events: GameEvent[];
	winner?: string;
	out: string[];
	server_time: string;
	kinds: string[];
	me: { hand: Card[]; peek: Card[]; peek_seq: number; out: boolean; can_nope: boolean };
}

export interface CardInfo {
	name: string;
	text: string;
	color: string;
}

export const cardInfo: Record<string, CardInfo> = {
	jaguar: { name: 'Onça', color: '#c97a16', text: 'Comprou a onça sem um Mergulho? Você está fora.' },
	dive: { name: 'Mergulho', color: '#2a9fd6', text: 'Te salva da onça. Depois você esconde ela de volta no monte.' },
	nope: { name: 'Nem Vem', color: '#e5383b', text: 'Cancela a carta que alguém acabou de jogar. Vale até fora da sua vez.' },
	current: { name: 'Correnteza', color: '#1f7fb3', text: 'Acaba sua vez sem comprar. A próxima capivara joga 2 vezes.' },
	nap: { name: 'Soneca', color: '#6c63d9', text: 'Acaba uma vez sua sem comprar carta.' },
	peek: { name: 'Binóculo', color: '#2e9e5b', text: 'Veja as 3 cartas do topo do monte, só você.' },
	swirl: { name: 'Redemoinho', color: '#2b5f8a', text: 'Embaralha o monte.' },
	cuddle: { name: 'Cafuné', color: '#d94f8a', text: 'Alguém escolhe uma carta da mão e te dá.' },
	orange: { name: 'Capivara de Laranja', color: '#ff9f1c', text: 'Sozinha não faz nada. Par rouba uma carta; trinca pede a carta que quiser.' },
	shades: { name: 'Capivara de Óculos', color: '#444b5a', text: 'Sozinha não faz nada. Par rouba uma carta; trinca pede a carta que quiser.' },
	dj: { name: 'Capivara DJ', color: '#9b3dd1', text: 'Sozinha não faz nada. Par rouba uma carta; trinca pede a carta que quiser.' },
	spa: { name: 'Capivara no Ofurô', color: '#1fa3a3', text: 'Sozinha não faz nada. Par rouba uma carta; trinca pede a carta que quiser.' },
	surf: { name: 'Capivara Surfista', color: '#3d7be5', text: 'Sozinha não faz nada. Par rouba uma carta; trinca pede a carta que quiser.' }
};

export const capys = ['orange', 'shades', 'dj', 'spa', 'surf'];
export const actions = ['current', 'nap', 'peek', 'swirl', 'cuddle'];
