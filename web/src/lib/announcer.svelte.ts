// Full-screen announcements between moments of a match ("Vez de Ana", "A onça pegou
// Bruno!"), shared by every game so they all behave the same: each stays at least
// MIN_HOLD on screen, only a double tap skips it, and the table never sits through a
// backlog. Games push announcements; <Announcements> (components/Announcements.svelte)
// shows them.
import { untrack } from 'svelte';

/** Every announcement stays at least this long, so the whole table can read it. */
export const MIN_HOLD = 5000;

export interface Announcement {
	kicker: string;
	title: string;
	text?: string;
	/** A CSS color for the glow, the kicker and the timer bar. */
	tone?: string;
	/** User ids whose photos are shown. */
	faces?: string[];
	/** Something the game draws above the title (see the `art` snippet). */
	art?: string;
	/** How long it stays, in ms; never less than MIN_HOLD. */
	hold?: number;
	/** When set, the announcement is dropped from the queue once the match is no longer
	 * in this phase (the one already on screen still gets its full time). */
	phase?: string;
}

/** The longest queue kept: older announcements give way to newer ones. */
const MAX_QUEUE = 3;

export class Announcer {
	current = $state<Announcement | null>(null);
	#queue: Announcement[] = [];
	#phase: () => string | undefined;

	/** phase tells which phase the match is in now (to drop stale announcements). */
	constructor(phase: () => string | undefined = () => undefined) {
		this.#phase = phase;
	}

	/** Queues announcements; safe to call from an effect (it tracks nothing). */
	push(...items: Announcement[]) {
		untrack(() => {
			this.#queue = [...this.#queue, ...items].slice(-MAX_QUEUE);
			if (!this.current) this.current = this.#take();
		});
	}

	/** Closes the one on screen and shows the next that still fits the match. */
	next = () => {
		this.current = untrack(() => this.#take());
	};

	#take(): Announcement | null {
		const phase = this.#phase();
		const i = this.#queue.findIndex((a) => !a.phase || a.phase === phase);
		const item = i >= 0 ? this.#queue[i] : null;
		this.#queue = i >= 0 ? this.#queue.slice(i + 1) : [];
		return item;
	}
}
