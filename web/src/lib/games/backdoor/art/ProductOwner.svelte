<script lang="ts">
	let { size = 200 }: { size?: number } = $props();
	const id = $props.id();
</script>

<svg viewBox="0 0 200 200" width={size} height={size} role="img" aria-label="Product Owner: pessoa de cabelo black power e óculos redondos segurando um post-it escrito MVP, com um roadmap ao fundo">
	<defs><clipPath id="c{id}"><circle cx="100" cy="100" r="96" /></clipPath></defs>
	<circle cx="100" cy="100" r="96" fill="#3A2A05" />
	<g clip-path="url(#c{id})">
		<path d="M14 120 C40 96 60 128 86 104" fill="none" stroke="#6B5212" stroke-width="3" stroke-dasharray="2 7" stroke-linecap="round" />
		<path d="M118 40 C140 30 160 52 186 36" fill="none" stroke="#6B5212" stroke-width="3" stroke-dasharray="2 7" stroke-linecap="round" />
		<path d="M24 118 V96 L36 100 L24 104" fill="#F2CC60" />
		<path d="M172 46 V24 L184 28 L172 32" fill="#F2CC60" />
		<path d="M26 212 C26 160 58 140 100 140 C142 140 174 160 174 212 Z" fill="#D29922" />
		<path d="M84 141 L100 172 L116 141 Z" fill="#F3EAD7" />
		<path d="M84 141 L72 172 L92 158 Z M116 141 L128 172 L108 158 Z" fill="#B07F12" />
		<rect x="89" y="114" width="22" height="30" rx="8" fill="#B5793F" />
		<circle cx="100" cy="74" r="44" fill="#1F120A" />
		<circle cx="66" cy="84" r="16" fill="#1F120A" />
		<circle cx="134" cy="84" r="16" fill="#1F120A" />
		<circle cx="100" cy="94" r="28" fill="#B5793F" />
		<circle cx="89" cy="93" r="8" fill="none" stroke="#F2CC60" stroke-width="3" />
		<circle cx="111" cy="93" r="8" fill="none" stroke="#F2CC60" stroke-width="3" />
		<path d="M97 93 H103" stroke="#F2CC60" stroke-width="3" />
		<circle cx="89" cy="94" r="2.5" fill="#1B1B1B" />
		<circle cx="111" cy="94" r="2.5" fill="#1B1B1B" />
		<path d="M91 108 Q100 115 109 108" fill="none" stroke="#1B1B1B" stroke-width="3" stroke-linecap="round" />
		<g transform="rotate(-8 132 168)">
			<rect x="104" y="140" width="58" height="54" rx="3" fill="#F2CC60" />
			<rect x="104" y="140" width="58" height="10" rx="3" fill="#E3B341" />
			<text x="133" y="181" text-anchor="middle" font-family="Inter, Arial, sans-serif" font-size="19" font-weight="800" fill="#3A2A05">MVP</text>
		</g>
		<circle cx="108" cy="178" r="9" fill="#B5793F" />
		<circle cx="54" cy="182" r="9" fill="#B5793F" />
	</g>
</svg>
