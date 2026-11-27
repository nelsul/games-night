<script lang="ts">
	// A Língua Presa match. Everything shown comes from the server's view for this player;
	// this file never decides a rule. The card (the word and the forbidden ones) only
	// arrives for the narrator and the teams that are not playing, so the narrator's own
	// team can't see it. The clock is the server's: the phone counts down from the time
	// left in each snapshot and asks the server to end the turn when it hits zero.
	import { onDestroy } from 'svelte';
	import { flip } from 'svelte/animate';
	import { fade, fly, scale } from 'svelte/transition';
	import { cubicOut, backOut } from 'svelte/easing';
	import Title from './Title.svelte';
	import { feel, haptic, ms, saveFeel, sound } from '$lib/feel.svelte';
	import { api, ApiError, type Room } from '$lib/api';
	import Avatar from '$lib/components/Avatar.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import Announcements from '$lib/components/Announcements.svelte';
	import { Announcer } from '$lib/announcer.svelte';
	import { teamColors, teamNames, type MatchView } from './types';

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
	const color = (t: number) => teamColors[t % teamColors.length];
	const teamName = (t: number) => teamNames[t % teamNames.length];

	let error = $state('');
	let busy = $state(false);
	let confirmStop = $state(false);

	async function send(action: Record<string, unknown>, quiet = false) {
		if (!quiet) error = '';
		busy = true;
		try {
			onroom(await api.post<Room>(`/rooms/${room.code}/action`, action));
		} catch (err) {
			if (!quiet) {
				error = err instanceof ApiError ? err.message : 'Algo deu errado.';
				haptic('warning');
			}
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

	// ---- The clock ----
	let deadline = 0;
	let now = $state(performance.now());
	$effect(() => {
		// Every snapshot restarts the count from what the server says is left.
		deadline = performance.now() + match.remaining_ms;
		now = performance.now();
	});
	const tick = setInterval(() => (now = performance.now()), 200);
	const left = $derived(match.phase === 'turn' ? Math.max(deadline - now, 0) : 0);
	const secondsLeft = $derived(Math.ceil(left / 1000));
	const share = $derived(match.seconds ? left / (match.seconds * 1000) : 0);

	// When the clock runs out any phone asks the server to end the turn. The first one
	// wins; the rest are told it is already over, which is fine.
	let askedAt = 0;
	$effect(() => {
		if (match.phase !== 'turn' || left > 0) return;
		const t = performance.now();
		if (t - askedAt < 2000) return;
		askedAt = t;
		send({ type: 'time_up' }, true);
	});

	// ---- What just happened ----
	const announcer = new Announcer(() => match.phase);
	let prevTurn = -1;
	let prevPhase = '';
	let prevPlayed = 0;
	$effect(() => {
		const phase = match.phase;
		const turn = match.turn;
		if (phase === 'ready' && turn !== prevTurn) {
			const t = match.active_team;
			announcer.push({
				phase: 'ready',
				kicker: `Turno ${turn + 1} de ${match.turns}`,
				title: `Vez do time ${teamName(t)}`,
				text: `${match.narrator === me ? 'Você narra' : `${first(match.narrator)} narra`}. Os outros times ficam de olho nas palavras proibidas.`,
				tone: color(t),
				faces: [match.narrator]
			});
			sound('turn');
			haptic('tap');
		}
		// The clock is running: nothing may cover the card or the buzzer.
		if (phase === 'turn' && prevPhase !== 'turn') announcer.next();
		if (phase === 'summary' && prevPhase === 'turn') {
			sound(match.played.some((p) => p.outcome === 'hit') ? 'reveal' : 'lose');
		}
		if (phase === 'over' && prevPhase !== 'over') {
			sound('win');
			haptic('success');
		}
		prevTurn = turn;
		prevPhase = phase;
	});
	$effect(() => {
		const n = match.phase === 'turn' ? match.played.length : 0;
		if (n > prevPlayed) {
			const last = match.played[n - 1];
			if (last.outcome === 'hit') sound('select');
			else if (last.outcome === 'buzz') {
				sound('rollback');
				haptic('warning');
			}
		}
		prevPlayed = n;
	});
	onDestroy(() => clearInterval(tick));

	const turnScore = $derived(
		match.played.reduce((n, p) => n + (p.outcome === 'hit' ? 1 : p.outcome === 'buzz' ? -1 : 0), 0)
	);
	const ranking = $derived(
		match.teams.map((t, i) => ({ ...t, i })).sort((a, b) => b.score - a.score)
	);
	const iWon = $derived(match.winners?.includes(match.my_team) ?? false);
	const nextTeam = $derived((match.turn + 1) % match.teams.length);
	const outcomeLabel = { hit: 'Acertou', buzz: 'Proibida', skip: 'Pulou' } as const;
</script>

<div class="lp">
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

	<!-- The scoreboard: every team, the one playing is lit up -->
	<section class="board" aria-label="Placar">
		{#each match.teams as t, i (i)}
			<div class="team" class:active={i === match.active_team} class:mine={i === match.my_team} style="--c:{color(i)}">
				<div class="team-top">
					<span class="dot"></span>
					<strong>{teamName(i)}</strong>
					{#if i === match.my_team}<span class="you">seu time</span>{/if}
				</div>
				<div class="faces">
					{#each t.players as p (p)}
						<span class="face" class:narr={p === match.narrator && match.phase !== 'over'}>
							<Avatar user={people[p]} size={30} />
						</span>
					{/each}
				</div>
				{#key t.score}
					<b class="score" in:scale={{ start: 1.5, duration: motion(260), easing: backOut }}>{t.score}</b>
				{/key}
			</div>
		{/each}
	</section>

	{#if match.phase !== 'over'}
		<p class="turnline small center">Turno {match.turn + 1} de {match.turns}</p>
	{/if}

	{#key match.phase + ':' + match.turn}
		<div class="phase" in:fly={{ y: 16, duration: motion(300), easing: cubicOut }}>
			{#if match.phase === 'ready'}
				<section class="panel hero" style="--c:{color(match.active_team)}">
					<span class="kicker">Vez do time {teamName(match.active_team)}</span>
					{#if match.role === 'narrator'}
						<h1>Você narra</h1>
						<p>
							Você vai ver uma palavra e cinco proibidas. Faça seu time falar a palavra, sem dizer nenhuma delas, e pule só se travar (até 3 vezes).
						</p>
						<button class="btn primary block big" onclick={() => send({ type: 'start' })} disabled={busy}>Começar o turno de {match.seconds}s</button>
						<p class="small center">Só aperte quando o seu time estiver pronto para ouvir.</p>
					{:else if match.role === 'mate'}
						<h1>{first(match.narrator)} vai narrar</h1>
						<p>Você não vê a carta. Escute as dicas e diga a palavra em voz alta, o mais rápido que puder.</p>
						<p class="small center">Esperando {first(match.narrator)} começar.</p>
					{:else}
						<h1>Você é fiscal</h1>
						<p>
							{first(match.narrator)} vai narrar para o time {teamName(match.active_team)}. Você verá a carta: se ouvir uma palavra proibida, aperte "Proibida!".
						</p>
						<p class="small center">Esperando {first(match.narrator)} começar.</p>
					{/if}
				</section>
			{:else if match.phase === 'turn'}
				<div class="clock" class:low={secondsLeft <= 10} style="--c:{color(match.active_team)}" role="timer" aria-label="{secondsLeft} segundos">
					<span class="clock-n">{secondsLeft}</span>
					<span class="clock-bar"><i style="width:{share * 100}%"></i></span>
				</div>

				{#if match.card}
					{#key match.card_seq}
						<section class="card" in:fly={{ x: 60, duration: motion(260), easing: cubicOut }} aria-label="Carta">
							<span class="card-word">{match.card.word}</span>
							<ul class="taboo" aria-label="Palavras proibidas">
								{#each match.card.taboo as w (w)}
									<li>{w}</li>
								{/each}
							</ul>
						</section>
					{/key}
				{/if}

				{#if match.role === 'narrator'}
					<div class="actions">
						<button class="btn skip" onclick={() => send({ type: 'skip', seq: match.card_seq })} disabled={busy || match.skips_left === 0}>
							Pular ({match.skips_left})
						</button>
						<button class="btn hit" onclick={() => send({ type: 'hit', seq: match.card_seq })} disabled={busy}>
							<Icon name="check" size={22} />Acertou
						</button>
					</div>
					<p class="small center">Quando o seu time disser a palavra, toque em Acertou. Não diga as proibidas.</p>
				{:else if match.role === 'watcher'}
					<button class="btn buzz block" onclick={() => send({ type: 'buzz', seq: match.card_seq })} disabled={busy}>Proibida!</button>
					<p class="small center">{first(match.narrator)} disse uma palavra proibida (ou a própria palavra)? Aperte.</p>
				{:else}
					<section class="panel guess">
						<h2>Adivinhe!</h2>
						<p>Você não vê a carta. Fale em voz alta o que {first(match.narrator)} está descrevendo.</p>
					</section>
				{/if}

				{#if match.played.length}
					<ul class="played" aria-label="Cartas deste turno">
						{#each match.played.slice().reverse() as p, i (match.played.length - i)}
							<li class={p.outcome} animate:flip={{ duration: motion(200) }} in:fly={{ y: -10, duration: motion(220) }}>
								<span class="tag">{outcomeLabel[p.outcome]}</span>{p.word}
							</li>
						{/each}
					</ul>
				{/if}
			{:else if match.phase === 'summary'}
				<section class="panel hero" style="--c:{color(match.active_team)}">
					<span class="kicker">Fim do turno · time {teamName(match.active_team)}</span>
					<h1 class:neg={turnScore < 0}>{turnScore > 0 ? '+' : ''}{turnScore} {Math.abs(turnScore) === 1 ? 'ponto' : 'pontos'}</h1>
					{#if match.played.length}
						<ul class="played full">
							{#each match.played as p, i (i)}
								<li class={p.outcome} in:fly={{ y: 10, duration: motion(220), delay: motion(60 * i) }}>
									<span class="tag">{outcomeLabel[p.outcome]}</span>{p.word}
								</li>
							{/each}
						</ul>
					{:else}
						<p>Nenhuma carta neste turno.</p>
					{/if}
					<button class="btn primary block big" onclick={() => send({ type: 'next' })} disabled={busy}>
						{match.turn + 1 >= match.turns ? 'Ver o resultado' : `Próximo turno: time ${teamName(nextTeam)}`}
					</button>
				</section>
			{:else}
				<section class="panel hero over" style="--c:{color(match.winners?.[0] ?? 0)}">
					<span class="kicker">{iWon ? 'Vitória!' : 'Fim de jogo'}</span>
					<h1>
						{match.winners && match.winners.length > 1
							? `Empate entre ${match.winners.map((w) => teamName(w)).join(' e ')}`
							: `Time ${teamName(match.winners?.[0] ?? 0)} ganhou`}
					</h1>
					<ol class="rank">
						{#each ranking as t, k (t.i)}
							<li style="--c:{color(t.i)}" class:win={match.winners?.includes(t.i)} in:fly={{ y: 14, duration: motion(260), delay: motion(120 * k) }}>
								<span class="pos">{k + 1}</span>
								<span class="rank-team">
									<strong>Time {teamName(t.i)}</strong>
									<span class="rank-faces">{t.players.map((p) => first(p)).join(', ')}</span>
								</span>
								<b>{t.score}</b>
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

<Announcements {announcer} {people} />

<style>
	.lp {
		--gold: #ffb347;
		--ink: #1f1038;
		position: relative;
		overflow-x: clip;
		background:
			radial-gradient(120% 50% at 50% -10%, #5a2a8a 0%, transparent 60%),
			linear-gradient(180deg, #140a24, #1f1038 60%, #2a1650);
		color: #f3f7ff;
		min-height: 100dvh;
		max-width: 560px;
		margin: 0 auto;
		padding: calc(10px + var(--safe-top)) 16px calc(24px + var(--safe-bottom));
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	:global(body:has(.lp)) {
		background: #140a24;
	}
	.phase {
		display: flex;
		flex-direction: column;
		gap: 14px;
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
		border: 1px solid #5a3a8a;
		background: transparent;
		color: inherit;
		transition: transform 0.12s, background 0.2s;
	}
	.pill:active {
		transform: scale(0.92);
		background: #3a2268;
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

	/* ---- scoreboard ---- */
	.board {
		display: grid;
		grid-auto-flow: column;
		grid-auto-columns: minmax(0, 1fr);
		gap: 8px;
	}
	.team {
		position: relative;
		padding: 10px;
		border-radius: 14px;
		background: #2a1850;
		border: 2px solid transparent;
		display: flex;
		flex-direction: column;
		gap: 6px;
		opacity: 0.72;
		transition: opacity 0.25s, border-color 0.25s, transform 0.25s;
	}
	.team.active {
		opacity: 1;
		border-color: var(--c);
		transform: translateY(-2px);
		box-shadow: 0 8px 22px #0006;
	}
	.team-top {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		min-width: 0;
	}
	.dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
		background: var(--c);
		flex: none;
	}
	.you {
		font-size: 10px;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: #c9aee8;
	}
	.faces {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
	}
	.face {
		border-radius: 8px;
		line-height: 0;
	}
	.face.narr {
		outline: 2px solid var(--c);
		outline-offset: 1px;
	}
	.score {
		font-family: var(--display);
		font-size: 34px;
		line-height: 1;
		color: var(--c);
	}
	.turnline {
		margin: -4px 0 0;
	}

	/* ---- panels ---- */
	.panel {
		background: #2a1850;
		border: 1px solid #4a2a7a;
		border-radius: 18px;
		padding: 18px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.panel p {
		margin: 0;
	}
	.hero {
		border-color: var(--c);
		background: linear-gradient(180deg, color-mix(in srgb, var(--c) 22%, #2a1850), #2a1850 70%);
	}
	.kicker {
		font-size: 12px;
		font-weight: 800;
		letter-spacing: 1.5px;
		text-transform: uppercase;
		color: var(--c);
	}
	.hero h1 {
		margin: 0;
		font-family: var(--display);
		font-size: clamp(36px, 11vw, 48px);
		line-height: 0.95;
		letter-spacing: 1px;
	}
	.hero h1.neg {
		color: #ff8a8a;
	}
	.guess h2 {
		margin: 0;
		font-family: var(--display);
		font-size: 34px;
		letter-spacing: 1px;
	}

	/* ---- the clock ---- */
	.clock {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.clock-n {
		font-family: var(--display);
		font-size: 54px;
		line-height: 1;
		min-width: 64px;
		color: var(--c);
	}
	.clock.low .clock-n {
		color: #ff6b6b;
		animation: breathe 0.8s ease-in-out infinite;
	}
	.clock-bar {
		flex: 1;
		height: 12px;
		border-radius: 999px;
		background: #3a2268;
		overflow: hidden;
	}
	.clock-bar i {
		display: block;
		height: 100%;
		background: var(--c);
		border-radius: 999px;
		transition: width 0.2s linear;
	}
	.clock.low .clock-bar i {
		background: #ff6b6b;
	}
	@keyframes breathe {
		50% {
			transform: scale(1.12);
		}
	}

	/* ---- the card ---- */
	.card {
		background: #fff4e0;
		color: #3a1a22;
		border-radius: 20px;
		overflow: hidden;
		box-shadow: 0 14px 34px #0008;
		border: 3px solid #ff4fa3;
	}
	.card-word {
		display: block;
		background: #ff4fa3;
		color: #fff;
		text-align: center;
		font-family: var(--display);
		font-size: clamp(40px, 13vw, 58px);
		letter-spacing: 2px;
		line-height: 1;
		padding: 16px 10px 12px;
		text-transform: uppercase;
		overflow-wrap: anywhere;
	}
	.taboo {
		list-style: none;
		margin: 0;
		padding: 10px 16px 16px;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.taboo li {
		text-align: center;
		font-weight: 800;
		font-size: 22px;
		line-height: 1.5;
		color: #a3263a;
		text-transform: uppercase;
		letter-spacing: 1px;
		border-bottom: 1px dashed #e8c9a6;
	}
	.taboo li:last-child {
		border-bottom: 0;
	}

	/* ---- buttons ---- */
	.actions {
		display: grid;
		grid-template-columns: 1fr 2fr;
		gap: 10px;
	}
	.small {
		font-size: 13px;
		color: #d0b8ec;
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
		transition: transform 0.12s, opacity 0.2s, filter 0.2s;
	}
	.btn:active:not(:disabled) {
		transform: scale(0.97);
	}
	.btn:disabled {
		opacity: 0.5;
	}
	.btn.primary {
		background: linear-gradient(180deg, #ff7cbf, #ff4fa3);
		color: #2a0a1e;
		box-shadow: 0 4px 14px #ff4fa340;
	}
	.btn.ghost {
		border-color: #5a3a8a;
		background: transparent;
		color: inherit;
	}
	.btn.hit {
		min-height: 72px;
		font-size: 22px;
		background: linear-gradient(180deg, #7be8a0, #4fcf7e);
		color: #06240f;
		box-shadow: 0 4px 0 #2c8a50;
	}
	.btn.skip {
		min-height: 72px;
		background: #3a2268;
		border-color: #5a3a8a;
		color: inherit;
	}
	.btn.buzz {
		min-height: 88px;
		font-size: 28px;
		font-family: var(--display);
		letter-spacing: 2px;
		text-transform: uppercase;
		background: linear-gradient(180deg, #ff6b6b, #e23d3d);
		color: #fff;
		box-shadow: 0 5px 0 #8f1f1f;
	}
	.btn.big {
		min-height: 60px;
		font-size: 18px;
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
		color: #d0b8ec;
		text-decoration: underline;
		font: inherit;
	}
	.error {
		margin: 0;
		color: #ff8a8a;
		font-size: 14px;
		text-align: center;
	}

	/* ---- cards played this turn ---- */
	.played {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.played.full {
		flex-direction: column;
		flex-wrap: nowrap;
	}
	.played li {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 12px 6px 6px;
		border-radius: 999px;
		font-weight: 700;
		background: #3a2268;
		border: 1px solid #5a3a8a;
	}
	.played li .tag {
		font-size: 11px;
		letter-spacing: 0.5px;
		text-transform: uppercase;
		padding: 3px 8px;
		border-radius: 999px;
		background: #5a3a8a;
	}
	.played li.hit .tag {
		background: #4fcf7e;
		color: #06240f;
	}
	.played li.buzz .tag {
		background: #e23d3d;
		color: #fff;
	}
	.played li.buzz {
		text-decoration: line-through;
		text-decoration-color: #e23d3d88;
	}
	.played li.skip {
		opacity: 0.75;
	}

	/* ---- the end ---- */
	.rank {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.rank li {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px;
		border-radius: 14px;
		background: #1f1038;
		border: 2px solid transparent;
	}
	.rank li.win {
		border-color: var(--c);
		background: color-mix(in srgb, var(--c) 18%, #1f1038);
	}
	.pos {
		font-family: var(--display);
		font-size: 26px;
		width: 28px;
		color: var(--c);
	}
	.rank-team {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
	}
	.rank-faces {
		font-size: 13px;
		color: #d0b8ec;
	}
	.rank b {
		font-family: var(--display);
		font-size: 36px;
		color: var(--c);
	}

	@media (prefers-reduced-motion: reduce) {
		.clock.low .clock-n {
			animation: none;
		}
	}
</style>
