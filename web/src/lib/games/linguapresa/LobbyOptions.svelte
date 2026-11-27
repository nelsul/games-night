<script lang="ts">
	// The host picks how the table splits into teams, how many laps and how long a turn is.
	import { api, ApiError, type Room } from '$lib/api';
	import { lapChoices, modeChoices, secondChoices, type Options } from './types';

	let { room, isHost, onchange }: { room: Room; isHost: boolean; onchange: (r: Room) => void } = $props();

	const opts = $derived(((room.options ?? {}) as Partial<Options>) ?? {});
	const mode = $derived(opts.mode || 'two');
	const laps = $derived(opts.laps || 1);
	const seconds = $derived(opts.seconds || 60);
	const hint = $derived(modeChoices.find((m) => m.id === mode)?.hint ?? '');
	let error = $state('');
	let busy = $state(false);

	async function save(patch: Partial<Options>) {
		error = '';
		busy = true;
		try {
			onchange(await api.put<Room>(`/rooms/${room.code}/options`, { mode, laps, seconds, ...patch }));
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
		} finally {
			busy = false;
		}
	}
</script>

<section class="opts" aria-labelledby="opts-title">
	<h2 id="opts-title">Regras da partida</h2>
	<p class="desc">Os times são sorteados quando o jogo começa.</p>

	<div class="group">
		<span class="label" id="mode-l">Times</span>
		<div class="row" role="radiogroup" aria-labelledby="mode-l">
			{#each modeChoices as m (m.id)}
				<button type="button" role="radio" aria-checked={mode === m.id} class:on={mode === m.id} disabled={!isHost || busy} onclick={() => save({ mode: m.id })}>{m.label}</button>
			{/each}
		</div>
		<p class="hint">{hint}</p>
	</div>

	<div class="group">
		<span class="label" id="laps-l">Voltas (quantas vezes cada pessoa narra)</span>
		<div class="row" role="radiogroup" aria-labelledby="laps-l">
			{#each lapChoices as n (n)}
				<button type="button" role="radio" aria-checked={laps === n} class:on={laps === n} disabled={!isHost || busy} onclick={() => save({ laps: n })}>{n} {n === 1 ? 'volta' : 'voltas'}</button>
			{/each}
		</div>
	</div>

	<div class="group">
		<span class="label" id="sec-l">Tempo de cada turno</span>
		<div class="row" role="radiogroup" aria-labelledby="sec-l">
			{#each secondChoices as n (n)}
				<button type="button" role="radio" aria-checked={seconds === n} class:on={seconds === n} disabled={!isHost || busy} onclick={() => save({ seconds: n })}>{n} s</button>
			{/each}
		</div>
	</div>
	{#if error}<p class="error" role="alert">{error}</p>{/if}
</section>

<style>
	.opts {
		background: var(--surface);
		border: 1px solid var(--line);
		border-radius: var(--radius);
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	h2 {
		margin: 0;
		font-size: 16px;
	}
	.desc,
	.hint {
		margin: 0;
		color: var(--muted);
		font-size: 14px;
	}
	.group {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.label {
		font-size: 14px;
		font-weight: 600;
	}
	.row {
		display: flex;
		gap: 8px;
	}
	button {
		flex: 1;
		min-height: 44px;
		border-radius: var(--radius);
		border: 1px solid var(--line-strong);
		background: transparent;
		color: var(--text);
		font: inherit;
		font-weight: 600;
	}
	button.on {
		background: #ff4fa3;
		border-color: #ff4fa3;
		color: #2a0a1e;
	}
	button:disabled:not(.on) {
		opacity: 0.6;
	}
	.error {
		color: var(--danger);
		margin: 0;
		font-size: 14px;
	}
</style>
