<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type RoomSummary } from '$lib/api';
	import Icon from '$lib/components/Icon.svelte';

	let rooms = $state<RoomSummary[] | null>(null);
	onMount(async () => {
		rooms = await api.get<RoomSummary[]>('/rooms').catch(() => []);
	});
</script>

<svelte:head><title>Salas · Games Night</title></svelte:head>

<main class="page rooms">
	<h1>Suas salas</h1>
	<a class="btn btn-primary btn-block" href="/codigo"><Icon name="plus" size={18} />Entrar com código</a>
	{#if rooms === null}
		<p class="muted">Carregando…</p>
	{:else if rooms.length === 0}
		<p class="muted">Você não está em nenhuma sala. Crie uma no início ou entre com o código de um amigo.</p>
	{:else}
		<ul>
			{#each rooms as room (room.code)}
				<li>
					<a href="/sala/{room.code}">
						<span class="code">{room.code}</span>
						<span class="text">
							<strong>{room.game.name}</strong>
							<span class="muted"
								>{room.status === 'playing' ? 'Em partida' : 'No lobby'} · {room.players}
								{room.players === 1 ? 'pessoa' : 'pessoas'}{room.is_host ? ' · você é host' : ''}</span
							>
						</span>
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</main>

<style>
	.rooms {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	h1 {
		margin: 8px 0 0;
		font-size: 28px;
	}
	p {
		margin: 0;
		line-height: 1.45;
	}
	ul {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	li a {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 12px;
		border-radius: var(--radius);
		background: var(--surface);
		text-decoration: none;
		min-height: 64px;
	}
	.code {
		font-family: var(--display);
		font-size: 26px;
		letter-spacing: 3px;
		width: 74px;
		text-align: center;
		background: #1b0b0e;
		border-radius: 4px;
		padding: 6px 0 2px;
	}
	.text {
		display: flex;
		flex-direction: column;
		gap: 2px;
		font-size: 14px;
	}
</style>
