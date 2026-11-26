<script lang="ts">
	let { size = 200 }: { size?: number } = $props();
	const id = $props.id();
</script>

<svg viewBox="0 0 200 200" width={size} height={size} role="img" aria-label="Dev: pessoa de fone de ouvido com um notebook cheio de adesivos">
	<defs><clipPath id="c{id}"><circle cx="100" cy="100" r="96" /></clipPath></defs>
	<circle cx="100" cy="100" r="96" fill="#12261E" />
	<g clip-path="url(#c{id})">
		<text x="26" y="62" font-family="JetBrains Mono, monospace" font-size="18" font-weight="700" fill="#1F4D35">{'{ }'}</text>
		<text x="136" y="48" font-family="JetBrains Mono, monospace" font-size="16" font-weight="700" fill="#1F4D35">&lt;/&gt;</text>
		<path d="M26 212 C26 160 58 138 100 138 C142 138 174 160 174 212 Z" fill="#3FB950" />
		<rect x="89" y="112" width="22" height="30" rx="8" fill="#C68642" />
		<circle cx="100" cy="90" r="32" fill="#C68642" />
		<circle cx="100" cy="50" r="13" fill="#2B1A10" />
		<path d="M68 90 C64 62 82 56 100 56 C118 56 136 62 132 90 C126 76 112 70 100 70 C88 70 74 76 68 90 Z" fill="#2B1A10" />
		<path d="M64 94 C60 46 140 46 136 94" fill="none" stroke="#E6EDF3" stroke-width="6" stroke-linecap="round" />
		<rect x="56" y="84" width="15" height="26" rx="7" fill="#E6EDF3" />
		<rect x="129" y="84" width="15" height="26" rx="7" fill="#E6EDF3" />
		<circle cx="89" cy="93" r="3.5" fill="#1B1B1B" />
		<circle cx="111" cy="93" r="3.5" fill="#1B1B1B" />
		<path d="M90 104 Q100 113 110 104" fill="none" stroke="#1B1B1B" stroke-width="3" stroke-linecap="round" />
		<rect x="56" y="156" width="88" height="58" rx="7" fill="#D0D7DE" />
		<circle cx="80" cy="178" r="9" fill="#3FB950" />
		<path d="M76 178 L79 181 L85 175" fill="none" stroke="#0D1117" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" />
		<rect x="96" y="171" width="38" height="14" rx="7" fill="#4493F8" />
		<text x="115" y="181.5" text-anchor="middle" font-family="JetBrains Mono, monospace" font-size="9" font-weight="700" fill="#FFFFFF">LGTM</text>
		<circle cx="56" cy="176" r="9" fill="#C68642" />
		<circle cx="144" cy="176" r="9" fill="#C68642" />
	</g>
</svg>
