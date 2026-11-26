<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { onDestroy, onMount } from 'svelte';
	import QRCode from 'qrcode';
	import { api, ApiError, type Room } from '$lib/api';
	import { connectRoom, type RoomSocket } from '$lib/room-socket';
	import { session } from '$lib/session.svelte';
	import Avatar from '$lib/components/Avatar.svelte';
	import BackdoorTitle from '$lib/components/BackdoorTitle.svelte';
	import Icon from '$lib/components/Icon.svelte';

	const code = $derived((page.params.code ?? '').toUpperCase());

	let room = $state<Room | null>(null);
	let error = $state('');
	let connected = $state(true);
	let showQR = $state(false);
	let qr = $state('');
	let copied = $state(false);
	let socket: RoomSocket | null = null;

	const me = $derived(session.me?.id ?? '');
	const isHost = $derived(room?.host_id === me);
	const missing = $derived(room ? Math.max(0, room.game.min_players - room.players.length) : 0);
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

	async function invite() {
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
					<button class="tile invite" onclick={invite} aria-label="Convidar">
						<Icon name="plus" size={40} />
					</button>
					<span class="name">{copied ? 'Link copiado' : 'Convidar'}</span>
				</li>
			{/if}
		</ul>

		<div class="bottom">
			<p class="muted center">
				{room.players.length} de {room.game.max_players} ·
				{#if missing > 0}
					{missing === 1 ? 'falta 1 pessoa' : `faltam ${missing} pessoas`} para começar
				{:else if !room.game.available}
					o {room.game.name} chega na próxima versão
				{:else if isHost}
					tudo pronto
				{:else}
					esperando o host começar
				{/if}
			</p>
			{#if isHost}
				<button class="btn btn-primary btn-block" disabled={missing > 0 || !room.game.available}>
					<Icon name="play" size={18} />Começar partida
				</button>
			{/if}
		</div>
	{/if}
</main>

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
