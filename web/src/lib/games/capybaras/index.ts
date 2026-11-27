// Crazy Capybaras, as the hub sees it: copy, art and screens. The hub finds this file by
// itself (lib/games.ts), so the game never has to edit hub code.
import type { GameInfo, GameModule } from '$lib/games';
import Match from './Capybaras.svelte';
import Title from './Title.svelte';
import KeyArt from './KeyArt.svelte';
import RoleJaguar from './art/RoleJaguar.svelte';
import RoleDive from './art/RoleDive.svelte';
import RoleNope from './art/RoleNope.svelte';
import RoleCurrent from './art/RoleCurrent.svelte';
import RoleOrange from './art/RoleOrange.svelte';

export const info: GameInfo = {
	id: 'capybaras',
	name: 'Crazy Capybaras',
	creators: ['nelsul'],
	tagline: 'Tem onça no monte. Boa sorte.',
	tags: 'Cartas · Sorte · Sacanagem',
	players: '2–8 jogadores',
	minutes: '~15 min',
	synopsis:
		'As capivaras estão de boa na lagoa, mas tem onça escondida no monte de cartas. Na sua vez, jogue cartas para se proteger e atrapalhar os outros, e depois compre uma. Comprou a onça sem um Mergulho na mão? Virou janta. A última capivara de pé vence.',
	howTo: [
		'Cada pessoa começa com 7 cartas e um Mergulho. No monte estão as onças: uma a menos que o número de capivaras.',
		'Na sua vez, jogue quantas cartas quiser (ou nenhuma) e termine comprando uma carta do monte.',
		'Comprou uma Onça? Se tiver um Mergulho, você se salva e esconde a onça de volta no monte, onde quiser, sem ninguém ver. Se não tiver, está fora.',
		'Correnteza, Soneca, Binóculo e Redemoinho ajudam a fugir da onça. Pares de capivaras iguais roubam cartas, trincas pedem a carta que você quiser.',
		'Qualquer carta jogada pode levar um Nem Vem de qualquer pessoa nos 3 segundos seguintes. E Nem Vem cancela Nem Vem.'
	],
	roles: [
		{ name: 'Onça', team: 'Perigo', tone: 'bad', art: RoleJaguar, text: 'Comprou sem um Mergulho? Fora do jogo.' },
		{ name: 'Mergulho', team: 'Salva-vidas', tone: 'good', art: RoleDive, text: 'Te salva da onça. Depois você esconde ela de volta no monte, onde quiser.' },
		{ name: 'Nem Vem', team: 'Reação', tone: 'bad', art: RoleNope, text: 'Cancela a carta que alguém acabou de jogar, até fora da sua vez.' },
		{ name: 'Correnteza', team: 'Ação', tone: 'good', art: RoleCurrent, text: 'Acaba sua vez sem comprar. A próxima capivara joga duas vezes.' },
		{ name: 'Capivaras', team: 'Combos', tone: 'good', art: RoleOrange, text: 'De Laranja, de Óculos, DJ, no Ofurô e Surfista. Par rouba uma carta; trinca pede qualquer carta.' }
	]
};

const game: GameModule = {
	info,
	Title,
	KeyArt,
	Match,
	screenshots: [
		{ src: '/games/capybaras/screens/mesa.jpg', caption: 'A mesa: o monte, as onças que faltam e a sua mão' },
		{ src: '/games/capybaras/screens/jogada.jpg', caption: 'Escolha a carta e em quem jogar' },
		{ src: '/games/capybaras/screens/nemvem.jpg', caption: 'Três segundos para alguém gritar Nem Vem' },
		{ src: '/games/capybaras/screens/binoculo.jpg', caption: 'O Binóculo mostra o topo do monte, só para você' },
		{ src: '/games/capybaras/screens/esconder.jpg', caption: 'Escapou da onça? Esconda ela onde quiser' },
		{ src: '/games/capybaras/screens/fim.jpg', caption: 'A última capivara de pé vence' }
	]
};

export default game;
