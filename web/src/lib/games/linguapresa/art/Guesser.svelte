<script lang="ts">
	let { size = 80 }: { size?: number } = $props();
</script>

<svg viewBox="0 0 80 80" width={size} height={size} role="img" aria-label="Uma pessoa pensando, com um balão de pergunta">
	<rect width="80" height="80" rx="8" fill="#1f1038" />
	<circle cx="34" cy="46" r="15" fill="#ffd9b0" />
	<circle cx="29" cy="43" r="2" fill="#1f1038" /><circle cx="39" cy="43" r="2" fill="#1f1038" />
	<path d="M30 52h8" stroke="#7a2e3a" stroke-width="3" stroke-linecap="round" />
	<ellipse cx="54" cy="22" rx="16" ry="12" fill="#fff4e0" />
	<text x="54" y="29" text-anchor="middle" font-family="Bebas Neue, Impact, sans-serif" font-size="22" fill="#ff4fa3">?</text>
</svg>
