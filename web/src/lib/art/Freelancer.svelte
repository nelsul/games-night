<script lang="ts">
	let { size = 200 }: { size?: number } = $props();
	const id = $props.id();
</script>

<svg viewBox="0 0 200 200" width={size} height={size} role="img" aria-label="Freelancer: hacker de boné virado, óculos escuros e fone no pescoço, sozinho com mochila, café e um notebook com adesivo R$/h">
	<defs><clipPath id="c{id}"><circle cx="100" cy="100" r="96" /></clipPath></defs>
	<circle cx="100" cy="100" r="96" fill="#2A1708" />
	<g clip-path="url(#c{id})">
		<circle cx="160" cy="44" r="16" fill="#3D2410" />
		<path d="M152 44 H168 M160 36 V52" stroke="#2A1708" stroke-width="3" />
		<path d="M24 212 C24 160 56 138 100 138 C144 138 176 160 176 212 Z" fill="#F0883E" />
		<path d="M54 152 L62 212 M146 152 L138 212" stroke="#5C3A1A" stroke-width="9" stroke-linecap="round" />
		<rect x="89" y="112" width="22" height="30" rx="8" fill="#6B4226" />
		<path d="M70 128 C70 150 130 150 130 128" fill="none" stroke="#161B22" stroke-width="7" />
		<rect x="62" y="122" width="14" height="20" rx="6" fill="#161B22" />
		<rect x="124" y="122" width="14" height="20" rx="6" fill="#161B22" />
		<circle cx="100" cy="90" r="31" fill="#6B4226" />
		<path d="M68 82 C68 56 132 56 132 82 Z" fill="#DA3633" />
		<path d="M66 80 C56 80 50 84 50 88 L70 86 Z" fill="#B62324" />
		<rect x="66" y="78" width="68" height="7" rx="3" fill="#B62324" />
		<path d="M78 92 H96 L94 101 C93 104 82 104 80 101 Z M104 92 H122 L120 101 C119 104 108 104 106 101 Z" fill="#0D1117" />
		<path d="M96 94 H104" stroke="#0D1117" stroke-width="3" />
		<path d="M82 94 L86 94" stroke="#58A6FF" stroke-width="2" stroke-linecap="round" />
		<path d="M93 112 Q100 116 108 110" fill="none" stroke="#1B1B1B" stroke-width="3" stroke-linecap="round" />
		<rect x="40" y="160" width="84" height="54" rx="7" fill="#21262D" />
		<rect x="58" y="176" width="44" height="18" rx="9" fill="#F0883E" />
		<text x="80" y="189" text-anchor="middle" font-family="JetBrains Mono, monospace" font-size="11" font-weight="700" fill="#2A1708">R$/h</text>
		<rect x="136" y="164" width="26" height="30" rx="4" fill="#F0F6FC" />
		<rect x="136" y="164" width="26" height="7" fill="#8B949E" />
		<path d="M162 172 C173 172 173 187 162 187" fill="none" stroke="#F0F6FC" stroke-width="4" />
		<path d="M144 156 C141 150 147 148 144 142 M154 156 C151 150 157 148 154 142" fill="none" stroke="#8B5A2B" stroke-width="2.5" stroke-linecap="round" />
		<circle cx="40" cy="178" r="9" fill="#6B4226" />
		<circle cx="132" cy="184" r="9" fill="#6B4226" />
	</g>
</svg>
