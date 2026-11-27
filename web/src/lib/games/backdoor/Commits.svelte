<script lang="ts">
	// The reveal: the Scrum Master opens the sprint's commits one at a time, in a
	// shuffled order, and the whole table watches each one flip. Who pushed what never
	// shows; only how many backdoors made it in.
	import { scale } from 'svelte/transition';
	import { backOut, cubicOut } from 'svelte/easing';
	import { haptic, ms, sound } from '$lib/feel.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import type { MatchView } from './types';

	let {
		match,
		leaderName,
		busy,
		onopen,
		oncontinue
	}: { match: MatchView; leaderName: string; busy: boolean; onopen: () => void; oncontinue: () => void } = $props();

	const opened = $derived(match.commits.length);
	const total = $derived(match.commits_total);
	const backdoors = $derived(match.commits.filter(Boolean).length);
	const done = $derived(opened === total && total > 0);
	const rollback = $derived(backdoors >= match.backdoors_needed);
	const leader = $derived(match.me.is_leader);

	// A short fake hash per card, so the cards look like real commits.
	function hash(i: number) {
		const seed = [...match.leader].reduce((n, c) => Math.imul(n, 31) + c.charCodeAt(0), 0);
		let h = Math.imul(seed + match.sprint * 31 + i + 7, 2654435761) >>> 0;
		h = Math.imul(h ^ (h >>> 15), 2246822507) >>> 0;
		return ((h ^ (h >>> 13)) >>> 0).toString(16).padStart(8, '0').slice(0, 7);
	}

	// The card turning over: it lifts, spins on its vertical axis and lands.
	function flip(_: Element, { duration = 520 } = {}) {
		return {
			duration: ms(duration),
			easing: cubicOut,
			css: (t: number) => `transform: perspective(600px) rotateY(${(1 - t) * 180}deg) scale(${1 + Math.sin(t * Math.PI) * 0.12});`
		};
	}

	// Sounds and haptics for every commit that opens, and the verdict at the end.
	let seen = -1;
	let verdictTimer: ReturnType<typeof setTimeout> | undefined;
	$effect(() => {
		const n = opened;
		if (seen === -1) {
			seen = n; // joined mid-reveal: nothing "just" happened
			return;
		}
		if (n > seen) {
			const bad = match.commits[n - 1];
			sound(bad ? 'push' : 'select');
			haptic(bad ? 'warning' : 'tap');
			if (n === total) {
				const rb = rollback;
				verdictTimer = setTimeout(() => {
					sound(rb ? 'rollback' : 'deploy');
					haptic(rb ? 'warning' : 'success');
				}, ms(650));
			}
		}
		seen = n;
		return () => clearTimeout(verdictTimer);
	});
</script>

<h1>Conferindo os commits</h1>
<p class="muted">
	{#if leader}
		Você é o Scrum Master. Abra os commits um por um, na frente de todo mundo.
	{:else}
		<strong class="hl">{leaderName}</strong> está abrindo os commits um por um.
	{/if}
	{match.backdoors_needed === 1 ? 'Uma backdoor derruba a sprint.' : `Só ${match.backdoors_needed} backdoors derrubam esta sprint.`}
</p>

<ol class="commits" aria-label="Commits da sprint {match.sprint + 1}">
	{#each Array(total) as _, i (i)}
		{@const isOpen = i < opened}
		{@const bad = match.commits[i]}
		{@const next = i === opened}
		<li>
			{#if isOpen}
				<div class="commit open" class:bad in:flip>
					<span class="mono hash">{hash(i)}</span>
					<span class="mono cmd">{bad ? 'push --backdoor' : 'push'}</span>
					<span class="verdict">
						{#if bad}
							<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" aria-hidden="true"><path d="M6 6l12 12M18 6L6 18" /></svg>
							Backdoor
						{:else}
							<Icon name="check" size={20} />Limpo
						{/if}
					</span>
				</div>
			{:else}
				<button
					class="commit closed"
					class:next
					disabled={!leader || !next || busy}
					onclick={onopen}
					aria-label={next && leader ? 'Abrir este commit' : `Commit ${i + 1}, fechado`}
				>
					<span class="mono hash">{hash(i)}</span>
					<span class="mono cmd blur" aria-hidden="true">push ••••••••</span>
					<span class="lock" aria-hidden="true">
						<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><rect x="5" y="11" width="14" height="9" rx="2" /><path d="M8 11V8a4 4 0 0 1 8 0v3" /></svg>
					</span>
				</button>
			{/if}
		</li>
	{/each}
</ol>

<p class="tally muted small" aria-live="polite">
	{opened} de {total} abertos · <strong class:hot={backdoors > 0}>{backdoors} {backdoors === 1 ? 'backdoor' : 'backdoors'}</strong>
</p>

{#if done}
	<div class="stamp" class:bad={rollback} in:scale={{ start: 2.2, duration: ms(420), delay: ms(450), easing: backOut }}>
		<span>{rollback ? 'Rollback' : 'Deploy em produção'}</span>
		<small>Sprint {match.sprint + 1}</small>
	</div>
{/if}

<div class="bottom">
	{#if leader}
		{#if done}
			<button class="btn bd-primary btn-block" disabled={busy} onclick={oncontinue}>Fechar a sprint</button>
		{:else}
			<button class="btn bd-primary btn-block" disabled={busy} onclick={onopen}>
				{opened === 0 ? 'Abrir o primeiro commit' : opened === total - 1 ? 'Abrir o último commit' : 'Abrir o próximo commit'}
			</button>
		{/if}
	{:else}
		<p class="muted center">{done ? `Esperando ${leaderName} fechar a sprint…` : `Esperando ${leaderName} abrir o próximo…`}</p>
	{/if}
</div>

<style>
	/* The match screen's type and buttons (its styles are scoped, so repeated here). */
	h1 {
		margin: 8px 0 0;
		font-size: 32px;
		font-weight: 300;
		letter-spacing: -0.6px;
	}
	p {
		margin: 0;
		line-height: 1.45;
	}
	.muted {
		color: var(--bd-muted);
	}
	.hl {
		color: var(--bd-text);
	}
	.small {
		font-size: 13px;
	}
	.center {
		text-align: center;
	}
	.bottom {
		margin-top: auto;
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding-top: 12px;
	}
	.btn {
		border-radius: 999px;
		height: 54px;
	}
	.bd-primary {
		background: var(--bd-text);
		color: var(--bd-bg);
	}
	.commits {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.commit {
		width: 100%;
		min-height: 60px;
		display: grid;
		grid-template-columns: auto 1fr auto;
		align-items: center;
		gap: 12px;
		padding: 12px 14px;
		border-radius: 12px;
		border: 1px solid var(--bd-line);
		background: var(--bd-surface);
		color: var(--bd-text);
		font: inherit;
		text-align: left;
		backface-visibility: hidden;
	}
	.closed {
		cursor: default;
	}
	.closed:disabled {
		opacity: 1;
	}
	.closed.next {
		border-color: var(--bd-blue);
		animation: wiggle 1.8s ease-in-out infinite;
	}
	.closed.next:not(:disabled) {
		cursor: pointer;
	}
	.mono {
		font-family: var(--mono);
	}
	.hash {
		font-size: 13px;
		color: var(--bd-blue);
	}
	.cmd {
		font-size: 14px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.blur {
		filter: blur(4px);
		color: var(--bd-muted);
	}
	.lock {
		color: var(--bd-muted);
		display: flex;
	}
	.open {
		border-color: color-mix(in srgb, var(--bd-green) 60%, transparent);
		background: color-mix(in srgb, var(--bd-green) 12%, var(--bd-surface));
	}
	.open.bad {
		border-color: color-mix(in srgb, var(--bd-red) 70%, transparent);
		background: color-mix(in srgb, var(--bd-red) 16%, var(--bd-surface));
	}
	.verdict {
		display: flex;
		align-items: center;
		gap: 4px;
		font-weight: 700;
		font-size: 14px;
		color: var(--bd-green);
	}
	.bad .verdict {
		color: var(--bd-red);
	}
	.bad .cmd {
		color: #ffa198;
	}
	.tally {
		margin: 0;
		text-align: center;
	}
	.hot {
		color: var(--bd-red);
	}
	.stamp {
		align-self: center;
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 10px 22px;
		border: 3px solid var(--bd-green);
		border-radius: 10px;
		color: var(--bd-green);
		transform: rotate(-4deg);
		text-transform: uppercase;
		text-align: center;
		white-space: nowrap;
		font-weight: 800;
		font-size: 22px;
		letter-spacing: 1px;
		box-shadow: 0 0 32px color-mix(in srgb, var(--bd-green) 35%, transparent);
	}
	.stamp.bad {
		border-color: var(--bd-red);
		color: var(--bd-red);
		box-shadow: 0 0 32px color-mix(in srgb, var(--bd-red) 40%, transparent);
	}
	.stamp small {
		font-size: 12px;
		letter-spacing: 3px;
		opacity: 0.8;
	}
	@keyframes wiggle {
		0%,
		70%,
		100% {
			transform: none;
		}
		78% {
			transform: rotate(-1deg) scale(1.01);
		}
		86% {
			transform: rotate(1deg) scale(1.01);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.closed.next {
			animation: none;
		}
	}
</style>
