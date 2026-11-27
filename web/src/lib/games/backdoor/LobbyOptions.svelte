<script lang="ts">
	// The host turns optional roles and the timer on or off; everyone else sees what's on.
	import { api, ApiError, type Room } from '$lib/api';
	import { hackersFor, optionInfo, optionsProblem, timerChoices, type OptionKey, type Options } from './types';

	let { room, isHost, onchange }: { room: Room; isHost: boolean; onchange: (r: Room) => void } = $props();

	const opts = $derived<Options>({
		product_owner: false,
		zero_day: false,
		deepfake: false,
		freelancer: false,
		timer_seconds: 0,
		...((room.options ?? {}) as Partial<Options>)
	});
	const count = $derived(room.players.length);
	const problem = $derived(optionsProblem(opts, count));

	// Turning the Product Owner off takes the Deepfake with it (it only exists for the PO).
	function toggle(key: OptionKey) {
		const on = !opts[key];
		const patch: Partial<Options> = { [key]: on };
		for (const o of optionInfo) if (!on && o.needs === key) patch[o.key] = false;
		save(patch);
	}
	let error = $state('');
	let busy = $state(false);

	async function save(patch: Partial<Options>) {
		error = '';
		busy = true;
		try {
			onchange(await api.put<Room>(`/rooms/${room.code}/options`, { ...opts, ...patch }));
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
		} finally {
			busy = false;
		}
	}
</script>

<section class="opts" aria-labelledby="opts-title">
	<h2 id="opts-title">Regras da partida</h2>

	<div class="always">
		<strong>Sempre na mesa:</strong> Analista de Segurança (vê os hackers) e Engenharia Social (se os devs vencerem, tenta
		adivinhar quem é o Analista).
	</div>

	<ul>
		{#each optionInfo as o (o.key)}
			{@const on = opts[o.key]}
			{@const tooFew = count < o.min}
			{@const locked = !!o.needs && !opts[o.needs]}
			<li>
				<div class="text">
					<span class="name">{o.name}{#if o.min > 5}<span class="min">&nbsp;· 7+ pessoas</span>{/if}{#if o.needs}<span class="min">&nbsp;· com o Product Owner</span>{/if}</span>
					<span class="desc">{o.text}</span>
					{#if on && tooFew}<span class="warn">Precisa de pelo menos {o.min} pessoas para começar.</span>{/if}
				</div>
				{#if isHost}
					<button
						type="button"
						class="switch"
						role="switch"
						aria-checked={on}
						aria-label={o.name}
						disabled={busy || (locked && !on)}
						onclick={() => toggle(o.key)}
					><span></span></button>
				{:else}
					<span class="state" class:on>{on ? 'Ligado' : 'Desligado'}</span>
				{/if}
			</li>
		{/each}
		<li class="timer">
			<div class="text">
				<span class="name">Timer de discussão</span>
				<span class="desc">Contagem regressiva enquanto o Scrum Master monta o squad. Só avisa, não pula a vez.</span>
			</div>
			{#if isHost}
				<div class="seg" role="radiogroup" aria-label="Timer de discussão">
					{#each timerChoices as t (t.value)}
						<button
							type="button"
							role="radio"
							aria-checked={opts.timer_seconds === t.value}
							disabled={busy}
							onclick={() => save({ timer_seconds: t.value })}>{t.label}</button
						>
					{/each}
				</div>
			{:else}
				<span class="state" class:on={opts.timer_seconds > 0}
					>{timerChoices.find((t) => t.value === opts.timer_seconds)?.label ?? 'Nenhum'}</span
				>
			{/if}
		</li>
	</ul>
	<p class="seats">
		Todo papel ligado entra na partida. Com {count} {count === 1 ? 'pessoa' : 'pessoas'} são {hackersFor[count] ?? '?'} hackers, e a
		Engenharia Social já ocupa uma vaga.
	</p>
	{#if problem}<p class="warn" role="status">{problem}.</p>{/if}
	{#if error}<p class="error" role="alert">{error}</p>{/if}
</section>

<style>
	.opts {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	h2 {
		margin: 0;
		font-size: 18px;
	}
	.always {
		font-size: 13px;
		line-height: 1.45;
		color: var(--muted);
		background: var(--surface);
		border-radius: var(--radius);
		padding: 10px 12px;
	}
	.always strong {
		color: #fff;
	}
	ul {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		background: var(--surface);
		border-radius: var(--radius);
	}
	li {
		display: flex;
		gap: 12px;
		align-items: center;
		padding: 12px;
	}
	li + li {
		border-top: 1px solid var(--line);
	}
	li.timer {
		flex-direction: column;
		align-items: stretch;
	}
	.text {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 3px;
	}
	.name {
		font-weight: 600;
		font-size: 15px;
	}
	.min {
		font-weight: 400;
		color: var(--muted);
		font-size: 13px;
	}
	.desc {
		font-size: 13px;
		color: var(--muted);
		line-height: 1.4;
	}
	.seats {
		margin: 4px 0 0;
		font-size: 13px;
		color: var(--muted);
		line-height: 1.4;
	}
	.warn {
		font-size: 13px;
		color: #e8a317;
	}
	.switch {
		flex-shrink: 0;
		width: 52px;
		height: 32px;
		border-radius: 16px;
		border: none;
		background: #4d4d4d;
		position: relative;
		cursor: pointer;
		padding: 0;
	}
	.switch span {
		position: absolute;
		top: 3px;
		left: 3px;
		width: 26px;
		height: 26px;
		border-radius: 50%;
		background: #fff;
		transition: transform 0.15s;
	}
	.switch[aria-checked='true'] {
		background: var(--brand);
	}
	.switch[aria-checked='true'] span {
		transform: translateX(20px);
	}
	.state {
		font-size: 13px;
		color: var(--muted);
		flex-shrink: 0;
	}
	.state.on {
		color: #fff;
		font-weight: 600;
	}
	.seg {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: 6px;
	}
	.seg button {
		height: 40px;
		border-radius: var(--radius);
		border: 1px solid var(--line-strong);
		background: transparent;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
	}
	.seg button[aria-checked='true'] {
		background: #fff;
		color: #141414;
		border-color: #fff;
	}
</style>
