<script lang="ts">
	// The back of a card: a lagoon green with a little capybara.
	import Capy from './Capy.svelte';
	let { width = 64 }: { width?: number } = $props();
</script>

<svg viewBox="0 0 70 100" width={width} height={(width * 100) / 70} aria-hidden="true" style="display: block">
	<rect x="1.5" y="1.5" width="67" height="97" rx="9" fill="#1d6b4f" stroke="#f6d365" stroke-width="3" />
	<rect x="7" y="7" width="56" height="86" rx="6" fill="none" stroke="#f6d365" stroke-opacity="0.45" stroke-dasharray="3 4" />
	<circle cx="35" cy="50" r="22" fill="#2b8a66" />
	<Capy x={17} y={32} s={0.36} />
</svg>
