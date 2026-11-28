<script lang="ts">
	// The host picks how long the match is; everyone else sees the choice.
	import { api, ApiError, type Room } from '$lib/api';
	import { roundChoices, type Options } from './types';

	let { room, isHost, onchange }: { room: Room; isHost: boolean; onchange: (r: Room) => void } = $props();

	const rounds = $derived(((room.options ?? {}) as Partial<Options>).rounds ?? 5);
	let error = $state('');
	let busy = $state(false);

	async function save(value: number) {
		error = '';
		busy = true;
		try {
			onchange(await api.put<Room>(`/rooms/${room.code}/options`, { rounds: value }));
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
		} finally {
			busy = false;
		}
	}
</script>

<section class="opts" aria-labelledby="opts-title">
	<h2 id="opts-title">Regras da partida</h2>
	<p class="desc">Uma pessoa por rodada conta a história de uma carta. Em "uma volta", todo mundo conta uma vez.</p>
	<div class="row" role="radiogroup" aria-label="Tamanho da partida">
		{#each roundChoices as c (c.value)}
			<button type="button" role="radio" aria-checked={rounds === c.value} class:on={rounds === c.value} disabled={!isHost || busy} onclick={() => save(c.value)}>{c.label}</button>
		{/each}
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
		gap: 10px;
	}
	h2 {
		margin: 0;
		font-size: 16px;
	}
	.desc {
		margin: 0;
		color: var(--muted);
		font-size: 14px;
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
		background: #ff8f6b;
		border-color: #ff8f6b;
		color: #2a0f08;
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
