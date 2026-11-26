<script lang="ts">
	let { size = 200 }: { size?: number } = $props();
	const id = $props.id();
</script>

<svg viewBox="0 0 200 200" width={size} height={size} role="img" aria-label="Hacker: figura de capuz com olhos vermelhos e notebook">
	<defs><clipPath id="c{id}"><circle cx="100" cy="100" r="96" /></clipPath></defs>
	<circle cx="100" cy="100" r="96" fill="#2A1215" />
	<g clip-path="url(#c{id})">
		<g font-family="JetBrains Mono, monospace" font-size="13" font-weight="700" fill="#5A2025">
			<text x="22" y="56">0110</text><text x="138" y="44">1001</text><text x="150" y="118">01</text><text x="20" y="128">10</text>
		</g>
		<path d="M24 212 C24 158 56 132 100 132 C144 132 176 158 176 212 Z" fill="#30363D" />
		<path d="M58 130 C48 74 70 40 100 40 C130 40 152 74 142 130 Z" fill="#3D444D" />
		<ellipse cx="100" cy="94" rx="27" ry="32" fill="#07090C" />
		<ellipse cx="89" cy="94" rx="6" ry="3.5" fill="#F85149" />
		<ellipse cx="111" cy="94" rx="6" ry="3.5" fill="#F85149" />
		<path d="M90 130 L87 152 M110 130 L113 152" stroke="#9198A1" stroke-width="3" stroke-linecap="round" />
		<rect x="56" y="156" width="88" height="58" rx="7" fill="#161B22" stroke="#F85149" stroke-width="2.5" />
		<text x="100" y="190" text-anchor="middle" font-family="JetBrains Mono, monospace" font-size="20" font-weight="700" fill="#F85149">&lt;/&gt;</text>
		<circle cx="56" cy="176" r="9" fill="#21262D" />
		<circle cx="144" cy="176" r="9" fill="#21262D" />
	</g>
</svg>
