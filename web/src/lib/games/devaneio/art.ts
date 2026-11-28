// The pictures of Devaneio. Every card is a little scene drawn from its number alone, so
// the server only has to know numbers and every phone draws the same picture. A scene is
// a sky, something in the sky, a landscape, one main object and a few small ones, all in
// one of eight palettes. Everything here is original.

export const CARD_COUNT = 160;
export const CARD_W = 200;
export const CARD_H = 300;

type Rand = () => number;

/** A small seeded random generator (mulberry32), so a card always looks the same. */
function seeded(seed: number): Rand {
	let a = (seed * 2654435761 + 1013904223) >>> 0;
	return () => {
		a = (a + 0x6d2b79f5) >>> 0;
		let t = a;
		t = Math.imul(t ^ (t >>> 15), t | 1);
		t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
		return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
	};
}

interface Palette {
	name: string;
	sky1: string;
	sky2: string;
	far: string;
	near: string;
	glow: string;
	accent: string;
	ink: string;
	dark: boolean;
}

const palettes: Palette[] = [
	{ name: 'crepúsculo', sky1: '#2b1b5a', sky2: '#e0567a', far: '#3d2468', near: '#1d1236', glow: '#ffd36e', accent: '#ff9ec4', ink: '#120a28', dark: true },
	{ name: 'maré', sky1: '#0e3b5e', sky2: '#3fb6c9', far: '#145073', near: '#0a2640', glow: '#ffe29a', accent: '#ff8f6b', ink: '#071a2c', dark: false },
	{ name: 'aurora', sky1: '#0b1b3a', sky2: '#2fd6a3', far: '#1a2e57', near: '#0a1226', glow: '#b7ff7a', accent: '#9a7bff', ink: '#060c1c', dark: true },
	{ name: 'deserto', sky1: '#f7b267', sky2: '#f4845f', far: '#c25b3d', near: '#8c3b2f', glow: '#fff1c9', accent: '#5b2a5a', ink: '#3a1a24', dark: false },
	{ name: 'floresta', sky1: '#1f4d3a', sky2: '#a8d672', far: '#2a6b4c', near: '#123226', glow: '#ffe58a', accent: '#ff9f68', ink: '#0b2018', dark: false },
	{ name: 'algodão', sky1: '#ffd1e8', sky2: '#b8c5ff', far: '#a98cf2', near: '#7f6bd6', glow: '#fff6b0', accent: '#ff7fb0', ink: '#3b2a6b', dark: false },
	{ name: 'carvão', sky1: '#1a1a24', sky2: '#4a4a68', far: '#262638', near: '#0c0c14', glow: '#ffcf5a', accent: '#ff6b5a', ink: '#07070c', dark: true },
	{ name: 'verão', sky1: '#7ad7ff', sky2: '#fff3b0', far: '#59c98f', near: '#2aa876', glow: '#ffffff', accent: '#ff7a59', ink: '#1c4a5e', dark: false }
];

const n = (v: number) => Math.round(v * 10) / 10;

/** Something that is drawn around (0,0), standing on y = 0, about 80 wide and 120 tall. */
type Motif = (p: Palette, r: Rand) => string;

const motifs: Record<string, Motif> = {
	torre: (p) => `
		<path d="M-16 0V-70H16V0Z" fill="${p.near}"/>
		<path d="M-22 -70H22L14 -84H-14Z" fill="${p.far}"/>
		<path d="M0 -84V-110" stroke="${p.ink}" stroke-width="2"/>
		<path d="M0 -110l16 5-16 5Z" fill="${p.accent}"/>
		<rect x="-5" y="-58" width="10" height="14" rx="5" fill="${p.glow}"/>
		<rect x="-5" y="-34" width="10" height="14" rx="5" fill="${p.glow}" opacity=".7"/>`,
	barco: (p) => `
		<path d="M-34 -6H34L24 10H-24Z" fill="${p.near}"/>
		<path d="M0 -8V-78" stroke="${p.ink}" stroke-width="3"/>
		<path d="M3 -76L30 -14H3Z" fill="${p.glow}"/>
		<path d="M-3 -62L-24 -14H-3Z" fill="${p.accent}"/>`,
	arvore: (p) => `
		<path d="M-4 0L-3 -40H3L4 0Z" fill="${p.ink}"/>
		<circle cx="0" cy="-60" r="26" fill="${p.far}"/>
		<circle cx="-18" cy="-46" r="16" fill="${p.far}"/>
		<circle cx="18" cy="-48" r="17" fill="${p.far}"/>
		<circle cx="6" cy="-70" r="14" fill="${p.accent}" opacity=".55"/>
		<circle cx="-10" cy="-52" r="3" fill="${p.glow}"/><circle cx="14" cy="-60" r="3" fill="${p.glow}"/>`,
	porta: (p) => `
		<path d="M-26 0V-58A26 26 0 0 1 26 -58V0Z" fill="${p.ink}"/>
		<path d="M-19 0V-56A19 19 0 0 1 19 -56V0Z" fill="${p.glow}"/>
		<path d="M-19 0L0 -30L19 0Z" fill="${p.accent}" opacity=".6"/>
		<circle cx="10" cy="-26" r="2.5" fill="${p.ink}"/>`,
	balao: (p) => `
		<path d="M0 -118C-34 -118 -40 -72 -14 -50L-8 -42H8L14 -50C40 -72 34 -118 0 -118Z" fill="${p.accent}"/>
		<path d="M0 -118C-14 -100 -14 -64 -8 -42M0 -118C14 -100 14 -64 8 -42" stroke="${p.glow}" stroke-width="2" fill="none" opacity=".75"/>
		<path d="M-7 -42L-9 -24M7 -42L9 -24" stroke="${p.ink}" stroke-width="1.5"/>
		<rect x="-11" y="-24" width="22" height="14" rx="2" fill="${p.near}"/>`,
	peixe: (p) => `
		<path d="M-34 -30C-14 -62 22 -62 36 -30C22 2 -14 2 -34 -30Z" fill="${p.accent}"/>
		<path d="M30 -30L52 -50V-10Z" fill="${p.glow}"/>
		<path d="M-6 -56C0 -66 10 -62 14 -56Z" fill="${p.glow}"/>
		<circle cx="-20" cy="-34" r="4" fill="${p.ink}"/>
		<path d="M4 -48C10 -38 10 -22 4 -12" stroke="${p.ink}" stroke-width="2" fill="none" opacity=".4"/>`,
	passaros: (p) => `
		<path d="M-34 -50q9 -14 18 0q9 -14 18 0M6 -78q8 -12 16 0q8 -12 16 0M-6 -22q6 -9 12 0q6 -9 12 0" stroke="${p.ink}" stroke-width="3" fill="none" stroke-linecap="round"/>`,
	chave: (p) => `
		<circle cx="0" cy="-86" r="20" fill="none" stroke="${p.glow}" stroke-width="9"/>
		<circle cx="0" cy="-86" r="5" fill="${p.accent}"/>
		<path d="M0 -66V-4" stroke="${p.glow}" stroke-width="9" stroke-linecap="round"/>
		<path d="M0 -16H16M0 -32H12M0 -48H20" stroke="${p.glow}" stroke-width="8" stroke-linecap="round"/>`,
	lanterna: (p) => `
		<path d="M0 -112V-90" stroke="${p.ink}" stroke-width="2.5"/>
		<path d="M-14 -90H14L18 -86V-82H-18V-86Z" fill="${p.near}"/>
		<path d="M-16 -82H16L12 -44H-12Z" fill="${p.glow}"/>
		<path d="M-12 -44H12L16 -38H-16Z" fill="${p.near}"/>
		<circle cx="0" cy="-62" r="34" fill="${p.glow}" opacity=".22"/>
		<path d="M-4 -76V-52M4 -76V-52" stroke="${p.near}" stroke-width="1.5" opacity=".6"/>`,
	escada: (p) => `
		<path d="M-14 0L-8 -126M14 0L8 -126" stroke="${p.near}" stroke-width="5" stroke-linecap="round"/>
		<path d="M-13 -20H13M-12 -44H12M-11 -68H11M-10 -92H10M-9 -114H9" stroke="${p.glow}" stroke-width="4" stroke-linecap="round"/>`,
	olho: (p) => `
		<path d="M-44 -50C-20 -86 20 -86 44 -50C20 -14 -20 -14 -44 -50Z" fill="${p.glow}"/>
		<circle cx="0" cy="-50" r="19" fill="${p.accent}"/>
		<circle cx="0" cy="-50" r="9" fill="${p.ink}"/>
		<circle cx="6" cy="-56" r="3.5" fill="#fff"/>
		<path d="M-30 -80l-6 -12M0 -86v-14M30 -80l6 -12" stroke="${p.glow}" stroke-width="3" stroke-linecap="round"/>`,
	nuvem: (p) => `
		<circle cx="-20" cy="-30" r="18" fill="${p.glow}"/>
		<circle cx="2" cy="-42" r="24" fill="${p.glow}"/>
		<circle cx="26" cy="-30" r="17" fill="${p.glow}"/>
		<rect x="-38" y="-30" width="80" height="18" rx="9" fill="${p.glow}"/>`,
	casa: (p) => `
		<rect x="-26" y="-44" width="52" height="44" fill="${p.near}"/>
		<path d="M-34 -44L0 -80L34 -44Z" fill="${p.accent}"/>
		<rect x="-8" y="-24" width="16" height="24" rx="8" fill="${p.glow}"/>
		<rect x="12" y="-36" width="9" height="9" fill="${p.glow}"/><rect x="-21" y="-36" width="9" height="9" fill="${p.glow}"/>
		<rect x="14" y="-92" width="8" height="22" fill="${p.far}"/>`,
	pipa: (p) => `
		<path d="M0 -120L22 -92L0 -58L-22 -92Z" fill="${p.accent}"/>
		<path d="M0 -120V-58M-22 -92H22" stroke="${p.glow}" stroke-width="2"/>
		<path d="M0 -58C10 -44 -10 -34 0 -20C8 -8 -6 0 2 6" stroke="${p.ink}" stroke-width="2" fill="none"/>
		<path d="M-6 -32l12 0M-5 -22l10 0" stroke="${p.glow}" stroke-width="3" opacity="0"/>
		<circle cx="2" cy="-30" r="4" fill="${p.glow}"/><circle cx="-1" cy="-14" r="3.5" fill="${p.accent}"/>`,
	guardachuva: (p) => `
		<path d="M-44 -64A44 44 0 0 1 44 -64Q30 -72 22 -64Q11 -72 0 -64Q-11 -72 -22 -64Q-30 -72 -44 -64Z" fill="${p.accent}"/>
		<path d="M0 -108V-64M-24 -98L-14 -64M24 -98L14 -64" stroke="${p.glow}" stroke-width="1.5" opacity=".6" fill="none"/>
		<path d="M0 -64V-8Q0 4 -9 4" stroke="${p.ink}" stroke-width="4" fill="none" stroke-linecap="round"/>`,
	farol: (p) => `
		<path d="M-18 0L-11 -84H11L18 0Z" fill="${p.glow}"/>
		<path d="M-15 -26H15L13 -42H-13ZM-12 -62H12L11 -76H-11Z" fill="${p.accent}"/>
		<rect x="-12" y="-100" width="24" height="16" fill="${p.ink}"/>
		<path d="M-14 -100L0 -114L14 -100Z" fill="${p.accent}"/>
		<circle cx="0" cy="-92" r="5" fill="${p.glow}"/>
		<path d="M10 -92L90 -116V-68Z" fill="${p.glow}" opacity=".28"/>`,
	degraus: (p) => `
		<path d="M-44 0V-14H-22V-30H0V-46H22V-62H44V0Z" fill="${p.near}"/>
		<path d="M-44 -14H-22M-22 -30H0M0 -46H22M22 -62H44" stroke="${p.glow}" stroke-width="3"/>
		<circle cx="52" cy="-86" r="9" fill="${p.glow}"/>`,
	cogumelo: (p) => `
		<path d="M-8 0L-6 -34H6L8 0Z" fill="${p.glow}"/>
		<path d="M-40 -34A40 32 0 0 1 40 -34Z" fill="${p.accent}"/>
		<circle cx="-18" cy="-48" r="6" fill="${p.glow}"/><circle cx="8" cy="-58" r="7" fill="${p.glow}"/><circle cx="24" cy="-44" r="4.5" fill="${p.glow}"/>`,
	coroa: (p) => `
		<path d="M-40 -8L-44 -52L-22 -34L0 -66L22 -34L44 -52L40 -8Z" fill="${p.glow}"/>
		<rect x="-40" y="-12" width="80" height="12" rx="3" fill="${p.accent}"/>
		<circle cx="-44" cy="-54" r="5" fill="${p.accent}"/><circle cx="0" cy="-70" r="5" fill="${p.accent}"/><circle cx="44" cy="-54" r="5" fill="${p.accent}"/>`,
	livro: (p) => `
		<path d="M-44 -6C-26 -22 -10 -20 0 -10C10 -20 26 -22 44 -6V-34C26 -50 10 -48 0 -38C-10 -48 -26 -50 -44 -34Z" fill="${p.glow}"/>
		<path d="M0 -38V-10" stroke="${p.ink}" stroke-width="2"/>
		<path d="M-34 -32C-24 -38 -14 -36 -8 -32M-34 -24C-24 -30 -14 -28 -8 -24M8 -32C14 -36 24 -38 34 -32M8 -24C14 -28 24 -30 34 -24" stroke="${p.ink}" stroke-width="1.5" fill="none" opacity=".5"/>
		<circle cx="0" cy="-70" r="4" fill="${p.glow}"/><circle cx="-16" cy="-58" r="2.5" fill="${p.glow}"/><circle cx="18" cy="-62" r="3" fill="${p.glow}"/>`,
	relogio: (p) => `
		<circle cx="0" cy="-64" r="38" fill="${p.glow}" stroke="${p.ink}" stroke-width="5"/>
		<path d="M0 -64V-88M0 -64L18 -54" stroke="${p.ink}" stroke-width="4" stroke-linecap="round"/>
		<circle cx="0" cy="-64" r="4" fill="${p.accent}"/>
		<path d="M0 -98V-102M0 -26V-30M-34 -64H-30M34 -64H30" stroke="${p.ink}" stroke-width="3"/>`,
	cadeira: (p) => `
		<path d="M-16 -50V-96M16 -50V-96" stroke="${p.near}" stroke-width="6" stroke-linecap="round"/>
		<rect x="-20" y="-96" width="40" height="10" rx="3" fill="${p.accent}"/>
		<rect x="-24" y="-50" width="48" height="9" rx="3" fill="${p.near}"/>
		<path d="M-18 -41L-22 0M18 -41L22 0" stroke="${p.near}" stroke-width="6" stroke-linecap="round"/>`
};
const motifNames = Object.keys(motifs);
/** Motifs that float in the sky instead of standing on the ground. */
const floaters = new Set(['balao', 'peixe', 'passaros', 'chave', 'lanterna', 'olho', 'nuvem', 'pipa', 'coroa', 'livro', 'relogio', 'guardachuva']);

function hills(p: Palette, r: Rand, y: number, fill: string, amp: number): string {
	const pts: string[] = [];
	const steps = 5;
	let d = `M0 ${n(y)}`;
	for (let i = 0; i <= steps; i++) {
		const x = (CARD_W / steps) * i;
		const yy = y + (r() - 0.5) * amp * 2;
		pts.push(`${n(x)} ${n(yy)}`);
	}
	d += ` L${pts[0]}`;
	for (let i = 1; i < pts.length; i++) {
		const [px, py] = pts[i - 1].split(' ').map(Number);
		const [x, yy] = pts[i].split(' ').map(Number);
		const cx = (px + x) / 2;
		d += ` C${n(cx)} ${n(py)} ${n(cx)} ${n(yy)} ${n(x)} ${n(yy)}`;
	}
	d += ` V${CARD_H} H0Z`;
	return `<path d="${d}" fill="${fill}"/>`;
}

function waves(p: Palette, r: Rand, y: number): string {
	let out = `<rect x="0" y="${n(y)}" width="${CARD_W}" height="${CARD_H - y}" fill="${p.far}"/>`;
	for (let i = 0; i < 6; i++) {
		const yy = y + 10 + i * ((CARD_H - y) / 6);
		const x0 = r() * 40 - 20;
		let d = `M${n(x0)} ${n(yy)}`;
		for (let x = x0; x < CARD_W + 20; x += 36) d += ` q9 -6 18 0 t18 0`;
		out += `<path d="${d}" stroke="${p.glow}" stroke-width="1.8" fill="none" opacity="${n(0.18 + r() * 0.2)}"/>`;
	}
	return out;
}

function skyline(p: Palette, r: Rand, y: number): string {
	let out = '';
	let x = -6;
	while (x < CARD_W) {
		const w = 16 + r() * 22;
		const h = 20 + r() * 52;
		out += `<rect x="${n(x)}" y="${n(y - h + 18)}" width="${n(w)}" height="${n(h + CARD_H)}" fill="${p.far}"/>`;
		for (let k = 0; k < 3; k++) {
			if (r() < 0.55) out += `<rect x="${n(x + 3 + r() * (w - 9))}" y="${n(y - h + 26 + k * 12)}" width="4" height="5" fill="${p.glow}" opacity=".8"/>`;
		}
		x += w + 2;
	}
	return out;
}

function dunes(p: Palette, r: Rand, y: number): string {
	return hills(p, r, y, p.far, 12) + hills(p, r, y + 36, p.near, 14);
}

function stars(p: Palette, r: Rand, count: number, maxY: number): string {
	let out = '';
	for (let i = 0; i < count; i++) {
		out += `<circle cx="${n(r() * CARD_W)}" cy="${n(r() * maxY)}" r="${n(0.6 + r() * 1.4)}" fill="${p.glow}" opacity="${n(0.4 + r() * 0.6)}"/>`;
	}
	return out;
}

/** The inside of the SVG for card `id` (a 200 by 300 picture). */
export function cardSvg(id: number): string {
	const r = seeded(id + 1);
	const p = palettes[Math.floor(r() * palettes.length)];
	const uid = `d${id}`;
	const horizon = 150 + Math.floor(r() * 70);
	const landscape = ['colinas', 'mar', 'cidade', 'dunas', 'plano'][Math.floor(r() * 5)];

	let s = `<defs>
		<linearGradient id="${uid}s" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="${p.sky1}"/><stop offset="1" stop-color="${p.sky2}"/></linearGradient>
		<radialGradient id="${uid}g"><stop offset="0" stop-color="${p.glow}" stop-opacity=".9"/><stop offset="1" stop-color="${p.glow}" stop-opacity="0"/></radialGradient>
	</defs>`;
	s += `<rect width="${CARD_W}" height="${CARD_H}" fill="url(#${uid}s)"/>`;
	if (p.dark || r() < 0.25) s += stars(p, r, 14 + Math.floor(r() * 24), horizon - 30);

	// Something big in the sky: a sun, a moon, a ring or nothing.
	const sky = r();
	const cx = 40 + r() * 120;
	const cy = 40 + r() * (horizon - 110);
	if (sky < 0.55) {
		const rad = 16 + r() * 22;
		s += `<circle cx="${n(cx)}" cy="${n(cy)}" r="${n(rad * 2.6)}" fill="url(#${uid}g)"/><circle cx="${n(cx)}" cy="${n(cy)}" r="${n(rad)}" fill="${p.glow}"/>`;
		if (r() < 0.4) s += `<circle cx="${n(cx + rad * 0.45)}" cy="${n(cy - rad * 0.2)}" r="${n(rad * 0.85)}" fill="${p.sky1}" opacity=".9"/>`;
	} else if (sky < 0.75) {
		const rad = 22 + r() * 20;
		s += `<circle cx="${n(cx)}" cy="${n(cy)}" r="${n(rad)}" fill="none" stroke="${p.glow}" stroke-width="3" opacity=".8"/><circle cx="${n(cx)}" cy="${n(cy)}" r="${n(rad * 0.62)}" fill="none" stroke="${p.accent}" stroke-width="2" opacity=".7"/>`;
	}

	// The landscape.
	if (landscape === 'colinas') s += hills(p, r, horizon + 6, p.far, 16) + hills(p, r, horizon + 40, p.near, 14);
	else if (landscape === 'mar') s += waves(p, r, horizon);
	else if (landscape === 'cidade') s += skyline(p, r, horizon) + `<rect x="0" y="${horizon + 24}" width="${CARD_W}" height="${CARD_H}" fill="${p.near}"/>`;
	else if (landscape === 'dunas') s += dunes(p, r, horizon + 4);
	else s += `<rect x="0" y="${horizon}" width="${CARD_W}" height="${CARD_H - horizon}" fill="${p.far}"/><rect x="0" y="${horizon + 50}" width="${CARD_W}" height="${CARD_H}" fill="${p.near}"/>`;

	// The main object: it stands on the ground (or floats, if that fits it better).
	const hero = motifNames[Math.floor(r() * motifNames.length)];
	const ground = landscape === 'mar' ? horizon + 22 : horizon + 40 + r() * 24;
	const scale = 0.95 + r() * 0.55;
	const hx = 62 + r() * 76;
	const hy = floaters.has(hero) ? 150 + r() * 40 : ground;
	const flip = r() < 0.5 ? -1 : 1;
	s += `<g transform="translate(${n(hx)} ${n(hy)}) scale(${n(scale * flip)} ${n(scale)})">${motifs[hero](p, r)}</g>`;

	// One or two small things around it.
	const extras = 1 + Math.floor(r() * 2);
	for (let i = 0; i < extras; i++) {
		const m = motifNames[Math.floor(r() * motifNames.length)];
		if (m === hero) continue;
		const left = (hx > 100) === (i % 2 === 0);
		const x = left ? 22 + r() * 34 : 146 + r() * 34;
		const sc = 0.32 + r() * 0.28;
		const y = floaters.has(m) ? 30 + r() * 110 : ground + 6 + r() * 30;
		s += `<g transform="translate(${n(x)} ${n(y)}) scale(${n(sc)})" opacity="${p.dark ? 0.95 : 0.9}">${motifs[m](p, r)}</g>`;
	}

	// Floating specks, and a soft edge so the card feels like a card.
	for (let i = 0; i < 8; i++) {
		s += `<circle cx="${n(r() * CARD_W)}" cy="${n(r() * CARD_H)}" r="${n(0.8 + r() * 1.8)}" fill="${p.glow}" opacity="${n(0.15 + r() * 0.3)}"/>`;
	}
	s += `<rect width="${CARD_W}" height="${CARD_H}" fill="none" stroke="#00000030" stroke-width="2"/>`;
	return s;
}
