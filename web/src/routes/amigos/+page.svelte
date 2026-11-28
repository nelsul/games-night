<script lang="ts">
	import { onMount } from 'svelte';
	import { api, ApiError, type FriendRequest, type PublicUser, type UserWithRelation } from '$lib/api';
	import { inbox, refreshInbox } from '$lib/inbox.svelte';
	import Avatar from '$lib/components/Avatar.svelte';
	import Icon from '$lib/components/Icon.svelte';

	let query = $state('');
	let results = $state<UserWithRelation[]>([]);
	let searching = $state(false);
	let friends = $state<PublicUser[] | null>(null);
	let incoming = $state<FriendRequest[]>([]);
	let outgoing = $state<FriendRequest[]>([]);
	let error = $state('');
	let removing = $state<PublicUser | null>(null);

	async function load() {
		const [f, r] = await Promise.all([
			api.get<PublicUser[]>('/friends').catch(() => []),
			api
				.get<{ incoming: FriendRequest[]; outgoing: FriendRequest[] }>('/friend-requests')
				.catch(() => ({ incoming: [], outgoing: [] }))
		]);
		friends = f;
		incoming = r.incoming;
		outgoing = r.outgoing;
	}

	onMount(load);

	// Someone asked or answered while this page is open: the inbox socket pokes us.
	let seen = -1;
	$effect(() => {
		const v = inbox.version;
		if (seen !== -1 && v !== seen) load();
		seen = v;
	});

	$effect(() => {
		const q = query.trim().replace(/^@/, '').toLowerCase();
		if (q.length < 2) {
			results = [];
			searching = false;
			return;
		}
		searching = true;
		const t = setTimeout(async () => {
			try {
				const found = await api.get<UserWithRelation[]>(`/users/search?q=${encodeURIComponent(q)}`);
				if (query.trim().replace(/^@/, '').toLowerCase() === q) results = found;
			} catch {
				results = [];
			} finally {
				searching = false;
			}
		}, 300);
		return () => clearTimeout(t);
	});

	async function run(fn: () => Promise<unknown>) {
		error = '';
		try {
			await fn();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
		}
		await refreshInbox(); // also reloads this page's lists through the effect above
	}

	async function add(u: UserWithRelation) {
		await run(async () => {
			const r = await api.post<{ relation: UserWithRelation['relation'] }>('/friend-requests', { handle: u.handle });
			u.relation = r.relation;
		});
	}

	const answer = (id: string, action: 'accept' | 'decline' | 'cancel') =>
		run(() => api.post(`/friend-requests/${id}/${action}`));

	async function confirmRemove() {
		const u = removing;
		removing = null;
		if (u) await run(() => api.del(`/friends/${u.id}`));
	}
</script>

<svelte:head><title>Amigos · Games Night</title></svelte:head>

<main class="page friends">
	<h1>Amigos</h1>

	<div class="search">
		<label for="search" class="sr-only">Buscar pelo @</label>
		<span class="at" aria-hidden="true">@</span>
		<input
			id="search"
			class="input"
			type="search"
			placeholder="nome de usuário"
			autocapitalize="none"
			autocomplete="off"
			spellcheck="false"
			bind:value={query}
		/>
	</div>
	{#if error}<p class="error" role="alert">{error}</p>{/if}

	{#if query.trim().replace(/^@/, '').length >= 2}
		<section aria-label="Resultados">
			{#if results.length === 0}
				<p class="muted">{searching ? 'Buscando…' : 'Ninguém com esse @.'}</p>
			{:else}
				<ul class="list">
					{#each results as u (u.id)}
						<li>
							<Avatar user={u} size={44} />
							<span class="who"><strong>{u.name}</strong><span class="muted">@{u.handle}</span></span>
							{#if u.relation === 'friend'}
								<span class="tag">Amigos</span>
							{:else if u.relation === 'request_sent'}
								<span class="tag">Pedido enviado</span>
							{:else if u.relation === 'request_received'}
								<button class="btn btn-primary small" onclick={() => add(u)}>Aceitar</button>
							{:else}
								<button class="btn btn-primary small" onclick={() => add(u)} aria-label="Adicionar {u.name}">
									<Icon name="userplus" size={18} />Adicionar
								</button>
							{/if}
						</li>
					{/each}
				</ul>
			{/if}
		</section>
	{/if}

	{#if incoming.length}
		<section>
			<h2 class="row-title">Pedidos de amizade</h2>
			<ul class="list">
				{#each incoming as rq (rq.id)}
					<li>
						<Avatar user={rq.user} size={44} />
						<span class="who"><strong>{rq.user.name}</strong><span class="muted">@{rq.user.handle}</span></span>
						<button class="btn btn-ghost small" onclick={() => answer(rq.id, 'decline')}>Recusar</button>
						<button class="btn btn-primary small" onclick={() => answer(rq.id, 'accept')}>Aceitar</button>
					</li>
				{/each}
			</ul>
		</section>
	{/if}

	<section>
		<h2 class="row-title">Seus amigos</h2>
		{#if friends === null}
			<p class="muted">Carregando…</p>
		{:else if friends.length === 0}
			<p class="muted">Busque a galera pelo @ para adicionar. Depois é só convidar para as salas.</p>
		{:else}
			<ul class="list">
				{#each friends as f (f.id)}
					<li>
						<Avatar user={f} size={44} />
						<span class="who"><strong>{f.name}</strong><span class="muted">@{f.handle}</span></span>
						<button class="icon-btn" onclick={() => (removing = f)} aria-label="Desfazer amizade com {f.name}">
							<Icon name="trash" size={20} />
						</button>
					</li>
				{/each}
			</ul>
		{/if}
	</section>

	{#if outgoing.length}
		<section>
			<h2 class="row-title">Pedidos enviados</h2>
			<ul class="list">
				{#each outgoing as rq (rq.id)}
					<li>
						<Avatar user={rq.user} size={44} />
						<span class="who"><strong>{rq.user.name}</strong><span class="muted">@{rq.user.handle}</span></span>
						<button class="btn btn-ghost small" onclick={() => answer(rq.id, 'cancel')}>Cancelar</button>
					</li>
				{/each}
			</ul>
		</section>
	{/if}
</main>

{#if removing}
	<div class="sheet" role="dialog" aria-modal="true" aria-labelledby="remove-title">
		<div class="sheet-card">
			<h2 id="remove-title">Desfazer amizade com {removing.name}?</h2>
			<p class="muted">Vocês podem se adicionar de novo depois.</p>
			<button class="btn btn-danger btn-block" onclick={confirmRemove}>Desfazer amizade</button>
			<button class="btn btn-ghost btn-block" onclick={() => (removing = null)}>Cancelar</button>
		</div>
	</div>
{/if}

<style>
	.friends {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}
	h1 {
		margin: 8px 0 0;
		font-size: 28px;
	}
	section {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	p {
		margin: 0;
		line-height: 1.45;
	}
	.search {
		position: relative;
	}
	.search .input {
		padding-left: 32px;
	}
	.at {
		position: absolute;
		left: 14px;
		top: 50%;
		transform: translateY(-50%);
		color: var(--muted);
	}
	.list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.list li {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 12px;
		border-radius: var(--radius);
		background: var(--surface);
		min-height: 64px;
	}
	.who {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
		font-size: 14px;
	}
	.who strong {
		font-size: 15px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.small {
		height: 40px;
		padding: 0 12px;
		font-size: 14px;
		min-width: 44px;
	}
	.tag {
		font-size: 13px;
		color: var(--muted);
	}
	.icon-btn {
		width: 44px;
		height: 44px;
		border: none;
		background: none;
		color: var(--muted);
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
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
		gap: 12px;
	}
	.sheet-card h2 {
		margin: 0;
		font-size: 20px;
	}
	.btn-danger {
		background: var(--brand);
		color: #fff;
	}
</style>
