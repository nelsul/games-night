// Mirrors api/games/backdoor/view.go: what one player may see of a match.

export type Role = 'dev' | 'analyst' | 'product_owner' | 'hacker' | 'social_engineer' | 'zero_day' | 'deepfake' | 'freelancer';
export type Phase = 'badges' | 'proposal' | 'review' | 'push' | 'reveal' | 'guess' | 'over';

export interface Options {
	product_owner: boolean;
	zero_day: boolean;
	deepfake: boolean;
	freelancer: boolean;
	timer_seconds: number;
}

export interface Known {
	id: string;
	as: 'hacker' | 'social_engineer' | 'zero_day' | 'deepfake' | 'analyst_candidate';
}

export interface Proposal {
	sprint: number;
	leader: string;
	squad: string[];
	votes: Record<string, boolean>;
	approved: boolean;
}

export interface SprintResult {
	squad: string[];
	backdoors: number;
	rollback: boolean;
}

export interface MatchView {
	phase: Phase;
	sprint: number;
	squad_sizes: number[];
	/** Backdoors that roll back each sprint (1 or 2). */
	backdoors_per_sprint: number[];
	squad_size: number;
	leader: string;
	failed_meetings: number;
	max_failed: number;
	players: { id: string; is_leader: boolean; ready: boolean; voted: boolean; pushed: boolean }[];
	selection: string[];
	results: SprintResult[];
	proposals: Proposal[];
	options: Options;
	phase_started_at: string;
	/** Commits opened so far in the reveal phase (true = backdoor). */
	commits: boolean[];
	commits_total: number;
	backdoors_needed: number;
	social_engineer?: string;
	me: {
		role: Role;
		team: 'devs' | 'hackers';
		knows: Known[];
		ready: boolean;
		vote: boolean | null;
		pushed: boolean;
		in_squad: boolean;
		can_backdoor: boolean;
		is_leader: boolean;
	};
	winner?: 'devs' | 'hackers';
	reason?: 'rollbacks' | 'meetings' | 'guess' | 'guess_missed';
	guess?: string;
	roles?: Record<string, Role>;
}

export type ArtKey = 'dev' | 'security' | 'hacker' | 'social' | 'po' | 'zeroday' | 'deepfake' | 'freelancer';

export const roleInfo: Record<Role, { name: string; team: 'devs' | 'hackers'; text: string; art: ArtKey }> = {
	dev: { name: 'Dev', team: 'devs', art: 'dev', text: 'Entregue 3 sprints. Você só faz push limpo e não sabe em quem confiar.' },
	analyst: {
		name: 'Analista de Segurança',
		team: 'devs',
		art: 'security',
		text: 'Você sabe quem são os hackers. Guie o time sem se entregar: se os devs vencerem, a Engenharia Social vai tentar te achar.'
	},
	product_owner: {
		name: 'Product Owner',
		team: 'devs',
		art: 'po',
		text: 'Você sabe quem é o Analista. Se aparecerem dois nomes, um deles é o Deepfake. Proteja o Analista.'
	},
	hacker: { name: 'Hacker', team: 'hackers', art: 'hacker', text: 'Derrube 3 sprints com push --backdoor sem ser descoberto.' },
	social_engineer: {
		name: 'Engenharia Social',
		team: 'hackers',
		art: 'social',
		text: 'Você é hacker. Se os devs entregarem 3 sprints, você tem um palpite: acerte quem é o Analista e os hackers vencem.'
	},
	zero_day: {
		name: 'Zero-day',
		team: 'hackers',
		art: 'zeroday',
		text: 'Você é hacker e o Analista não te enxerga. Use isso para entrar nos squads.'
	},
	deepfake: {
		name: 'Deepfake',
		team: 'hackers',
		art: 'deepfake',
		text: 'Você é hacker e o Product Owner te vê como um possível Analista. Finja ser o Analista para atrair o palpite errado.'
	},
	freelancer: {
		name: 'Freelancer',
		team: 'hackers',
		art: 'freelancer',
		text: 'Você é hacker, mas trabalha sozinho: não conhece os outros hackers e eles não te conhecem.'
	}
};

export type OptionKey = 'product_owner' | 'zero_day' | 'deepfake' | 'freelancer';

export const optionInfo: { key: OptionKey; name: string; text: string; min: number; needs?: OptionKey; hacker: boolean }[] = [
	{
		key: 'product_owner',
		name: 'Product Owner',
		min: 5,
		hacker: false,
		text: 'Um dev que sabe quem é o Analista de Segurança. Ajuda os devs a proteger o Analista no palpite final.'
	},
	{
		key: 'zero_day',
		name: 'Zero-day',
		min: 7,
		hacker: true,
		text: 'Um hacker invisível para o Analista de Segurança. Ajuda os hackers.'
	},
	{
		key: 'deepfake',
		name: 'Deepfake',
		min: 5,
		hacker: true,
		needs: 'product_owner',
		text: 'Um hacker que o Product Owner vê como um possível Analista: ele recebe dois nomes e não sabe qual é o verdadeiro.'
	},
	{
		key: 'freelancer',
		hacker: true,
		name: 'Freelancer',
		min: 7,
		text: 'Um hacker que trabalha sozinho: não conhece os outros hackers e eles não o conhecem. Ajuda os devs.'
	}
];

export const timerChoices = [
	{ value: 0, label: 'Nenhum' },
	{ value: 60, label: '1 min' },
	{ value: 120, label: '2 min' },
	{ value: 180, label: '3 min' }
];

/** Hackers at each table size (api/games/backdoor/engine.go). */
export const hackersFor: Record<number, number> = { 5: 2, 6: 2, 7: 3, 8: 3, 9: 3, 10: 4, 11: 4, 12: 5, 13: 5, 14: 6, 15: 6, 16: 7 };

/** Why these options can't start with this many people, or null. Mirrors CheckStart. */
export function optionsProblem(o: Partial<Options>, players: number): string | null {
	if ((o.zero_day || o.freelancer) && players < 7) return 'Zero-day e Freelancer precisam de 7 pessoas';
	if (o.deepfake && !o.product_owner) return 'O Deepfake só entra com o Product Owner';
	const special = 1 + [o.zero_day, o.deepfake, o.freelancer].filter(Boolean).length;
	const seats = hackersFor[players];
	if (seats && special > seats) return `Com ${players} pessoas cabem só ${seats} hackers: desligue algum papel de hacker`;
	return null;
}
