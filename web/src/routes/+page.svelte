<script lang="ts">
	import { onMount } from 'svelte';
	import { api, ApiError, type RoomSummary } from '$lib/api';
	import { createRoom, gameList } from '$lib/games';
	import { session } from '$lib/session.svelte';
	import Avatar from '$lib/components/Avatar.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import Logo from '$lib/components/Logo.svelte';
	import RoomCard from '$lib/components/RoomCard.svelte';
	import Contributors from '$lib/components/Contributors.svelte';
	import { inbox, refreshInbox } from '$lib/inbox.svelte';
	import { goto } from '$app/navigation';

	// The billboard shows one game; with more games it rotates daily.
	const featured = gameList[Math.floor(Date.now() / 86_400_000) % Math.max(gameList.length, 1)];

	let rooms = $state<RoomSummary[]>([]);
	let error = $state('');
	let creating = $state(false);

	onMount(async () => {
		rooms = await api.get<RoomSummary[]>('/rooms').catch(() => []);
	});

	async function decline(id: string) {
		await api.post(`/invites/${id}/decline`).catch(() => {});
		await refreshInbox();
	}

	async function create() {
		error = '';
		creating = true;
		try {
			await createRoom(featured.info.id);
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
			creating = false;
		}
	}
</script>

<svelte:head><title>Games Night</title></svelte:head>

<main class="page home">
	<header>
		<Logo />
		{#if session.me}
			<a href="/perfil" aria-label="Seu perfil"><Avatar user={session.me} size={32} /></a>
		{/if}
	</header>

	{#if inbox.invites.length}
		<section class="invites" aria-label="Convites">
			{#each inbox.invites as inv (inv.id)}
				<div class="invite">
					<Avatar user={inv.from} size={44} />
					<span class="invite-text">
						<strong>{inv.from.name} te chamou para {inv.game.name}</strong>
						<span class="muted">Sala {inv.code} · {inv.players} {inv.players === 1 ? 'pessoa' : 'pessoas'}</span>
					</span>
					<span class="invite-actions">
						<button class="btn btn-ghost small" onclick={() => decline(inv.id)} aria-label="Recusar convite de {inv.from.name}">
							<Icon name="close" size={18} />
						</button>
						<button class="btn btn-primary small" onclick={() => goto(`/sala/${inv.code}`)}>Entrar</button>
					</span>
				</div>
			{/each}
		</section>
	{/if}

	{#if featured}
	<section class="billboard" aria-label="Destaque">
		<a href="/jogos/{featured.info.id}" aria-label="Ver o {featured.info.name}" class="art"><featured.KeyArt /></a>
		<div class="info">
			<featured.Title />
			<span class="tags">{featured.info.tags} · {featured.info.players}</span>
			<div class="actions">
				<button class="btn btn-primary" onclick={create} disabled={creating}>
					<Icon name="play" size={18} />{creating ? 'Criando…' : 'Criar sala'}
				</button>
				<a class="btn btn-secondary" href="/codigo"><Icon name="plus" size={18} />Código</a>
			</div>
			{#if error}<p class="error" role="alert">{error}</p>{/if}
		</div>
	</section>
	{/if}

	{#if rooms.length}
		<section class="row">
			<h2 class="row-title">Continue jogando</h2>
			<div class="scroller">
				{#each rooms as room (room.code)}
					<RoomCard {room} />
				{/each}
			</div>
		</section>
	{/if}

	<section class="row">
		<h2 class="row-title">Jogos</h2>
		<div class="scroller">
			{#each gameList as g (g.info.id)}
				<a class="poster" href="/jogos/{g.info.id}">
					<span class="poster-art"><g.Title size={34} /></span>
					<span class="poster-caption">{g.info.tagline}</span>
				</a>
			{/each}
			<div class="poster soon" aria-label="Próximo jogo em breve">
				<span class="poster-art">EM BREVE</span>
				<span class="poster-caption">Próximo jogo</span>
			</div>
		</div>
	</section>

	<Contributors />
</main>

<style>
	.home {
		display: flex;
		flex-direction: column;
		gap: 22px;
	}
	header {
		height: 44px;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	header a {
		display: flex;
		min-width: 44px;
		min-height: 44px;
		align-items: center;
		justify-content: flex-end;
		text-decoration: none;
	}
	.invites {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.invite {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 12px;
		border-radius: var(--radius);
		background: var(--surface);
		border-left: 3px solid var(--brand);
	}
	.invite-text {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
		font-size: 13px;
	}
	.invite-text strong {
		font-size: 14px;
		line-height: 1.3;
	}
	.invite-actions {
		display: flex;
		gap: 6px;
	}
	.small {
		height: 40px;
		padding: 0 12px;
		font-size: 14px;
		min-width: 44px;
	}
	.billboard {
		border-radius: 12px;
		overflow: hidden;
		background: #1b0b0e;
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
	}
	.art {
		display: block;
	}
	.info {
		padding: 0 16px 16px;
		margin-top: -54px;
		position: relative;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 10px;
		text-align: center;
	}
	.tags {
		font-size: 13px;
		color: #d6d6d6;
	}
	.actions {
		display: flex;
		gap: 10px;
		width: 100%;
	}
	.actions .btn {
		flex: 1;
		height: 44px;
	}
	.row {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.scroller {
		display: flex;
		gap: 8px;
		overflow-x: auto;
		margin: 0 -16px;
		padding: 0 16px 4px;
		scroll-snap-type: x mandatory;
		scrollbar-width: none;
	}
	.scroller::-webkit-scrollbar {
		display: none;
	}
	.poster {
		width: 150px;
		flex-shrink: 0;
		border-radius: var(--radius);
		overflow: hidden;
		background: var(--surface-2);
		color: #fff;
		text-decoration: none;
		display: flex;
		flex-direction: column;
		scroll-snap-align: start;
	}
	.poster-art {
		height: 200px;
		background: #1b0b0e;
		display: flex;
		align-items: center;
		justify-content: center;
		font-family: var(--display);
		font-size: 26px;
		letter-spacing: 1px;
	}
	.poster-caption {
		padding: 8px;
		font-size: 12px;
		color: #d6d6d6;
	}
	.soon .poster-art {
		background: var(--surface);
		color: #8c8c8c;
		border: 2px dashed var(--line-strong);
		border-bottom: none;
		border-radius: var(--radius) var(--radius) 0 0;
	}
</style>
