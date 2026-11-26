<script lang="ts">
	// Someone's card, opened by tapping them in a room: who they are and, if you
	// aren't friends yet, a button to ask.
	import { fly } from 'svelte/transition';
	import { api, ApiError, type UserWithRelation } from '$lib/api';
	import { refreshInbox } from '$lib/inbox.svelte';
	import { haptic, ms, sound } from '$lib/feel.svelte';
	import Avatar from './Avatar.svelte';
	import Icon from './Icon.svelte';

	let { userId, online, onclose }: { userId: string; online?: boolean; onclose: () => void } = $props();

	type Card = UserWithRelation & { is_bot: boolean };
	let user = $state<Card | null>(null);
	let error = $state('');
	let busy = $state(false);

	$effect(() => {
		const id = userId;
		user = null;
		error = '';
		api
			.get<Card>(`/users/${id}`)
			.then((u) => (user = u))
			.catch((err) => (error = err instanceof ApiError ? err.message : 'Algo deu errado.'));
	});

	async function befriend() {
		if (!user?.handle) return;
		error = '';
		busy = true;
		try {
			const r = await api.post<{ relation: Card['relation'] }>('/friend-requests', { handle: user.handle });
			user.relation = r.relation;
			sound('select');
			haptic('success');
			await refreshInbox();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
		} finally {
			busy = false;
		}
	}

	function onKey(e: KeyboardEvent) {
		if (e.key === 'Escape') onclose();
	}
</script>

<svelte:window onkeydown={onKey} />

<div class="backdrop" role="presentation" onclick={(e) => e.target === e.currentTarget && onclose()}>
	<div class="card" role="dialog" aria-modal="true" aria-labelledby="player-name" tabindex="-1" in:fly={{ y: 40, duration: ms(220) }}>
		<button class="close" onclick={onclose} aria-label="Fechar"><Icon name="close" size={20} /></button>
		{#if error && !user}
			<p class="error" role="alert">{error}</p>
		{:else if !user}
			<p class="muted">Carregando…</p>
		{:else}
			<Avatar {user} size={112} radius={10} {online} />
			<h2 id="player-name">{user.name}</h2>
			{#if user.is_bot}
				<p class="muted">Bot da casa. Joga sozinho, com uma estratégia simples.</p>
			{:else}
				<p class="handle">@{user.handle}</p>
				{#if online !== undefined}<p class="muted small">{online ? 'Online agora' : 'Fora do app agora'}</p>{/if}
				<div class="action">
					{#if user.relation === 'self'}
						<a class="btn btn-ghost btn-block" href="/perfil">Esse é você · editar perfil</a>
					{:else if user.relation === 'friend'}
						<p class="status"><Icon name="check" size={18} />Vocês são amigos</p>
					{:else if user.relation === 'request_sent'}
						<p class="status muted">Pedido de amizade enviado</p>
					{:else if user.relation === 'request_received'}
						<button class="btn btn-primary btn-block" onclick={befriend} disabled={busy}>
							<Icon name="check" size={18} />Aceitar pedido de amizade
						</button>
					{:else}
						<button class="btn btn-primary btn-block" onclick={befriend} disabled={busy}>
							<Icon name="userplus" size={18} />{busy ? 'Enviando…' : 'Pedir amizade'}
						</button>
					{/if}
				</div>
			{/if}
			{#if error}<p class="error" role="alert">{error}</p>{/if}
		{/if}
	</div>
</div>

<style>
	.backdrop {
		position: fixed;
		inset: 0;
		z-index: 80;
		background: rgba(0, 0, 0, 0.7);
		display: flex;
		align-items: flex-end;
		justify-content: center;
	}
	.card {
		position: relative;
		width: 100%;
		max-width: 560px;
		background: #1f1f1f;
		color: #fff;
		border-radius: 18px 18px 0 0;
		padding: 28px 20px calc(24px + var(--safe-bottom));
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		text-align: center;
	}
	.close {
		position: absolute;
		top: 8px;
		right: 8px;
		width: 44px;
		height: 44px;
		border: none;
		background: none;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
	}
	h2 {
		margin: 8px 0 0;
		font-size: 24px;
	}
	p {
		margin: 0;
		line-height: 1.4;
	}
	.handle {
		color: var(--muted);
		font-size: 15px;
	}
	.small {
		font-size: 13px;
	}
	.action {
		width: 100%;
		margin-top: 12px;
	}
	.status {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		min-height: 48px;
		font-weight: 600;
		color: #46d369;
	}
	.status.muted {
		color: var(--muted);
	}
</style>
