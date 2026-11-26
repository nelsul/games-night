<script lang="ts">
	let { size = 200 }: { size?: number } = $props();
	const id = $props.id();
</script>

<svg viewBox="0 0 200 200" width={size} height={size} role="img" aria-label="Analista de Segurança: pessoa de óculos com crachá de escudo">
	<defs><clipPath id="c{id}"><circle cx="100" cy="100" r="96" /></clipPath></defs>
	<circle cx="100" cy="100" r="96" fill="#0F1E33" />
	<g clip-path="url(#c{id})">
		<path d="M100 18 L158 38 L158 92 C158 128 132 152 100 166 C68 152 42 128 42 92 L42 38 Z" fill="#15294A" />
		<path d="M26 212 C26 160 58 138 100 138 C142 138 174 160 174 212 Z" fill="#4493F8" />
		<path d="M82 139 L100 160 L118 139 Z" fill="#E6EDF3" />
		<rect x="89" y="112" width="22" height="30" rx="8" fill="#8D5524" />
		<circle cx="100" cy="90" r="32" fill="#8D5524" />
		<path d="M68 86 C66 60 84 54 100 54 C118 54 134 60 132 86 C126 74 112 70 100 70 C86 70 74 74 68 86 Z" fill="#111111" />
		<rect x="79" y="85" width="18" height="14" rx="5" fill="#1C2A40" stroke="#E6EDF3" stroke-width="3" />
		<rect x="103" y="85" width="18" height="14" rx="5" fill="#1C2A40" stroke="#E6EDF3" stroke-width="3" />
		<path d="M97 91 L103 91" stroke="#E6EDF3" stroke-width="3" />
		<circle cx="88" cy="92" r="2.5" fill="#79B8FF" />
		<circle cx="112" cy="92" r="2.5" fill="#79B8FF" />
		<path d="M92 108 L108 108" stroke="#1B1B1B" stroke-width="3" stroke-linecap="round" />
		<path d="M84 142 L100 172 L116 142" fill="none" stroke="#0D1117" stroke-width="3" />
		<rect x="87" y="168" width="26" height="32" rx="5" fill="#E6EDF3" />
		<path d="M100 174 L108 177 L108 184 C108 189 104 192 100 194 C96 192 92 189 92 184 L92 177 Z" fill="#4493F8" />
	</g>
</svg>
