// Devaneio, as the hub sees it: copy, art and screens. The hub finds this file by
// itself (lib/games.ts), so the game never has to edit hub code.
import type { GameInfo, GameModule } from '$lib/games';
import Match from './Devaneio.svelte';
import LobbyOptions from './LobbyOptions.svelte';
import Title from './Title.svelte';
import KeyArt from './KeyArt.svelte';
import Teller from './art/Teller.svelte';
import Guesser from './art/Guesser.svelte';

export const info: GameInfo = {
	id: 'devaneio',
	name: 'Devaneio',
	creators: ['EmmusSouza'],
	tagline: 'Uma carta, uma frase e uma mesa tentando adivinhar',
	tags: 'Cartas · Imaginação · Votação',
	players: '4–14 jogadores',
	minutes: '~25 min',
	synopsis:
		'Cada pessoa tem uma mão de cartas desenhadas, cheias de cenas estranhas. A cada rodada, alguém escolhe uma carta e diz uma frase sobre ela. Todo mundo joga a carta que mais combina com a frase, e então a mesa tenta descobrir qual era a carta de verdade.',
	howTo: [
		'A cada rodada, uma pessoa é o contador. Ela escolhe uma carta da mão e escreve uma dica: uma palavra, uma frase, um som, qualquer coisa.',
		'Cada outra pessoa joga, às escondidas, a carta da própria mão que mais combina com a dica. As cartas são embaralhadas e viradas na mesa.',
		'Todo mundo, menos o contador, vota na carta que acha que é a dele. Não vale votar na própria carta.',
		'Se todos ou ninguém acharem a carta certa, o contador não pontua e os outros ganham 2. Senão, o contador e quem acertou ganham 3, e cada carta que enganou alguém dá 1 ponto ao seu dono, até 3. A dica precisa ser boa na medida: nem fácil demais, nem difícil demais.'
	],
	roles: [
		{
			name: 'Contador',
			team: 'A pessoa da vez',
			tone: 'good',
			art: Teller,
			text: 'Escolhe uma carta e escreve uma dica que alguns acertem, mas não todos.'
		},
		{
			name: 'Jogador',
			team: 'Todo mundo',
			tone: 'good',
			art: Guesser,
			text: 'Joga a carta que combina com a dica e tenta achar a carta do contador entre as da mesa.'
		}
	]
};

const game: GameModule = {
	info,
	Title,
	KeyArt,
	Match,
	LobbyOptions,
	screenshots: [
		{ src: '/games/devaneio/screens/historia.jpg', caption: 'O contador escolhe uma carta e escreve a dica' },
		{ src: '/games/devaneio/screens/jogar.jpg', caption: 'Cada pessoa joga a carta que combina com a dica' },
		{ src: '/games/devaneio/screens/votar.jpg', caption: 'As cartas vão para a mesa e todo mundo vota' },
		{ src: '/games/devaneio/screens/revelacao.jpg', caption: 'A revelação mostra quem jogou cada carta' },
		{ src: '/games/devaneio/screens/catorze.jpg', caption: 'Uma mesa com 14 pessoas' },
		{ src: '/games/devaneio/screens/fim.jpg', caption: 'O placar final' }
	]
};

export default game;
