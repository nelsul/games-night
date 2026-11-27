<script lang="ts">
	// Admins only: read what people sent, mark it done or delete it.
	import { onMount } from 'svelte';
	import { fly, slide } from 'svelte/transition';
	import { api, ApiError, type FeedbackItem, type FeedbackKind } from '$lib/api';
	import { session } from '$lib/session.svelte';
	import { haptic, ms } from '$lib/feel.svelte';
	import Avatar from '$lib/components/Avatar.svelte';
	import Icon from '$lib/components/Icon.svelte';

	const kindInfo: Record<FeedbackKind, { name: string; icon: 'chat' | 'bug' | 'bulb' }> = {
		feedback: { name: 'Feedback', icon: 'chat' },
		bug: { name: 'Bug', icon: 'bug' },
		idea: { name: 'Ideia de jogo', icon: 'bulb' }
	};

	let status = $state<'open' | 'done'>('open');
	let items = $state<FeedbackItem[] | null>(null);
	let open = $state(0);
	let error = $state('');
	let busy = $state('');
	let deleting = $state<FeedbackItem | null>(null);
	let filter = $state<FeedbackKind | 'all'>('all');

	const shown = $derived((items ?? []).filter((it) => filter === 'all' || it.kind === filter));

	async function load() {
		error = '';
		try {
			const r = await api.get<{ items: FeedbackItem[]; open: number }>(`/admin/feedback?status=${status}`);
			items = r.items;
			open = r.open;
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
			items = [];
		}
	}

	onMount(() => {
		if (session.me?.is_admin) load();
	});

	function show(s: 'open' | 'done') {
		if (status === s) return;
		status = s;
		items = null;
		load();
	}

	async function finish(it: FeedbackItem) {
		busy = it.id;
		try {
			await api.post(`/admin/feedback/${it.id}/done`);
			items = (items ?? []).filter((x) => x.id !== it.id);
			open = Math.max(0, open - 1);
			haptic('success');
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
		} finally {
			busy = '';
		}
	}

	async function remove() {
		const it = deleting;
		deleting = null;
		if (!it) return;
		busy = it.id;
		try {
			await api.del(`/admin/feedback/${it.id}`);
			items = (items ?? []).filter((x) => x.id !== it.id);
			if (!it.done_at) open = Math.max(0, open - 1);
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
		} finally {
			busy = '';
		}
	}

	const fmt = new Intl.DateTimeFormat('pt-BR', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' });
	const when = (iso: string) => fmt.format(new Date(iso));
</script>

<svelte:head><title>Feedbacks · Games Night</title></svelte:head>

<main class="page admin">
	<header>
		<a href="/perfil" aria-label="Voltar"><Icon name="back" /></a>
	</header>

	{#if !session.me?.is_admin}
		<h1>Feedbacks</h1>
		<p class="muted">Só admins podem ver esta página.</p>
	{:else}
		<h1>Feedbacks</h1>
		<div class="tabs" role="tablist">
			<button role="tab" aria-selected={status === 'open'} onclick={() => show('open')}>
				Abertos{#if open}<span class="pill">{open}</span>{/if}
			</button>
			<button role="tab" aria-selected={status === 'done'} onclick={() => show('done')}>Concluídos</button>
		</div>
		<div class="filters" role="radiogroup" aria-label="Tipo">
			<button role="radio" aria-checked={filter === 'all'} onclick={() => (filter = 'all')}>Todos</button>
			{#each Object.entries(kindInfo) as [id, k] (id)}
				<button role="radio" aria-checked={filter === id} class={id} onclick={() => (filter = id as FeedbackKind)}>
					<Icon name={k.icon} size={15} />{k.name}
				</button>
			{/each}
		</div>

		{#if error}<p class="error" role="alert">{error}</p>{/if}

		{#if items === null}
			<p class="muted">Carregando…</p>
		{:else if shown.length === 0}
			<div class="empty">
				<Icon name="inbox" size={40} />
				<p class="muted">{status === 'open' ? 'Nada novo por aqui.' : 'Nenhum feedback concluído.'}</p>
			</div>
		{:else}
			<ul class="list">
				{#each shown as it, i (it.id)}
					<li class="item {it.kind}" in:fly={{ y: 10, duration: ms(220), delay: ms(Math.min(i, 8) * 40) }} out:slide={{ duration: ms(220) }}>
						<div class="top">
							<span class="kind"><Icon name={kindInfo[it.kind].icon} size={15} />{kindInfo[it.kind].name}</span>
							<span class="muted date">{when(it.created_at)}</span>
						</div>
						<p class="body">{it.body}</p>
						<div class="author">
							{#if it.author}
								<Avatar user={it.author} size={24} radius={12} />
								<span>{it.author.name} <span class="muted">@{it.author.handle}</span></span>
							{:else}
								<span class="mask"><Icon name="mask" size={15} /></span>
								<span class="muted">Anônimo</span>
							{/if}
						</div>
						{#if it.done_at}
							<p class="muted done">Concluído por {it.done_by || 'um admin'} em {when(it.done_at)}</p>
						{/if}
						<div class="actions">
							<button class="btn btn-ghost small" disabled={busy === it.id} onclick={() => (deleting = it)}>
								<Icon name="trash" size={16} />Excluir
							</button>
							{#if !it.done_at}
								<button class="btn btn-primary small" disabled={busy === it.id} onclick={() => finish(it)}>
									<Icon name="check" size={16} />Concluir
								</button>
							{/if}
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	{/if}
</main>

{#if deleting}
	<div class="sheet" role="dialog" aria-modal="true" aria-labelledby="del-title">
		<div class="sheet-card">
			<h2 id="del-title">Excluir este feedback?</h2>
			<p class="muted">Ele some de vez, para todos os admins. Não dá para desfazer.</p>
			<button class="btn btn-primary btn-block danger" onclick={remove}>Excluir</button>
			<button class="btn btn-ghost btn-block" onclick={() => (deleting = null)}>Cancelar</button>
		</div>
	</div>
{/if}

<style>
	.admin {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	header {
		height: 44px;
		display: flex;
		align-items: center;
	}
	header a {
		width: 44px;
		height: 44px;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-left: -10px;
	}
	h1 {
		margin: 4px 0 0;
		font-size: 28px;
	}
	p {
		margin: 0;
	}
	.tabs {
		display: flex;
		gap: 24px;
		border-bottom: 1px solid var(--line);
	}
	.tabs button {
		background: none;
		border: none;
		border-bottom: 3px solid transparent;
		padding: 10px 0;
		min-height: 44px;
		color: var(--muted);
		font: inherit;
		font-weight: 600;
		font-size: 15px;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.tabs button[aria-selected='true'] {
		color: #fff;
		border-bottom-color: var(--brand);
	}
	.pill {
		min-width: 20px;
		height: 20px;
		padding: 0 6px;
		border-radius: 999px;
		background: var(--brand);
		color: #fff;
		font-size: 12px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}
	.filters {
		display: flex;
		gap: 6px;
		overflow-x: auto;
		scrollbar-width: none;
		margin: 0 -16px;
		padding: 0 16px;
	}
	.filters button {
		flex-shrink: 0;
		display: flex;
		align-items: center;
		gap: 5px;
		height: 34px;
		padding: 0 12px;
		border-radius: 999px;
		border: 1px solid var(--line-strong);
		background: none;
		color: var(--muted);
		font: inherit;
		font-size: 13px;
		cursor: pointer;
	}
	.filters button[aria-checked='true'] {
		background: #fff;
		color: #111;
		border-color: #fff;
	}
	.list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.item {
		--tone: #4493f8;
		background: var(--surface);
		border-radius: var(--radius);
		border-left: 3px solid var(--tone);
		padding: 12px 14px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.item.bug {
		--tone: #ff7b72;
	}
	.item.idea {
		--tone: #e3b341;
	}
	.top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}
	.kind {
		display: flex;
		align-items: center;
		gap: 5px;
		color: var(--tone);
		font-size: 13px;
		font-weight: 700;
	}
	.date {
		font-size: 12px;
	}
	.body {
		white-space: pre-wrap;
		overflow-wrap: anywhere;
		line-height: 1.5;
		font-size: 15px;
	}
	.author {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13px;
	}
	.mask {
		width: 24px;
		height: 24px;
		border-radius: 50%;
		background: var(--surface-2);
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.done {
		font-size: 12px;
	}
	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
	}
	.small {
		height: 40px;
		padding: 0 14px;
		font-size: 14px;
	}
	.empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 10px;
		padding: 48px 0;
		color: var(--muted);
	}
	.sheet {
		position: fixed;
		inset: 0;
		z-index: 80;
		background: rgba(0, 0, 0, 0.7);
		display: flex;
		align-items: flex-end;
		justify-content: center;
	}
	.sheet-card {
		width: 100%;
		max-width: 560px;
		background: #1f1f1f;
		border-radius: 18px 18px 0 0;
		padding: 24px 20px calc(24px + var(--safe-bottom));
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.sheet-card h2 {
		margin: 0;
		font-size: 20px;
	}
	.danger {
		background: #da3633;
		color: #fff;
	}
</style>
