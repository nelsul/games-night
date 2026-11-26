// What the hub shows about each game: copy and art live here, rules live in the API.
import { goto } from '$app/navigation';
import { api, type Room } from './api';

export interface Role {
	name: string;
	team: 'devs' | 'hackers';
	text: string;
	art: 'dev' | 'security' | 'hacker' | 'social';
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

export const backdoor: GameInfo = {
	id: 'backdoor',
	name: 'Backdoor',
	tagline: 'Tem hacker no squad da Cudia',
	tags: 'Dedução · Blefe · Vida de dev',
	players: '5–10 jogadores',
	minutes: '~30 min',
	synopsis:
		'Os devs da Cudia precisam entregar o app da Samgung. O problema: tem hacker infiltrado no squad, e ninguém sabe quem é. Aprove PRs, desconfie de todo mundo e não caia na engenharia social.',
	howTo: [
		'Cada pessoa recebe um crachá secreto: dev ou hacker. Os hackers sabem quem são os outros hackers.',
		'A cada sprint, o Scrum Master da vez monta um squad e todo mundo aprova ou pede mudanças. Cinco reuniões seguidas sem decisão e os hackers vencem.',
		'Quem está no squad faz um push em segredo. Basta uma backdoor para a sprint levar rollback.',
		'Três sprints com rollback e os hackers vencem. Três deploys e os devs quase vencem: a Engenharia Social ainda tem um palpite para achar o Analista.'
	],
	roles: [
		{ name: 'Dev', team: 'devs', art: 'dev', text: 'Entrega as sprints. Só pode fazer push limpo.' },
		{
			name: 'Analista de Segurança',
			team: 'devs',
			art: 'security',
			text: 'Sabe quem são os hackers. Precisa guiar o time sem dar bandeira.'
		},
		{ name: 'Hacker', team: 'hackers', art: 'hacker', text: 'Conhece os outros hackers. Pode derrubar a sprint.' },
		{
			name: 'Engenharia Social',
			team: 'hackers',
			art: 'social',
			text: 'Se os devs vencerem, tem uma última chance: apontar o Analista.'
		},
		{
			name: 'Product Owner',
			team: 'devs',
			art: 'dev',
			text: 'Opcional. Sabe quem é o Analista e ajuda a protegê-lo.'
		},
		{ name: 'Zero-day', team: 'hackers', art: 'hacker', text: 'Opcional, 7+. Um hacker que o Analista não enxerga.' },
		{
			name: 'Freelancer',
			team: 'hackers',
			art: 'hacker',
			text: 'Opcional, 7+. Hacker que trabalha sozinho: não conhece os outros hackers.'
		}
	]
};

export const games: Record<string, GameInfo> = { backdoor };

export async function createRoom(game: string): Promise<void> {
	const room = await api.post<Room>('/rooms', { game });
	await goto(`/sala/${room.code}`);
}
