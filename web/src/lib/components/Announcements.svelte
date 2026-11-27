<script lang="ts">
	// Shows a game's announcements (lib/announcer.svelte.ts) full screen. The backdrop
	// stays up while one follows another; only the content swaps. A double tap skips
	// (a single tap doesn't, so a stray touch can't), and a bar shows the time left.
	import type { Snippet } from 'svelte';
	import { fade, fly, scale } from 'svelte/transition';
	import { backOut } from 'svelte/easing';
	import { ms } from '$lib/feel.svelte';
	import { MIN_HOLD, type Announcement, type Announcer } from '$lib/announcer.svelte';
	import type { RoomPlayer } from '$lib/api';
	import Avatar from './Avatar.svelte';

	let {
		announcer,
		people,
		art
	}: {
		announcer: Announcer;
		people: Record<string, RoomPlayer>;
		/** Draws an announcement's `art` above its title. */
		art?: Snippet<[Announcement]>;
	} = $props();

	const item = $derived(announcer.current);
	const hold = $derived(Math.max(MIN_HOLD, item?.hold ?? MIN_HOLD));

	// One timer per announcement: reading item makes each new one start its full time,
	// even when it lasts as long as the one before.
	$effect(() => {
		if (!item) return;
		const t = setTimeout(announcer.next, hold);
		return () => clearTimeout(t);
	});

	let lastTap = 0;
	function tap() {
		const now = Date.now();
		if (now - lastTap < 400) {
			lastTap = 0;
			announcer.next();
		} else {
			lastTap = now;
		}
	}
</script>

{#if item}
	<button
		class="announce"
		style="--tone: {item.tone ?? '#4493f8'}"
		onclick={tap}
		aria-live="assertive"
		aria-label="{item.kicker}. {item.title}. {item.text ?? ''} Toque duas vezes para fechar."
		in:fade={{ duration: ms(140) }}
		out:fade={{ duration: ms(220) }}
	>
		{#key item}
			<span class="glow" aria-hidden="true"></span>
			<span class="kicker" in:fly={{ y: -14, duration: ms(260), delay: ms(60) }}>{item.kicker}</span>
			{#if item.art && art}
				<span class="art" in:scale={{ start: 0.5, duration: ms(420), delay: ms(80), easing: backOut }}>{@render art(item)}</span>
			{/if}
			{#if item.faces?.length}
				<span class="faces" aria-hidden="true">
					{#each item.faces as id, i (id)}
						<span class="face" in:scale={{ start: 0.4, duration: ms(320), delay: ms(140 + i * 60), easing: backOut }}>
							<Avatar user={people[id] ?? { id, name: '?', avatar_url: '' }} size={item.faces.length === 1 ? 88 : 44} radius={8} />
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
{/if}

<style>
	.announce {
		position: fixed;
		inset: 0;
		z-index: 70;
		border: none;
		padding: 24px;
		margin: 0;
		background: rgba(3, 6, 9, 0.9);
		backdrop-filter: blur(6px);
		-webkit-backdrop-filter: blur(6px);
		color: #f0f3f6;
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
	.glow {
		position: absolute;
		width: 520px;
		height: 520px;
		border-radius: 50%;
		background: radial-gradient(circle, color-mix(in srgb, var(--tone) 30%, transparent), transparent 65%);
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
	.art {
		position: relative;
		display: flex;
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
		font-size: clamp(32px, 10vw, 50px);
		line-height: 1.05;
		font-weight: 800;
		letter-spacing: -0.5px;
	}
	.text {
		position: relative;
		max-width: 320px;
		font-size: 16px;
		line-height: 1.4;
		color: #9aa4ae;
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
