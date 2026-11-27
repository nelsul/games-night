<script lang="ts">
	let { size = 80 }: { size?: number } = $props();
</script>

<svg viewBox="0 0 80 80" width={size} height={size} role="img" aria-label="Uma campainha vermelha">
	<rect width="80" height="80" rx="8" fill="#10213a" />
	<rect x="18" y="52" width="44" height="10" rx="4" fill="#5a6b85" />
	<path d="M24 52q0-24 16-24t16 24z" fill="#e23d3d" />
	<circle cx="40" cy="26" r="4" fill="#ff8a8a" />
	<path d="M14 30l-6-4M66 30l6-4M40 14V8" stroke="#ffb347" stroke-width="3" stroke-linecap="round" />
</svg>
