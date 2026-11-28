// Escala, as the hub sees it: copy, art and screens. The hub finds this file by
// itself (lib/games.ts), so the game never has to edit hub code.
import type { GameInfo, GameModule } from '$lib/games';
import Match from './Escala.svelte';
import LobbyOptions from './LobbyOptions.svelte';
import Title from './Title.svelte';
import KeyArt from './KeyArt.svelte';
import Player from './art/Player.svelte';

export const info: GameInfo = {
	id: 'escala',
	name: 'Escala',
	creators: ['EmmusSouza'],
	tagline: 'Todo mundo no mesmo time, cada um com um número',
	tags: 'Cooperativo · Dicas · Conversa',
	players: '3–16 jogadores',
	minutes: '~15 min',
	synopsis:
		'Cada pessoa recebe um número secreto de 1 a 100 e um tema. Vocês precisam se ordenar do menor para o maior sem dizer o número, só com uma dica criativa. Dá para confiar na dica do amigo?',
	howTo: [
		'Cada rodada tem um tema, como "tamanho de animais". O 1 é o menor extremo e o 100 é o maior.',
		'Todo mundo escreve uma dica que mostre o tamanho do seu número naquele tema, sem falar números.',
		'Conversem e coloquem as cartas em fila, do menor para o maior. Ninguém vê o número dos outros.',
		'Antes de revelar, dá para trocar a sua dica e mover a sua carta. Ao revelar, cada rodada com a fila certa vale um ponto para o time. O jogo é cooperativo: ninguém perde, o grupo só tenta acertar mais rodadas.'
	],
	roles: [
		{
			name: 'Jogador',
			team: 'Todo mundo junto',
			tone: 'good',
			art: Player,
			text: 'Guarda um número secreto, dá uma dica criativa e ajuda o time a montar a fila certa.'
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
		{ src: '/games/escala/screens/tema.jpg', caption: 'O tema da rodada e o seu número secreto' },
		{ src: '/games/escala/screens/dicas.jpg', caption: 'Cada pessoa escreve uma dica sem falar o número' },
		{ src: '/games/escala/screens/fila.jpg', caption: 'A mesa monta a fila do menor para o maior' },
		{ src: '/games/escala/screens/revelacao.jpg', caption: 'As cartas viram uma a uma e mostram quem acertou' },
		{ src: '/games/escala/screens/fim.jpg', caption: 'No fim, quantas rodadas o time acertou' }
	]
};

export default game;
