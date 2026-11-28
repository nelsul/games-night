// Língua Presa, as the hub sees it: copy, art and screens. The hub finds this file by
// itself (lib/games.ts), so the game never has to edit hub code.
import type { GameInfo, GameModule } from '$lib/games';
import type { Room } from '$lib/api';
import Match from './LinguaPresa.svelte';
import LobbyOptions from './LobbyOptions.svelte';
import Title from './Title.svelte';
import KeyArt from './KeyArt.svelte';
import Narrator from './art/Narrator.svelte';
import Guesser from './art/Guesser.svelte';
import Watcher from './art/Watcher.svelte';

export const info: GameInfo = {
	id: 'linguapresa',
	name: 'Língua Presa',
	creators: ['EmmusSouza'],
	tagline: 'Faça o seu time falar a palavra sem dizer as proibidas',
	tags: 'Times · Palavras · Conversa',
	players: '4–16 jogadores',
	minutes: '~20 min',
	synopsis:
		'Em duplas, trios ou dois times, uma pessoa por vez dá dicas em voz alta para o time descobrir uma palavra. O problema: cada carta tem cinco palavras proibidas, e o time rival está de olho, pronto para apertar a campainha.',
	howTo: [
		'Os times são sorteados ao começar. Cada turno joga um time, e o narrador muda a cada volta.',
		'O narrador vê a palavra e as 5 proibidas, e dá dicas falando. Os parceiros adivinham em voz alta, sem digitar nada. No começo do turno ele tem 10 segundos para ler a carta antes de o relógio correr.',
		'Cada acerto vale 1 ponto. Os outros times veem a carta e apertam "Proibida!" quando ouvem uma palavra proibida: o time perde 1 ponto e a carta é descartada. Depois de cada "Proibida!" a campainha descansa 5 segundos, para ninguém tirar dois pontos pelo mesmo erro.',
		'O narrador pode pular até 3 cartas por turno. Quando o tempo acaba, todo mundo vê como foi, e a vez passa para o próximo time. No fim, o time com mais pontos ganha. Se algo for marcado errado, o host da sala pode ajustar o placar.'
	],
	roles: [
		{
			name: 'Narrador',
			team: 'Time da vez',
			tone: 'good',
			art: Narrator,
			text: 'Vê a palavra e as proibidas e dá dicas falando, sem dizer nenhuma delas.'
		},
		{
			name: 'Parceiro',
			team: 'Time da vez',
			tone: 'good',
			art: Guesser,
			text: 'Não vê a carta. Escuta o narrador e tenta dizer a palavra certa em voz alta.'
		},
		{
			name: 'Fiscal',
			team: 'Os outros times',
			tone: 'bad',
			art: Watcher,
			text: 'Vê a carta e escuta o narrador. Aperta "Proibida!" se ele disser uma palavra proibida.'
		}
	]
};

const game: GameModule = {
	info,
	Title,
	KeyArt,
	Match,
	LobbyOptions,
	whyNotStart(room: Room) {
		const n = room.players.length;
		const mode = ((room.options ?? {}) as { mode?: string }).mode ?? 'two';
		if (mode === 'pairs' && n % 2 !== 0) return 'Duplas precisam de um número par de pessoas.';
		if (mode === 'trios' && (n % 3 !== 0 || n < 6)) return 'Trios precisam de 6, 9, 12 ou 15 pessoas.';
		return null;
	},
	screenshots: [
		{ src: '/games/linguapresa/screens/vez.jpg', caption: 'O narrador começa o turno quando o time está pronto' },
		{ src: '/games/linguapresa/screens/narrador.jpg', caption: 'A palavra, as cinco proibidas e o tempo correndo' },
		{ src: '/games/linguapresa/screens/fiscal.jpg', caption: 'Os outros times veem a carta e apertam Proibida!' },
		{ src: '/games/linguapresa/screens/turno.jpg', caption: 'No fim do turno, todo mundo vê como foi' },
		{ src: '/games/linguapresa/screens/fim.jpg', caption: 'O placar final dos times' }
	]
};

export default game;
