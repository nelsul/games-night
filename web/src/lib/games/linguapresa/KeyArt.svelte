<script lang="ts">
	// The Língua Presa poster used by the billboard and the game page: a card with a word
	// and a row of crossed-out forbidden words, under a stopwatch.
	let { height = 290 }: { height?: number } = $props();
	const id = $props.id();
	const taboo = ['pão', 'forno', 'queijo', 'massa'];
</script>

<svg viewBox="0 0 358 290" width="100%" {height} preserveAspectRatio="xMidYMid slice" style="display: block" role="img" aria-label="Arte do Língua Presa: uma carta com uma palavra e quatro palavras proibidas riscadas">
	<defs>
		<linearGradient id="g{id}" x1="0" y1="0" x2="0" y2="1">
			<stop offset="0" stop-color="#1f1038" />
			<stop offset="1" stop-color="#5a2a8a" />
		</linearGradient>
		<linearGradient id="f{id}" x1="0" y1="0" x2="0" y2="1">
			<stop offset="0.65" stop-color="#140a24" stop-opacity="0" />
			<stop offset="1" stop-color="#140a24" stop-opacity="1" />
		</linearGradient>
	</defs>
	<rect width="358" height="290" fill="url(#g{id})" />
	<circle cx="298" cy="62" r="30" fill="none" stroke="#ff4fa3" stroke-width="6" />
	<path d="M298 62V44M298 62l12 8" stroke="#ff4fa3" stroke-width="5" stroke-linecap="round" />
	<g transform="rotate(-4 179 150)">
		<rect x="74" y="38" width="210" height="222" rx="14" fill="#fff4e0" stroke="#ff4fa3" stroke-width="4" />
		<rect x="74" y="38" width="210" height="52" rx="14" fill="#ff4fa3" />
		<text x="179" y="76" text-anchor="middle" font-family="Bebas Neue, Impact, sans-serif" font-size="36" fill="#fff" letter-spacing="2">PIZZA</text>
		{#each taboo as w, i (w)}
			<g transform="translate(0 {i * 36})">
				<text x="179" y="130" text-anchor="middle" font-family="Inter, sans-serif" font-weight="700" font-size="20" fill="#7a2e3a">{w}</text>
				<path d="M126 123h106" stroke="#e23d3d" stroke-width="3" stroke-linecap="round" />
			</g>
		{/each}
	</g>
	<rect width="358" height="290" fill="url(#f{id})" />
</svg>
