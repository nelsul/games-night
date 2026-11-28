<script lang="ts">
	let { size = 80 }: { size?: number } = $props();
</script>

<svg viewBox="0 0 80 80" width={size} height={size} role="img" aria-label="Três cartas lado a lado e uma lupa">
	<rect width="80" height="80" rx="8" fill="#08222b" />
	<g fill="#fff1d6"><rect x="8" y="24" width="14" height="22" rx="2" /><rect x="25" y="24" width="14" height="22" rx="2" /><rect x="42" y="24" width="14" height="22" rx="2" /></g>
	<g fill="#1c5a66"><circle cx="15" cy="35" r="3" /><circle cx="32" cy="35" r="3" /><circle cx="49" cy="35" r="3" /></g>
	<circle cx="56" cy="52" r="12" fill="#ffd36e22" stroke="#ffd36e" stroke-width="4" />
	<path d="M65 61l9 9" stroke="#ffd36e" stroke-width="5" stroke-linecap="round" />
</svg>
