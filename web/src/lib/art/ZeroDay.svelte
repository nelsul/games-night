<script lang="ts">
	let { size = 200 }: { size?: number } = $props();
	const id = $props.id();
</script>

<svg viewBox="0 0 200 200" width={size} height={size} role="img" aria-label="Zero-day: hacker disfarçado de dev, de camiseta verde e máscara branca sorridente, com a imagem falhando em glitch">
	<defs><clipPath id="c{id}"><circle cx="100" cy="100" r="96" /></clipPath></defs>
	<circle cx="100" cy="100" r="96" fill="#071C1F" />
	<g clip-path="url(#c{id})">
		<g fill="#39C5CF" opacity="0.14">
			<rect x="0" y="40" width="200" height="3" /><rect x="0" y="62" width="200" height="2" /><rect x="0" y="118" width="200" height="3" /><rect x="0" y="150" width="200" height="2" />
		</g>
		<text x="22" y="54" font-family="JetBrains Mono, monospace" font-size="12" font-weight="700" fill="#145A60">CVE-????</text>
		<text x="140" y="132" font-family="JetBrains Mono, monospace" font-size="12" font-weight="700" fill="#145A60">0x00</text>
		<!-- chromatic ghost copies: the disguise doesn't quite hold -->
		<g opacity="0.55">
			<path d="M20 212 C20 162 52 140 96 140 C140 140 172 162 172 212 Z" fill="#F85149" transform="translate(-6 0)" />
			<path d="M30 212 C30 162 62 140 106 140 C150 140 182 162 182 212 Z" fill="#39C5CF" />
		</g>
		<path d="M26 212 C26 160 58 140 100 140 C142 140 174 160 174 212 Z" fill="#3FB950" />
		<rect x="118" y="160" width="30" height="16" rx="4" fill="#071C1F" />
		<text x="133" y="172" text-anchor="middle" font-family="JetBrains Mono, monospace" font-size="10" font-weight="700" fill="#39C5CF">0day</text>
		<rect x="89" y="112" width="22" height="30" rx="8" fill="#D8A27A" />
		<circle cx="100" cy="90" r="32" fill="#D8A27A" />
		<path d="M68 86 C66 58 90 50 112 56 C128 60 136 72 132 86 C120 70 96 66 72 76 Z" fill="#0F2A2D" />
		<path d="M70 86 C70 72 84 66 100 66 C116 66 130 72 130 86 C130 108 116 122 100 122 C84 122 70 108 70 86 Z" fill="#F0F6FC" />
		<path d="M80 88 Q88 82 96 88 Q88 92 80 88 Z M104 88 Q112 82 120 88 Q112 92 104 88 Z" fill="#071C1F" />
		<circle cx="88" cy="88" r="2" fill="#F85149" />
		<circle cx="112" cy="88" r="2" fill="#F85149" />
		<path d="M86 104 Q100 116 114 104" fill="none" stroke="#071C1F" stroke-width="3" stroke-linecap="round" />
		<path d="M70 86 L60 80 M130 86 L140 80" stroke="#F0F6FC" stroke-width="2" />
		<!-- glitch slices -->
		<rect x="66" y="96" width="70" height="5" fill="#071C1F" />
		<rect x="74" y="96" width="70" height="5" fill="#F0F6FC" opacity="0.85" />
		<rect x="34" y="176" width="60" height="4" fill="#39C5CF" />
		<rect x="110" y="132" width="40" height="3" fill="#F85149" />
	</g>
</svg>
