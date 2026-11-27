<script lang="ts">
	import { onMount, untrack } from 'svelte';
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
	import { reducedMotion } from '$lib/feel.svelte';

	// The billboard is a carousel of every game. It starts on a different game each day,
	// moves on by itself every few seconds, and stops moving once someone swipes it.
	const start = Math.floor(Date.now() / 86_400_000) % Math.max(gameList.length, 1);
	let current = $state(start);
	let track = $state<HTMLElement>();
	let touched = false;
	const AUTOPLAY_MS = 6000;

	function go(i: number, smooth = true) {
		if (!track) return;
		const next = (i + gameList.length) % gameList.length;
		current = next;
		track.scrollTo({ left: next * track.clientWidth, behavior: smooth && !reducedMotion ? 'smooth' : 'auto' });
	}

	function onScroll() {
		if (!track) return;
		current = Math.round(track.scrollLeft / track.clientWidth);
	}

	// A swipe (or a wheel) means someone is looking: stop the autoplay.
	$effect(() => {
		const el = track;
		if (!el) return;
		const stop = () => (touched = true);
		el.addEventListener('pointerdown', stop, { passive: true });
		el.addEventListener('wheel', stop, { passive: true });
		return () => {
			el.removeEventListener('pointerdown', stop);
			el.removeEventListener('wheel', stop);
		};
	});

	// Runs once the track exists; untrack keeps it from re-running (and jumping back to the
	// first slide) every time the current slide changes.
	$effect(() => {
		if (!track) return;
		untrack(() => go(start, false));
		if (gameList.length < 2) return;
		const t = setInterval(() => {
			if (!touched && !document.hidden) go(untrack(() => current) + 1);
		}, AUTOPLAY_MS);
		return () => clearInterval(t);
	});

	let rooms = $state<RoomSummary[]>([]);
	let error = $state('');
	let creating = $state('');

	onMount(async () => {
		rooms = await api.get<RoomSummary[]>('/rooms').catch(() => []);
	});

	async function decline(id: string) {
		await api.post(`/invites/${id}/decline`).catch(() => {});
		await refreshInbox();
	}

	async function create(id: string) {
		error = '';
		creating = id;
		try {
			await createRoom(id);
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
			creating = '';
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

	{#if gameList.length}
		<section class="billboard" aria-roledescription="carrossel" aria-label="Jogos em destaque">
			<div
				class="track"
				bind:this={track}
				onscroll={onScroll}
			>
				{#each gameList as g, i (g.info.id)}
					<div class="slide" role="group" aria-roledescription="slide" aria-label="{i + 1} de {gameList.length}: {g.info.name}" inert={i !== current}>
						<a href="/jogos/{g.info.id}" aria-label="Ver o {g.info.name}" class="art"><g.KeyArt /></a>
						<div class="info">
							<g.Title />
							<span class="tags">{g.info.tags} · {g.info.players}</span>
							<div class="actions">
								<button class="btn btn-primary" onclick={() => create(g.info.id)} disabled={!!creating}>
									<Icon name="play" size={18} />{creating === g.info.id ? 'Criando…' : 'Criar sala'}
								</button>
								<a class="btn btn-secondary" href="/codigo"><Icon name="plus" size={18} />Código</a>
							</div>
						</div>
					</div>
				{/each}
			</div>
			{#if gameList.length > 1}
				<div class="dots">
					{#each gameList as g, i (g.info.id)}
						<button
							class="dot"
							class:on={i === current}
							aria-label="Mostrar {g.info.name}"
							aria-current={i === current ? 'true' : undefined}
							onclick={() => {
								touched = true;
								go(i);
							}}
						></button>
					{/each}
				</div>
			{/if}
			{#if error}<p class="error" role="alert">{error}</p>{/if}
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
					<span class="poster-art"><g.KeyArt height={200} /><span class="poster-title"><g.Title size={30} /></span></span>
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

	<a class="talk" href="/feedback">
		<span class="talk-icons" aria-hidden="true"><Icon name="bug" size={20} /><Icon name="bulb" size={20} /></span>
		<span><strong>Achou um bug ou tem uma ideia?</strong><span class="muted">Fale com a gente, até anônimo.</span></span>
		<Icon name="next" size={18} />
	</a>
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
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.track {
		display: flex;
		overflow-x: auto;
		scroll-snap-type: x mandatory;
		scrollbar-width: none;
		border-radius: 12px;
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
		overscroll-behavior-x: contain;
	}
	.track::-webkit-scrollbar {
		display: none;
	}
	.slide {
		flex: 0 0 100%;
		scroll-snap-align: start;
		scroll-snap-stop: always;
		background: #1b0b0e;
	}
	.dots {
		display: flex;
		justify-content: center;
		gap: 2px;
	}
	.dot {
		width: 28px;
		height: 24px;
		padding: 0;
		border: none;
		background: none;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
	}
	.dot::before {
		content: '';
		width: 8px;
		height: 8px;
		border-radius: 999px;
		background: var(--line-strong);
		transition:
			width 0.25s ease,
			background-color 0.25s ease;
	}
	.dot.on::before {
		width: 20px;
		background: #fff;
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
		/* Snapping lines cards up with the padding, not with the screen's edge. */
		scroll-padding-inline: 16px;
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
		position: relative;
		height: 200px;
		overflow: hidden;
		background: #1b0b0e;
		display: flex;
		align-items: center;
		justify-content: center;
		font-family: var(--display);
		font-size: 26px;
		letter-spacing: 1px;
	}
	.poster-title {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		padding: 28px 8px 10px;
		display: flex;
		justify-content: center;
		background: linear-gradient(transparent, rgba(0, 0, 0, 0.75));
	}
	.poster-art > :global(svg) {
		position: absolute;
		inset: 0;
		height: 100%;
	}
	.poster-caption {
		padding: 8px;
		font-size: 12px;
		color: #d6d6d6;
	}
	.talk {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 14px;
		border-radius: 12px;
		background: var(--surface);
		color: #fff;
		text-decoration: none;
		margin-bottom: 8px;
	}
	.talk > span:nth-child(2) {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 2px;
		font-size: 13px;
	}
	.talk strong {
		font-size: 15px;
	}
	.talk-icons {
		display: flex;
		gap: 2px;
		color: #e3b341;
	}
	.soon .poster-art {
		background: var(--surface);
		color: #8c8c8c;
		border: 2px dashed var(--line-strong);
		border-bottom: none;
		border-radius: var(--radius) var(--radius) 0 0;
	}
</style>
