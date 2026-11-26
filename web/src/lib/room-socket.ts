// Live room updates over a WebSocket. Every message is a full snapshot of the room,
// so after a reconnect (phone locked, app in the background, network change) the
// page is simply up to date again.
import type { Room } from './api';

const LEFT = 4000; // the server closed us because we left the room

export interface RoomSocket {
	close(): void;
}

export function connectRoom(
	code: string,
	handlers: { onRoom: (room: Room) => void; onStatus: (connected: boolean) => void }
): RoomSocket {
	let ws: WebSocket | null = null;
	let closed = false;
	let retry = 0;
	let timer: ReturnType<typeof setTimeout> | undefined;

	function open() {
		if (closed) return;
		const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
		ws = new WebSocket(`${proto}//${location.host}/api/rooms/${code}/ws`);
		ws.onopen = () => {
			retry = 0;
			handlers.onStatus(true);
		};
		ws.onmessage = (ev) => {
			try {
				const msg = JSON.parse(ev.data);
				if (msg.type === 'room') handlers.onRoom(msg.room as Room);
			} catch {
				/* ignore malformed messages */
			}
		};
		ws.onclose = (ev) => {
			handlers.onStatus(false);
			ws = null;
			if (closed || ev.code === LEFT) return;
			const delay = Math.min(10_000, 500 * 2 ** retry++);
			timer = setTimeout(open, delay);
		};
	}

	// Safari suspends sockets in the background; reconnect right away when the page is back.
	function onVisible() {
		if (document.visibilityState === 'visible' && !ws && !closed) {
			clearTimeout(timer);
			retry = 0;
			open();
		}
	}
	document.addEventListener('visibilitychange', onVisible);
	window.addEventListener('online', onVisible);
	open();

	return {
		close() {
			closed = true;
			clearTimeout(timer);
			document.removeEventListener('visibilitychange', onVisible);
			window.removeEventListener('online', onVisible);
			ws?.close(1000);
		}
	};
}
