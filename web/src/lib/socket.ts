// A WebSocket that keeps itself open. Messages are full snapshots or "refetch" pokes,
// so after a reconnect (phone locked, app in the background, network change) the
// page simply catches up.

const LEFT = 4000; // the server closed us on purpose (we left the room): don't come back

export interface LiveSocket {
	close(): void;
}

export function liveSocket(
	path: string,
	handlers: { onMessage: (msg: { type: string; [k: string]: unknown }) => void; onStatus?: (connected: boolean) => void }
): LiveSocket {
	let ws: WebSocket | null = null;
	let closed = false;
	let retry = 0;
	let timer: ReturnType<typeof setTimeout> | undefined;

	function open() {
		if (closed) return;
		const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
		ws = new WebSocket(`${proto}//${location.host}${path}`);
		ws.onopen = () => {
			retry = 0;
			handlers.onStatus?.(true);
		};
		ws.onmessage = (ev) => {
			try {
				handlers.onMessage(JSON.parse(ev.data));
			} catch {
				/* ignore malformed messages */
			}
		};
		ws.onclose = (ev) => {
			handlers.onStatus?.(false);
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
