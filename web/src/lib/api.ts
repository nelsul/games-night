// Thin client for the Go API. Errors carry the API's Portuguese message.

export class ApiError extends Error {
	constructor(
		public status: number,
		message: string
	) {
		super(message);
	}
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
	const init: RequestInit = { method, credentials: 'same-origin', headers: {} };
	if (body instanceof FormData) {
		init.body = body;
	} else if (body !== undefined) {
		init.body = JSON.stringify(body);
		(init.headers as Record<string, string>)['Content-Type'] = 'application/json';
	}
	let res: Response;
	try {
		res = await fetch('/api' + path, init);
	} catch {
		throw new ApiError(0, 'Sem conexão. Verifique a internet.');
	}
	const data = await res.json().catch(() => ({}));
	if (!res.ok) {
		// A session that expired mid-use goes back to login. The session check
		// itself (/me) and the login page handle 401 on their own.
		if (res.status === 401 && path !== '/me' && !path.startsWith('/auth') && location.pathname !== '/entrar') {
			location.href = '/entrar';
		}
		throw new ApiError(res.status, data.error ?? 'Algo deu errado.');
	}
	return data as T;
}

export const api = {
	get: <T>(path: string) => request<T>('GET', path),
	post: <T>(path: string, body?: unknown) => request<T>('POST', path, body ?? {}),
	patch: <T>(path: string, body: unknown) => request<T>('PATCH', path, body),
	put: <T>(path: string, body: unknown) => request<T>('PUT', path, body),
	del: <T>(path: string) => request<T>('DELETE', path)
};

export interface PublicUser {
	id: string;
	name: string;
	handle: string;
	avatar_url: string;
}

export type Relation = 'self' | 'friend' | 'request_sent' | 'request_received' | 'none';

export interface UserWithRelation extends PublicUser {
	relation: Relation;
}

export interface FriendRequest {
	id: string;
	user: PublicUser;
	created_at: string;
}

export interface Invite {
	id: string;
	code: string;
	game: Game;
	from: PublicUser;
	players: number;
	created_at: string;
}

export interface Inbox {
	requests: number;
	invites: Invite[];
}

export interface InvitableFriend extends PublicUser {
	state: 'in_room' | 'invited' | 'none';
}

export interface Me extends PublicUser {
	email: string;
	onboarded: boolean;
	photos_enabled: boolean;
}

export interface Game {
	id: string;
	name: string;
	min_players: number;
	max_players: number;
	available: boolean;
}

export type RoomStatus = 'lobby' | 'playing' | 'finished' | 'closed';

export interface RoomPlayer extends PublicUser {
	is_host: boolean;
	online: boolean;
	is_bot: boolean;
}

export interface Room {
	code: string;
	game: Game;
	status: RoomStatus;
	host_id: string;
	created_at: string;
	options: Record<string, unknown>;
	players: RoomPlayer[];
	match: unknown; // the game's view for me, or null outside a match
}

export interface Contributor {
	login: string;
	avatar_url: string;
	url: string;
	contributions: number;
}

export interface Contributors {
	repo: string;
	contributors: Contributor[];
}

export interface RoomSummary {
	code: string;
	game: Game;
	status: RoomStatus;
	is_host: boolean;
	players: number;
	updated_at: string;
}

export interface ChatMessage {
	id: number;
	user: PublicUser;
	body: string;
	created_at: string;
}
