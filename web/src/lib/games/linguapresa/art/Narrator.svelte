<script lang="ts">
	let { size = 80 }: { size?: number } = $props();
</script>

<svg viewBox="0 0 80 80" width={size} height={size} role="img" aria-label="Uma pessoa falando, com a boca aberta">
	<rect width="80" height="80" rx="8" fill="#10213a" />
	<circle cx="34" cy="36" r="16" fill="#ffd9b0" />
	<path d="M30 42q4 6 8 0" stroke="#7a2e3a" stroke-width="3" fill="#7a2e3a" stroke-linecap="round" />
	<circle cx="29" cy="32" r="2" fill="#10213a" /><circle cx="39" cy="32" r="2" fill="#10213a" />
	<g stroke="#ff7a45" stroke-width="3" stroke-linecap="round" fill="none">
		<path d="M54 30q5 6 0 12" /><path d="M60 25q9 11 0 22" />
	</g>
</svg>
