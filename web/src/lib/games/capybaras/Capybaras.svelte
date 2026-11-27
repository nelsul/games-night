<script lang="ts">
	// A Crazy Capybaras match. Everything shown comes from the server's view for this
	// player; this file never decides a rule. It only runs the Nem Vem countdown on the
	// phone and asks the server to apply the card once the countdown ends.
	import { onDestroy } from 'svelte';
	import { fade, fly, scale } from 'svelte/transition';
	import { backOut } from 'svelte/easing';
	import { feel, haptic, ms, saveFeel, sound } from '$lib/feel.svelte';
	import { api, ApiError, type Room, type RoomPlayer } from '$lib/api';
	import Avatar from '$lib/components/Avatar.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import Title from './Title.svelte';
	import CardFace from './CardFace.svelte';
	import CardArt from './art/CardArt.svelte';
	import Back from './art/Back.svelte';
	import { capys, cardInfo, type Card, type GameEvent, type MatchView } from './types';

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

	const people = $derived(Object.fromEntries(room.players.map((p) => [p.id, p])) as Record<string, RoomPlayer>);
	const name = (id?: string) => (!id ? '?' : id === me ? 'você' : (people[id]?.name ?? 'Alguém'));
	const Name = (id?: string) => {
		const n = name(id);
		return n.charAt(0).toUpperCase() + n.slice(1);
	};
	const cardName = (kind?: string) => (kind ? (cardInfo[kind]?.name ?? kind) : 'uma carta');

	const myTurn = $derived(match.turn === me && match.phase === 'play' && !match.pending && !match.me.out);
	const alive = $derived(match.players.filter((p) => !p.out));

	let error = $state('');
	let busy = $state(false);
	let confirmStop = $state(false);

	async function act(action: Record<string, unknown>, quiet = false) {
		if (!quiet) {
			error = '';
			busy = true;
		}
		try {
			onroom(await api.post<Room>(`/rooms/${room.code}/action`, action));
			return true;
		} catch (err) {
			if (!quiet) {
				error = err instanceof ApiError ? err.message : 'Algo deu errado.';
				haptic('warning');
			}
			return false;
		} finally {
			if (!quiet) busy = false;
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

	// ---- Clock: the server's time, whatever this phone's clock says ----
	let offset = $state(0);
	$effect(() => {
		offset = Date.parse(match.server_time) - Date.now();
	});
	let now = $state(Date.now());
	const tick = setInterval(() => (now = Date.now()), 100);
	onDestroy(() => clearInterval(tick));

	const remaining = $derived(match.pending ? Math.max(0, Date.parse(match.pending.deadline) - (now + offset)) : 0);
	const windowMs = 3000;

	// When the countdown ends, ask the server to apply the card. The player who played it
	// asks first; everyone else a moment later, in case that phone is asleep.
	let asked = '';
	$effect(() => {
		const pd = match.pending;
		if (!pd || remaining > 0 || asked === pd.deadline) return;
		asked = pd.deadline;
		const wait = pd.player === me ? 0 : 400 + Math.random() * 900;
		setTimeout(() => {
			if (match.pending?.deadline === pd.deadline) act({ type: 'resolve' }, true);
		}, wait);
	});

	// ---- Hand: pick a card, a pair or a trio ----
	let selected = $state<number[]>([]);
	let target = $state('');
	let ask = $state('');
	$effect(() => {
		// A new hand (cards played, stolen or drawn) drops picks that are gone.
		const ids = new Set(match.me.hand.map((c) => c.id));
		if (selected.some((id) => !ids.has(id))) selected = selected.filter((id) => ids.has(id));
	});
	const picked = $derived(match.me.hand.filter((c) => selected.includes(c.id)));
	const pickedKinds = $derived(picked.map((c) => c.kind));

	type Plan = { ok: boolean; label: string; needsTarget: boolean; needsAsk: boolean; why?: string };
	const plan = $derived.by((): Plan => {
		const k = pickedKinds;
		if (k.length === 0) return { ok: false, label: '', needsTarget: false, needsAsk: false };
		if (k.length === 1) {
			const one = k[0];
			if (one === 'cuddle') return { ok: true, label: 'Pedir cafuné', needsTarget: true, needsAsk: false };
			if (['current', 'nap', 'peek', 'swirl'].includes(one)) return { ok: true, label: `Jogar ${cardName(one)}`, needsTarget: false, needsAsk: false };
			if (one === 'nope') return { ok: false, label: '', needsTarget: false, needsAsk: false, why: 'O Nem Vem é usado quando alguém joga uma carta.' };
			if (one === 'dive') return { ok: false, label: '', needsTarget: false, needsAsk: false, why: 'O Mergulho é usado sozinho quando a onça aparece.' };
			return { ok: false, label: '', needsTarget: false, needsAsk: false, why: 'Junte um par ou uma trinca de capivaras iguais.' };
		}
		const same = k.every((x) => x === k[0]) && capys.includes(k[0]);
		if (!same || k.length > 3) return { ok: false, label: '', needsTarget: false, needsAsk: false, why: 'Pares e trincas são de capivaras iguais.' };
		return k.length === 2
			? { ok: true, label: 'Roubar uma carta', needsTarget: true, needsAsk: false }
			: { ok: true, label: 'Pedir uma carta', needsTarget: true, needsAsk: true };
	});
	const targets = $derived(match.players.filter((p) => p.id !== me && !p.out && p.cards > 0));
	const ready = $derived(plan.ok && (!plan.needsTarget || targets.some((t) => t.id === target)) && (!plan.needsAsk || !!ask));

	function toggle(c: Card) {
		if (match.phase === 'favor' && match.favor?.from === me) {
			giveCard(c);
			return;
		}
		sound('tap');
		selected = selected.includes(c.id) ? selected.filter((x) => x !== c.id) : [...selected, c.id];
		if (selected.length === 0) {
			target = '';
			ask = '';
		}
	}

	async function playSelected() {
		const ok = await act({ type: 'play', cards: selected, target: plan.needsTarget ? target : undefined, ask: plan.needsAsk ? ask : undefined });
		if (ok) {
			sound('push');
			haptic('select');
			selected = [];
			target = '';
			ask = '';
		}
	}

	async function draw() {
		sound('reveal');
		await act({ type: 'draw' });
	}

	async function nope() {
		haptic('warning');
		sound('rollback');
		await act({ type: 'nope' });
	}

	// ---- Hiding the jaguar back ----
	let hideAt = $state(0);
	const hideLabel = (i: number) =>
		i === 0 ? 'No topo: a próxima carta' : i === match.deck ? 'No fundo do monte' : `${i + 1}ª carta de cima`;
	async function hide() {
		if (await act({ type: 'hide', index: hideAt })) {
			sound('select');
			hideAt = 0;
		}
	}

	async function giveCard(c: Card) {
		if (await act({ type: 'give', cards: [c.id] })) sound('select');
	}

	// ---- What just happened ----
	function story(e: GameEvent): string {
		const p = Name(e.player);
		const t = name(e.target);
		switch (e.type) {
			case 'start':
				return `${p} começa.`;
			case 'play': {
				const k = e.kinds ?? [];
				if (k.length === 2) return `${p} jogou um par de ${cardName(k[0])} em ${t}.`;
				if (k.length === 3) return `${p} jogou uma trinca e pediu ${cardName(e.card?.kind)} para ${t}.`;
				if (k[0] === 'cuddle') return `${p} pediu um cafuné para ${t}.`;
				return `${p} jogou ${cardName(k[0])}.`;
			}
			case 'nope':
				return e.ok ? `${p}: Nem Vem!` : `${p}: Nem Vem no Nem Vem!`;
			case 'canceled':
				return e.player === me ? 'Sua carta foi cancelada.' : `A carta de ${name(e.player)} foi cancelada.`;
			case 'nap':
				return `${p} tirou uma soneca e não comprou.`;
			case 'current':
				return `A correnteza levou a vez para ${t}: 2 jogadas seguidas.`;
			case 'peek':
				return `${p} espiou o monte de binóculo.`;
			case 'swirl':
				return `${p} embaralhou o monte.`;
			case 'cuddle':
				return `${Name(e.target)} escolhe uma carta para dar a ${name(e.player)}.`;
			case 'favor':
				return `${Name(e.player)} não tinha cartas para dar.`;
			case 'steal':
				return e.ok ? `${p} roubou ${e.card ? cardName(e.card.kind) : 'uma carta'} de ${t}.` : `${Name(e.target)} não tinha cartas.`;
			case 'ask':
				return e.ok ? `${Name(e.target)} entregou ${cardName(e.card?.kind)} para ${name(e.player)}.` : `${Name(e.target)} não tinha ${cardName(e.card?.kind)}.`;
			case 'give':
				return `${p} deu ${e.card ? cardName(e.card.kind) : 'uma carta'} para ${t}.`;
			case 'draw':
				return e.player === me && e.card ? `Você comprou ${cardName(e.card.kind)}.` : `${p} comprou uma carta.`;
			case 'dive':
				return `${p} achou uma onça e mergulhou!`;
			case 'hide':
				return `${p} escondeu a onça de volta no monte.`;
			case 'eaten':
				return `A onça pegou ${name(e.player)}!`;
			case 'win':
				return `${p} venceu!`;
		}
		return '';
	}

	const recent = $derived(match.events.slice(-4).reverse());

	// Big moments get a full-screen card: a jaguar eating someone, a dive, your turn.
	type Moment = { key: number; kind: 'eaten' | 'dive' | 'turn' | 'drew' | 'win'; text: string; sub?: string; card?: string };
	let moment = $state<Moment | null>(null);
	let momentTimer: ReturnType<typeof setTimeout> | undefined;
	function show(m: Moment, hold = 1800) {
		moment = m;
		clearTimeout(momentTimer);
		momentTimer = setTimeout(() => (moment = null), ms(hold) || 1200);
	}
	onDestroy(() => clearTimeout(momentTimer));

	let lastSeq = -1;
	let lastTurnKey = '';
	$effect(() => {
		const evs = match.events;
		const top = evs.at(-1)?.seq ?? 0;
		if (lastSeq < 0) {
			lastSeq = top; // first render: nothing just happened
			lastTurnKey = `${match.turn}:${top}`;
			return;
		}
		for (const e of evs.filter((x) => x.seq > lastSeq)) {
			switch (e.type) {
				case 'eaten':
					sound('lose');
					haptic('warning');
					show({ key: e.seq, kind: 'eaten', text: e.player === me ? 'A onça te pegou!' : `A onça pegou ${name(e.player)}!`, sub: e.player === me ? 'Agora é torcer pelos outros.' : undefined }, 2400);
					break;
				case 'dive':
					sound('push');
					haptic('success');
					show({ key: e.seq, kind: 'dive', text: e.player === me ? 'Onça! Você mergulhou a tempo.' : `${Name(e.player)} mergulhou!`, sub: e.player === me ? 'Agora esconda a onça no monte.' : 'Escapou de uma onça.' }, 2000);
					break;
				case 'nope':
					sound('vote');
					haptic('tap');
					break;
				case 'steal':
				case 'ask':
				case 'give':
					sound('select');
					break;
				case 'draw':
					if (e.player === me && e.card) show({ key: e.seq, kind: 'drew', text: 'Você comprou', card: e.card.kind }, 1300);
					else sound('tap');
					break;
				case 'win':
					sound(e.player === me ? 'win' : 'lose');
					haptic(e.player === me ? 'success' : 'warning');
					break;
				case 'play':
					sound('push');
					break;
			}
		}
		lastSeq = top;
		const key = `${match.turn}:${match.turns_left}`;
		if (key !== lastTurnKey && match.turn === me && match.phase === 'play' && !match.me.out && !moment) {
			sound('turn');
			haptic('select');
			show({ key: top + 0.5, kind: 'turn', text: 'Sua vez', sub: match.turns_left > 1 ? `Você joga ${match.turns_left} vezes.` : 'Jogue o que quiser e compre uma carta.' }, 1400);
		}
		lastTurnKey = key;
	});

	// Binóculo: show the top cards when my peek arrives.
	let peekSeen = $state(-1);
	let showPeek = $state(false);
	$effect(() => {
		const seq = match.me.peek_seq;
		if (peekSeen < 0) {
			peekSeen = seq;
			return;
		}
		if (seq !== peekSeen && match.me.peek.length) {
			showPeek = true;
			sound('reveal');
		}
		peekSeen = seq;
	});

	const pendingWho = $derived(match.pending ? Name(match.pending.player) : '');
	const pendingText = $derived.by(() => {
		const pd = match.pending;
		if (!pd) return '';
		if (pd.effect === 'pair') return `quer roubar uma carta de ${name(pd.target)}`;
		if (pd.effect === 'trio') return `quer ${cardName(pd.ask)} de ${name(pd.target)}`;
		if (pd.effect === 'cuddle') return `pediu um cafuné para ${name(pd.target)}`;
		return `jogou ${cardName(pd.kinds[0])}`;
	});

	const ranking = $derived(match.winner ? [match.winner, ...[...match.out].reverse()] : [...match.out].reverse());
</script>

<div class="cc">
	<header class="bar">
		<a href="/" aria-label="Voltar ao início"><Icon name="back" /></a>
		<Title size={22} />
		<span class="bar-actions">
			<button class="pill" onclick={toggleSound} aria-label={feel.sound ? 'Desligar sons' : 'Ligar sons'} aria-pressed={feel.sound}>
				<Icon name={feel.sound ? 'sound' : 'mute'} size={18} />
			</button>
			<button class="pill chat" onclick={onchat} aria-label={unread ? `Chat, ${unread} novas` : 'Chat'}>
				<Icon name="chat" size={18} />{#if unread}<span class="badge" aria-hidden="true">{unread}</span>{/if}
			</button>
		</span>
	</header>

	<ul class="seats" aria-label="Capivaras na mesa">
		{#each match.players as p (p.id)}
			<li class:turn={p.id === match.turn && match.phase !== 'over'} class:out={p.out}>
				<button onclick={() => onplayer?.(p.id)} aria-label="{name(p.id)}: {p.out ? 'fora' : `${p.cards} cartas`}">
					<span class="face">
						<Avatar user={people[p.id]} size={44} radius={22} online={people[p.id]?.online} />
						{#if p.out}<span class="eaten" aria-hidden="true"><CardArt kind="jaguar" size={30} /></span>{/if}
					</span>
					<span class="who">{p.id === me ? 'Você' : (people[p.id]?.name ?? '?').split(' ')[0]}</span>
					<span class="count">{p.out ? 'fora' : `${p.cards} cartas`}{#if p.turns > 1} · {p.turns}x{/if}</span>
				</button>
			</li>
		{/each}
	</ul>

	{#if match.phase === 'over'}
		<section class="over" in:fly={{ y: 20, duration: ms(400) }}>
			<span class="crown" in:scale={{ start: 0.3, duration: ms(500), easing: backOut }}>👑</span>
			<h1>{match.winner === me ? 'Você é a última capivara de pé!' : `${Name(match.winner)} venceu!`}</h1>
			<ol class="ranking">
				{#each ranking as id, i (id)}
					<li in:fly={{ x: -16, duration: ms(260), delay: ms(200 + i * 90) }}>
						<span class="place">{i + 1}º</span>
						<Avatar user={people[id]} size={32} radius={16} />
						<span class="rname">{Name(id)}</span>
						{#if i > 0}<span class="muted small">virou janta</span>{/if}
					</li>
				{/each}
			</ol>
			{#if isHost}
				<button class="btn primary block" onclick={stop}>Jogar de novo</button>
			{:else}
				<p class="muted center">Esperando o host começar outra partida.</p>
			{/if}
			<button class="btn ghost block" onclick={onleave}><Icon name="door" size={18} />Sair da sala</button>
		</section>
	{:else}
		<section class="table" aria-label="Mesa">
			<button class="deck" class:ready={myTurn} disabled={!myTurn || busy} onclick={draw} aria-label="Comprar carta do monte ({match.deck} cartas)">
				<span class="stack"><Back width={78} /></span>
				<span class="deck-count">{match.deck}</span>
			</button>
			<div class="middle">
				<span class="jaguars" title="Onças no monte"><CardArt kind="jaguar" size={22} /> {match.jaguars} {match.jaguars === 1 ? 'onça' : 'onças'} no monte</span>
				<span class="odds muted small">{match.deck ? Math.round((match.jaguars / match.deck) * 100) : 0}% de chance na próxima</span>
			</div>
			<div class="discard" aria-label="Descarte">
				{#if match.discard.length}
					{#key match.discard_len}
						<span in:fly={{ y: -30, duration: ms(260) }}><CardFace kind={match.discard[match.discard.length - 1].kind} width={70} /></span>
					{/key}
				{:else}
					<span class="empty">descarte</span>
				{/if}
			</div>
		</section>

		{#if match.pending}
			{@const pd = match.pending}
			<section class="pending" class:noped={pd.nopes % 2 === 1} in:scale={{ start: 0.85, duration: ms(220) }}>
				<div class="pcards">
					{#each pd.kinds as k, i (i)}
						<span style="transform: rotate({(i - (pd.kinds.length - 1) / 2) * 8}deg)"><CardFace kind={k} width={64} /></span>
					{/each}
				</div>
				<div class="ptext">
					<strong>{pendingWho} {pendingText}</strong>
					{#if pd.nopes > 0}<span class="nopes">{pd.nopes % 2 === 1 ? 'Cancelada por Nem Vem' : 'Valendo de novo'} ({pd.nopes} Nem Vem)</span>{/if}
					<span class="timer"><span class="fill" style="width: {(remaining / windowMs) * 100}%"></span></span>
				</div>
				{#if match.me.can_nope && remaining > 0}
					<button class="nemvem" onclick={nope} disabled={busy}>NEM VEM!</button>
				{/if}
			</section>
		{/if}

		<ol class="story" aria-live="polite">
			{#each recent as e, i (e.seq)}
				<li class:latest={i === 0} in:fly={{ y: -10, duration: ms(220) }}>{story(e)}</li>
			{/each}
		</ol>

		{#if match.me.out}
			<p class="center muted spectator">A onça te pegou. Você segue vendo a mesa até o fim.</p>
		{:else}
			<section class="hand-area" aria-label="Sua mão">
				<div class="hand-head">
					<strong>Sua mão</strong>
					<span class="muted small">
						{#if match.phase === 'favor' && match.favor?.from === me}Toque numa carta para dar a {name(match.favor.to)}
						{:else if myTurn}Toque para escolher, ou compre do monte
						{:else if match.turn !== me}Vez de {name(match.turn)}
						{/if}
					</span>
				</div>
				<ul class="hand" class:giving={match.phase === 'favor' && match.favor?.from === me}>
					{#each match.me.hand as c (c.id)}
						<li class:picked={selected.includes(c.id)} in:fly={{ y: 40, duration: ms(260) }}>
							<button onclick={() => toggle(c)} disabled={!myTurn && !(match.phase === 'favor' && match.favor?.from === me)} aria-pressed={selected.includes(c.id)} aria-label={cardInfo[c.kind]?.name}>
								<CardFace kind={c.kind} width={84} />
							</button>
						</li>
					{/each}
				</ul>
			</section>

			<div class="dock">
				{#if myTurn && selected.length}
					{#if plan.why}
						<p class="why">{plan.why}</p>
					{:else}
						{#if plan.needsTarget}
							<div class="chips" role="radiogroup" aria-label="De quem">
								{#each targets as t (t.id)}
									<button role="radio" aria-checked={target === t.id} class:on={target === t.id} onclick={() => (target = t.id)}>
										<Avatar user={people[t.id]} size={22} radius={11} />{name(t.id)}
									</button>
								{/each}
							</div>
						{/if}
						{#if plan.needsAsk}
							<div class="chips" role="radiogroup" aria-label="Qual carta pedir">
								{#each match.kinds as k (k)}
									<button role="radio" aria-checked={ask === k} class:on={ask === k} onclick={() => (ask = k)}>{cardName(k)}</button>
								{/each}
							</div>
						{/if}
						<button class="btn primary block" disabled={!ready || busy} onclick={playSelected}>{plan.label}</button>
					{/if}
				{:else if myTurn}
					<button class="btn draw block" disabled={busy} onclick={draw}>
						Comprar carta{#if match.turns_left > 1} <span class="small">(vez 1 de {match.turns_left})</span>{/if}
					</button>
				{/if}
			</div>
		{/if}

		{#if error}<p class="error" role="alert">{error}</p>{/if}
		{#if isHost}
			<button class="link" onclick={() => (confirmStop = true)}>Encerrar partida</button>
		{/if}
	{/if}
</div>

{#if match.phase === 'defuse' && match.turn === me}
	<div class="sheet" role="dialog" aria-modal="true" aria-labelledby="hide-title" transition:fade={{ duration: ms(150) }}>
		<div class="sheet-card" in:fly={{ y: 60, duration: ms(260) }}>
			<CardArt kind="jaguar" size={72} />
			<h2 id="hide-title">Esconda a onça</h2>
			<p class="muted">Ninguém vê onde. Bem no topo, a próxima capivara compra ela.</p>
			<label class="slider">
				<span>{hideLabel(hideAt)}</span>
				<input type="range" min="0" max={match.deck} bind:value={hideAt} />
			</label>
			<div class="quick">
				<button onclick={() => (hideAt = 0)}>Topo</button>
				<button onclick={() => (hideAt = Math.floor(match.deck / 2))}>Meio</button>
				<button onclick={() => (hideAt = match.deck)}>Fundo</button>
				<button onclick={() => (hideAt = Math.floor(Math.random() * (match.deck + 1)))}>Sorteio</button>
			</div>
			<button class="btn primary block" disabled={busy} onclick={hide}>Esconder aqui</button>
		</div>
	</div>
{/if}

{#if showPeek && match.me.peek.length}
	<div class="sheet" role="dialog" aria-modal="true" aria-labelledby="peek-title" transition:fade={{ duration: ms(150) }}>
		<div class="sheet-card">
			<h2 id="peek-title">Topo do monte</h2>
			<p class="muted">Só você está vendo. A primeira é a próxima a sair.</p>
			<div class="peek">
				{#each match.me.peek as c, i (c.id)}
					<span class="pk" in:fly={{ y: 30, duration: ms(260), delay: ms(i * 140) }}>
						<span class="pos">{i + 1}ª</span><CardFace kind={c.kind} width={88} />
					</span>
				{/each}
			</div>
			<button class="btn primary block" onclick={() => (showPeek = false)}>Entendi</button>
		</div>
	</div>
{/if}

{#if moment}
	<!-- The backdrop stays while moments follow each other; only the content swaps. -->
	<button class="moment {moment.kind}" onclick={() => (moment = null)} in:fade={{ duration: ms(120) }} out:fade={{ duration: ms(200) }} aria-live="assertive">
		{#key moment.key}
			{#if moment.kind === 'eaten'}
				<span class="big-art" in:scale={{ start: 2.4, duration: ms(420), easing: backOut }}><CardArt kind="jaguar" size={150} /></span>
			{:else if moment.kind === 'dive'}
				<span class="big-art" in:fly={{ y: 60, duration: ms(420) }}><CardArt kind="dive" size={140} /></span>
			{:else if moment.kind === 'drew' && moment.card}
				<span in:fly={{ x: -120, y: -60, duration: ms(380), easing: backOut }}><CardFace kind={moment.card} width={150} text /></span>
			{/if}
			<strong in:scale={{ start: 0.7, duration: ms(320), delay: ms(120), easing: backOut }}>{moment.text}</strong>
			{#if moment.sub}<span class="sub">{moment.sub}</span>{/if}
		{/key}
	</button>
{/if}

{#if confirmStop}
	<div class="sheet" role="dialog" aria-modal="true" aria-labelledby="stop-title">
		<div class="sheet-card">
			<h2 id="stop-title">Encerrar a partida?</h2>
			<p class="muted">Todo mundo volta para o lobby e a partida não conta.</p>
			<button class="btn danger block" onclick={stop}>Encerrar</button>
			<button class="btn ghost block" onclick={() => (confirmStop = false)}>Continuar jogando</button>
		</div>
	</div>
{/if}

<style>
	/* The game's colors, also for the sheets and big moments, which sit outside .cc. */
	.cc,
	.sheet,
	.moment {
		--sand: #f6e7c8;
		--lagoon: #0f2a1f;
		--leaf: #1d6b4f;
		--orange: #ff9f1c;
		--muted-c: #a8c3b5;
	}
	.cc {
		background: radial-gradient(120% 60% at 50% 0%, #17402f, var(--lagoon) 60%);
		color: var(--sand);
		min-height: 100dvh;
		max-width: 560px;
		margin: 0 auto;
		padding: calc(10px + var(--safe-top)) 16px calc(24px + var(--safe-bottom));
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	:global(body:has(.cc)) {
		background: #0f2a1f;
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
		color: var(--sand);
	}
	.bar-actions {
		display: flex;
		gap: 6px;
	}
	.pill {
		position: relative;
		width: 40px;
		height: 40px;
		border-radius: 50%;
		border: 1px solid rgba(246, 231, 200, 0.25);
		background: rgba(0, 0, 0, 0.2);
		color: var(--sand);
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
	}
	.badge {
		position: absolute;
		top: -4px;
		right: -4px;
		min-width: 18px;
		height: 18px;
		border-radius: 999px;
		background: var(--orange);
		color: #2b1a0e;
		font-size: 11px;
		font-weight: 800;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.seats {
		list-style: none;
		margin: 0 -16px;
		padding: 2px 16px 4px;
		display: flex;
		gap: 8px;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.seats li button {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2px;
		width: 62px;
		padding: 6px 2px;
		border-radius: 12px;
		border: 2px solid transparent;
		background: rgba(0, 0, 0, 0.18);
		color: var(--sand);
		cursor: pointer;
		font: inherit;
	}
	.seats li.turn button {
		border-color: var(--orange);
		background: rgba(255, 159, 28, 0.12);
		animation: bob 1.6s ease-in-out infinite;
	}
	.seats li.out button {
		opacity: 0.45;
	}
	.face {
		position: relative;
		display: flex;
	}
	.eaten {
		position: absolute;
		right: -10px;
		bottom: -8px;
	}
	.who {
		font-size: 12px;
		font-weight: 700;
		max-width: 58px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.count {
		font-size: 11px;
		color: var(--muted-c);
	}
	.table {
		display: grid;
		grid-template-columns: auto 1fr auto;
		align-items: center;
		gap: 10px;
		padding: 14px;
		border-radius: 18px;
		background: radial-gradient(circle at 50% 40%, #2a8fb8 0%, #1f6f93 60%, #175a78 100%);
		box-shadow: inset 0 0 0 3px rgba(255, 255, 255, 0.08);
	}
	.deck {
		position: relative;
		border: none;
		background: none;
		padding: 0;
		cursor: pointer;
	}
	.deck:disabled {
		cursor: default;
	}
	.stack {
		display: flex;
		filter: drop-shadow(3px 3px 0 #13513c) drop-shadow(3px 3px 0 #0d3b2b);
	}
	.deck.ready .stack {
		animation: wiggle 1.4s ease-in-out infinite;
	}
	.deck-count {
		position: absolute;
		right: -6px;
		top: -6px;
		min-width: 26px;
		height: 26px;
		padding: 0 6px;
		border-radius: 999px;
		background: var(--sand);
		color: #2b1a0e;
		font-weight: 800;
		font-size: 13px;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.middle {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
		text-align: center;
	}
	.jaguars {
		display: flex;
		align-items: center;
		gap: 4px;
		font-weight: 800;
		font-size: 14px;
		padding: 4px 10px 4px 4px;
		border-radius: 999px;
		background: rgba(0, 0, 0, 0.3);
	}
	.odds {
		color: #d7f1ff;
	}
	.discard {
		width: 70px;
		display: flex;
		justify-content: center;
	}
	.empty {
		width: 70px;
		height: 100px;
		border-radius: 8px;
		border: 2px dashed rgba(255, 255, 255, 0.35);
		color: rgba(255, 255, 255, 0.6);
		font-size: 11px;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.pending {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px;
		border-radius: 16px;
		background: #fbf1dc;
		color: #2b1a0e;
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.35);
	}
	.pending.noped {
		background: #ffd9d6;
	}
	.pcards {
		display: flex;
	}
	.pcards span + span {
		margin-left: -40px;
	}
	.ptext {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 14px;
	}
	.nopes {
		font-size: 12px;
		font-weight: 700;
		color: #b3261e;
	}
	.timer {
		height: 6px;
		border-radius: 3px;
		background: rgba(0, 0, 0, 0.12);
		overflow: hidden;
	}
	.fill {
		display: block;
		height: 100%;
		background: var(--orange);
		transition: width 0.1s linear;
	}
	.nemvem {
		flex-shrink: 0;
		width: 76px;
		height: 76px;
		border-radius: 50%;
		border: 4px solid #fff;
		background: #e5383b;
		color: #fff;
		font: inherit;
		font-weight: 900;
		font-size: 14px;
		line-height: 1;
		cursor: pointer;
		box-shadow: 0 0 0 0 rgba(229, 56, 59, 0.7);
		animation: pulse 0.9s ease-out infinite;
	}
	.story {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 3px;
		min-height: 64px;
	}
	.story li {
		font-size: 13px;
		color: var(--muted-c);
	}
	.story li.latest {
		font-size: 15px;
		font-weight: 700;
		color: var(--sand);
	}
	.hand-area {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.hand-head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 8px;
	}
	.hand {
		list-style: none;
		margin: 0 -16px;
		padding: 22px 16px 6px;
		display: flex;
		gap: 6px;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.hand li {
		flex-shrink: 0;
		transition: transform 0.18s cubic-bezier(0.34, 1.56, 0.64, 1);
	}
	.hand li.picked {
		transform: translateY(-16px);
	}
	.hand li.picked :global(.cf) {
		box-shadow:
			0 0 0 3px var(--orange),
			0 10px 20px rgba(0, 0, 0, 0.45);
	}
	.hand button {
		border: none;
		background: none;
		padding: 0;
		cursor: pointer;
	}
	.hand button:disabled {
		cursor: default;
		opacity: 0.9;
	}
	.hand.giving li {
		animation: bob 1.2s ease-in-out infinite;
	}
	.dock {
		position: sticky;
		bottom: calc(8px + var(--safe-bottom));
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.why {
		margin: 0;
		padding: 10px 12px;
		border-radius: 12px;
		background: rgba(0, 0, 0, 0.35);
		font-size: 14px;
		text-align: center;
	}
	.chips {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.chips button {
		display: flex;
		align-items: center;
		gap: 6px;
		height: 36px;
		padding: 0 12px 0 6px;
		border-radius: 999px;
		border: 1px solid rgba(246, 231, 200, 0.35);
		background: rgba(0, 0, 0, 0.3);
		color: var(--sand);
		font: inherit;
		font-size: 13px;
		cursor: pointer;
	}
	.chips button.on {
		background: var(--orange);
		border-color: var(--orange);
		color: #2b1a0e;
		font-weight: 700;
	}
	.btn {
		height: 54px;
		border-radius: 999px;
		border: none;
		font: inherit;
		font-weight: 800;
		font-size: 16px;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		cursor: pointer;
	}
	.btn:disabled {
		opacity: 0.5;
		cursor: default;
	}
	.block {
		width: 100%;
	}
	.primary {
		background: var(--orange);
		color: #2b1a0e;
	}
	.draw {
		background: var(--sand);
		color: #2b1a0e;
		box-shadow: 0 6px 18px rgba(0, 0, 0, 0.35);
	}
	.ghost {
		background: transparent;
		color: var(--sand);
		border: 1px solid rgba(246, 231, 200, 0.35);
	}
	.danger {
		background: #da3633;
		color: #fff;
	}
	.link {
		align-self: center;
		background: none;
		border: none;
		color: var(--muted-c);
		text-decoration: underline;
		min-height: 44px;
		cursor: pointer;
		font: inherit;
	}
	.muted {
		color: var(--muted-c);
	}
	.small {
		font-size: 13px;
	}
	.center {
		text-align: center;
	}
	.spectator {
		padding: 16px;
		border-radius: 14px;
		background: rgba(0, 0, 0, 0.25);
	}
	.error {
		margin: 0;
		color: #ffb4ab;
		text-align: center;
	}
	.over {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		text-align: center;
		margin-top: 12px;
	}
	.crown {
		font-size: 56px;
	}
	.over h1 {
		margin: 0;
		font-size: 26px;
	}
	.ranking {
		list-style: none;
		margin: 0;
		padding: 0;
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.ranking li {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 12px;
		border-radius: 12px;
		background: rgba(0, 0, 0, 0.25);
		text-align: left;
	}
	.ranking li:first-child {
		background: rgba(255, 159, 28, 0.2);
		border: 1px solid var(--orange);
	}
	.place {
		width: 28px;
		font-weight: 800;
	}
	.rname {
		flex: 1;
		font-weight: 700;
	}
	.sheet {
		position: fixed;
		inset: 0;
		z-index: 60;
		background: rgba(0, 0, 0, 0.65);
		display: flex;
		align-items: flex-end;
		justify-content: center;
	}
	.sheet-card {
		width: 100%;
		max-width: 560px;
		background: #143528;
		color: var(--sand);
		border-radius: 20px 20px 0 0;
		padding: 22px 20px calc(22px + var(--safe-bottom));
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		text-align: center;
	}
	.sheet-card h2 {
		margin: 0;
		font-size: 22px;
	}
	.sheet-card p {
		margin: 0;
	}
	.slider {
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: 8px;
		font-weight: 800;
	}
	.slider input {
		width: 100%;
		accent-color: var(--orange);
	}
	.quick {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
		justify-content: center;
	}
	.quick button {
		height: 36px;
		padding: 0 14px;
		border-radius: 999px;
		border: 1px solid rgba(246, 231, 200, 0.35);
		background: rgba(0, 0, 0, 0.25);
		color: var(--sand);
		font: inherit;
		font-size: 13px;
		cursor: pointer;
	}
	.peek {
		display: flex;
		gap: 8px;
		justify-content: center;
	}
	.pk {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
	}
	.pos {
		font-weight: 800;
		font-size: 13px;
	}
	.moment {
		position: fixed;
		inset: 0;
		z-index: 70;
		border: none;
		margin: 0;
		padding: 24px;
		background: rgba(8, 26, 19, 0.9);
		backdrop-filter: blur(4px);
		-webkit-backdrop-filter: blur(4px);
		color: var(--sand);
		font: inherit;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 14px;
		text-align: center;
		cursor: pointer;
	}
	.moment.eaten {
		background: rgba(60, 16, 6, 0.92);
	}
	.moment.dive {
		background: rgba(8, 48, 70, 0.92);
	}
	.moment.drew strong {
		order: -1;
		font-size: 22px !important;
	}
	.moment strong {
		font-size: clamp(30px, 9vw, 44px);
		line-height: 1.1;
		font-weight: 900;
	}
	.moment .sub {
		color: var(--muted-c);
		font-size: 16px;
	}
	.big-art {
		display: flex;
		border-radius: 50%;
		background: #fbf1dc;
		padding: 10px;
	}
	@keyframes bob {
		0%,
		100% {
			transform: translateY(0);
		}
		50% {
			transform: translateY(-3px);
		}
	}
	@keyframes wiggle {
		0%,
		80%,
		100% {
			transform: rotate(0);
		}
		85% {
			transform: rotate(-4deg);
		}
		90% {
			transform: rotate(4deg);
		}
		95% {
			transform: rotate(-2deg);
		}
	}
	@keyframes pulse {
		to {
			box-shadow: 0 0 0 18px rgba(229, 56, 59, 0);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.seats li.turn button,
		.deck.ready .stack,
		.hand.giving li,
		.nemvem {
			animation: none;
		}
	}
</style>
