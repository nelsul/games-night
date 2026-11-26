// The hub's side of games: what every game module provides, and the registry that
// finds them. A game lives entirely in lib/games/<id>/ and exports its module from
// lib/games/<id>/index.ts; nothing here names a specific game.
import type { Component } from 'svelte';
import { goto } from '$app/navigation';
import { api, type Room } from './api';

export interface Role {
	name: string;
	/** The side as shown to people, e.g. "Time Devs". */
	team: string;
	/** Colors the side: good guys green, bad guys red. */
	tone: 'good' | 'bad';
	text: string;
	art: Component<{ size?: number }>;
}

export interface GameInfo {
	id: string;
	name: string;
	tagline: string;
	tags: string;
	players: string;
	minutes: string;
	synopsis: string;
	howTo: string[];
	roles: Role[];
}

/** Props the hub passes to a game's match screen. `match` is the game's own view. */
export interface MatchProps {
	room: Room;
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	match: any;
	me: string;
	isHost: boolean;
	onroom: (r: Room) => void;
	onchat: () => void;
	onleave: () => void;
	unread?: number;
}

export interface LobbyOptionsProps {
	room: Room;
	isHost: boolean;
	onchange: (r: Room) => void;
}

export interface GameModule {
	info: GameInfo;
	/** The game's wordmark. */
	Title: Component<{ size?: number }>;
	/** Poster art for the billboard and the game page. */
	KeyArt: Component<{ height?: number }>;
	/** The whole match screen. */
	Match: Component<MatchProps>;
	/** Host settings shown in the lobby, if the game has any. */
	LobbyOptions?: Component<LobbyOptionsProps>;
	/** A reason the host can't start yet (beyond the player count), or null. */
	whyNotStart?: (room: Room) => string | null;
}

const found = import.meta.glob<GameModule>('./games/*/index.ts', { eager: true, import: 'default' });

/** Every game, by id. */
export const games: Record<string, GameModule> = Object.fromEntries(Object.values(found).map((g) => [g.info.id, g]));

/** Every game, in name order. */
export const gameList: GameModule[] = Object.values(games).sort((a, b) => a.info.name.localeCompare(b.info.name));

export async function createRoom(game: string): Promise<void> {
	const room = await api.post<Room>('/rooms', { game });
	await goto(`/sala/${room.code}`);
}
