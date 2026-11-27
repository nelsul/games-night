<script lang="ts">
	// The Escala poster used by the hub's billboard and the game page: cards climbing a scale.
	let { height = 290 }: { height?: number } = $props();
	const id = $props.id();
	const cards = [
		{ x: 22, h: 70, n: '12' },
		{ x: 90, h: 100, n: '34' },
		{ x: 158, h: 130, n: '58' },
		{ x: 226, h: 160, n: '77' }
	];
</script>

<svg viewBox="0 0 358 290" width="100%" {height} preserveAspectRatio="xMidYMid slice" style="display: block" role="img" aria-label="Arte do Escala: cartas com números em ordem crescente">
	<defs>
		<linearGradient id="g{id}" x1="0" y1="0" x2="0" y2="1">
			<stop offset="0" stop-color="#2b1636" />
			<stop offset="1" stop-color="#7a2e3a" />
		</linearGradient>
		<linearGradient id="f{id}" x1="0" y1="0" x2="0" y2="1">
			<stop offset="0.6" stop-color="#1a0f22" stop-opacity="0" />
			<stop offset="1" stop-color="#1a0f22" stop-opacity="1" />
		</linearGradient>
	</defs>
	<rect width="358" height="290" fill="url(#g{id})" />
	<circle cx="290" cy="70" r="34" fill="#ffb347" opacity="0.9" />
	{#each cards as c (c.x)}
		<g transform="translate({c.x} {250 - c.h})">
			<rect width="62" height={c.h + 20} rx="8" fill="#fff4e0" stroke="#ffb347" stroke-width="3" />
			<text x="31" y="38" text-anchor="middle" font-family="Bebas Neue, Impact, sans-serif" font-size="32" fill="#7a2e3a">{c.n}</text>
		</g>
	{/each}
	<path d="M10 262h338" stroke="#ffb347" stroke-width="3" stroke-linecap="round" />
	<rect width="358" height="290" fill="url(#f{id})" />
</svg>
