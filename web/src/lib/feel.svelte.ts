// Game feel: haptics, sound effects and motion settings, shared by the hub and games.
//
// Haptics: Android browsers have navigator.vibrate. iPhone Safari doesn't, but since
// iOS 18 toggling a <input type="checkbox" switch> gives a light haptic tick, and only
// inside a tap; events that arrive on their own (a websocket message) can't vibrate there.
// Sounds are synthesized with Web Audio, so there are no files to download.

const KEY = 'gn-feel';

export const feel = $state({ sound: true, haptics: true });

try {
	Object.assign(feel, JSON.parse(localStorage.getItem(KEY) ?? '{}'));
} catch {
	/* private mode or bad JSON: keep the defaults */
}

export function saveFeel() {
	try {
		localStorage.setItem(KEY, JSON.stringify(feel));
	} catch {
		/* not saved; fine */
	}
}

/** True when the person asked their system for less motion. */
export const reducedMotion =
	typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches;

/** A transition duration that respects reduced motion. */
export const ms = (n: number) => (reducedMotion ? 0 : n);

// ---- Haptics ----

type Haptic = 'tap' | 'select' | 'success' | 'warning';
const patterns: Record<Haptic, number | number[]> = {
	tap: 8,
	select: 14,
	success: [12, 60, 18],
	warning: [30, 50, 30, 50, 40]
};

let iosSwitch: HTMLLabelElement | null = null;

function iosTick() {
	if (!iosSwitch) {
		iosSwitch = document.createElement('label');
		iosSwitch.setAttribute('aria-hidden', 'true');
		iosSwitch.style.cssText = 'position:fixed;left:-100px;top:0;width:1px;height:1px;overflow:hidden;opacity:0';
		const input = document.createElement('input');
		input.type = 'checkbox';
		input.setAttribute('switch', '');
		input.tabIndex = -1;
		iosSwitch.appendChild(input);
		document.body.appendChild(iosSwitch);
	}
	iosSwitch.click();
}

export function haptic(kind: Haptic = 'tap') {
	if (!feel.haptics) return;
	if (typeof navigator.vibrate === 'function') {
		navigator.vibrate(patterns[kind]);
	} else {
		iosTick();
		if (kind === 'success' || kind === 'warning') setTimeout(iosTick, 90);
	}
}

// ---- Sounds ----

let ctx: AudioContext | null = null;

function audio(): AudioContext | null {
	if (!feel.sound) return null;
	if (!ctx) {
		const AC = window.AudioContext ?? (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
		if (!AC) return null;
		ctx = new AC();
	}
	if (ctx.state === 'suspended') ctx.resume().catch(() => {});
	return ctx;
}

/** One short note: frequency (Hz), start offset and length (s), wave and volume. */
function note(ac: AudioContext, freq: number, at: number, len: number, type: OscillatorType = 'sine', vol = 0.12, slideTo?: number) {
	const t = ac.currentTime + at;
	const osc = ac.createOscillator();
	const gain = ac.createGain();
	osc.type = type;
	osc.frequency.setValueAtTime(freq, t);
	if (slideTo) osc.frequency.exponentialRampToValueAtTime(slideTo, t + len);
	gain.gain.setValueAtTime(0.0001, t);
	gain.gain.exponentialRampToValueAtTime(vol, t + 0.01);
	gain.gain.exponentialRampToValueAtTime(0.0001, t + len);
	osc.connect(gain).connect(ac.destination);
	osc.start(t);
	osc.stop(t + len + 0.02);
}

export type Sound = 'tap' | 'select' | 'vote' | 'push' | 'deploy' | 'rollback' | 'turn' | 'chat' | 'join' | 'win' | 'lose' | 'reveal';

export function sound(name: Sound) {
	const ac = audio();
	if (!ac) return;
	switch (name) {
		case 'tap':
			note(ac, 520, 0, 0.05, 'triangle', 0.04);
			break;
		case 'select':
			note(ac, 660, 0, 0.07, 'triangle', 0.06);
			break;
		case 'vote':
			note(ac, 440, 0, 0.08, 'square', 0.04);
			note(ac, 660, 0.07, 0.1, 'square', 0.04);
			break;
		case 'push':
			note(ac, 300, 0, 0.18, 'sawtooth', 0.05, 900);
			break;
		case 'deploy':
			[523, 659, 784, 1047].forEach((f, i) => note(ac, f, i * 0.09, 0.22, 'triangle', 0.1));
			break;
		case 'rollback':
			note(ac, 440, 0, 0.35, 'sawtooth', 0.08, 110);
			note(ac, 220, 0.18, 0.4, 'square', 0.05, 80);
			break;
		case 'turn':
			note(ac, 880, 0, 0.12, 'sine', 0.1);
			note(ac, 1175, 0.12, 0.18, 'sine', 0.1);
			break;
		case 'chat':
			note(ac, 1046, 0, 0.06, 'sine', 0.06);
			note(ac, 1318, 0.06, 0.08, 'sine', 0.05);
			break;
		case 'join':
			note(ac, 587, 0, 0.1, 'triangle', 0.07);
			note(ac, 880, 0.08, 0.14, 'triangle', 0.07);
			break;
		case 'reveal':
			note(ac, 220, 0, 0.25, 'sine', 0.08, 440);
			break;
		case 'win':
			[523, 659, 784, 659, 784, 1047].forEach((f, i) => note(ac, f, i * 0.11, 0.25, 'triangle', 0.1));
			break;
		case 'lose':
			[392, 370, 349, 262].forEach((f, i) => note(ac, f, i * 0.2, 0.35, 'sawtooth', 0.06));
			break;
	}
}

/** Feedback for a tap on any button: a tick you feel and, softly, hear. */
export function tapFeedback() {
	haptic('tap');
	sound('tap');
}
