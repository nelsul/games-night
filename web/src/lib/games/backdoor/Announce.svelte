<script lang="ts" module>
	export interface Announcement {
		/** The phase it belongs to: once the match has moved past it, it is dropped. */
		phase: string;
		kicker: string;
		title: string;
		text?: string;
		tone?: 'turn' | 'good' | 'bad';
		faces?: string[];
		/** How long it stays, in ms (at least MIN_HOLD; a double tap closes it sooner). */
		hold?: number;
	}
</script>

<script lang="ts">
	// A full-screen card between phases ("Vez de Ana montar o squad"), so a change of
	// turn is an event instead of a silent swap of screens. It stays at least 5 seconds,
	// so everyone at the table can read it; a double tap closes it sooner (a single tap
	// doesn't, so a stray touch can't skip it).
	import { fade, fly, scale } from 'svelte/transition';
	import { backOut } from 'svelte/easing';
	import { ms } from '$lib/feel.svelte';
	import type { RoomPlayer } from '$lib/api';
	import Avatar from '$lib/components/Avatar.svelte';

	let { item, people, onclose }: { item: Announcement; people: Record<string, RoomPlayer>; onclose: () => void } = $props();

	const MIN_HOLD = 5000;
	const hold = $derived(Math.max(MIN_HOLD, item.hold ?? MIN_HOLD));

	// Two taps within this long close it.
	let lastTap = 0;
	function tap() {
		const now = Date.now();
		if (now - lastTap < 400) onclose();
		lastTap = now;
	}
	// One timer per announcement: reading item makes a new one start its own full time,
	// even when it lasts as long as the one before.
	$effect(() => {
		void item;
		const t = setTimeout(onclose, hold);
		return () => clearTimeout(t);
	});
</script>

<button
	class="announce {item.tone ?? ''}"
	onclick={tap}
	aria-live="assertive"
	aria-label="{item.kicker}. {item.title}. {item.text ?? ''} Toque duas vezes para fechar."
	in:fade={{ duration: ms(140) }}
	out:fade={{ duration: ms(220) }}
>
	{#key item}
		<span class="glow" aria-hidden="true"></span>
		<span class="kicker" in:fly={{ y: -14, duration: ms(260), delay: ms(60) }}>{item.kicker}</span>
		{#if item.faces?.length}
			<span class="faces" aria-hidden="true">
				{#each item.faces as id, i (id)}
					<span class="face" in:scale={{ start: 0.4, duration: ms(320), delay: ms(140 + i * 60), easing: backOut }}>
						<Avatar user={people[id]} size={item.faces.length === 1 ? 88 : 44} radius={8} />
					</span>
				{/each}
			</span>
		{/if}
		<strong class="title" in:scale={{ start: 0.7, duration: ms(360), delay: ms(120), easing: backOut }}>{item.title}</strong>
		{#if item.text}<span class="text" in:fly={{ y: 12, duration: ms(260), delay: ms(260) }}>{item.text}</span>{/if}
		<span class="hint" aria-hidden="true">Toque duas vezes para pular</span>
		<span class="bar" aria-hidden="true" style="animation-duration: {hold}ms"></span>
	{/key}
</button>

<style>
	.announce {
		--tone: #4493f8;
		position: fixed;
		inset: 0;
		z-index: 70;
		border: none;
		padding: 24px;
		margin: 0;
		background: rgba(1, 4, 9, 0.88);
		backdrop-filter: blur(6px);
		-webkit-backdrop-filter: blur(6px);
		color: #e6edf3;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 14px;
		text-align: center;
		cursor: pointer;
		font: inherit;
		overflow: hidden;
		/* No double-tap zoom on phones: two taps skip the announcement. */
		touch-action: manipulation;
		user-select: none;
		-webkit-user-select: none;
	}
	.announce.good {
		--tone: #3fb950;
	}
	.announce.bad {
		--tone: #f85149;
	}
	.announce.turn {
		--tone: #d29922;
	}
	.glow {
		position: absolute;
		width: 520px;
		height: 520px;
		border-radius: 50%;
		background: radial-gradient(circle, color-mix(in srgb, var(--tone) 28%, transparent), transparent 65%);
		animation: pulse 1.6s ease-in-out infinite;
		pointer-events: none;
	}
	.kicker {
		position: relative;
		font-size: 13px;
		font-weight: 700;
		letter-spacing: 3px;
		text-transform: uppercase;
		color: var(--tone);
	}
	.faces {
		position: relative;
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 6px;
		max-width: 320px;
	}
	.face {
		display: flex;
		border-radius: 8px;
		box-shadow: 0 0 0 3px var(--tone);
	}
	.title {
		position: relative;
		font-size: clamp(34px, 11vw, 52px);
		line-height: 1.05;
		font-weight: 800;
		letter-spacing: -0.5px;
	}
	.text {
		position: relative;
		max-width: 320px;
		font-size: 16px;
		line-height: 1.4;
		color: #9198a1;
	}
	.hint {
		position: absolute;
		left: 0;
		right: 0;
		bottom: calc(18px + var(--safe-bottom));
		font-size: 12px;
		color: #6e7681;
	}
	.bar {
		position: absolute;
		left: 0;
		bottom: 0;
		height: 4px;
		width: 100%;
		background: var(--tone);
		transform-origin: left;
		animation: drain linear forwards;
	}
	@keyframes drain {
		from {
			transform: scaleX(1);
		}
		to {
			transform: scaleX(0);
		}
	}
	@keyframes pulse {
		0%,
		100% {
			transform: scale(0.9);
			opacity: 0.7;
		}
		50% {
			transform: scale(1.05);
			opacity: 1;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.glow {
			animation: none;
		}
	}
</style>
