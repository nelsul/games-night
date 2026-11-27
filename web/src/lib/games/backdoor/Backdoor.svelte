<script lang="ts">
	// A Backdoor match. The look borrows from phone settings screens (big titles, actions
	// at the bottom) and code review tools (PRs, diffs, pipelines). Everything shown comes
	// from the server's view for this player; this file never decides a rule.
	import { onDestroy } from 'svelte';
	import { fly, scale } from 'svelte/transition';
	import Title from './Title.svelte';
	import { feel, haptic, ms, saveFeel, sound } from '$lib/feel.svelte';
	import { api, ApiError, type Room, type RoomPlayer } from '$lib/api';
	import Avatar from '$lib/components/Avatar.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import Badge from './Badge.svelte';
	import { roleInfo, type MatchView } from './types';

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

	const onlineCount = $derived(room.players.filter((p) => p.online).length);

	// Sounds and haptics for what happens at the table (not only for taps).
	let prev: { phase: string; results: number; proposals: number } | null = null;
	$effect(() => {
		const cur = { phase: match.phase, results: match.results.length, proposals: match.proposals.length };
		if (!prev) {
			prev = cur; // first render: nothing happened yet
			return;
		}
		if (cur.results > prev.results) {
			const r = match.results.at(-1);
			sound(r?.rollback ? 'rollback' : 'deploy');
			haptic(r?.rollback ? 'warning' : 'success');
		} else if (cur.proposals > prev.proposals) {
			sound('vote');
		}
		if (cur.phase !== prev.phase) {
			if (cur.phase === 'over') {
				const won = match.winner === match.me.team;
				sound(won ? 'win' : 'lose');
				haptic(won ? 'success' : 'warning');
			} else if (
				(cur.phase === 'proposal' && match.me.is_leader) ||
				(cur.phase === 'push' && match.me.in_squad) ||
				(cur.phase === 'guess' && match.me.role === 'social_engineer')
			) {
				sound('turn');
				haptic('select');
			}
		}
		prev = cur;
	});

	function toggleSound() {
		feel.sound = !feel.sound;
		saveFeel();
		if (feel.sound) sound('select');
	}

	const people = $derived(Object.fromEntries(room.players.map((p) => [p.id, p])) as Record<string, RoomPlayer>);
	const name = (id: string) => (id === me ? 'você' : (people[id]?.name ?? '?'));
	const leaderName = $derived(name(match.leader));

	let error = $state('');
	let busy = $state(false);
	let showBadge = $state(false);
	let confirmStop = $state(false);

	async function act(action: Record<string, unknown>) {
		error = '';
		busy = true;
		try {
			onroom(await api.post<Room>(`/rooms/${room.code}/action`, action));
			if (action.type === 'push') sound('push');
			if (action.type === 'select') sound('select');
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
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

	// Squad picking: the Scrum Master's taps are sent at once so everyone sees them live.
	function toggle(id: string) {
		if (!match.me.is_leader || busy) return;
		const sel = match.selection.includes(id)
			? match.selection.filter((x) => x !== id)
			: match.selection.length < match.squad_size
				? [...match.selection, id]
				: match.selection;
		if (sel !== match.selection) act({ type: 'select', squad: sel });
	}

	let pushChoice = $state<'clean' | 'backdoor'>('clean');
	let guessTarget = $state('');
	// Every sprint starts from a clean push, so a backdoor is never chosen by accident.
	$effect(() => {
		void match.sprint;
		pushChoice = 'clean';
	});

	// Discussion timer: advisory countdown while the squad is being built.
	let now = $state(Date.now());
	const tick = setInterval(() => (now = Date.now()), 500);
	onDestroy(() => clearInterval(tick));
	const remaining = $derived(
		match.options.timer_seconds > 0 && match.phase === 'proposal'
			? Math.max(0, match.options.timer_seconds - Math.floor((now - Date.parse(match.phase_started_at)) / 1000))
			: null
	);
	let buzzed = '';
	$effect(() => {
		if (remaining === 0 && buzzed !== match.phase_started_at) {
			buzzed = match.phase_started_at;
			navigator.vibrate?.(300);
		}
	});
	const mmss = (s: number) => `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;

	const lastProposal = $derived(match.proposals.at(-1));
	const lastResult = $derived(match.results.at(-1));
	// Show the latest review right after it happened (push phase, or a new squad after a refusal).
	const showReview = $derived(
		lastProposal &&
			((match.phase === 'push' && lastProposal.approved) ||
				(match.phase === 'proposal' && !lastProposal.approved && match.failed_meetings > 0))
	);
	// Show the latest sprint's result at the start of the next one.
	const showResult = $derived(lastResult && match.phase === 'proposal' && match.failed_meetings === 0 && match.results.length === match.sprint);

	const readyCount = $derived(match.players.filter((p) => p.ready).length);
	const votedCount = $derived(match.players.filter((p) => p.voted).length);
	const pushedCount = $derived(match.players.filter((p) => p.pushed).length);
	const deploys = $derived(match.results.filter((r) => !r.rollback).length);
	const rollbacks = $derived(match.results.filter((r) => r.rollback).length);

	const reasonText: Record<string, string> = {
		rollbacks: '3 sprints com rollback. Os hackers derrubaram a release.',
		meetings: '5 reuniões seguidas sem decisão. O projeto foi cancelado.',
		guess: 'A Engenharia Social achou o Analista de Segurança.',
		guess_missed: 'A Engenharia Social errou o palpite. A release está no ar.'
	};
</script>

<div class="bd">
	<header class="bar">
		<a href="/" aria-label="Voltar ao início"><Icon name="back" /></a>
		<Title size={28} />
		<span class="bar-actions">
		<button class="pill icon" onclick={toggleSound} aria-label={feel.sound ? 'Desligar sons' : 'Ligar sons'} aria-pressed={feel.sound}>
			<Icon name={feel.sound ? 'sound' : 'mute'} size={18} />
		</button>
		<button class="pill icon chat" onclick={onchat} aria-label={unread ? `Chat, ${unread} novas` : 'Chat'}>
			<Icon name="chat" size={18} />{#if unread}<span class="badge" aria-hidden="true">{unread}</span>{/if}
		</button>
		<button class="pill" onclick={() => (showBadge = true)} aria-label="Ver meu crachá">
			<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"
				><rect x="5" y="3" width="14" height="18" rx="3" /><circle cx="12" cy="10" r="3" /><path d="M8 17h8" /></svg
			>
			Crachá
		</button>
		</span>
	</header>

	<div class="presence" aria-label="Quem está na partida">
		<span class="faces">
			{#each room.players as p (p.id)}
				<button class="face" onclick={() => onplayer?.(p.id)} aria-label="Ver {name(p.id)}"
					><Avatar user={p} size={24} online={p.online} /></button
				>
			{/each}
		</span>
		<span class="muted tiny">{onlineCount}/{room.players.length} online</span>
	</div>

	<section class="card sprints" aria-label="Sprints">
		<div class="row muted small">
			<span>Sprint {Math.min(match.sprint + 1, 5)} de 5 · {deploys} deploy · {rollbacks} rollback</span>
			<span>Reuniões <strong class:warn={match.failed_meetings >= 3}>{match.failed_meetings}</strong>/{match.max_failed}</span>
		</div>
		<ol class="track">
			{#each match.squad_sizes as size, i (i)}
				{@const r = match.results[i]}
				<li
					class:ok={r && !r.rollback}
					class:bad={r?.rollback}
					class:now={!r && i === match.sprint && match.phase !== 'over'}
					aria-label="Sprint {i + 1}: {r ? (r.rollback ? 'rollback' : 'deploy') : `${size} pessoas`}{i === match.double_sprint
						? ', precisa de 2 backdoors'
						: ''}"
				>
					{#if r && !r.rollback}<Icon name="check" size={18} />
					{:else if r}<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" aria-hidden="true"><path d="M4 12a8 8 0 1 0 2.3-5.7M4 4v5h5" /></svg>
					{:else}{size}{#if i === match.double_sprint}<sup>★</sup>{/if}{/if}
				</li>
			{/each}
		</ol>
		{#if match.double_sprint >= 0 && match.sprint <= match.double_sprint && match.phase !== 'over' && match.phase !== 'guess'}
			<p class="muted tiny">★ Na sprint {match.double_sprint + 1}, só dá rollback com 2 backdoors.</p>
		{/if}
	</section>

	{#if showResult && lastResult}
		<section class="card event" class:bad={lastResult.rollback} in:fly={{ y: -12, duration: ms(300) }}>
			<strong>Sprint {match.results.length}: {lastResult.rollback ? 'Rollback' : 'Deploy em produção'}</strong>
			<span class="muted small"
				>{lastResult.backdoors === 0
					? 'Nenhuma backdoor no código.'
					: lastResult.backdoors === 1
						? '1 backdoor encontrada no código.'
						: `${lastResult.backdoors} backdoors encontradas no código.`}
				Squad: {lastResult.squad.map(name).join(', ')}</span
			>
		</section>
	{/if}

	{#if showReview && lastProposal}
		<section class="card event" class:bad={!lastProposal.approved} in:fly={{ y: -12, duration: ms(300) }}>
			<strong>PR de {name(lastProposal.leader)}: {lastProposal.approved ? 'aprovado' : 'recusado'}</strong>
			<div class="votes">
				{#each Object.entries(lastProposal.votes) as [id, yes], i (id)}
					<span class="vote" class:yes in:scale={{ start: 0.6, duration: ms(220), delay: ms(120 + i * 70) }}><Avatar user={people[id]} size={22} /> {name(id)}: {yes ? 'Approve' : 'Changes'}</span>
				{/each}
			</div>
		</section>
	{/if}

	<main class="phase">
		{#key match.phase}
		<div class="phase-body" in:fly={{ y: 18, duration: ms(280) }}>
		{#if match.phase === 'badges'}
			<h1>Seu crachá</h1>
			<p class="muted">Só você está vendo isto.</p>
			<Badge {match} {people} />
			<div class="bottom">
				<p class="muted small center">{readyCount} de {match.players.length} prontos</p>
				<button class="btn bd-primary btn-block" disabled={match.me.ready || busy} onclick={() => act({ type: 'ready' })}>
					{match.me.ready ? 'Esperando os outros…' : 'Entendi, bora codar'}
				</button>
			</div>
		{:else if match.phase === 'proposal'}
			<div class="head">
				<h1>{match.me.is_leader ? 'Monte o squad' : 'Montando o squad'}</h1>
				{#if remaining !== null}
					<span class="timer" class:done={remaining === 0} role="timer" aria-live="off">{remaining === 0 ? 'Tempo!' : mmss(remaining)}</span>
				{/if}
			</div>
			<p class="muted">
				{#if match.me.is_leader}
					Você é o Scrum Master. Escolha {match.squad_size} pessoas e abra o PR.
				{:else}
					<strong class="hl">{leaderName}</strong> é o Scrum Master e está escolhendo {match.squad_size} pessoas.
				{/if}
			</p>
			<ul class="pick" aria-label="Jogadores">
				{#each match.players as p (p.id)}
					{@const on = match.selection.includes(p.id)}
					<li>
						<button
							type="button"
							class="person"
							class:on
							aria-pressed={on}
							disabled={!match.me.is_leader || busy}
							onclick={() => toggle(p.id)}
						>
							<Avatar user={people[p.id]} size={36} online={people[p.id]?.online} />
							<span class="pname">{name(p.id)}{#if p.is_leader}<span class="tag">Scrum Master</span>{/if}</span>
							<span class="check" aria-hidden="true">{#if on}<Icon name="check" size={18} />{/if}</span>
						</button>
					</li>
				{/each}
			</ul>
			{#if match.me.is_leader}
				<div class="bottom">
					<p class="muted small center">{match.selection.length} de {match.squad_size} escolhidos</p>
					<button
						class="btn bd-primary btn-block"
						disabled={match.selection.length !== match.squad_size || busy}
						onclick={() => act({ type: 'propose' })}>Abrir PR</button
					>
				</div>
			{/if}
		{:else if match.phase === 'review'}
			<h1>Review do squad</h1>
			<section class="card pr">
				<div class="pr-head">
					<span class="open">Open</span>
					<strong>Squad da sprint {match.sprint + 1}</strong>
				</div>
				<span class="muted small">{leaderName} quer escalar {match.squad_size} pessoas</span>
				<div class="diff">
					{#each match.selection as id (id)}
						<div><span class="plus">+</span><Avatar user={people[id]} size={24} />{name(id)}</div>
					{/each}
				</div>
				<div class="muted small">Reviews {votedCount}/{match.players.length} · aparecem todos juntos no final</div>
			</section>
			<div class="bottom">
				{#if match.me.vote === null}
					<div class="two">
						<button class="btn bd-ghost red" disabled={busy} onclick={() => act({ type: 'vote', approve: false })}>Request changes</button>
						<button class="btn bd-green" disabled={busy} onclick={() => act({ type: 'vote', approve: true })}>
							<Icon name="check" size={18} />Approve
						</button>
					</div>
				{:else}
					<p class="muted center">Você votou <strong class="hl">{match.me.vote ? 'Approve' : 'Request changes'}</strong>. Esperando os outros…</p>
				{/if}
			</div>
		{:else if match.phase === 'push'}
			<h1>Hora do push</h1>
			<p class="muted">Ninguém vê o seu commit. Só quantas backdoors entraram.</p>
			<section class="card pipeline mono small">
				{#each match.selection as id (id)}
					{@const pushed = match.players.find((p) => p.id === id)?.pushed}
					<div class="step">
						<span class="dot" class:done={pushed}>{#if pushed}<Icon name="check" size={12} />{/if}</span>
						<span>{name(id)}</span>
						<span class="muted right">{pushed ? 'push feito' : 'aguardando'}</span>
					</div>
				{/each}
			</section>
			{#if match.me.in_squad && !match.me.pushed}
				<div class="choices" role="radiogroup" aria-label="Seu commit">
					<button type="button" role="radio" aria-checked={pushChoice === 'clean'} class="choice ok" onclick={() => (pushChoice = 'clean')}>
						<span class="mono big">git push</span><span class="small">Commit limpo. A sprint segue.</span>
					</button>
					<button
						type="button"
						role="radio"
						aria-checked={pushChoice === 'backdoor'}
						class="choice bad"
						disabled={!match.me.can_backdoor}
						onclick={() => (pushChoice = 'backdoor')}
					>
						<span class="mono big">git push --backdoor</span>
						<span class="small">{match.me.can_backdoor ? 'Derruba a sprint.' : 'Só hackers. Você é dev.'}</span>
					</button>
				</div>
				<div class="bottom">
					<button class="btn bd-primary btn-block" disabled={busy} onclick={() => act({ type: 'push', backdoor: pushChoice === 'backdoor' })}>
						Fazer push
					</button>
				</div>
			{:else}
				<p class="muted center bottom">
					{match.me.in_squad ? 'Push feito.' : 'Você não está no squad.'} Esperando {match.selection.length - pushedCount}
					{match.selection.length - pushedCount === 1 ? 'push' : 'pushes'}…
				</p>
			{/if}
		{:else if match.phase === 'guess'}
			<h1>Última chance</h1>
			<p class="muted">
				Os devs entregaram 3 sprints. Agora <strong class="hl">{name(match.social_engineer ?? '')}</strong>, a Engenharia Social,
				aponta quem acha que é o Analista de Segurança. Se acertar, os hackers vencem.
			</p>
			{#if match.me.role === 'social_engineer'}
				<ul class="pick" aria-label="Quem é o Analista?">
					{#each match.players.filter((p) => p.id !== me) as p (p.id)}
						<li>
							<button type="button" class="person" class:on={guessTarget === p.id} aria-pressed={guessTarget === p.id} onclick={() => (guessTarget = p.id)}>
								<Avatar user={people[p.id]} size={36} />
								<span class="pname">{name(p.id)}</span>
								<span class="check" aria-hidden="true">{#if guessTarget === p.id}<Icon name="check" size={18} />{/if}</span>
							</button>
						</li>
					{/each}
				</ul>
				<div class="bottom">
					<button class="btn bd-red btn-block" disabled={!guessTarget || busy} onclick={() => act({ type: 'guess', target: guessTarget })}>
						É essa pessoa
					</button>
				</div>
			{:else}
				<p class="muted center bottom">{match.me.team === 'hackers' ? 'Ajudem a Engenharia Social a escolher.' : 'Esperando o palpite…'}</p>
			{/if}
		{:else if match.phase === 'over'}
			<span class="sev" class:won={match.winner === 'devs'}>{match.winner === 'devs' ? 'RELEASE 1.0 NO AR' : 'INCIDENTE SEV-1'}</span>
			<h1>Post-mortem</h1>
			<p class="muted">
				<strong class="hl">{match.winner === 'devs' ? 'Os devs venceram.' : 'Os hackers venceram.'}</strong>
				{reasonText[match.reason ?? ''] ?? ''}
				{#if match.guess}Palpite: {name(match.guess)}.{/if}
			</p>
			<ul class="reveal">
				{#each match.players as p, i (p.id)}
					{@const role = match.roles?.[p.id] ?? 'dev'}
					<li in:fly={{ x: -16, duration: ms(260), delay: ms(150 + i * 90) }}>
						<button class="reveal-row" onclick={() => onplayer?.(p.id)} aria-label="Ver {name(p.id)}">
							<Avatar user={people[p.id]} size={32} />
							<span class="pname">{name(p.id)}</span>
							<span class="role-tag" class:hk={roleInfo[role].team === 'hackers'}>{roleInfo[role].name}</span>
						</button>
					</li>
				{/each}
			</ul>
			<details class="history">
				<summary>Histórico das sprints</summary>
				{#each match.proposals as pr, i (i)}
					<div class="h-item">
						<strong>Sprint {pr.sprint + 1} · PR de {name(pr.leader)} · {pr.approved ? 'aprovado' : 'recusado'}</strong>
						<span class="muted small">Squad: {pr.squad.map(name).join(', ')}</span>
						<span class="muted small"
							>Approve: {Object.entries(pr.votes).filter(([, y]) => y).map(([id]) => name(id)).join(', ') || 'ninguém'}</span
						>
					</div>
				{/each}
			</details>
			<div class="bottom">
				{#if isHost}
					<button class="btn bd-primary btn-block" onclick={stop}>Jogar de novo</button>
				{:else}
					<p class="muted center">Esperando o host começar outra partida.</p>
				{/if}
				<button class="btn bd-ghost btn-block" onclick={onleave}><Icon name="door" size={18} />Sair da sala</button>
			</div>
		{/if}

		</div>
		{/key}

		{#if error}<p class="error" role="alert">{error}</p>{/if}

		{#if isHost && match.phase !== 'over'}
			<button class="link" onclick={() => (confirmStop = true)}>Encerrar partida</button>
		{/if}
	</main>
</div>

{#if showBadge}
	<div class="sheet" role="dialog" aria-modal="true" aria-label="Seu crachá">
		<div class="sheet-card bd">
			<Badge {match} {people} />
			<button class="btn bd-primary btn-block" onclick={() => (showBadge = false)}>Fechar</button>
		</div>
	</div>
{/if}

{#if confirmStop}
	<div class="sheet" role="dialog" aria-modal="true" aria-labelledby="stop-title">
		<div class="sheet-card bd">
			<h2 id="stop-title">Encerrar a partida?</h2>
			<p class="muted">Todo mundo volta para o lobby e a partida não conta.</p>
			<button class="btn bd-red btn-block" onclick={stop}>Encerrar</button>
			<button class="btn bd-ghost btn-block" onclick={() => (confirmStop = false)}>Continuar jogando</button>
		</div>
	</div>
{/if}

<style>
	.bd {
		--bd-bg: #0d1117;
		--bd-surface: #161b22;
		--bd-line: #30363d;
		--bd-text: #e6edf3;
		--bd-muted: #9198a1;
		--bd-green: #3fb950;
		--bd-green-btn: #238636;
		--bd-red: #f85149;
		--bd-amber: #d29922;
		--bd-blue: #4493f8;
		background: var(--bd-bg);
		color: var(--bd-text);
		min-height: 100dvh;
		max-width: 560px;
		margin: 0 auto;
		padding: calc(10px + var(--safe-top)) 16px calc(24px + var(--safe-bottom));
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	:global(body:has(.bd)) {
		background: #0d1117;
	}
	.bar {
		height: 48px;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	.bar a {
		width: 44px;
		height: 44px;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-left: -10px;
	}
	.presence {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		margin-top: -4px;
	}
	.faces {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.face {
		display: inline-flex;
		border: none;
		background: none;
		padding: 2px;
		border-radius: 6px;
		cursor: pointer;
	}
	.reveal-row {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 12px;
		border: none;
		background: none;
		padding: 0;
		text-align: left;
		cursor: pointer;
	}
	.phase-body {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.pill.icon {
		width: 40px;
		justify-content: center;
		padding: 0;
	}
	.track li.ok,
	.track li.bad {
		animation: gn-pop 0.45s ease-out;
	}
	.track li.now {
		animation: bd-now 2s ease-in-out infinite;
	}
	@keyframes bd-now {
		0%,
		100% {
			box-shadow: 0 0 0 0 rgba(210, 153, 34, 0.45);
		}
		50% {
			box-shadow: 0 0 0 6px rgba(210, 153, 34, 0);
		}
	}
	.event.bad {
		animation: gn-shake 0.5s ease-in-out;
	}
	.sev {
		animation: gn-pop 0.5s ease-out;
	}
	.person {
		transition:
			border-color 0.15s ease,
			background-color 0.15s ease,
			transform 0.12s ease;
	}
	.pill {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		height: 36px;
		padding: 0 12px;
		border-radius: 999px;
		border: 1px solid var(--bd-line);
		background: var(--bd-surface);
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
	}
	.bar-actions {
		display: flex;
		gap: 8px;
	}
	.pill.chat {
		position: relative;
		justify-content: center;
		padding: 0;
	}
	.badge {
		position: absolute;
		top: -6px;
		right: -6px;
		min-width: 18px;
		height: 18px;
		padding: 0 5px;
		border-radius: 9px;
		background: var(--bd-red);
		color: #fff;
		font-size: 11px;
		font-weight: 700;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.card {
		background: var(--bd-surface);
		border: 1px solid var(--bd-line);
		border-radius: 20px;
		padding: 12px 14px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.row {
		display: flex;
		justify-content: space-between;
		gap: 8px;
	}
	.warn {
		color: var(--bd-amber);
	}
	.track {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		grid-template-columns: repeat(5, minmax(0, 1fr));
		gap: 8px;
	}
	.track li {
		height: 40px;
		border-radius: 12px;
		border: 2px dashed var(--bd-line);
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
		color: var(--bd-muted);
	}
	.track li.now {
		border: 2px solid var(--bd-amber);
		color: var(--bd-amber);
	}
	.track li.ok {
		border: none;
		background: var(--bd-green-btn);
		color: #fff;
	}
	.track li.bad {
		border: none;
		background: #da3633;
		color: #fff;
	}
	.event {
		border-left: 3px solid var(--bd-green);
	}
	.event.bad {
		border-left-color: var(--bd-red);
	}
	.votes {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.vote {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		padding: 3px 8px 3px 3px;
		border-radius: 999px;
		background: #2a1215;
		color: #ffa198;
	}
	.vote.yes {
		background: #12261e;
		color: #8bd5a0;
	}
	.phase {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding-top: 6px;
	}
	h1 {
		margin: 8px 0 0;
		font-size: 32px;
		font-weight: 300;
		letter-spacing: -0.6px;
	}
	.head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 8px;
	}
	.timer {
		font-family: var(--mono);
		font-size: 20px;
		font-weight: 700;
		padding: 4px 12px;
		border-radius: 999px;
		background: #2b2111;
		color: var(--bd-amber);
	}
	.timer.done {
		background: #da3633;
		color: #fff;
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
	.tiny {
		font-size: 12px;
	}
	.center {
		text-align: center;
	}
	.mono {
		font-family: var(--mono);
	}
	.pick,
	.reveal {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.person {
		width: 100%;
		min-height: 56px;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 8px 12px;
		border-radius: 16px;
		border: 1px solid var(--bd-line);
		background: var(--bd-surface);
		text-align: left;
		cursor: pointer;
	}
	.person:disabled {
		cursor: default;
		opacity: 1;
	}
	.person.on {
		border-color: var(--bd-green);
		background: rgba(46, 160, 67, 0.15);
	}
	.pname {
		flex: 1;
		display: flex;
		align-items: center;
		gap: 8px;
		font-weight: 600;
		min-width: 0;
	}
	.tag {
		font-size: 11px;
		font-weight: 700;
		color: var(--bd-blue);
	}
	.check {
		width: 24px;
		height: 24px;
		border-radius: 50%;
		border: 2px solid var(--bd-line);
		display: flex;
		align-items: center;
		justify-content: center;
		color: #fff;
	}
	.on .check {
		border-color: var(--bd-green);
		background: var(--bd-green-btn);
	}
	.bottom {
		margin-top: auto;
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding-top: 12px;
	}
	.two {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 10px;
	}
	.btn {
		border-radius: 999px;
		height: 54px;
	}
	.bd-primary {
		background: var(--bd-text);
		color: var(--bd-bg);
	}
	.bd-green {
		background: var(--bd-green-btn);
		color: #fff;
	}
	.bd-red {
		background: #da3633;
		color: #fff;
	}
	.bd-ghost {
		background: #21262d;
		border: 1px solid var(--bd-line);
		color: var(--bd-text);
	}
	.bd-ghost.red {
		color: var(--bd-red);
	}
	.pr-head {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.open {
		font-size: 12px;
		font-weight: 600;
		padding: 3px 10px;
		border-radius: 999px;
		background: var(--bd-green-btn);
		color: #fff;
	}
	.diff {
		font-family: var(--mono);
		font-size: 14px;
		border-radius: 10px;
		overflow: hidden;
	}
	.diff div {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 10px;
		background: rgba(46, 160, 67, 0.15);
	}
	.diff div + div {
		border-top: 1px solid var(--bd-bg);
	}
	.plus {
		color: var(--bd-green);
		font-weight: 700;
	}
	.step {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.right {
		margin-left: auto;
	}
	.dot {
		width: 20px;
		height: 20px;
		border-radius: 50%;
		border: 3px solid var(--bd-amber);
		border-top-color: transparent;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.dot.done {
		border: none;
		background: var(--bd-green-btn);
		color: #fff;
	}
	.choices {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.choice {
		display: flex;
		flex-direction: column;
		gap: 4px;
		align-items: flex-start;
		padding: 14px 16px;
		border-radius: 18px;
		border: 2px solid var(--bd-line);
		background: var(--bd-surface);
		text-align: left;
		cursor: pointer;
	}
	.choice .big {
		font-size: 17px;
		font-weight: 700;
	}
	.choice.ok .big {
		color: var(--bd-green);
	}
	.choice.bad .big {
		color: var(--bd-red);
	}
	.choice.ok[aria-checked='true'] {
		border-color: var(--bd-green);
		background: rgba(46, 160, 67, 0.15);
	}
	.choice.bad[aria-checked='true'] {
		border-color: var(--bd-red);
		background: rgba(248, 81, 73, 0.15);
	}
	.choice:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.sev {
		align-self: flex-start;
		font-family: var(--mono);
		font-size: 12px;
		font-weight: 700;
		padding: 4px 10px;
		border-radius: 999px;
		background: #da3633;
		color: #fff;
		margin-top: 12px;
	}
	.sev.won {
		background: var(--bd-green-btn);
	}
	.reveal li {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 8px 12px;
		border-radius: 14px;
		background: var(--bd-surface);
	}
	.role-tag {
		font-size: 12px;
		font-weight: 600;
		padding: 3px 10px;
		border-radius: 999px;
		background: #12261e;
		color: var(--bd-green);
	}
	.role-tag.hk {
		background: #2a1215;
		color: var(--bd-red);
	}
	.history {
		background: var(--bd-surface);
		border-radius: 14px;
		padding: 10px 12px;
	}
	.history summary {
		cursor: pointer;
		font-weight: 600;
		min-height: 32px;
		display: flex;
		align-items: center;
	}
	.h-item {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: 8px 0;
		border-top: 1px solid var(--bd-line);
		font-size: 14px;
	}
	.link {
		align-self: center;
		background: none;
		border: none;
		color: var(--bd-muted);
		text-decoration: underline;
		min-height: 44px;
		cursor: pointer;
		font-size: 14px;
	}
	.sheet {
		position: fixed;
		inset: 0;
		z-index: 50;
		background: rgba(0, 0, 0, 0.8);
		display: flex;
		align-items: flex-end;
		justify-content: center;
	}
	.sheet-card {
		width: 100%;
		min-height: auto;
		border-radius: 24px 24px 0 0;
		padding: 20px 16px calc(20px + var(--safe-bottom));
		gap: 14px;
	}
	.sheet-card h2 {
		margin: 0;
		font-size: 20px;
	}
</style>
