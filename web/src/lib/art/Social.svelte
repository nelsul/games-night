<script lang="ts">
	let { size = 200 }: { size?: number } = $props();
	const id = $props.id();
</script>

<svg viewBox="0 0 200 200" width={size} height={size} role="img" aria-label="Engenharia Social: pessoa sorridente piscando ao telefone, com um anzol ao lado">
	<defs><clipPath id="c{id}"><circle cx="100" cy="100" r="96" /></clipPath></defs>
	<circle cx="100" cy="100" r="96" fill="#221836" />
	<g clip-path="url(#c{id})">
		<rect x="20" y="34" width="82" height="26" rx="13" fill="#A371F7" />
		<path d="M44 58 L40 70 L54 59 Z" fill="#A371F7" />
		<text x="61" y="51.5" text-anchor="middle" font-family="Inter, sans-serif" font-size="11" font-weight="700" fill="#0D1117">oi, é do TI</text>
		<path d="M160 34 L160 62 C160 74 146 74 146 62 L152 66" fill="none" stroke="#A371F7" stroke-width="4" stroke-linecap="round" stroke-linejoin="round" />
		<circle cx="160" cy="30" r="4" fill="none" stroke="#A371F7" stroke-width="3" />
		<path d="M26 212 C26 160 58 138 100 138 C142 138 174 160 174 212 Z" fill="#6E40C9" />
		<path d="M84 139 L100 166 L116 139 Z" fill="#E6EDF3" />
		<path d="M84 139 L74 168 L92 152 Z M116 139 L126 168 L108 152 Z" fill="#8957E5" />
		<rect x="89" y="112" width="22" height="30" rx="8" fill="#F1C27D" />
		<circle cx="100" cy="90" r="32" fill="#F1C27D" />
		<path d="M66 90 C62 60 88 50 106 54 C124 58 138 70 132 90 C122 72 100 66 82 74 C76 78 70 84 66 90 Z" fill="#8B5A2B" />
		<path d="M83 93 Q89 88 95 93" fill="none" stroke="#1B1B1B" stroke-width="3" stroke-linecap="round" />
		<circle cx="111" cy="92" r="3.5" fill="#1B1B1B" />
		<path d="M87 103 Q100 118 113 103" fill="none" stroke="#1B1B1B" stroke-width="3" stroke-linecap="round" />
		<rect x="128" y="76" width="15" height="28" rx="4" fill="#161B22" stroke="#A371F7" stroke-width="2.5" />
		<circle cx="136" cy="106" r="9" fill="#F1C27D" />
	</g>
</svg>
