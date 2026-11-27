<script lang="ts">
	// One card, face up: the kind's color, its picture and its name (and, when big, what it does).
	import CardArt from './art/CardArt.svelte';
	import { cardInfo } from './types';
	let { kind, width = 92, text = false }: { kind: string; width?: number; text?: boolean } = $props();
	const info = $derived(cardInfo[kind] ?? { name: kind, text: '', color: '#666' });
</script>

<span class="cf" style="--w: {width}px; --c: {info.color}">
	<span class="art"><CardArt {kind} size={width * 0.78} /></span>
	<span class="name">{info.name}</span>
	{#if text}<span class="text">{info.text}</span>{/if}
</span>

<style>
	.cf {
		width: var(--w);
		min-height: calc(var(--w) * 1.3);
		border-radius: calc(var(--w) * 0.11);
		background: #fbf1dc;
		border: calc(var(--w) * 0.045) solid var(--c);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.35);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2px;
		padding: calc(var(--w) * 0.06) calc(var(--w) * 0.05);
		color: #2b1a0e;
		box-sizing: border-box;
		overflow: hidden;
	}
	.art {
		display: flex;
		border-radius: calc(var(--w) * 0.08);
		background: color-mix(in srgb, var(--c) 18%, #fff7e6);
	}
	.name {
		font-weight: 800;
		font-size: calc(var(--w) * 0.125);
		line-height: 1.05;
		text-align: center;
		color: color-mix(in srgb, var(--c) 70%, #2b1a0e);
	}
	.text {
		font-size: calc(var(--w) * 0.085);
		line-height: 1.25;
		text-align: center;
		color: #5a4636;
	}
</style>
