<script lang="ts">
	let { size = 200 }: { size?: number } = $props();
	const id = $props.id();
</script>

<svg viewBox="0 0 200 200" width={size} height={size} role="img" aria-label="Freelancer: hacker de gorro e fone, sozinho com um café e um notebook">
	<defs><clipPath id="c{id}"><circle cx="100" cy="100" r="96" /></clipPath></defs>
	<circle cx="100" cy="100" r="96" fill="#2A1708" />
	<g clip-path="url(#c{id})">
		<path d="M150 58 C146 50 154 46 150 38 M162 58 C158 50 166 46 162 38" fill="none" stroke="#6B4A2E" stroke-width="3" stroke-linecap="round" />
		<path d="M24 212 C24 158 56 134 100 134 C144 134 176 158 176 212 Z" fill="#DB6D28" />
		<path d="M70 140 C74 156 126 156 130 140" fill="none" stroke="#B4521A" stroke-width="5" />
		<rect x="89" y="112" width="22" height="28" rx="8" fill="#A0673F" />
		<circle cx="100" cy="92" r="31" fill="#A0673F" />
		<path d="M66 88 C64 54 136 54 134 88 Z" fill="#30363D" />
		<rect x="64" y="80" width="72" height="12" rx="6" fill="#484F58" />
		<circle cx="100" cy="52" r="7" fill="#484F58" />
		<path d="M89 101 L95 101 M105 101 L111 101" stroke="#1B1B1B" stroke-width="3.5" stroke-linecap="round" />
		<path d="M93 113 L107 112" stroke="#1B1B1B" stroke-width="3" stroke-linecap="round" />
		<path d="M70 98 C68 110 72 118 78 122" fill="none" stroke="#E6EDF3" stroke-width="2.5" />
		<circle cx="69" cy="98" r="4" fill="#E6EDF3" />
		<rect x="40" y="158" width="88" height="56" rx="7" fill="#161B22" stroke="#F0883E" stroke-width="2.5" />
		<text x="84" y="191" text-anchor="middle" font-family="JetBrains Mono, monospace" font-size="15" font-weight="700" fill="#F0883E">$ solo</text>
		<rect x="138" y="166" width="24" height="28" rx="4" fill="#E6EDF3" />
		<path d="M162 172 C172 172 172 186 162 186" fill="none" stroke="#E6EDF3" stroke-width="4" />
		<rect x="138" y="166" width="24" height="6" fill="#C9D1D9" />
		<circle cx="40" cy="178" r="9" fill="#A0673F" />
		<circle cx="134" cy="182" r="9" fill="#A0673F" />
	</g>
</svg>
