<script lang="ts">
	let { size = 80 }: { size?: number } = $props();
</script>

<svg viewBox="0 0 80 80" width={size} height={size} role="img" aria-label="Três cartas com números">
	<rect width="80" height="80" rx="8" fill="#2b1636" />
	<g stroke="#ffb347" stroke-width="2" fill="#fff4e0">
		<rect x="10" y="30" width="22" height="34" rx="4" />
		<rect x="29" y="22" width="22" height="42" rx="4" />
		<rect x="48" y="14" width="22" height="50" rx="4" />
	</g>
	<g font-family="Bebas Neue, Impact, sans-serif" font-size="14" fill="#7a2e3a" text-anchor="middle">
		<text x="21" y="52">8</text><text x="40" y="52">42</text><text x="59" y="52">91</text>
	</g>
</svg>
