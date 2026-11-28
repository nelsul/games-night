<script lang="ts">
	// One card picture, drawn from its number (see art.ts). It fills whatever box it is in.
	import { cardSvg } from './art';

	let { id, label = '' }: { id: number; label?: string } = $props();
	const svg = $derived(cardSvg(id));
</script>

<svg class="card-art" viewBox="0 0 200 300" preserveAspectRatio="xMidYMid slice" role="img" aria-label={label || 'Carta com uma cena desenhada'}>
	{@html svg}
</svg>

<style>
	.card-art {
		display: block;
		width: 100%;
		height: 100%;
	}
</style>
