// Pending friend requests and room invites, kept fresh by the person's own socket.
import { api, type Inbox } from './api';
import { liveSocket, type LiveSocket } from './socket';

// version goes up on every refresh, so pages can reload what the inbox doesn't carry (like the friends list).
export const inbox = $state<Inbox & { version: number }>({ requests: 0, invites: [], version: 0 });

let socket: LiveSocket | null = null;

export async function refreshInbox(): Promise<void> {
	try {
		const fresh = await api.get<Inbox>('/inbox');
		inbox.requests = fresh.requests;
		inbox.invites = fresh.invites;
		inbox.version++;
	} catch {
		/* keep what we had */
	}
}

export function startInbox(): void {
	if (socket) return;
	socket = liveSocket('/api/me/ws', {
		onMessage: (msg) => {
			if (msg.type === 'inbox') refreshInbox();
		}
	});
}

export function stopInbox(): void {
	socket?.close();
	socket = null;
	inbox.requests = 0;
	inbox.invites = [];
}
