<script lang="ts">
	// The Devaneio poster used by the billboard and the game page: three cards fanned out.
	import { cardSvg } from './art';

	let { height = 290 }: { height?: number } = $props();
	const id = $props.id();
	const picks = [7, 23, 41];
	const xs = [88, 160, 232];
	const ys = [196, 176, 196];
	const tilt = [-14, 0, 14];
</script>

<svg viewBox="0 0 358 290" width="100%" {height} preserveAspectRatio="xMidYMid slice" style="display: block" role="img" aria-label="Arte do Devaneio: três cartas desenhadas abertas em leque">
	<defs>
		<linearGradient id="g{id}" x1="0" y1="0" x2="0" y2="1">
			<stop offset="0" stop-color="#08222b" />
			<stop offset="1" stop-color="#1c5a66" />
		</linearGradient>
		<linearGradient id="f{id}" x1="0" y1="0" x2="0" y2="1">
			<stop offset="0.7" stop-color="#06161c" stop-opacity="0" />
			<stop offset="1" stop-color="#06161c" stop-opacity="1" />
		</linearGradient>
	</defs>
	<rect width="358" height="290" fill="url(#g{id})" />
	{#each picks as c, i (c)}
		<g transform="translate({xs[i]} {ys[i]}) rotate({tilt[i]}) translate(-52 -150)">
			<clipPath id="c{id}{i}"><rect width="104" height="156" rx="9" /></clipPath>
			<g clip-path="url(#c{id}{i})"><g transform="scale(0.52)">{@html cardSvg(c)}</g></g>
			<rect width="104" height="156" rx="9" fill="none" stroke="#fff1d6" stroke-width="3" />
		</g>
	{/each}
	<rect width="358" height="290" fill="url(#f{id})" />
</svg>
