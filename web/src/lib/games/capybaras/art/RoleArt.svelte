<script lang="ts">
	// A card's picture for the game page's "Personagens" list.
	import CardArt from './CardArt.svelte';
	let { size = 88, kind }: { size?: number; kind: string } = $props();
</script>

<span style="display: flex; width: {size}px; height: {size}px; border-radius: 50%; overflow: hidden; background: #f6e7c8; align-items: center; justify-content: center">
	<CardArt {kind} size={size * 0.92} />
</span>
