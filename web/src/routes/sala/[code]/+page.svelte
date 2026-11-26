<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { onDestroy, onMount } from 'svelte';
	import QRCode from 'qrcode';
	import { api, ApiError, type InvitableFriend, type Room } from '$lib/api';
	import { connectRoom, type RoomSocket } from '$lib/room-socket';
	import { session } from '$lib/session.svelte';
	import Avatar from '$lib/components/Avatar.svelte';
	import BackdoorTitle from '$lib/components/BackdoorTitle.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import Backdoor from '$lib/games/backdoor/Backdoor.svelte';
	import LobbyOptions from '$lib/games/backdoor/LobbyOptions.svelte';
	import type { MatchView } from '$lib/games/backdoor/types';

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
	const match = $derived(room?.status === 'playing' ? (room.match as MatchView | null) : null);
	// Zero-day and Freelancer need 7+ people; the server checks it too.
	const specialsBlocked = $derived(
		!!room && (room.options?.zero_day === true || room.options?.freelancer === true) && room.players.length < 7
	);
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
			onStatus: (c) => (connected = c)
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
	<Backdoor {room} {match} {me} isHost={isHost} onroom={(r) => (room = r)} />
{:else}
<main class="page lobby">
	<header>
		<a href="/" aria-label="Voltar ao início"><Icon name="back" /></a>
		<BackdoorTitle size={26} />
		{#if room}
			<button class="icon-btn" onclick={leave} aria-label="Sair da sala"><Icon name="door" size={22} /></button>
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
			{#if !connected}<p class="reconnecting" role="status">Reconectando…</p>{/if}
		</div>

		<ul class="players">
			{#each room.players as p (p.id)}
				<li class:offline={!p.online}>
					<span class="tile">
						<Avatar user={p} size={96} radius={6} />
						{#if p.is_host}<span class="host">HOST</span>{/if}
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

		<LobbyOptions {room} {isHost} onchange={(r) => (room = r)} />

		<div class="bottom">
			<p class="muted center">
				{room.players.length} de {room.game.max_players} ·
				{#if missing > 0}
					{missing === 1 ? 'falta 1 pessoa' : `faltam ${missing} pessoas`} para começar
				{:else if specialsBlocked}
					Zero-day e Freelancer precisam de 7 pessoas
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
					disabled={missing > 0 || specialsBlocked || !room.game.available || starting}
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
	.icon-btn {
		margin-right: -10px;
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
