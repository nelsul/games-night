<script lang="ts">
	let { size = 80 }: { size?: number } = $props();
</script>

<svg viewBox="0 0 80 80" width={size} height={size} role="img" aria-label="Uma carta com um balão de fala">
	<rect width="80" height="80" rx="8" fill="#08222b" />
	<rect x="14" y="22" width="32" height="46" rx="4" fill="#fff1d6" />
	<circle cx="30" cy="40" r="7" fill="#ff8f6b" /><path d="M18 62l9-10 6 6 7-9 6 13Z" fill="#1c5a66" />
	<path d="M42 10h26a4 4 0 0 1 4 4v14a4 4 0 0 1-4 4H58l-8 8v-8h-8a4 4 0 0 1-4-4V14a4 4 0 0 1 4-4Z" fill="#ffd36e" />
	<path d="M48 18h18M48 24h12" stroke="#08222b" stroke-width="2.5" stroke-linecap="round" />
</svg>
