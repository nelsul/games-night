<script lang="ts">
	// A Devaneio match. Everything shown comes from the server's view for this player;
	// this file never decides a rule. Only your own hand is ever sent to your phone, and
	// during the vote nobody can tell whose card is whose until the round is scored.
	import { onDestroy } from 'svelte';
	import { flip } from 'svelte/animate';
	import { fly, scale } from 'svelte/transition';
	import { cubicOut, backOut } from 'svelte/easing';
	import Title from './Title.svelte';
	import CardArt from './CardArt.svelte';
	import { feel, haptic, ms, saveFeel, sound } from '$lib/feel.svelte';
	import { api, ApiError, type Room } from '$lib/api';
	import Avatar from '$lib/components/Avatar.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import type { MatchView } from './types';

	let {
		room,
		match,
		me,
		isHost,
		onroom,
		onchat,
		onleave,
		unread = 0,
		onplayer
	}: {
		room: Room;
		match: MatchView;
		me: string;
		isHost: boolean;
		onroom: (r: Room) => void;
		onchat: () => void;
		onleave: () => void;
		unread?: number;
		onplayer?: (userId: string) => void;
	} = $props();

	const motion = (n: number) => ms(n);
	const people = $derived(Object.fromEntries(room.players.map((p) => [p.id, p])));
	const name = (id: string) => people[id]?.name ?? 'Alguém';
	const first = (id: string) => name(id).split(' ')[0];
	const iAmTeller = $derived(match.teller === me);
	const myState = $derived(match.players.find((p) => p.id === me));
	const waiting = $derived(match.players.filter((p) => !p.done));
	const ranking = $derived([...match.players].sort((a, b) => b.score - a.score));

	let error = $state('');
	let busy = $state(false);
	let confirmStop = $state(false);
	let picked = $state<number | null>(null);
	let clue = $state('');
	let prevKey = '';

	// A new phase or round starts with nothing picked.
	$effect(() => {
		const key = `${match.round}:${match.phase}`;
		if (key !== prevKey) {
			picked = null;
			if (match.phase === 'story') clue = '';
			prevKey = key;
		}
	});

	async function send(action: Record<string, unknown>) {
		error = '';
		busy = true;
		try {
			onroom(await api.post<Room>(`/rooms/${room.code}/action`, action));
			if (action.type === 'story' || action.type === 'play' || action.type === 'vote') {
				sound('select');
				haptic('select');
			}
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
			haptic('warning');
		} finally {
			busy = false;
		}
	}

	async function stop() {
		confirmStop = false;
		try {
			onroom(await api.post<Room>(`/rooms/${room.code}/stop`));
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
		}
	}

	function toggleSound() {
		feel.sound = !feel.sound;
		saveFeel();
		if (feel.sound) sound('select');
	}

	function pick(c: number) {
		picked = picked === c ? null : c;
		sound('tap');
		haptic('tap');
	}

	// ---- What just happened ----
	let prevPhase = '';
	$effect(() => {
		const phase = match.phase;
		if (prevPhase) {
			if (phase === 'play' || phase === 'story') {
				if (phase === 'story') {
					sound('turn');
					haptic('tap');
				}
			} else if (phase === 'vote') {
				sound('vote');
				haptic('tap');
			} else if (phase === 'reveal') {
				sound('reveal');
			} else if (phase === 'over') {
				sound('win');
				haptic('success');
			}
		}
		prevPhase = phase;
	});
	onDestroy(() => {});

	const phaseTitle = $derived(
		match.phase === 'story'
			? iAmTeller
				? 'Conte a sua história'
				: `${first(match.teller)} está escolhendo`
			: match.phase === 'play'
				? iAmTeller
					? 'Todo mundo está escolhendo'
					: myState?.done
						? 'Carta jogada'
						: 'Jogue uma carta'
				: match.phase === 'vote'
					? iAmTeller
						? 'A mesa está votando'
						: 'Qual é a carta de ' + first(match.teller) + '?'
					: match.phase === 'reveal'
						? 'A carta de ' + first(match.teller) + ' era...'
						: ''
	);
</script>

<div class="dv">
	<header class="bar">
		<a href="/" aria-label="Voltar ao início"><Icon name="back" /></a>
		<Title size={26} />
		<span class="bar-actions">
			<button class="pill" onclick={toggleSound} aria-label={feel.sound ? 'Desligar sons' : 'Ligar sons'} aria-pressed={feel.sound}>
				<Icon name={feel.sound ? 'sound' : 'mute'} size={18} />
			</button>
			<button class="pill" onclick={onchat} aria-label={unread ? `Chat, ${unread} novas` : 'Chat'}>
				<Icon name="chat" size={18} />{#if unread}<span class="badge" aria-hidden="true">{unread}</span>{/if}
			</button>
		</span>
	</header>

	<!-- Everyone, with their score; a tick means they did their part -->
	<section class="board" class:compact={match.players.length > 8} aria-label="Placar">
		{#each match.players as p (p.id)}
			<button class="who" class:teller={p.id === match.teller && match.phase !== 'over'} class:me={p.id === me} onclick={() => onplayer?.(p.id)} aria-label="{name(p.id)}, {p.score} pontos">
				<span class="who-face">
					<Avatar user={people[p.id]} size={match.players.length > 8 ? 30 : 40} />
					{#if p.done && match.phase !== 'story' && match.phase !== 'reveal' && match.phase !== 'over'}
						<span class="tick" in:scale={{ duration: motion(220), easing: backOut }}><Icon name="check" size={11} /></span>
					{/if}
				</span>
				<span class="who-name">{first(p.id)}</span>
				{#key p.score}
					<b class="who-score" in:scale={{ start: 1.5, duration: motion(260), easing: backOut }}>{p.score}</b>
				{/key}
				{#if match.phase === 'reveal' && match.gains[p.id]}
					<span class="gain" in:fly={{ y: 8, duration: motion(260) }}>+{match.gains[p.id]}</span>
				{/if}
			</button>
		{/each}
	</section>

	{#if match.phase !== 'over'}
		<p class="roundline small center">Rodada {match.round + 1} de {match.rounds} · contador: <strong>{iAmTeller ? 'você' : name(match.teller)}</strong></p>
	{/if}

	{#key match.phase + ':' + match.round}
		<div class="phase" in:fly={{ y: 16, duration: motion(300), easing: cubicOut }}>
			{#if match.phase !== 'over'}
				<h1 class="title">{phaseTitle}</h1>
			{/if}

			{#if match.clue && match.phase !== 'story'}
				<blockquote class="clue">“{match.clue}”</blockquote>
			{/if}

			{#if match.phase === 'story' || match.phase === 'play'}
				{#if match.phase === 'story' && iAmTeller}
					<p class="small center">Escolha uma carta e escreva uma dica. Boa é a que alguns acertam, mas não todos.</p>
				{:else if match.phase === 'story'}
					<p class="small center">Enquanto isso, veja a sua mão.</p>
				{:else if iAmTeller}
					<p class="small center">Esperando os outros jogarem uma carta. Faltam {waiting.length}.</p>
				{:else if !myState?.done}
					<p class="small center">Escolha a carta da sua mão que mais combina com a dica.</p>
				{:else}
					<p class="small center">Esperando os outros. Faltam {waiting.length}.</p>
				{/if}

				<div class="hand" role="listbox" aria-label="Sua mão">
					{#each match.hand as c (c)}
						{@const can = (match.phase === 'story' && iAmTeller) || (match.phase === 'play' && !iAmTeller && !myState?.done)}
						<button
							class="hcard"
							class:picked={picked === c}
							class:dim={!can}
							role="option"
							aria-selected={picked === c}
							disabled={!can || busy}
							onclick={() => pick(c)}
							animate:flip={{ duration: motion(220) }}
						>
							<CardArt id={c} label="Carta {c + 1}" />
						</button>
					{/each}
				</div>

				{#if match.phase === 'story' && iAmTeller}
					<form
						class="clue-form"
						onsubmit={(e) => {
							e.preventDefault();
							if (picked !== null) send({ type: 'story', card: picked, text: clue });
						}}
					>
						<label class="sr" for="clue">Sua dica</label>
						<input id="clue" bind:value={clue} maxlength="60" placeholder="Ex.: saudade de casa" autocomplete="off" />
						<button class="btn primary block" disabled={busy || picked === null || !clue.trim()}>
							{picked === null ? 'Toque numa carta' : 'Contar a história'}
						</button>
					</form>
				{:else if match.phase === 'play' && !iAmTeller && !myState?.done}
					<button class="btn primary block" disabled={busy || picked === null} onclick={() => picked !== null && send({ type: 'play', card: picked })}>
						{picked === null ? 'Toque numa carta' : 'Jogar esta carta'}
					</button>
				{/if}
			{:else if match.phase === 'vote'}
				{#if iAmTeller}
					<p class="small center">Você não vota. Esperando {waiting.length} {waiting.length === 1 ? 'pessoa' : 'pessoas'}.</p>
				{:else if match.my_vote >= 0}
					<p class="small center">Voto enviado. Esperando {waiting.length}.</p>
				{:else}
					<p class="small center">Toque na carta que você acha que é a dele. Não vale a sua.</p>
				{/if}
				<div class="table">
					{#each match.table as t, i (i)}
						{@const voted = match.my_vote === i}
						<button
							class="tcard"
							class:picked={picked === i || voted}
							class:own={t.mine}
							disabled={t.mine || iAmTeller || match.my_vote >= 0 || busy}
							onclick={() => (picked = picked === i ? null : i)}
							in:fly={{ y: 18, duration: motion(260), delay: motion(50 * i) }}
							aria-label="Carta {i + 1}{t.mine ? ', a sua' : ''}"
						>
							<CardArt id={t.card} />
							{#if t.mine}<span class="flag">Sua</span>{/if}
							{#if voted}<span class="flag vote">Seu voto</span>{/if}
						</button>
					{/each}
				</div>
				{#if !iAmTeller && match.my_vote < 0}
					<button class="btn primary block" disabled={busy || picked === null} onclick={() => picked !== null && send({ type: 'vote', index: picked })}>
						{picked === null ? 'Toque numa carta' : 'Votar nesta carta'}
					</button>
				{/if}
			{:else if match.phase === 'reveal'}
				<div class="table">
					{#each match.table as t, i (i)}
						<div class="tcard shown" class:truth={t.teller} in:fly={{ y: 18, duration: motion(260), delay: motion(70 * i) }}>
							<CardArt id={t.card} />
							{#if t.voters?.length}
								<span class="voters" aria-label="Votaram nesta carta">
									{#each t.voters as v (v)}<span class="voter"><Avatar user={people[v]} size={24} /></span>{/each}
								</span>
							{/if}
							{#if t.owner}
								<span class="tag" class:truth={t.teller}>
									<Avatar user={people[t.owner]} size={26} />
									<span class="tag-text">
										<strong>{t.owner === me ? 'Você' : first(t.owner)}</strong>
										<small>{t.teller ? 'contador' : 'jogou'}</small>
									</span>
									{#if match.gains[t.owner]}<em>+{match.gains[t.owner]}</em>{/if}
								</span>
							{/if}
						</div>
					{/each}
				</div>
				<button class="btn primary block" disabled={busy} onclick={() => send({ type: 'next' })}>
					{match.round + 1 >= match.rounds ? 'Ver o resultado' : 'Próxima rodada'}
				</button>
			{:else}
				<section class="panel over">
					<span class="kicker">Fim de jogo</span>
					<h1>
						{match.winners && match.winners.length > 1
							? 'Empate!'
							: `${match.winners?.[0] === me ? 'Você' : first(match.winners?.[0] ?? '')} ${match.winners?.[0] === me ? 'ganhou' : 'ganhou'}`}
					</h1>
					<ol class="rank">
						{#each ranking as p, k (p.id)}
							<li class:win={match.winners?.includes(p.id)} in:fly={{ y: 14, duration: motion(260), delay: motion(70 * k) }}>
								<span class="pos">{k + 1}</span>
								<Avatar user={people[p.id]} size={34} />
								<span class="rank-name">{p.id === me ? 'Você' : name(p.id)}</span>
								<b>{p.score}</b>
							</li>
						{/each}
					</ol>
					{#if isHost}
						<button class="btn primary block" onclick={stop}>Jogar de novo</button>
					{:else}
						<p class="small center">Esperando o host começar outra partida.</p>
					{/if}
					<button class="btn ghost block" onclick={onleave}><Icon name="door" size={18} />Sair da sala</button>
				</section>
			{/if}
		</div>
	{/key}

	{#if error}<p class="error" role="alert" in:fly={{ y: 8, duration: motion(200) }}>{error}</p>{/if}

	{#if isHost && match.phase !== 'over'}
		{#if confirmStop}
			<p class="small center">Encerrar a partida para todo mundo?</p>
			<div class="two">
				<button class="btn ghost" onclick={() => (confirmStop = false)}>Cancelar</button>
				<button class="btn primary" onclick={stop}>Encerrar</button>
			</div>
		{:else}
			<button class="link" onclick={() => (confirmStop = true)}>Encerrar partida</button>
		{/if}
	{/if}
</div>

<style>
	.dv {
		--gold: #ffd36e;
		--coral: #ff8f6b;
		--ink: #08222b;
		position: relative;
		overflow-x: clip;
		background:
			radial-gradient(120% 50% at 50% -10%, #1f6a78 0%, transparent 60%),
			linear-gradient(180deg, #06161c, #08222b 60%, #0c2f3b);
		color: #fff1d6;
		min-height: 100dvh;
		max-width: 560px;
		margin: 0 auto;
		padding: calc(10px + var(--safe-top)) 14px calc(24px + var(--safe-bottom));
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	:global(body:has(.dv)) {
		background: #06161c;
	}
	.phase {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.bar {
		height: 48px;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	.bar-actions {
		display: flex;
		gap: 6px;
	}
	.bar a {
		min-width: 44px;
		min-height: 44px;
		display: grid;
		place-items: center;
		color: inherit;
	}
	.pill {
		position: relative;
		min-width: 44px;
		min-height: 44px;
		display: grid;
		place-items: center;
		border-radius: 999px;
		border: 1px solid #2a5a68;
		background: transparent;
		color: inherit;
		transition: transform 0.12s, background 0.2s;
	}
	.pill:active {
		transform: scale(0.92);
		background: #12404d;
	}
	.badge {
		position: absolute;
		top: 2px;
		right: 2px;
		background: var(--gold);
		color: var(--ink);
		border-radius: 999px;
		font-size: 11px;
		font-weight: 700;
		padding: 1px 5px;
	}

	/* ---- the players ---- */
	.board {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		justify-content: center;
	}
	.who {
		position: relative;
		width: 52px;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1px;
		padding: 4px 0 2px;
		border: 2px solid transparent;
		border-radius: 12px;
		background: #0c2f3b;
		color: inherit;
		font: inherit;
	}
	.compact {
		gap: 4px;
	}
	.compact .who {
		width: 44px;
		padding-top: 3px;
	}
	.compact .who-name {
		font-size: 10px;
		max-width: 42px;
	}
	.compact .who-score {
		font-size: 15px;
	}
	.who.teller {
		border-color: var(--gold);
		background: #174553;
	}
	.who.me .who-name {
		color: var(--coral);
	}
	.who-face {
		position: relative;
		line-height: 0;
	}
	.tick {
		position: absolute;
		right: -4px;
		bottom: -4px;
		width: 18px;
		height: 18px;
		border-radius: 50%;
		display: grid;
		place-items: center;
		background: #4fcf7e;
		color: #06240f;
	}
	.who-name {
		font-size: 11px;
		max-width: 48px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.who-score {
		font-family: var(--display);
		font-size: 18px;
		line-height: 1;
		color: var(--gold);
	}
	.gain {
		position: absolute;
		top: -8px;
		right: -2px;
		font-size: 11px;
		font-weight: 800;
		padding: 1px 6px;
		border-radius: 999px;
		background: #4fcf7e;
		color: #06240f;
	}
	.roundline {
		margin: 0;
	}
	.title {
		margin: 0;
		text-align: center;
		font-family: var(--display);
		font-size: clamp(30px, 9vw, 40px);
		line-height: 1;
		letter-spacing: 1px;
	}
	.clue {
		margin: 0;
		padding: 14px 16px;
		text-align: center;
		font-size: 20px;
		font-weight: 800;
		line-height: 1.25;
		border-radius: 16px;
		background: var(--gold);
		color: var(--ink);
		overflow-wrap: anywhere;
		box-shadow: 0 6px 0 #b98d1f;
	}

	/* ---- your hand ---- */
	.hand {
		display: flex;
		gap: 10px;
		overflow-x: auto;
		padding: 10px 4px 14px;
		scroll-snap-type: x proximity;
		-webkit-overflow-scrolling: touch;
	}
	.hcard {
		flex: none;
		width: 128px;
		aspect-ratio: 2 / 3;
		padding: 0;
		border: 3px solid transparent;
		border-radius: 14px;
		overflow: hidden;
		background: #0c2f3b;
		scroll-snap-align: center;
		box-shadow: 0 6px 16px #0007;
		transition: transform 0.18s, border-color 0.18s, opacity 0.2s;
	}
	.hcard.picked {
		border-color: var(--gold);
		transform: translateY(-12px) scale(1.05);
	}
	.hcard.dim {
		opacity: 0.75;
	}
	.hcard:active:not(:disabled) {
		transform: scale(0.97);
	}

	/* ---- the table ---- */
	.table {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 8px 12px;
	}
	.tcard {
		position: relative;
		aspect-ratio: 2 / 3;
		padding: 0;
		border: 3px solid transparent;
		border-radius: 14px;
		overflow: hidden;
		background: #0c2f3b;
		box-shadow: 0 6px 16px #0007;
		transition: transform 0.15s, border-color 0.15s, opacity 0.2s;
	}
	.tcard.picked {
		border-color: var(--gold);
		transform: scale(1.03);
	}
	.tcard.own {
		opacity: 0.55;
	}
	.tcard:disabled:not(.own):not(.picked) {
		opacity: 0.85;
	}
	.tcard.shown.truth {
		border-color: var(--gold);
		box-shadow: 0 0 0 3px #ffd36e55, 0 6px 16px #0007;
	}
	.flag {
		position: absolute;
		left: 8px;
		bottom: 8px;
		padding: 3px 9px;
		border-radius: 999px;
		font-size: 12px;
		font-weight: 800;
		background: #08222bcc;
		color: #fff1d6;
	}
	.flag.vote {
		background: var(--gold);
		color: var(--ink);
	}
	.voters {
		position: absolute;
		top: 6px;
		left: 6px;
		right: 6px;
		display: flex;
		flex-wrap: wrap;
		gap: 3px;
	}
	.voter {
		line-height: 0;
		border-radius: 6px;
		box-shadow: 0 2px 6px #0008;
	}
	.tag {
		position: absolute;
		left: 6px;
		right: 6px;
		bottom: 6px;
		display: flex;
		align-items: center;
		gap: 7px;
		padding: 5px 9px 5px 5px;
		border-radius: 12px;
		background: #06161ca6;
		backdrop-filter: blur(6px);
		-webkit-backdrop-filter: blur(6px);
		border: 1px solid #ffffff2e;
		text-align: left;
	}
	.tag.truth {
		background: #ffd36ecc;
		color: var(--ink);
		border-color: #ffd36e;
	}
	.tag-text {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		line-height: 1.1;
	}
	.tag-text strong {
		font-size: 13px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.tag-text small {
		font-size: 10px;
		opacity: 0.8;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}
	.tag em {
		font-style: normal;
		font-weight: 800;
		font-size: 13px;
		padding: 2px 7px;
		border-radius: 999px;
		background: #4fcf7e;
		color: #06240f;
	}

	/* ---- forms and buttons ---- */
	.clue-form {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	input {
		min-height: 52px;
		font-size: 17px;
		padding: 0 14px;
		border-radius: 12px;
		border: 1px solid #2a5a68;
		background: #06161c;
		color: inherit;
	}
	input:focus-visible {
		outline: none;
		border-color: var(--gold);
		box-shadow: 0 0 0 3px #ffd36e40;
	}
	.sr {
		position: absolute;
		left: -9999px;
	}
	.small {
		font-size: 13px;
		color: #a8cdd6;
		margin: 0;
	}
	.center {
		text-align: center;
	}
	.btn {
		min-height: 52px;
		padding: 0 20px;
		border-radius: 14px;
		border: 1px solid transparent;
		font: inherit;
		font-weight: 800;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		transition: transform 0.12s, opacity 0.2s;
	}
	.btn:active:not(:disabled) {
		transform: scale(0.97);
	}
	.btn:disabled {
		opacity: 0.5;
	}
	.btn.primary {
		background: linear-gradient(180deg, #ffa98c, var(--coral));
		color: #2a0f08;
		box-shadow: 0 4px 14px #ff8f6b40;
	}
	.btn.ghost {
		border-color: #2a5a68;
		background: transparent;
		color: inherit;
	}
	.block {
		width: 100%;
	}
	.two {
		display: flex;
		gap: 8px;
	}
	.two .btn {
		flex: 1;
	}
	.link {
		align-self: center;
		min-height: 44px;
		background: none;
		border: 0;
		color: #a8cdd6;
		text-decoration: underline;
		font: inherit;
	}
	.error {
		margin: 0;
		color: #ff8a8a;
		font-size: 14px;
		text-align: center;
	}

	/* ---- the end ---- */
	.panel {
		background: #0c2f3b;
		border: 1px solid #2a5a68;
		border-radius: 18px;
		padding: 18px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.kicker {
		font-size: 12px;
		font-weight: 800;
		letter-spacing: 1.5px;
		text-transform: uppercase;
		color: var(--coral);
	}
	.panel h1 {
		margin: 0;
		font-family: var(--display);
		font-size: clamp(36px, 11vw, 48px);
		line-height: 0.95;
	}
	.rank {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.rank li {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 12px;
		border-radius: 12px;
		background: #08222b;
		border: 2px solid transparent;
	}
	.rank li.win {
		border-color: var(--gold);
		background: #174553;
	}
	.pos {
		font-family: var(--display);
		font-size: 22px;
		width: 22px;
		color: var(--gold);
	}
	.rank-name {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.rank b {
		font-family: var(--display);
		font-size: 28px;
		color: var(--gold);
	}
</style>
