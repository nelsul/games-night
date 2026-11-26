<script lang="ts">
	let { size = 200 }: { size?: number } = $props();
	const id = $props.id();
</script>

<svg viewBox="0 0 200 200" width={size} height={size} role="img" aria-label="Zero-day: hacker quase invisível, um contorno tracejado com um olho vermelho">
	<defs><clipPath id="c{id}"><circle cx="100" cy="100" r="96" /></clipPath></defs>
	<circle cx="100" cy="100" r="96" fill="#1E1024" />
	<g clip-path="url(#c{id})">
		<g font-family="JetBrains Mono, monospace" font-weight="700" fill="#3D2347">
			<text x="20" y="54" font-size="13">0x00</text>
			<text x="146" y="46" font-size="13">CVE-?</text>
			<text x="18" y="132" font-size="12">null</text>
		</g>
		<rect x="0" y="70" width="200" height="4" fill="#BC8CFF" opacity="0.18" />
		<rect x="0" y="124" width="200" height="3" fill="#F85149" opacity="0.2" />
		<path d="M24 212 C24 158 56 132 100 132 C144 132 176 158 176 212" fill="#2A1733" fill-opacity="0.6" stroke="#BC8CFF" stroke-width="2.5" stroke-dasharray="7 6" />
		<path d="M58 130 C48 74 70 40 100 40 C130 40 152 74 142 130" fill="#2A1733" fill-opacity="0.6" stroke="#BC8CFF" stroke-width="2.5" stroke-dasharray="7 6" />
		<ellipse cx="100" cy="94" rx="27" ry="32" fill="#0A060C" opacity="0.85" />
		<ellipse cx="111" cy="94" rx="6" ry="3.5" fill="#F85149" />
		<path d="M83 94 L95 94" stroke="#BC8CFF" stroke-width="2" stroke-linecap="round" opacity="0.6" />
		<rect x="62" y="80" width="30" height="5" fill="#1E1024" />
		<rect x="104" y="106" width="24" height="4" fill="#1E1024" />
		<rect x="58" y="160" width="84" height="40" rx="7" fill="none" stroke="#BC8CFF" stroke-width="2" stroke-dasharray="5 5" />
		<text x="100" y="188" text-anchor="middle" font-family="JetBrains Mono, monospace" font-size="18" font-weight="700" fill="#BC8CFF">0day</text>
	</g>
</svg>
