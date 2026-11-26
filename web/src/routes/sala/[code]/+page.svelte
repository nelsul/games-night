<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { onDestroy, onMount } from 'svelte';
	import QRCode from 'qrcode';
	import { api, ApiError, type ChatMessage, type InvitableFriend, type Room } from '$lib/api';
	import RoomChat from '$lib/components/RoomChat.svelte';
	import PlayerCard from '$lib/components/PlayerCard.svelte';
	import { flip } from 'svelte/animate';
	import { scale } from 'svelte/transition';
	import { haptic, ms, sound } from '$lib/feel.svelte';
	import { connectRoom, type RoomSocket } from '$lib/room-socket';
	import { session } from '$lib/session.svelte';
	import Avatar from '$lib/components/Avatar.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import { games } from '$lib/games';

	const code = $derived((page.params.code ?? '').toUpperCase());

	let room = $state<Room | null>(null);
	let error = $state('');
	let connected = $state(true);
	let showQR = $state(false);
	let qr = $state('');
	let copied = $state(false);
	let showInvite = $state(false);
	let friends = $state<InvitableFriend[] | null>(null);
	let inviteError = $state('');
	let socket: RoomSocket | null = null;

	const me = $derived(session.me?.id ?? '');
	const isHost = $derived(room?.host_id === me);
	const missing = $derived(room ? Math.max(0, room.game.min_players - room.players.length) : 0);
	// The room's game: its screens come from lib/games/<id>/index.ts.
	const game = $derived(room ? games[room.game.id] : undefined);
	const match = $derived(room?.status === 'playing' || room?.status === 'finished' ? room.match : null);
	// A game-specific reason the host can't start yet (the server checks it too).
	const notReady = $derived(room && game?.whyNotStart ? game.whyNotStart(room) : null);
	const onlineCount = $derived(room ? room.players.filter((p) => p.online).length : 0);

	// A chime when someone (or a bot) arrives in the lobby.
	let seenPlayers = -1;
	$effect(() => {
		const n = room?.players.length ?? 0;
		if (seenPlayers >= 0 && n > seenPlayers && room?.status === 'lobby') sound('join');
		seenPlayers = n;
	});

	let botsBusy = $state(false);
	let botsError = $state('');

	async function addBots(count: number) {
		botsError = '';
		botsBusy = true;
		try {
			room = await api.post<Room>(`/rooms/${code}/bots`, { count });
		} catch (err) {
			botsError = err instanceof ApiError ? err.message : 'Algo deu errado.';
		} finally {
			botsBusy = false;
		}
	}

	async function removeBots(id?: string) {
		botsError = '';
		botsBusy = true;
		try {
			room = await api.del<Room>(`/rooms/${code}/bots${id ? `?id=${id}` : ''}`);
		} catch (err) {
			botsError = err instanceof ApiError ? err.message : 'Algo deu errado.';
		} finally {
			botsBusy = false;
		}
	}

	let starting = $state(false);
	let startError = $state('');

	async function start() {
		startError = '';
		starting = true;
		try {
			room = await api.post<Room>(`/rooms/${code}/start`);
		} catch (err) {
			startError = err instanceof ApiError ? err.message : 'Algo deu errado.';
		} finally {
			starting = false;
		}
	}
	const full = $derived(room ? room.players.length >= room.game.max_players : false);
	const link = $derived(`${location.origin}/sala/${code}`);

	// Chat: history on join and after each reconnect, live messages over the socket.
	let messages = $state<ChatMessage[]>([]);
	let chatOpen = $state(false);
	let unread = $state(0);
	let toast = $state<ChatMessage | null>(null);
	let toastTimer: ReturnType<typeof setTimeout> | undefined;

	function addMessages(list: ChatMessage[]) {
		const seen = new Set(messages.map((m) => m.id));
		const fresh = list.filter((m) => !seen.has(m.id));
		if (!fresh.length) return;
		messages = [...messages, ...fresh].sort((a, b) => a.id - b.id).slice(-200);
		return fresh;
	}

	async function loadChat() {
		addMessages(await api.get<ChatMessage[]>(`/rooms/${code}/messages`).catch(() => []));
	}

	function onChat(m: ChatMessage) {
		if (!addMessages([m]) || m.user.id === me) return;
		sound('chat');
		if (chatOpen) return;
		haptic('tap');
		unread++;
		toast = m;
		clearTimeout(toastTimer);
		toastTimer = setTimeout(() => (toast = null), 3500);
	}

	// The player card: tap someone to see them and ask to be friends.
	let cardFor = $state<string | null>(null);
	const cardOnline = $derived(cardFor ? room?.players.find((p) => p.id === cardFor)?.online : undefined);

	function openChat() {
		chatOpen = true;
		unread = 0;
		toast = null;
	}

	onMount(async () => {
		try {
			room = await api.post<Room>(`/rooms/${code}/join`);
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
			return;
		}
		socket = connectRoom(code, {
			onRoom: (r) => {
				room = r;
				// Someone else removed us (we left on another device): go home.
				if (!r.players.some((p) => p.id === me)) goto('/', { replaceState: true });
			},
			onChat,
			onStatus: (c) => {
				connected = c;
				if (c) loadChat();
			}
		});
	});

	onDestroy(() => socket?.close());

	async function openQR() {
		qr = await QRCode.toDataURL(link, { margin: 1, width: 560, color: { dark: '#141414', light: '#ffffff' } });
		showQR = true;
	}

	async function openInvite() {
		showInvite = true;
		inviteError = '';
		friends = await api.get<InvitableFriend[]>(`/rooms/${code}/friends`).catch(() => []);
	}

	async function inviteFriend(f: InvitableFriend) {
		inviteError = '';
		try {
			const r = await api.post<{ state: InvitableFriend['state'] }>(`/rooms/${code}/invites`, { user_id: f.id });
			f.state = r.state;
		} catch (err) {
			inviteError = err instanceof ApiError ? err.message : 'Algo deu errado.';
		}
	}

	async function shareLink() {
		const data = { title: 'Games Night', text: `Bora jogar ${room?.game.name ?? ''}? Sala ${code}`, url: link };
		if (navigator.share) {
			try {
				await navigator.share(data);
				return;
			} catch {
				/* cancelled: fall back to copying */
			}
		}
		try {
			await navigator.clipboard.writeText(link);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch {
			await openQR();
		}
	}

	async function leave() {
		socket?.close();
		socket = null;
		await api.post(`/rooms/${code}/leave`).catch(() => {});
		await goto('/', { replaceState: true });
	}
</script>

<svelte:head><title>Sala {code} · Games Night</title></svelte:head>

{#if room && match}
	{#if game}
		<game.Match {room} {match} {me} {isHost} onroom={(r) => (room = r)} onchat={openChat} onleave={leave} {unread} onplayer={(id) => (cardFor = id)} />
	{:else}
		<main class="page"><p class="muted center">Esse jogo não está disponível nesta versão do site. Atualize a página.</p></main>
	{/if}
{:else}
<main class="page lobby">
	<header>
		<a href="/" aria-label="Voltar ao início"><Icon name="back" /></a>
		{#if game}<game.Title size={26} />{:else}<span></span>{/if}
		{#if room}
			<span class="actions">
				<button class="icon-btn chat-btn" onclick={openChat} aria-label={unread ? `Chat, ${unread} novas` : 'Chat'}>
					<Icon name="chat" size={22} />{#if unread}<span class="badge" aria-hidden="true">{unread}</span>{/if}
				</button>
				<button class="icon-btn" onclick={leave} aria-label="Sair da sala"><Icon name="door" size={22} /></button>
			</span>
		{:else}
			<span class="icon-btn" aria-hidden="true"></span>
		{/if}
	</header>

	{#if error}
		<div class="problem">
			<p>{error}</p>
			<a class="btn btn-primary" href="/codigo">Tentar outro código</a>
		</div>
	{:else if !room}
		<p class="muted center">Entrando na sala…</p>
	{:else}
		<div class="top">
			<h1>Quem está na sala?</h1>
			<button class="code" onclick={openQR} aria-label="Mostrar o QR da sala {code}">
				<span>{code}</span>
				<Icon name="qr" size={20} />
			</button>
			<p class="online"><span class="pulse" aria-hidden="true"></span>{onlineCount} de {room.players.length} online agora</p>
			{#if !connected}<p class="reconnecting" role="status">Reconectando…</p>{/if}
		</div>

		<ul class="players">
			{#each room.players as p (p.id)}
				<li class:offline={!p.online} animate:flip={{ duration: ms(250) }} in:scale={{ start: 0.5, duration: ms(300) }}>
					<span class="tile">
						<button class="who" onclick={() => (cardFor = p.id)} aria-label="Ver {p.id === me ? 'seu perfil' : p.name}">
							<Avatar user={p} size={96} radius={6} online={p.online} />
						</button>
						{#if p.is_host}<span class="host">HOST</span>{/if}
						{#if p.is_bot}<span class="host bot">BOT</span>{/if}
						{#if p.is_bot && isHost && room.status === 'lobby'}
							<button class="remove" onclick={() => removeBots(p.id)} aria-label="Tirar {p.name} da sala"
								><Icon name="close" size={14} /></button
							>
						{/if}
					</span>
					<span class="name" class:me={p.id === me}>{p.id === me ? 'você' : p.name}</span>
					{#if !p.online}<span class="sr-only">(fora do app)</span>{/if}
				</li>
			{/each}
			{#if !full}
				<li>
					<button class="tile invite" onclick={openInvite} aria-label="Convidar">
						<Icon name="plus" size={40} />
					</button>
					<span class="name">Convidar</span>
				</li>
			{/if}
		</ul>

		{#if isHost && room.status === 'lobby'}
			<section class="bots" aria-labelledby="bots-title">
				<div class="bots-text">
					<h2 id="bots-title">Bots</h2>
					<p class="muted">Completam a mesa e jogam sozinhos, com uma estratégia simples.</p>
				</div>
				<div class="bots-actions">
					{#if missing > 0}
						<button class="btn btn-primary" onclick={() => addBots(missing)} disabled={botsBusy}>
							Completar com {missing} {missing === 1 ? 'bot' : 'bots'}
						</button>
					{/if}
					<button class="btn btn-ghost" onclick={() => addBots(1)} disabled={botsBusy || full}>+ 1 bot</button>
					{#if room.players.some((p) => p.is_bot)}
						<button class="btn btn-ghost" onclick={() => removeBots()} disabled={botsBusy}>Tirar bots</button>
					{/if}
				</div>
				{#if botsError}<p class="error" role="alert">{botsError}</p>{/if}
			</section>
		{/if}

		{#if game?.LobbyOptions}
			<game.LobbyOptions {room} {isHost} onchange={(r) => (room = r)} />
		{/if}

		<div class="bottom">
			<p class="muted center">
				{room.players.length} de {room.game.max_players} ·
				{#if missing > 0}
					{missing === 1 ? 'falta 1 pessoa' : `faltam ${missing} pessoas`} para começar
				{:else if notReady}
					{notReady}
				{:else if room.status === 'playing'}
					carregando a partida…
				{:else if isHost}
					tudo pronto
				{:else}
					esperando o host começar
				{/if}
			</p>
			{#if startError}<p class="error center" role="alert">{startError}</p>{/if}
			{#if isHost}
				<button
					class="btn btn-primary btn-block"
					disabled={missing > 0 || !!notReady || !room.game.available || starting}
					onclick={start}
				>
					<Icon name="play" size={18} />{starting ? 'Distribuindo crachás…' : 'Começar partida'}
				</button>
			{/if}
		</div>
	{/if}
</main>
{/if}

{#if showInvite}
	<div class="sheet" role="dialog" aria-modal="true" aria-labelledby="invite-title">
		<div class="sheet-card invite-card">
			<h2 id="invite-title">Convidar amigos</h2>
			{#if friends === null}
				<p class="muted">Carregando…</p>
			{:else if friends.length === 0}
				<p class="muted">Você ainda não tem amigos aqui. <a href="/amigos">Busque a galera pelo @</a> ou mande o link da sala.</p>
			{:else}
				<ul class="friends">
					{#each friends as f (f.id)}
						<li>
							<Avatar user={f} size={40} />
							<span class="who"><strong>{f.name}</strong><span class="muted">@{f.handle}</span></span>
							{#if f.state === 'in_room'}
								<span class="tag">Na sala</span>
							{:else if f.state === 'invited'}
								<span class="tag"><Icon name="check" size={16} />Convidado</span>
							{:else}
								<button class="btn btn-primary small" onclick={() => inviteFriend(f)}>Convidar</button>
							{/if}
						</li>
					{/each}
				</ul>
			{/if}
			{#if inviteError}<p class="error" role="alert">{inviteError}</p>{/if}
			<button class="btn btn-ghost btn-block" onclick={shareLink}>
				<Icon name="share" size={18} />{copied ? 'Link copiado' : 'Mandar o link'}
			</button>
			<button class="btn btn-primary btn-block" onclick={() => (showInvite = false)}>Pronto</button>
		</div>
	</div>
{/if}

{#if showQR}
	<div class="sheet" role="dialog" aria-modal="true" aria-label="QR da sala {code}">
		<div class="sheet-card">
			<img src={qr} alt="QR code que abre a sala {code}" width="280" height="280" />
			<p class="sheet-code">{code}</p>
			<p class="muted">Aponte a câmera para entrar na sala.</p>
			<button class="btn btn-primary btn-block" onclick={() => (showQR = false)}>Fechar</button>
		</div>
	</div>
{/if}

{#if toast && !chatOpen}
	<button class="toast" onclick={openChat}>
		<strong>{toast.user.name}:</strong> {toast.body}
	</button>
{/if}

{#if cardFor}
	<PlayerCard userId={cardFor} online={cardOnline} onclose={() => (cardFor = null)} />
{/if}

{#if chatOpen && room}
	<RoomChat {code} {me} {messages} onclose={() => (chatOpen = false)} />
{/if}

<style>
	.lobby {
		display: flex;
		flex-direction: column;
		gap: 20px;
		padding-bottom: calc(24px + var(--safe-bottom));
	}
	header {
		height: 44px;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	header a,
	.icon-btn {
		width: 44px;
		height: 44px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: none;
		border: none;
		cursor: pointer;
	}
	header a {
		margin-left: -10px;
	}
	.actions {
		display: flex;
		margin-right: -10px;
	}
	.chat-btn {
		position: relative;
	}
	.badge {
		position: absolute;
		top: 4px;
		right: 2px;
		min-width: 18px;
		height: 18px;
		padding: 0 5px;
		border-radius: 9px;
		background: var(--brand);
		color: #fff;
		font-size: 11px;
		font-weight: 700;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.toast {
		position: fixed;
		left: 50%;
		top: calc(10px + var(--safe-top));
		transform: translateX(-50%);
		z-index: 60;
		width: min(92vw, 520px);
		text-align: left;
		background: rgba(40, 40, 40, 0.97);
		border: 1px solid var(--line-strong);
		border-radius: 14px;
		padding: 10px 14px;
		color: #fff;
		font-size: 14px;
		line-height: 1.35;
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		cursor: pointer;
	}
	.center {
		text-align: center;
	}
	.problem {
		display: flex;
		flex-direction: column;
		gap: 16px;
		align-items: center;
		text-align: center;
		margin-top: 60px;
	}
	.top {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 10px;
	}
	h1 {
		margin: 0;
		font-size: 26px;
		font-weight: 500;
	}
	.code {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		height: 48px;
		padding: 0 16px;
		border-radius: var(--radius);
		border: 1px solid var(--line-strong);
		background: var(--surface);
		cursor: pointer;
	}
	.code span {
		font-family: var(--display);
		font-size: 30px;
		letter-spacing: 6px;
		padding-top: 4px;
	}
	.online {
		margin: 0;
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13px;
		color: var(--muted);
	}
	.pulse {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: #3fb950;
		animation: gn-glow 2s ease-in-out infinite;
	}
	.reconnecting {
		margin: 0;
		font-size: 13px;
		color: #e8a317;
	}
	.players {
		list-style: none;
		margin: 0;
		padding: 0 4px;
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 18px 16px;
	}
	.players li {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		min-width: 0;
	}
	.who {
		border: none;
		background: none;
		padding: 0;
		cursor: pointer;
		border-radius: 6px;
		display: flex;
	}
	.host.bot {
		background: #2f80ed;
		right: auto;
		left: -8px;
	}
	.remove {
		position: absolute;
		bottom: -8px;
		left: -8px; /* the online dot sits bottom-right */
		width: 30px;
		height: 30px;
		border-radius: 50%;
		border: 2px solid var(--bg);
		background: #fff;
		color: #141414;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
	}
	.bots {
		background: var(--surface);
		border-radius: var(--radius);
		padding: 12px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.bots h2 {
		margin: 0;
		font-size: 18px;
	}
	.bots p {
		margin: 2px 0 0;
		font-size: 13px;
		line-height: 1.4;
	}
	.bots-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
	.bots-actions .btn {
		height: 42px;
		font-size: 14px;
		padding: 0 14px;
	}
	.offline .tile {
		opacity: 0.45;
	}
	.tile {
		position: relative;
		width: 96px;
		height: 96px;
		display: flex;
	}
	.host {
		position: absolute;
		top: -8px;
		right: -8px;
		font-size: 10px;
		font-weight: 700;
		background: var(--brand);
		padding: 3px 6px;
		border-radius: 3px;
	}
	.invite {
		border: 2px solid var(--line-strong);
		border-radius: 6px;
		background: none;
		align-items: center;
		justify-content: center;
		color: var(--muted);
		cursor: pointer;
	}
	.name {
		font-size: 14px;
		color: var(--muted);
		max-width: 100%;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.name.me {
		color: #fff;
		font-weight: 600;
	}
	.bottom {
		margin-top: auto;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.bottom p {
		margin: 0;
		font-size: 13px;
	}
	.sheet {
		position: fixed;
		inset: 0;
		z-index: 50;
		background: rgba(0, 0, 0, 0.75);
		display: flex;
		align-items: flex-end;
		justify-content: center;
	}
	.sheet-card {
		width: 100%;
		max-width: 560px;
		background: var(--surface);
		border-radius: 16px 16px 0 0;
		padding: 24px 16px calc(20px + var(--safe-bottom));
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
	}
	.sheet-card img {
		border-radius: 8px;
		background: #fff;
	}
	.invite-card {
		align-items: stretch;
		max-height: 85dvh;
	}
	.invite-card h2 {
		margin: 0;
		font-size: 20px;
	}
	.invite-card p {
		margin: 0;
		line-height: 1.45;
	}
	.friends {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 6px;
		overflow-y: auto;
	}
	.friends li {
		display: flex;
		align-items: center;
		gap: 12px;
		min-height: 56px;
	}
	.who {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
		font-size: 13px;
	}
	.who strong {
		font-size: 15px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.tag {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		font-size: 13px;
		color: var(--muted);
	}
	.small {
		height: 40px;
		padding: 0 12px;
		font-size: 14px;
	}
	.sheet-code {
		margin: 0;
		font-family: var(--display);
		font-size: 40px;
		letter-spacing: 8px;
	}
	.sheet-card .muted {
		margin: 0;
	}
</style>
