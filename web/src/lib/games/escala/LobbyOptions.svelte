<script lang="ts">
	// The host picks how many rounds the table plays; everyone else sees the choice.
	import { api, ApiError, type Room } from '$lib/api';
	import { roundChoices, type Options } from './types';

	let { room, isHost, onchange }: { room: Room; isHost: boolean; onchange: (r: Room) => void } = $props();

	const rounds = $derived(((room.options ?? {}) as Partial<Options>).rounds || 3);
	let error = $state('');
	let busy = $state(false);

	async function save(n: number) {
		error = '';
		busy = true;
		try {
			onchange(await api.put<Room>(`/rooms/${room.code}/options`, { rounds: n }));
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
		} finally {
			busy = false;
		}
	}
</script>

<section class="opts" aria-labelledby="opts-title">
	<h2 id="opts-title">Regras da partida</h2>
	<p class="desc">
		Todo mundo joga junto, no mesmo time. O grupo joga todas as rodadas e no fim vê quantas acertou.
	</p>
	<div class="row" role="radiogroup" aria-label="Número de rodadas">
		{#each roundChoices as n (n)}
			<button
				type="button"
				role="radio"
				aria-checked={rounds === n}
				class:on={rounds === n}
				disabled={!isHost || busy}
				onclick={() => save(n)}>{n} rodadas</button
			>
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
		background: #ffb347;
		border-color: #ffb347;
		color: #2b1636;
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
