<script lang="ts">
	// The Crazy Capybaras poster: capybaras chilling in a lagoon at sunset while a pair of
	// jaguar eyes watches from the bushes.
	import Capy from './art/Capy.svelte';
	let { height = 290 }: { height?: number } = $props();
	const id = $props.id();
</script>

<svg viewBox="0 0 358 290" width="100%" {height} preserveAspectRatio="xMidYMid slice" style="display: block" role="img" aria-label="Arte do Crazy Capybaras: capivaras numa lagoa e olhos de onça no mato">
	<defs>
		<linearGradient id="sky{id}" x1="0" y1="0" x2="0" y2="1">
			<stop offset="0" stop-color="#ff8a3d" />
			<stop offset="0.55" stop-color="#ffbf69" />
			<stop offset="1" stop-color="#ffd89b" />
		</linearGradient>
		<linearGradient id="fade{id}" x1="0" y1="0" x2="0" y2="1">
			<stop offset="0.6" stop-color="#0f2a1f" stop-opacity="0" />
			<stop offset="1" stop-color="#0f2a1f" stop-opacity="1" />
		</linearGradient>
	</defs>
	<rect width="358" height="290" fill="url(#sky{id})" />
	<circle cx="270" cy="92" r="40" fill="#fff1c1" opacity="0.9" />
	<!-- bushes with the jaguar -->
	<path d="M0 150q30-40 70-20q30-30 70 0q20-20 50 0v40H0z" fill="#1f5c40" />
	<path d="M210 150q30-36 70-16q30-24 78 4v32H210z" fill="#1f5c40" />
	<ellipse cx="322" cy="150" rx="6" ry="3.5" fill="#ffe14d" />
	<ellipse cx="340" cy="150" rx="6" ry="3.5" fill="#ffe14d" />
	<ellipse cx="322" cy="150" rx="1.6" ry="3" fill="#1a1a1a" />
	<ellipse cx="340" cy="150" rx="1.6" ry="3" fill="#1a1a1a" />
	<!-- lagoon -->
	<rect y="168" width="358" height="122" fill="#2a8fb8" />
	<path d="M0 172q20-8 40 0t40 0t40 0t40 0t40 0t40 0t40 0t40 0t40 0" fill="none" stroke="#bfeaff" stroke-width="3" opacity="0.8" />
	<Capy wear="orange" x={34} y={128} s={0.95} />
	<Capy wear="shades" x={140} y={138} s={0.82} />
	<Capy wear="spa" x={232} y={142} s={0.8} />
	<rect width="358" height="290" fill="url(#fade{id})" />
</svg>
