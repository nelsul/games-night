// Backdoor, as the hub sees it: copy, art and screens. The hub finds this file by
// itself (lib/games/registry.ts), so the game never has to edit hub code.
import type { GameInfo, GameModule } from '$lib/games';
import Match from './Backdoor.svelte';
import LobbyOptions from './LobbyOptions.svelte';
import Title from './Title.svelte';
import KeyArt from './art/KeyArt.svelte';
import Dev from './art/Dev.svelte';
import Security from './art/Security.svelte';
import Hacker from './art/Hacker.svelte';
import Social from './art/Social.svelte';
import ProductOwner from './art/ProductOwner.svelte';
import ZeroDay from './art/ZeroDay.svelte';
import Freelancer from './art/Freelancer.svelte';

export const info: GameInfo = {
	id: 'backdoor',
	name: 'Backdoor',
	creators: ['nelsul'],
	tagline: 'Tem hacker no squad da Cudia',
	tags: 'Dedução · Blefe · Vida de dev',
	players: '5–12 jogadores',
	minutes: '~30 min',
	synopsis:
		'Os devs da Cudia precisam entregar o app da Samgung. O problema: tem hacker infiltrado no squad, e ninguém sabe quem é. Aprove PRs, desconfie de todo mundo e não caia na engenharia social.',
	howTo: [
		'Cada pessoa recebe um crachá secreto: dev ou hacker. Os hackers sabem quem são os outros hackers.',
		'A cada sprint, o Scrum Master da vez monta um squad e todo mundo aprova ou pede mudanças. Cinco reuniões seguidas sem decisão e os hackers vencem.',
		'Quem está no squad faz um push em segredo. Depois o Scrum Master abre os commits um por um, embaralhados, na frente de todos. Basta uma backdoor para a sprint levar rollback.',
		'Três sprints com rollback e os hackers vencem. Três deploys e os devs quase vencem: a Engenharia Social ainda tem um palpite para achar o Analista.'
	],
	roles: [
		{ name: 'Dev', team: 'Time Devs', tone: 'good', art: Dev, text: 'Entrega as sprints. Só pode fazer push limpo.' },
		{
			name: 'Analista de Segurança',
			team: 'Time Devs', tone: 'good',
			art: Security,
			text: 'Sabe quem são os hackers. Precisa guiar o time sem dar bandeira.'
		},
		{ name: 'Hacker', team: 'Time Hackers', tone: 'bad', art: Hacker, text: 'Conhece os outros hackers. Pode derrubar a sprint.' },
		{
			name: 'Engenharia Social',
			team: 'Time Hackers', tone: 'bad',
			art: Social,
			text: 'Se os devs vencerem, tem uma última chance: apontar o Analista.'
		},
		{
			name: 'Product Owner',
			team: 'Time Devs', tone: 'good',
			art: ProductOwner,
			text: 'Opcional. Sabe quem é o Analista e ajuda a protegê-lo.'
		},
		{ name: 'Zero-day', team: 'Time Hackers', tone: 'bad', art: ZeroDay, text: 'Opcional, 7+. Um hacker que o Analista não enxerga.' },
		{
			name: 'Freelancer',
			team: 'Time Hackers', tone: 'bad',
			art: Freelancer,
			text: 'Opcional, 7+. Hacker que trabalha sozinho: não conhece os outros hackers.'
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
		{ src: '/games/backdoor/screens/cracha.jpg', caption: 'Seu crachá secreto, só enquanto você segura' },
		{ src: '/games/backdoor/screens/review.jpg', caption: 'O squad vira um PR: Approve ou Request changes' },
		{ src: '/games/backdoor/screens/push.jpg', caption: 'Votos revelados e o push em segredo' },
		{ src: '/games/backdoor/screens/postmortem.jpg', caption: 'Post-mortem: quem venceu e todos os crachás' }
	],
	// Zero-day and Freelancer need 7 people; the server checks it too.
	whyNotStart: (room) =>
		(room.options?.zero_day === true || room.options?.freelancer === true) && room.players.length < 7
			? 'Zero-day e Freelancer precisam de 7 pessoas'
			: null
};

export default game;
