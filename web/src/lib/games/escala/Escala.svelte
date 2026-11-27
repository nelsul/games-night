<script lang="ts">
	// An Escala match. Everything shown comes from the server's view for this player;
	// this file never decides a rule. Other people's numbers only arrive after the reveal.
	//
	// Look: the theme is a big poster card, your secret number is a big playing card you can
	// flip face down, the table's line is made of big cards that flip open on the reveal, and
	// a dock at the bottom keeps your card and the main button within thumb reach.
	import { onDestroy } from 'svelte';
	import { flip } from 'svelte/animate';
	import { fade, fly, scale, slide } from 'svelte/transition';
	import { cubicOut, backOut } from 'svelte/easing';
	import Title from './Title.svelte';
	import CountUp from './CountUp.svelte';
	import { feel, haptic, ms, reducedMotion, saveFeel, sound } from '$lib/feel.svelte';
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
	const player = (id: string) => match.players.find((p) => p.id === id);

	let error = $state('');
	let busy = $state(false);
	let faceDown = $state(false);
	let clue = $state('');
	let editing = $state(false);
	let confirmStop = $state(false);
	let toast = $state('');
	let toastTimer: ReturnType<typeof setTimeout> | undefined;

	const myPlayer = $derived(player(me));
	const myPos = $derived(match.line.findIndex((c) => c.player === me));
	const waiting = $derived(match.players.filter((p) => !p.placed));
	const lastRound = $derived(match.round + 1 >= match.rounds);
	const cluesDone = $derived(match.players.filter((p) => p.has_clue).length);
	const perfect = $derived(!match.line.some((c) => c.wrong));
	const markerLeft = $derived(((match.me.number - 1) / 99) * 100);
	const step = $derived(match.phase === 'clues' ? 0 : match.phase === 'order' ? 1 : 2);
	const steps = ['Dicas', 'Ordem', 'Revelação'];

	// Each theme gets its own color, so a swapped theme feels like a new card.
	const hue = $derived(
		[265, 335, 18, 192, 150, 42][[...match.theme.title].reduce((n, c) => n + c.charCodeAt(0), 0) % 6]
	);

	// The spot of a card in the line, from the cool end of the scale to the warm one.
	const spotHue = (i: number, n: number) => Math.round(210 - (n < 2 ? 0 : i / (n - 1)) * 210);

	const pips = $derived(
		Array.from({ length: match.rounds }, (_, i) => {
			const r = match.results[i];
			return r ? (r.ok ? 'ok' : 'miss') : i === match.round && match.phase !== 'over' ? 'now' : 'todo';
		})
	);

	// ---- Theme change vote ----
	const swap = $derived(match.swap);
	const iVoted = $derived(swap ? me in swap.votes : false);
	const canPropose = $derived(
		!swap && match.swaps_left > 0 && (match.phase === 'clues' || match.phase === 'order')
	);
	const yesCount = $derived(swap ? Object.values(swap.votes).filter(Boolean).length : 0);
	const needed = $derived(Math.floor(match.players.length / 2) + 1);

	async function act(action: Record<string, unknown>) {
		error = '';
		busy = true;
		try {
			onroom(await api.post<Room>(`/rooms/${room.code}/action`, action));
			if (action.type === 'clue') {
				clue = '';
				editing = false;
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

	// Slot k sits before the k-th card of the line. The server wants the index among the
	// other cards, so when my own card is to the left of the slot it doesn't count.
	function place(slot: number) {
		sound('tap');
		haptic('tap');
		act({ type: 'place', index: myPos >= 0 && slot > myPos ? slot - 1 : slot });
	}
	// Dropping my card where it already is would do nothing.
	const slotUseful = (slot: number) => myPos < 0 || (slot !== myPos && slot !== myPos + 1);

	function toggleSound() {
		feel.sound = !feel.sound;
		saveFeel();
		if (feel.sound) sound('select');
	}

	function editClue() {
		clue = myPlayer?.clue ?? '';
		editing = true;
	}

	function flash(text: string) {
		toast = text;
		clearTimeout(toastTimer);
		toastTimer = setTimeout(() => (toast = ''), 3200);
	}

	// ---- What just happened: sounds, announcements and the card-by-card reveal ----

	let prevPhase: string | null = null;
	let prevRound = -1;
	let prevSwapSeq = -1;
	let prevSwapProposer = '';
	$effect(() => {
		const phase = match.phase;
		const round = match.round;
		if (prevPhase !== null) {
			if (phase === 'clues' && round !== prevRound) {
				sound('turn');
				haptic('tap');
			} else if (phase === 'order' && prevPhase === 'clues') {
				sound('vote');
			}
		}
		prevPhase = phase;
		prevRound = round;
	});
	$effect(() => {
		const seq = match.swap_seq;
		if (prevSwapSeq >= 0 && seq !== prevSwapSeq) {
			flash(match.swap_ok ? 'Tema trocado! Escrevam novas dicas.' : 'A troca de tema foi recusada.');
			sound(match.swap_ok ? 'turn' : 'rollback');
			haptic(match.swap_ok ? 'success' : 'warning');
		}
		prevSwapSeq = seq;
	});
	$effect(() => {
		const proposer = match.swap?.proposer ?? '';
		if (proposer && proposer !== prevSwapProposer && proposer !== me) {
			sound('vote');
			haptic('tap');
		}
		prevSwapProposer = proposer;
	});

	// The reveal opens one card at a time, smallest first, so the table sees the order build up.
	let shown = $state(0);
	let timer: ReturnType<typeof setInterval> | undefined;
	let celebrate = $state(false);
	const allShown = $derived(shown >= match.line.length);

	$effect(() => {
		clearInterval(timer);
		celebrate = false;
		if (match.phase !== 'reveal' && match.phase !== 'over') {
			shown = 0;
			return;
		}
		if (match.phase === 'over' || reducedMotion) {
			shown = match.line.length;
			return;
		}
		shown = 0;
		timer = setInterval(() => {
			shown += 1;
			sound('reveal');
			haptic('tap');
			if (shown >= match.line.length) clearInterval(timer);
		}, 800);
	});

	$effect(() => {
		if (match.phase === 'reveal' && allShown && match.line.length) {
			if (perfect) {
				celebrate = true;
				sound('deploy');
				haptic('success');
			} else {
				sound('rollback');
				haptic('warning');
			}
		}
	});
	$effect(() => {
		if (match.phase === 'over') {
			if (match.score === match.rounds) {
				celebrate = true;
				sound('win');
				haptic('success');
			} else {
				sound(match.score === 0 ? 'lose' : 'win');
			}
		}
	});
	onDestroy(() => {
		clearInterval(timer);
		clearTimeout(toastTimer);
	});

	const stars = $derived(match.score === match.rounds ? 3 : match.score * 2 >= match.rounds ? 2 : match.score > 0 ? 1 : 0);

	const confetti = Array.from({ length: 34 }, (_, i) => ({
		left: (i * 37 + 11) % 100,
		delay: (i % 8) * 0.09,
		dur: 1.7 + (i % 5) * 0.25,
		color: ['#ffb347', '#6ee08f', '#4aa3ff', '#ff8a8a', '#fff4e0'][i % 5],
		spin: (i % 2 ? 1 : -1) * (180 + (i % 4) * 90)
	}));
</script>

{#snippet clueForm()}
	<form
		class="clue-form"
		onsubmit={(e) => {
			e.preventDefault();
			act({ type: 'clue', text: clue });
		}}
	>
		<label class="sr" for="clue">Sua dica</label>
		<input id="clue" bind:value={clue} maxlength="30" placeholder="Ex.: um gato gordo" autocomplete="off" />
		<button class="btn primary" disabled={busy || !clue.trim()}>Enviar</button>
		{#if editing}<button type="button" class="btn ghost" onclick={() => (editing = false)}>Cancelar</button>{/if}
	</form>
{/snippet}

<div class="es">
	{#if celebrate}
		<div class="confetti" aria-hidden="true">
			{#each confetti as c, i (i)}
				<i style="left:{c.left}%;animation-delay:{c.delay}s;animation-duration:{c.dur}s;background:{c.color};--spin:{c.spin}deg"></i>
			{/each}
		</div>
	{/if}

	{#if toast}
		<div class="toast" role="status" transition:fly={{ y: -20, duration: motion(260) }}>{toast}</div>
	{/if}

	<header class="bar">
		<a href="/" aria-label="Voltar ao início"><Icon name="back" /></a>
		<Title size={28} />
		<span class="bar-actions">
			<button class="pill" onclick={toggleSound} aria-label={feel.sound ? 'Desligar sons' : 'Ligar sons'} aria-pressed={feel.sound}>
				<Icon name={feel.sound ? 'sound' : 'mute'} size={18} />
			</button>
			<button class="pill" onclick={onchat} aria-label={unread ? `Chat, ${unread} novas` : 'Chat'}>
				<Icon name="chat" size={18} />{#if unread}<span class="badge" aria-hidden="true">{unread}</span>{/if}
			</button>
		</span>
	</header>

	{#if match.phase !== 'over'}
		<div class="status">
			<ol class="steps" aria-label="Etapa da rodada">
				{#each steps as s, i (s)}
					<li class:now={i === step} class:past={i < step} aria-current={i === step ? 'step' : undefined}>
						<span class="dot">{#if i < step}<Icon name="check" size={12} />{:else}{i + 1}{/if}</span>{s}
					</li>
				{/each}
			</ol>
			<span class="score">
				{#key match.score}<b in:scale={{ start: 1.6, duration: motion(300), easing: backOut }}>{match.score}</b>{/key}
				{match.score === 1 ? 'acerto' : 'acertos'}
			</span>
		</div>

		<!-- The theme: a big poster card in its own color -->
		{#key match.theme.title}
			<section
				class="theme"
				style="--h:{hue}"
				aria-label="Tema da rodada"
				in:fly={{ y: -28, duration: motion(480), easing: cubicOut }}
			>
				<div class="theme-top">
					<span class="kicker">
						Rodada {match.round + 1} de {match.rounds}
						<span class="pips" role="img" aria-label="Andamento das rodadas">
							{#each pips as p, i (i)}<span class="pip {p}"></span>{/each}
						</span>
					</span>
					{#if canPropose}
						<button class="swap-btn" onclick={() => act({ type: 'swap_propose' })} disabled={busy} aria-label="Pedir para trocar o tema, {match.swaps_left} restantes">
							<Icon name="rotate" size={16} />Trocar tema
						</button>
					{/if}
				</div>
				<h1>{match.theme.title}</h1>
				<div class="gauge">
					<div class="track">
						{#if !faceDown}
							<span class="marker" style="left:{markerLeft}%" in:fly={{ y: -14, duration: motion(500), delay: motion(350) }} aria-hidden="true">
								<em>{match.me.number}</em>
							</span>
						{/if}
					</div>
					<div class="ends">
						<span><b>1</b>{match.theme.low}</span>
						<span><b>100</b>{match.theme.high}</span>
					</div>
				</div>
			</section>
		{/key}

		<!-- A vote to change the theme -->
		{#if swap}
			<section class="vote" transition:slide={{ duration: motion(260) }} aria-live="polite">
				<p class="vote-title">
					<Icon name="rotate" size={18} />
					<span><strong>{swap.proposer === me ? 'Você quer' : `${first(swap.proposer)} quer`}</strong> trocar o tema. Precisa de {needed} votos.</span>
				</p>
				<div class="voters">
					{#each match.players as p (p.id)}
						{@const v = swap.votes[p.id]}
						<span class="voter" class:yes={v === true} class:no={v === false} title="{name(p.id)}">
							<Avatar user={people[p.id]} size={36} />
							<span class="vmark" aria-hidden="true">{v === true ? '✓' : v === false ? '✕' : '…'}</span>
						</span>
					{/each}
				</div>
				{#if !iVoted}
					<div class="vote-actions">
						<button class="btn ghost" onclick={() => act({ type: 'swap_vote', yes: false })} disabled={busy}>Manter o tema</button>
						<button class="btn primary" onclick={() => act({ type: 'swap_vote', yes: true })} disabled={busy}>Trocar</button>
					</div>
				{:else}
					<p class="small">{yesCount} de {needed} votos para trocar. Esperando o resto da mesa.</p>
				{/if}
			</section>
		{/if}
	{/if}

	{#key match.phase}
		<div class="phase" in:fly={{ y: 18, duration: motion(320), easing: cubicOut }}>
			{#if match.phase === 'clues'}
				<!-- Your secret card, big -->
				<section class="hand" aria-label="Seu número">
					<button
						class="bigcard"
						class:flipped={faceDown}
						onclick={() => {
							faceDown = !faceDown;
							sound('tap');
							haptic('tap');
						}}
						aria-pressed={faceDown}
						aria-label={faceDown ? 'Mostrar meu número' : 'Esconder meu número'}
					>
						<span class="inner">
							<span class="face front">
								<span class="corner tl">{match.me.number}</span>
								<span class="big">{match.me.number}</span>
								<span class="corner br">{match.me.number}</span>
								<span class="mini-scale"><i style="left:{markerLeft}%"></i></span>
							</span>
							<span class="face back"><span class="logo">ES<br />CA<br />LA</span></span>
						</span>
					</button>
					<span class="hint">{faceDown ? 'Toque para virar a carta' : 'Seu número secreto · toque para esconder'}</span>
				</section>

				<section class="panel">
					<h2>Dê uma dica</h2>
					<p class="small">Descreva o seu número usando o tema. Não vale falar números.</p>
					{#if myPlayer?.has_clue && !editing}
						<p class="sent"><Icon name="check" size={18} /> Sua dica: <strong>“{myPlayer.clue}”</strong></p>
						<button class="btn ghost" onclick={editClue}>Trocar a dica</button>
					{:else}
						{@render clueForm()}
					{/if}
					<div class="who-row" aria-label="Quem já deu a dica">
						{#each match.players as p (p.id)}
							<span class="chip" class:done={p.has_clue} title={name(p.id)}>
								<Avatar user={people[p.id]} size={40} />
								{#if p.has_clue}<span class="tick" in:scale={{ duration: motion(260), easing: backOut }}><Icon name="check" size={12} /></span>{/if}
							</span>
						{/each}
						<span class="small count">{cluesDone} de {match.players.length}</span>
					</div>
				</section>
			{:else if match.phase === 'order' || match.phase === 'reveal'}
				<div class="line-head">
					<h2>{match.phase === 'order' ? 'Montem a fila' : 'Revelando…'}</h2>
					{#if match.phase === 'order'}
						<p class="small">
							Combinem de viva voz. {myPos < 0 ? 'Toque num + para colocar a sua carta.' : 'Dá para mover a sua carta até revelar.'}
						</p>
					{/if}
				</div>

				<div class="rail-end low"><span>▲</span> menor · {match.theme.low}</div>
				<ol class="line">
					{#each match.line as c, i (c.player)}
						{@const open = match.phase === 'reveal' && i < shown}
						{@const isMe = c.player === me}
						<li class="item" animate:flip={{ duration: motion(340) }} in:fly={{ x: -28, duration: motion(340), easing: cubicOut }}>
							{#if match.phase === 'order' && slotUseful(i)}
								<button class="slot" onclick={() => place(i)} disabled={busy} aria-label="Colocar minha carta na posição {i + 1}" transition:fade={{ duration: motion(150) }}>
									<span class="plus">+</span> colocar aqui
								</button>
							{/if}
							<div
								class="lcard"
								class:wrong={open && c.wrong}
								class:right={open && !c.wrong}
								class:mine={isMe}
								style="--spot:{spotHue(i, match.line.length)}"
							>
								<span class="rank" aria-label="Posição {i + 1}">{i + 1}</span>
								<button class="who" onclick={() => onplayer?.(c.player)} aria-label="Ver {name(c.player)}">
									<Avatar user={people[c.player]} size={48} />
								</button>
								<span class="txt">
									<span class="pname">{name(c.player)}{isMe ? ' (você)' : ''}</span>
									<span class="clue">“{player(c.player)?.clue}”</span>
								</span>
								<span class="tiny" class:flipped={open || (isMe && !faceDown && match.phase === 'order')}>
									<span class="inner">
										<span class="face front" class:bad={c.wrong}>
											{#if open && c.number !== undefined}
												<span class="tn"><CountUp value={c.number} duration={500} /></span>
											{:else if isMe}
												<span class="tn gold">{match.me.number}</span>
											{/if}
										</span>
										<span class="face back"><span class="q">?</span></span>
									</span>
								</span>
							</div>
						</li>
					{/each}
					{#if match.phase === 'order' && slotUseful(match.line.length)}
						<li class="item" transition:fade={{ duration: motion(150) }}>
							<button class="slot" onclick={() => place(match.line.length)} disabled={busy} aria-label="Colocar minha carta no fim">
								<span class="plus">+</span> colocar aqui
							</button>
						</li>
					{/if}
				</ol>
				<div class="rail-end high"><span>▼</span> maior · {match.theme.high}</div>

				{#if match.phase === 'order' && waiting.length}
					<section class="panel">
						<h3>Falta colocar</h3>
						<ul class="pending">
							{#each waiting as p (p.id)}
								<li out:fly={{ x: 30, duration: motion(220) }}>
									<Avatar user={people[p.id]} size={32} />
									<span><strong>{name(p.id)}{p.id === me ? ' (você)' : ''}</strong> “{p.clue}”</span>
								</li>
							{/each}
						</ul>
					</section>
				{/if}

				{#if match.phase === 'reveal' && allShown}
					<p class="verdict" class:good={perfect} in:scale={{ start: 0.8, duration: motion(360), easing: backOut }}>
						{perfect ? 'Ordem perfeita! Rodada certa.' : 'Algumas cartas ficaram fora de ordem. Na próxima vocês acertam.'}
					</p>
				{/if}

				<!-- The dock: your card, your clue and the main button, always within reach -->
				<div class="dock" in:fly={{ y: 40, duration: motion(300) }}>
					{#if match.phase === 'order'}
						{#if editing}
							<div class="dock-edit">{@render clueForm()}</div>
						{:else}
							<div class="dock-row">
								<button class="dockcard" class:flipped={faceDown} onclick={() => (faceDown = !faceDown)} aria-pressed={faceDown} aria-label={faceDown ? 'Mostrar meu número' : 'Esconder meu número'}>
									<span class="inner">
										<span class="face front">{match.me.number}</span>
										<span class="face back">?</span>
									</span>
								</button>
								<span class="dock-text">
									<span class="small">Sua dica</span>
									<strong>“{myPlayer?.clue}”</strong>
								</span>
								<button class="ghost" onclick={editClue}>Trocar</button>
							</div>
						{/if}
						{#if !editing}
							{#if waiting.length === 0}
								<button class="btn primary block pulse" onclick={() => act({ type: 'reveal' })} disabled={busy || !!swap}>
									{swap ? 'Terminem a votação' : 'Revelar os números'}
								</button>
							{:else}
								<p class="small center">Faltam {waiting.length} {waiting.length === 1 ? 'pessoa colocar' : 'pessoas colocarem'} a carta.</p>
							{/if}
						{/if}
					{:else if allShown}
						<button class="btn primary block" onclick={() => act({ type: 'next' })} disabled={busy} in:fade={{ duration: motion(300), delay: motion(250) }}>
							{lastRound ? 'Ver resultado final' : 'Próxima rodada'}
						</button>
					{:else}
						<p class="small center">Abrindo as cartas…</p>
					{/if}
				</div>
			{:else}
				<section class="panel over">
					<div class="stars" aria-label="{stars} de 3 estrelas">
						{#each [0, 1, 2] as s (s)}
							<span class="star" class:on={s < stars} in:scale={{ start: 0, duration: motion(420), delay: motion(250 + s * 220), easing: backOut }}>★</span>
						{/each}
					</div>
					<h1 class:good={match.score === match.rounds}>
						<CountUp value={match.score} duration={900} /> de {match.rounds} rodadas certas
					</h1>
					<p>
						{match.score === match.rounds
							? 'Sincronia perfeita! Todo mundo se entendeu.'
							: match.score === 0
								? 'Nenhuma desta vez. Que tal outra partida?'
								: 'Bom trabalho em equipe. Dá para melhorar na próxima.'}
					</p>
					<ul class="history">
						{#each match.results as r, i (i)}
							<li in:fly={{ x: -16, duration: motion(260), delay: motion(300 + i * 90) }}>
								<span class="mark" class:good={r.ok}>{r.ok ? 'ok' : 'x'}</span>{r.theme.title}
							</li>
						{/each}
					</ul>
					{#if match.line.length}
						<h3>Última ordem</h3>
						<ol class="final">
							{#each match.line as c (c.player)}
								<li class:wrong={c.wrong}>{name(c.player)} <b>{c.number}</b></li>
							{/each}
						</ol>
					{/if}
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
	.es {
		--gold: #ffb347;
		--ink: #2b1636;
		--wine: #7a2e3a;
		position: relative;
		overflow-x: clip;
		background:
			radial-gradient(120% 50% at 50% -10%, #4a2260 0%, transparent 60%),
			linear-gradient(180deg, #1a0f22, #241230 60%, #2b1636);
		color: #fff4e0;
		min-height: 100dvh;
		max-width: 560px;
		margin: 0 auto;
		padding: calc(10px + var(--safe-top)) 16px calc(24px + var(--safe-bottom));
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	:global(body:has(.es)) {
		background: #1a0f22;
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
		border: 1px solid #5a3a68;
		background: transparent;
		color: inherit;
		transition: transform 0.12s, background 0.2s;
	}
	.pill:active {
		transform: scale(0.92);
		background: #3a2148;
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
	.toast {
		position: fixed;
		top: calc(12px + var(--safe-top));
		left: 50%;
		translate: -50% 0;
		z-index: 30;
		max-width: min(92vw, 420px);
		background: var(--gold);
		color: var(--ink);
		font-weight: 700;
		border-radius: 999px;
		padding: 10px 18px;
		box-shadow: 0 8px 24px #0008;
		text-align: center;
	}

	/* ---- steps ---- */
	.status {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 8px;
	}
	.steps {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		gap: 12px;
		font-size: 13px;
		font-weight: 600;
		color: #8f7a9b;
	}
	.steps li {
		display: flex;
		align-items: center;
		gap: 6px;
		transition: color 0.3s;
	}
	.steps .dot {
		width: 20px;
		height: 20px;
		border-radius: 50%;
		display: grid;
		place-items: center;
		font-size: 11px;
		border: 1px solid #5a3a68;
		transition: background 0.3s, color 0.3s, border-color 0.3s;
	}
	.steps .now {
		color: #fff4e0;
	}
	.steps .now .dot {
		background: var(--gold);
		border-color: var(--gold);
		color: var(--ink);
	}
	.steps .past .dot {
		background: #46d369;
		border-color: #46d369;
		color: #0c2a14;
	}
	.score {
		font-size: 13px;
		font-weight: 600;
		white-space: nowrap;
	}
	.score b {
		display: inline-block;
		color: var(--gold);
		font-family: var(--display);
		font-size: 22px;
		line-height: 1;
	}

	/* ---- theme poster ---- */
	.theme {
		position: relative;
		overflow: hidden;
		border-radius: 22px;
		padding: 16px 18px 18px;
		background: hsl(var(--h) 52% 30%);
		box-shadow: 0 14px 34px #0007, inset 0 0 0 1px #ffffff1f;
		color: #fff;
	}
	.theme-top {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 8px;
		min-height: 36px;
	}
	.kicker {
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 12px;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 1.5px;
		opacity: 0.9;
	}
	.pips {
		display: flex;
		gap: 5px;
	}
	.pip {
		width: 20px;
		height: 6px;
		border-radius: 999px;
		background: #ffffff40;
		transition: background 0.4s;
	}
	.pip.ok {
		background: #6ee08f;
	}
	.pip.miss {
		background: #ff8a8a;
	}
	.pip.now {
		background: #fff;
		animation: breathe 1.6s ease-in-out infinite;
	}
	.swap-btn {
		min-height: 44px;
		padding: 0 14px;
		display: inline-flex;
		align-items: center;
		gap: 6px;
		border-radius: 999px;
		border: 1px solid #ffffff66;
		background: #00000030;
		color: #fff;
		font: inherit;
		font-size: 13px;
		font-weight: 700;
		transition: transform 0.12s, background 0.2s;
	}
	.swap-btn:active:not(:disabled) {
		transform: scale(0.95);
		background: #00000055;
	}
	.theme h1 {
		margin: 10px 0 30px;
		font-family: var(--display);
		font-size: clamp(40px, 12vw, 54px);
		line-height: 0.95;
		letter-spacing: 1px;
		text-shadow: 0 2px 0 #0004;
		overflow-wrap: anywhere;
	}
	.gauge .track {
		position: relative;
		height: 14px;
		border-radius: 999px;
		background: linear-gradient(90deg, #4aa3ff, #ffb347, #e5383b);
		box-shadow: 0 0 0 3px #00000040;
	}
	.marker {
		position: absolute;
		top: -30px;
		transform: translateX(-50%);
		display: flex;
		flex-direction: column;
		align-items: center;
		transition: left 0.4s;
	}
	.marker em {
		font-style: normal;
		font-family: var(--display);
		font-size: 20px;
		line-height: 1;
		background: #fff4e0;
		color: var(--ink);
		border-radius: 7px;
		padding: 3px 7px;
	}
	.marker::after {
		content: '';
		border: 6px solid transparent;
		border-top-color: #fff4e0;
	}
	.ends {
		display: flex;
		justify-content: space-between;
		gap: 12px;
		margin-top: 10px;
		font-size: 14px;
		font-weight: 600;
	}
	.ends span {
		display: flex;
		flex-direction: column;
		max-width: 48%;
	}
	.ends span:last-child {
		text-align: right;
		align-items: flex-end;
	}
	.ends b {
		font-family: var(--display);
		font-size: 28px;
		line-height: 1;
		opacity: 0.85;
	}

	/* ---- theme vote ---- */
	.vote {
		background: #3a2148;
		border: 1px solid var(--gold);
		border-radius: 16px;
		padding: 14px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.vote-title {
		margin: 0;
		display: flex;
		gap: 8px;
		align-items: flex-start;
	}
	.voters {
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
	}
	.voter {
		position: relative;
		opacity: 0.55;
		transition: opacity 0.3s;
	}
	.voter.yes,
	.voter.no {
		opacity: 1;
	}
	.vmark {
		position: absolute;
		right: -5px;
		bottom: -5px;
		min-width: 20px;
		height: 20px;
		border-radius: 50%;
		display: grid;
		place-items: center;
		font-size: 12px;
		font-weight: 800;
		background: #5a3a68;
		color: #fff;
	}
	.voter.yes .vmark {
		background: #46d369;
		color: #0c2a14;
	}
	.voter.no .vmark {
		background: #ff6b6b;
		color: #3a0a0a;
	}
	.vote-actions {
		display: flex;
		gap: 8px;
	}
	.vote-actions .btn {
		flex: 1;
	}

	/* ---- your big card ---- */
	.hand {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 10px;
	}
	.hint {
		font-size: 13px;
		color: #cdb8d6;
	}
	.bigcard {
		position: relative;
		width: min(64vw, 240px);
		aspect-ratio: 5 / 7;
		perspective: 1000px;
		border: 0;
		padding: 0;
		background: none;
		cursor: pointer;
		transition: transform 0.15s;
	}
	.bigcard:active {
		transform: scale(0.97);
	}
	.inner {
		position: absolute;
		inset: 0;
		transform-style: preserve-3d;
		transition: transform 0.6s cubic-bezier(0.3, 0.8, 0.3, 1);
	}
	.flipped > .inner {
		transform: rotateY(180deg);
	}
	.face {
		position: absolute;
		inset: 0;
		backface-visibility: hidden;
		-webkit-backface-visibility: hidden;
		display: grid;
		place-items: center;
		border-radius: 18px;
	}
	.bigcard .front {
		background: linear-gradient(160deg, #fff8ea, #ffe6bd);
		color: var(--ink);
		box-shadow: 0 16px 36px #0008, inset 0 0 0 4px #fff8ea, inset 0 0 0 7px var(--gold);
		animation: float 4s ease-in-out infinite;
	}
	.big {
		font-family: var(--display);
		font-size: clamp(96px, 30vw, 130px);
		line-height: 1;
		color: var(--wine);
	}
	.corner {
		position: absolute;
		font-family: var(--display);
		font-size: 26px;
		line-height: 1;
		color: var(--wine);
	}
	.corner.tl {
		top: 16px;
		left: 18px;
	}
	.corner.br {
		bottom: 28px;
		right: 18px;
		rotate: 180deg;
	}
	.mini-scale {
		position: absolute;
		left: 18px;
		right: 18px;
		bottom: 14px;
		height: 6px;
		border-radius: 999px;
		background: linear-gradient(90deg, #4aa3ff, #ffb347, #e5383b);
	}
	.mini-scale i {
		position: absolute;
		top: -3px;
		width: 12px;
		height: 12px;
		margin-left: -6px;
		border-radius: 50%;
		background: var(--ink);
		border: 2px solid #fff8ea;
	}
	.bigcard .back,
	.tiny .back,
	.dockcard .back {
		transform: rotateY(180deg);
		background:
			repeating-linear-gradient(45deg, #ffffff10 0 8px, transparent 8px 16px),
			linear-gradient(160deg, #7a2e3a, #4a1a3a);
		box-shadow: 0 16px 36px #0008, inset 0 0 0 4px #4a1a3a, inset 0 0 0 6px var(--gold);
		color: var(--gold);
	}
	.logo {
		font-family: var(--display);
		font-size: 54px;
		line-height: 0.9;
		letter-spacing: 3px;
		text-align: center;
	}

	/* ---- panels ---- */
	.panel {
		background: #2f1a3d;
		border: 1px solid #4a2d59;
		border-radius: 18px;
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.panel h2,
	.panel h3,
	.line-head h2 {
		margin: 0;
		font-size: 18px;
	}
	.panel h3 {
		font-size: 14px;
		color: #cdb8d6;
	}
	.line-head {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.small {
		font-size: 13px;
		color: #cdb8d6;
		margin: 0;
	}
	.center {
		text-align: center;
	}
	.clue-form {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
		width: 100%;
	}
	input {
		flex: 1;
		min-width: 0;
		min-height: 48px;
		font-size: 16px;
		padding: 0 14px;
		border-radius: 12px;
		border: 1px solid #5a3a68;
		background: #1a0f22;
		color: inherit;
		transition: border-color 0.2s, box-shadow 0.2s;
	}
	input:focus-visible {
		outline: none;
		border-color: var(--gold);
		box-shadow: 0 0 0 3px #ffb34740;
	}
	.sr {
		position: absolute;
		left: -9999px;
	}
	.btn {
		min-height: 48px;
		padding: 0 20px;
		border-radius: 12px;
		border: 1px solid transparent;
		font: inherit;
		font-weight: 700;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		transition: transform 0.12s, opacity 0.2s, filter 0.2s;
	}
	.btn:active:not(:disabled) {
		transform: scale(0.97);
	}
	.btn.primary {
		background: linear-gradient(180deg, #ffc067, var(--gold));
		color: var(--ink);
		box-shadow: 0 4px 14px #ffb34733;
	}
	.btn.ghost {
		border-color: #5a3a68;
		background: transparent;
		color: inherit;
	}
	.btn:disabled {
		opacity: 0.5;
	}
	.block {
		width: 100%;
	}
	.ghost {
		min-height: 44px;
		padding: 0 14px;
		border-radius: 10px;
		border: 1px solid #5a3a68;
		background: transparent;
		color: inherit;
		font: inherit;
		transition: transform 0.12s, background 0.2s;
	}
	.ghost:active {
		transform: scale(0.96);
		background: #3a2148;
	}
	.pulse {
		animation: glow 1.8s ease-in-out infinite;
	}
	.sent {
		display: flex;
		align-items: center;
		gap: 6px;
		margin: 0;
	}
	.who-row {
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
		align-items: center;
	}
	.count {
		margin-left: 4px;
	}
	.chip {
		position: relative;
		opacity: 0.45;
		filter: grayscale(0.6);
		transition: opacity 0.3s, filter 0.3s;
		animation: breathe 1.8s ease-in-out infinite;
	}
	.chip.done {
		opacity: 1;
		filter: none;
		animation: none;
	}
	.tick {
		position: absolute;
		right: -4px;
		bottom: -4px;
		width: 18px;
		height: 18px;
		border-radius: 50%;
		background: #46d369;
		color: #0c2a14;
		display: grid;
		place-items: center;
	}

	/* ---- the line of big cards ---- */
	.rail-end {
		font-size: 12px;
		font-weight: 700;
		letter-spacing: 1px;
		text-transform: uppercase;
	}
	.rail-end.low {
		color: #7bb8ff;
	}
	.rail-end.high {
		color: #ff8a8a;
		text-align: right;
	}
	.line {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.item {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.slot {
		width: 100%;
		min-height: 48px;
		border-radius: 14px;
		border: 2px dashed #ffb34799;
		background: #ffb3470d;
		color: var(--gold);
		font: inherit;
		font-weight: 700;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 10px;
		transition: background 0.2s, transform 0.12s;
		animation: breathe 2.2s ease-in-out infinite;
	}
	.plus {
		font-size: 24px;
		line-height: 1;
	}
	.slot:active:not(:disabled) {
		background: #ffb34733;
		transform: scale(0.98);
	}
	.lcard {
		position: relative;
		display: grid;
		grid-template-columns: auto auto 1fr auto;
		align-items: center;
		gap: 12px;
		min-height: 96px;
		padding: 12px 14px 12px 16px;
		border-radius: 18px;
		background: #fff4e0;
		color: var(--ink);
		overflow: hidden;
		box-shadow: 0 8px 20px #0005;
		transition: background 0.4s, box-shadow 0.4s;
	}
	.lcard::before {
		content: '';
		position: absolute;
		left: 0;
		top: 0;
		bottom: 0;
		width: 7px;
		background: hsl(var(--spot) 85% 55%);
	}
	.lcard.mine {
		box-shadow: 0 0 0 3px var(--gold), 0 8px 20px #0005;
	}
	.lcard.wrong {
		background: #ffd9d9;
		animation: shake 0.5s;
	}
	.lcard.right {
		background: #d8f7e1;
		animation: pop 0.4s;
	}
	.rank {
		font-family: var(--display);
		font-size: 30px;
		line-height: 1;
		color: hsl(var(--spot) 70% 38%);
		min-width: 18px;
		text-align: center;
	}
	.who {
		border: 0;
		background: none;
		padding: 0;
		min-width: 48px;
		min-height: 48px;
		display: grid;
		place-items: center;
	}
	.txt {
		display: flex;
		flex-direction: column;
		min-width: 0;
		gap: 2px;
	}
	.pname {
		font-size: 13px;
		font-weight: 700;
		color: #6b4a7a;
	}
	.clue {
		font-size: 19px;
		font-weight: 600;
		font-style: italic;
		line-height: 1.15;
		overflow-wrap: anywhere;
	}
	.tiny {
		position: relative;
		width: 58px;
		height: 80px;
		perspective: 600px;
	}
	.tiny .face {
		border-radius: 10px;
	}
	/* Line cards start face down; "flipped" here means face up. */
	.tiny > .inner {
		transform: rotateY(180deg);
	}
	.tiny.flipped > .inner {
		transform: rotateY(0deg);
	}
	.tiny .front {
		background: linear-gradient(160deg, #fff, #f2e6d0);
		box-shadow: inset 0 0 0 2px #fff, inset 0 0 0 3px #d9c8a8;
		color: #1d7a3c;
	}
	.tiny .front.bad {
		color: #c0262b;
	}
	.tiny .back {
		box-shadow: inset 0 0 0 2px #4a1a3a, inset 0 0 0 3px var(--gold);
	}
	.tn {
		font-family: var(--display);
		font-size: 34px;
		line-height: 1;
	}
	.tn.gold {
		color: var(--wine);
	}
	.q {
		font-family: var(--display);
		font-size: 32px;
	}
	.verdict {
		margin: 0;
		font-weight: 700;
		font-size: 17px;
		color: #ff8a8a;
	}
	.verdict.good,
	.over h1.good {
		color: #6ee08f;
	}
	.pending {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.pending li {
		display: flex;
		align-items: center;
		gap: 10px;
		overflow-wrap: anywhere;
	}

	/* ---- dock ---- */
	.dock {
		position: sticky;
		bottom: 0;
		z-index: 10;
		margin: 6px -16px calc(-24px - var(--safe-bottom));
		padding: 12px 16px calc(14px + var(--safe-bottom));
		background: linear-gradient(180deg, #241230d9, #1a0f22);
		backdrop-filter: blur(10px);
		border-top: 1px solid #4a2d59;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.dock-row {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.dock-text {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		overflow-wrap: anywhere;
	}
	.dockcard {
		position: relative;
		width: 52px;
		height: 72px;
		perspective: 600px;
		border: 0;
		padding: 0;
		background: none;
		flex: none;
	}
	.dockcard .front {
		background: linear-gradient(160deg, #fff8ea, #ffe6bd);
		color: var(--wine);
		font-family: var(--display);
		font-size: 30px;
		border-radius: 10px;
		box-shadow: inset 0 0 0 2px #fff8ea, inset 0 0 0 3px var(--gold);
	}
	.dockcard .face {
		border-radius: 10px;
		font-family: var(--display);
		font-size: 28px;
	}

	/* ---- the end ---- */
	.over {
		text-align: center;
		align-items: stretch;
	}
	.stars {
		display: flex;
		justify-content: center;
		gap: 10px;
		font-size: 52px;
		line-height: 1;
	}
	.star {
		color: #4a2d59;
	}
	.star.on {
		color: var(--gold);
		text-shadow: 0 0 18px #ffb347aa;
	}
	.over h1 {
		margin: 0;
		font-family: var(--display);
		font-size: 40px;
		line-height: 1;
		color: #ffd29a;
	}
	.over p {
		margin: 0;
	}
	.history,
	.final {
		margin: 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 6px;
		text-align: left;
	}
	.mark {
		display: inline-block;
		min-width: 28px;
		text-align: center;
		border-radius: 6px;
		background: #7a2e3a;
		font-weight: 700;
		font-size: 12px;
		padding: 2px 6px;
		margin-right: 8px;
	}
	.mark.good {
		background: #1d7a3c;
	}
	.final li {
		display: flex;
		justify-content: space-between;
	}
	.final li.wrong {
		color: #ff8a8a;
	}
	.error {
		color: #ff8a8a;
		margin: 0;
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
		color: #cdb8d6;
		text-decoration: underline;
		font: inherit;
	}

	/* ---- confetti ---- */
	.confetti {
		position: fixed;
		inset: 0;
		pointer-events: none;
		overflow: hidden;
		z-index: 40;
	}
	.confetti i {
		position: absolute;
		top: -12px;
		width: 9px;
		height: 14px;
		border-radius: 2px;
		animation-name: fall;
		animation-timing-function: cubic-bezier(0.3, 0.6, 0.5, 1);
		animation-fill-mode: both;
	}

	@keyframes breathe {
		50% {
			opacity: 0.6;
		}
	}
	@keyframes float {
		50% {
			transform: translateY(-5px);
		}
	}
	@keyframes glow {
		50% {
			box-shadow: 0 0 0 6px #ffb34726, 0 4px 14px #ffb34744;
		}
	}
	@keyframes shake {
		20% {
			transform: translateX(-6px);
		}
		40% {
			transform: translateX(6px);
		}
		60% {
			transform: translateX(-4px);
		}
		80% {
			transform: translateX(3px);
		}
	}
	@keyframes pop {
		40% {
			transform: scale(1.03);
		}
	}
	@keyframes fall {
		to {
			transform: translateY(105vh) rotate(var(--spin));
			opacity: 0.9;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.es *,
		.es *::before {
			animation: none !important;
			transition: none !important;
		}
		.confetti {
			display: none;
		}
	}
</style>
