<script lang="ts">
	// A full-screen look at a game's screenshots: swipe or use the arrows, Esc closes.
	import { fade } from 'svelte/transition';
	import type { Screenshot } from '$lib/games';
	import { ms } from '$lib/feel.svelte';
	import Icon from './Icon.svelte';

	let { shots, index = $bindable(0), onclose }: { shots: Screenshot[]; index: number; onclose: () => void } = $props();

	const shot = $derived(shots[index]);
	const go = (d: number) => (index = (index + d + shots.length) % shots.length);

	let startX: number | null = null;
	function down(e: PointerEvent) {
		startX = e.clientX;
	}
	function up(e: PointerEvent) {
		if (startX === null) return;
		const dx = e.clientX - startX;
		startX = null;
		if (Math.abs(dx) > 40) go(dx < 0 ? 1 : -1);
	}

	function onKey(e: KeyboardEvent) {
		if (e.key === 'Escape') onclose();
		else if (e.key === 'ArrowRight') go(1);
		else if (e.key === 'ArrowLeft') go(-1);
	}
</script>

<svelte:window onkeydown={onKey} />

<div class="viewer" role="dialog" aria-modal="true" aria-label="Imagens do jogo" transition:fade={{ duration: ms(160) }}>
	<button class="close" onclick={onclose} aria-label="Fechar"><Icon name="close" size={22} /></button>
	<div class="stage" onpointerdown={down} onpointerup={up} role="presentation">
		{#key index}
			<img src={shot.src} alt={shot.caption} in:fade={{ duration: ms(140) }} draggable="false" />
		{/key}
	</div>
	<div class="bar">
		<button class="nav" onclick={() => go(-1)} aria-label="Anterior" disabled={shots.length < 2}><Icon name="back" size={22} /></button>
		<p><span>{shot.caption}</span><span class="count">{index + 1} de {shots.length}</span></p>
		<button class="nav next" onclick={() => go(1)} aria-label="Próxima" disabled={shots.length < 2}><Icon name="back" size={22} /></button>
	</div>
</div>

<style>
	.viewer {
		position: fixed;
		inset: 0;
		z-index: 90;
		background: rgba(0, 0, 0, 0.94);
		display: flex;
		flex-direction: column;
		padding: calc(12px + var(--safe-top)) 12px calc(12px + var(--safe-bottom));
	}
	.close {
		align-self: flex-end;
		width: 44px;
		height: 44px;
		border: none;
		background: rgba(255, 255, 255, 0.1);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		color: #fff;
	}
	.stage {
		flex: 1;
		min-height: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		touch-action: pan-y;
		user-select: none;
	}
	img {
		max-width: 100%;
		max-height: 100%;
		object-fit: contain;
		border-radius: 14px;
	}
	.bar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding-top: 10px;
	}
	.bar p {
		flex: 1;
		margin: 0;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2px;
		text-align: center;
		font-size: 14px;
		color: #fff;
	}
	.count {
		font-size: 12px;
		color: var(--muted);
	}
	.nav {
		width: 44px;
		height: 44px;
		border: none;
		background: rgba(255, 255, 255, 0.1);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		color: #fff;
	}
	.nav.next {
		transform: scaleX(-1);
	}
	.nav:disabled {
		opacity: 0.3;
	}
</style>
