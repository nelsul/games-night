<script lang="ts">
	// The picture on each card, in a 100×100 box.
	import Capy from './Capy.svelte';
	let { kind, size = 72 }: { kind: string; size?: number } = $props();
	const id = $props.id();
</script>

<svg viewBox="0 0 100 100" width={size} height={size} aria-hidden="true" style="display: block">
	{#if kind === 'jaguar'}
		<defs>
			<radialGradient id="j{id}" cx="50%" cy="45%" r="60%">
				<stop offset="0" stop-color="#f4b13d" />
				<stop offset="1" stop-color="#c97a16" />
			</radialGradient>
		</defs>
		<circle cx="24" cy="22" r="11" fill="#c97a16" />
		<circle cx="76" cy="22" r="11" fill="#c97a16" />
		<circle cx="24" cy="22" r="5" fill="#5a3208" />
		<circle cx="76" cy="22" r="5" fill="#5a3208" />
		<ellipse cx="50" cy="54" rx="40" ry="38" fill="url(#j{id})" />
		{#each [[30, 34], [68, 32], [22, 58], [78, 58], [40, 22], [60, 20], [50, 84]] as [cx, cy] (cx + '-' + cy)}
			<circle {cx} {cy} r="5" fill="none" stroke="#3b2208" stroke-width="2.6" />
		{/each}
		<ellipse cx="50" cy="70" rx="18" ry="13" fill="#fff3dc" />
		<path d="M35 44l10 4-10 4z" fill="#ffe14d" stroke="#3b2208" stroke-width="1.5" transform="rotate(-8 40 48)" />
		<path d="M65 44l-10 4 10 4z" fill="#ffe14d" stroke="#3b2208" stroke-width="1.5" transform="rotate(8 60 48)" />
		<path d="M44 62h12l-6 6z" fill="#3b2208" />
		<path d="M42 74l3 6 3-6M52 74l3 6 3-6" fill="#fff" stroke="#3b2208" stroke-width="1.2" />
	{:else if kind === 'nope'}
		<Capy />
		<g transform="translate(58 46)">
			<rect x="0" y="0" width="34" height="34" rx="8" fill="#e5383b" stroke="#fff" stroke-width="3" />
			<path d="M10 10l14 14M24 10L10 24" stroke="#fff" stroke-width="5" stroke-linecap="round" />
		</g>
	{:else if kind === 'current'}
		<rect width="100" height="100" fill="#1f7fb3" rx="8" />
		{#each [24, 46, 68] as yy (yy)}
			<path d="M6 {yy}q11-9 22 0t22 0t22 0t22 0" fill="none" stroke="#bfeaff" stroke-width="5" stroke-linecap="round" />
		{/each}
		<path d="M68 84h22l-8-8M90 84l-8 8" fill="none" stroke="#fff" stroke-width="5" stroke-linecap="round" stroke-linejoin="round" />
		<Capy x={8} y={52} s={0.42} />
	{:else if kind === 'swirl'}
		<rect width="100" height="100" fill="#2b5f8a" rx="8" />
		<path d="M50 50m0-6a6 6 0 1 1-6 6a12 12 0 0 1 12-12a18 18 0 0 1 18 18a24 24 0 0 1-24 24a30 30 0 0 1-30-30a36 36 0 0 1 36-36" fill="none" stroke="#bfeaff" stroke-width="5" stroke-linecap="round" />
	{:else if kind === 'cuddle'}
		<Capy x={-8} y={14} s={0.7} />
		<Capy x={38} y={14} s={0.7} />
		<path d="M50 18c-4-8-16-6-14 3 2 7 14 13 14 13s12-6 14-13c2-9-10-11-14-3z" fill="#ff5d8f" />
	{:else}
		<Capy wear={kind} />
	{/if}
</svg>
