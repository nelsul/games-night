<script lang="ts">
	// The Backdoor poster used by the hub's billboard and the game page.
	let { height = 290 }: { height?: number } = $props();
	const id = $props.id();
</script>

<svg viewBox="0 0 358 290" width="100%" {height} preserveAspectRatio="xMidYMid slice" style="display: block" role="img" aria-label="Arte do Backdoor: hacker de capuz em frente a linhas de código">
	<defs>
		<linearGradient id="f{id}" x1="0" y1="0" x2="0" y2="1">
			<stop offset="0.55" stop-color="#1B0B0E" stop-opacity="0" />
			<stop offset="1" stop-color="#1B0B0E" stop-opacity="1" />
		</linearGradient>
	</defs>
	<rect width="358" height="290" fill="#1B0B0E" />
	<g font-family="JetBrains Mono, monospace" font-size="11" fill="#4A1A1E">
		<text x="12" y="28">if (user.isDev) {'{'} deploy(); {'}'}</text>
		<text x="190" y="52">git push --force</text>
		<text x="12" y="76">0110 1001 0011</text>
		<text x="230" y="100">sudo rm -rf /</text>
		<text x="12" y="124">ssh root@samgung</text>
		<text x="236" y="148">1011 0101</text>
		<text x="12" y="172">// TODO: remover backdoor</text>
	</g>
	<g transform="translate(79 30)">
		<path d="M24 262 C24 190 56 160 100 160 C144 160 176 190 176 262 Z" fill="#2A2F36" />
		<path d="M46 168 C34 92 64 40 100 40 C136 40 166 92 154 168 Z" fill="#3D444D" />
		<ellipse cx="100" cy="110" rx="34" ry="42" fill="#07090C" />
		<ellipse cx="86" cy="110" rx="8" ry="4.5" fill="#FF4545" />
		<ellipse cx="114" cy="110" rx="8" ry="4.5" fill="#FF4545" />
		<path d="M88 166 L84 196 M112 166 L116 196" stroke="#8C8C8C" stroke-width="3" stroke-linecap="round" />
	</g>
	<rect width="358" height="290" fill="url(#f{id})" />
</svg>
