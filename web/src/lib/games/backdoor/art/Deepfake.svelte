<script lang="ts">
	let { size = 200 }: { size?: number } = $props();
	const id = $props.id();
</script>

<svg viewBox="0 0 200 200" width={size} height={size} role="img" aria-label="Deepfake: um rosto que é metade pessoa, metade malha digital, copiando o crachá do Analista">
	<defs>
		<clipPath id="c{id}"><circle cx="100" cy="100" r="96" /></clipPath>
		<clipPath id="r{id}"><rect x="100" y="0" width="100" height="200" /></clipPath>
		<linearGradient id="g{id}" x1="0" y1="0" x2="1" y2="1">
			<stop offset="0" stop-color="#2A0F3D" />
			<stop offset="1" stop-color="#12061C" />
		</linearGradient>
	</defs>
	<circle cx="100" cy="100" r="96" fill="url(#g{id})" />
	<g clip-path="url(#c{id})">
		<!-- scan lines -->
		<g fill="#C77DFF" opacity="0.12">
			<rect x="0" y="36" width="200" height="2" /><rect x="0" y="58" width="200" height="3" /><rect x="0" y="126" width="200" height="2" /><rect x="0" y="154" width="200" height="3" />
		</g>
		<text x="20" y="48" font-family="JetBrains Mono, monospace" font-size="11" font-weight="700" fill="#5A2D82">model.fit()</text>
		<text x="132" y="140" font-family="JetBrains Mono, monospace" font-size="11" font-weight="700" fill="#5A2D82">99.8%</text>
		<!-- body: a copy of the Analyst's shirt and badge -->
		<path d="M26 212 C26 160 58 140 100 140 C142 140 174 160 174 212 Z" fill="#2F81F7" />
		<rect x="58" y="160" width="30" height="20" rx="4" fill="#F0F6FC" />
		<rect x="62" y="165" width="14" height="3" fill="#2F81F7" />
		<rect x="62" y="171" width="20" height="3" fill="#8B949E" />
		<rect x="89" y="112" width="22" height="30" rx="8" fill="#C99369" />
		<!-- the real half of the face -->
		<circle cx="100" cy="88" r="32" fill="#C99369" />
		<path d="M68 82 C66 56 90 48 112 54 C128 58 136 70 132 84 C118 68 94 64 70 76 Z" fill="#3B2A1A" />
		<circle cx="88" cy="88" r="3.5" fill="#1B1B1B" />
		<path d="M82 80 Q88 77 94 80" fill="none" stroke="#3B2A1A" stroke-width="2.5" stroke-linecap="round" />
		<path d="M88 106 Q94 110 100 108" fill="none" stroke="#1B1B1B" stroke-width="3" stroke-linecap="round" />
		<!-- the generated half: a wireframe mesh that doesn't quite match -->
		<g clip-path="url(#r{id})">
			<circle cx="100" cy="88" r="32" fill="#1A0B28" />
			<g fill="none" stroke="#C77DFF" stroke-width="1.4" opacity="0.9">
				<path d="M100 56 Q122 60 132 88 Q122 116 100 120" />
				<path d="M100 66 Q116 70 122 88 Q116 106 100 110" />
				<path d="M100 76 Q110 80 112 88 Q110 96 100 100" />
				<path d="M100 56 L100 120 M100 72 L130 76 M100 88 L132 88 M100 104 L128 100" />
			</g>
			<circle cx="112" cy="88" r="4" fill="#FF4DD2" />
			<circle cx="112" cy="88" r="8" fill="none" stroke="#FF4DD2" stroke-width="1.5" opacity="0.6" />
			<path d="M100 108 Q108 112 114 106" fill="none" stroke="#FF4DD2" stroke-width="3" stroke-linecap="round" />
		</g>
		<!-- the seam, glitching -->
		<rect x="98" y="52" width="4" height="72" fill="#FF4DD2" opacity="0.8" />
		<rect x="86" y="94" width="30" height="4" fill="#C77DFF" />
		<rect x="104" y="70" width="22" height="3" fill="#F0F6FC" opacity="0.7" />
		<rect x="120" y="172" width="44" height="3" fill="#FF4DD2" />
	</g>
</svg>
