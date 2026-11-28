<script lang="ts">
	import type { RoomSummary } from '$lib/api';
	import { games } from '$lib/games';

	let { room }: { room: RoomSummary } = $props();
	const game = $derived(games[room.game.id]);
	const status = $derived(room.status === 'playing' ? 'Em partida' : 'No lobby');
	const fill = $derived(Math.min(100, Math.round((room.players / room.game.max_players) * 100)));
</script>

<a class="card" href="/sala/{room.code}">
	<span class="thumb">
		{#if game}<span class="title"><game.Title size={24} /></span>{:else}<span class="fallback">{room.game.name}</span>{/if}
		<span class="code">{room.code}</span>
	</span>
	<span class="bar" aria-hidden="true"><span style="width: {fill}%"></span></span>
	<span class="caption">{status} · {room.players} {room.players === 1 ? 'pessoa' : 'pessoas'}</span>
</a>

<style>
	.card {
		width: 160px;
		flex-shrink: 0;
		border-radius: var(--radius);
		overflow: hidden;
		background: var(--surface-2);
		text-decoration: none;
		color: #fff;
		display: flex;
		flex-direction: column;
		scroll-snap-align: start;
	}
	.thumb {
		height: 90px;
		background: #1b0b0e;
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.title {
		display: flex;
		justify-content: center;
		max-width: 100%;
		padding: 0 8px 18px;
		text-align: center;
		overflow: hidden;
	}
	.code {
		position: absolute;
		left: 8px;
		bottom: 6px;
		font-size: 11px;
		font-weight: 700;
		background: rgba(0, 0, 0, 0.6);
		padding: 2px 6px;
		border-radius: 3px;
		letter-spacing: 1px;
	}
	.bar {
		height: 3px;
		background: #4d4d4d;
	}
	.bar span {
		display: block;
		height: 3px;
		background: var(--brand);
	}
	.caption {
		padding: 8px;
		font-size: 12px;
		color: #d6d6d6;
	}
</style>
