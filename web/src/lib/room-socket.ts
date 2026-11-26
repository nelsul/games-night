// Live room updates: every message is a full snapshot of the room.
import type { ChatMessage, Room } from './api';
import { liveSocket, type LiveSocket } from './socket';

export type RoomSocket = LiveSocket;

export function connectRoom(
	code: string,
	handlers: { onRoom: (room: Room) => void; onChat: (m: ChatMessage) => void; onStatus: (connected: boolean) => void }
): RoomSocket {
	return liveSocket(`/api/rooms/${code}/ws`, {
		onMessage: (msg) => {
			if (msg.type === 'room') handlers.onRoom(msg.room as Room);
			else if (msg.type === 'chat') handlers.onChat(msg.message as ChatMessage);
		},
		onStatus: handlers.onStatus
	});
}
