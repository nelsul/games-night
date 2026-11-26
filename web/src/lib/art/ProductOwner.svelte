<script lang="ts">
	let { size = 200 }: { size?: number } = $props();
	const id = $props.id();
</script>

<svg viewBox="0 0 200 200" width={size} height={size} role="img" aria-label="Product Owner: pessoa segurando um quadro de tarefas com post-its">
	<defs><clipPath id="c{id}"><circle cx="100" cy="100" r="96" /></clipPath></defs>
	<circle cx="100" cy="100" r="96" fill="#2B2111" />
	<g clip-path="url(#c{id})">
		<rect x="22" y="34" width="14" height="14" rx="2" fill="#4A3A1A" transform="rotate(-8 29 41)" />
		<rect x="160" y="40" width="14" height="14" rx="2" fill="#4A3A1A" transform="rotate(10 167 47)" />
		<rect x="150" y="104" width="12" height="12" rx="2" fill="#4A3A1A" transform="rotate(-6 156 110)" />
		<path d="M26 212 C26 160 58 138 100 138 C142 138 174 160 174 212 Z" fill="#D29922" />
		<path d="M86 139 L100 156 L114 139 Z" fill="#E6EDF3" />
		<rect x="89" y="112" width="22" height="30" rx="8" fill="#E0AC69" />
		<circle cx="100" cy="90" r="32" fill="#E0AC69" />
		<path d="M66 92 C60 58 88 48 104 52 C124 56 140 70 134 94 C130 80 120 70 104 68 C92 74 78 76 66 92 Z" fill="#3B2314" />
		<path d="M132 78 C146 84 148 104 140 118 C138 104 134 94 130 88 Z" fill="#3B2314" />
		<circle cx="88" cy="94" r="3.5" fill="#1B1B1B" />
		<circle cx="112" cy="94" r="3.5" fill="#1B1B1B" />
		<path d="M82 85 L94 83 M106 83 L118 85" stroke="#3B2314" stroke-width="3" stroke-linecap="round" />
		<path d="M92 106 Q100 111 108 106" fill="none" stroke="#1B1B1B" stroke-width="3" stroke-linecap="round" />
		<rect x="48" y="150" width="104" height="62" rx="6" fill="#F3EAD7" />
		<path d="M83 150 V212 M117 150 V212" stroke="#C9BC9E" stroke-width="2" />
		<rect x="55" y="158" width="21" height="16" rx="2" fill="#F2CC60" />
		<rect x="55" y="178" width="21" height="16" rx="2" fill="#FF9BCE" />
		<rect x="90" y="158" width="21" height="16" rx="2" fill="#79C0FF" />
		<rect x="124" y="158" width="21" height="16" rx="2" fill="#7EE2A8" />
		<rect x="124" y="178" width="21" height="16" rx="2" fill="#7EE2A8" />
		<path d="M128 166 L131 169 L138 162" fill="none" stroke="#0D1117" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
		<circle cx="48" cy="170" r="9" fill="#E0AC69" />
		<circle cx="152" cy="170" r="9" fill="#E0AC69" />
	</g>
</svg>
