<script lang="ts">
	// A capybara, front-on, in a 100×100 box. The look changes with `wear`: an orange on
	// the head, shades, DJ headphones, a hot tub, a surfboard, asleep, binoculars or diving.
	let {
		wear = '',
		x = 0,
		y = 0,
		s = 1
	}: { wear?: string; x?: number; y?: number; s?: number } = $props();
	const fur = '#b07648';
	const dark = '#8a5631';
	const deep = '#3a2414';
	const sleeping = $derived(wear === 'nap');
</script>

<g transform="translate({x} {y}) scale({s})">
	{#if wear === 'surf'}
		<rect x="60" y="4" width="22" height="92" rx="11" fill="#4fc3d9" transform="rotate(24 71 50)" />
		<path d="M71 8v84" stroke="#fff" stroke-width="3" transform="rotate(24 71 50)" opacity="0.8" />
	{/if}
	{#if wear !== 'dive'}
		<ellipse cx="50" cy="80" rx="34" ry="20" fill={dark} />
	{/if}
	<!-- ears -->
	<circle cx="29" cy="25" r="7" fill={dark} />
	<circle cx="71" cy="25" r="7" fill={dark} />
	<circle cx="29" cy="25" r="3.5" fill={deep} opacity="0.5" />
	<circle cx="71" cy="25" r="3.5" fill={deep} opacity="0.5" />
	<!-- head and snout -->
	<rect x="23" y="20" width="54" height="52" rx="22" fill={fur} />
	<rect x="31" y="44" width="38" height="27" rx="13" fill={dark} />
	<ellipse cx="43" cy="53" rx="3" ry="2.2" fill={deep} />
	<ellipse cx="57" cy="53" rx="3" ry="2.2" fill={deep} />
	<path d="M44 62q6 4 12 0" fill="none" stroke={deep} stroke-width="2" stroke-linecap="round" />
	<!-- eyes -->
	{#if sleeping}
		<path d="M34 37q4 3 8 0M58 37q4 3 8 0" fill="none" stroke={deep} stroke-width="2.4" stroke-linecap="round" />
	{:else if wear !== 'shades' && wear !== 'peek'}
		<circle cx="38" cy="36" r="3.4" fill={deep} />
		<circle cx="62" cy="36" r="3.4" fill={deep} />
		<circle cx="39.2" cy="34.8" r="1.1" fill="#fff" />
		<circle cx="63.2" cy="34.8" r="1.1" fill="#fff" />
	{/if}

	{#if wear === 'orange'}
		<circle cx="50" cy="13" r="11" fill="#ff9f1c" />
		<circle cx="47" cy="10" r="2.5" fill="#ffc46b" />
		<path d="M50 3q6-6 12-2q-6 4-12 2z" fill="#4caf50" />
	{:else if wear === 'shades'}
		<rect x="28" y="30" width="19" height="12" rx="4" fill="#111" />
		<rect x="53" y="30" width="19" height="12" rx="4" fill="#111" />
		<path d="M47 34h6M23 33l5 1M72 34l5-1" stroke="#111" stroke-width="2.5" />
		<path d="M31 33l5 0" stroke="#fff" stroke-width="1.5" opacity="0.7" />
	{:else if wear === 'dj'}
		<path d="M20 40a30 30 0 0 1 60 0" fill="none" stroke="#2b2b2b" stroke-width="6" />
		<rect x="13" y="36" width="13" height="20" rx="5" fill="#e5383b" />
		<rect x="74" y="36" width="13" height="20" rx="5" fill="#e5383b" />
		<text x="86" y="18" font-size="14" fill="#ffd166">♪</text>
	{:else if wear === 'spa'}
		<rect x="2" y="70" width="96" height="30" rx="10" fill="#4fc3d9" opacity="0.9" />
		<path d="M10 76q8-5 16 0t16 0t16 0t16 0t16 0" fill="none" stroke="#e8fbff" stroke-width="2.5" />
		<path d="M30 18q20-14 40 0v6h-40z" fill="#fff" />
		<path d="M30 22h40" stroke="#ffd1dc" stroke-width="3" />
		<circle cx="20" cy="62" r="3" fill="#e8fbff" opacity="0.8" />
		<circle cx="84" cy="58" r="2.4" fill="#e8fbff" opacity="0.8" />
	{:else if wear === 'nap'}
		<text x="70" y="18" font-size="13" font-weight="800" fill="#cfd8ff">Z</text>
		<text x="80" y="9" font-size="9" font-weight="800" fill="#cfd8ff">z</text>
	{:else if wear === 'peek'}
		<rect x="27" y="28" width="18" height="16" rx="5" fill="#2b2b2b" />
		<rect x="55" y="28" width="18" height="16" rx="5" fill="#2b2b2b" />
		<rect x="44" y="32" width="12" height="6" fill="#2b2b2b" />
		<circle cx="36" cy="36" r="5" fill="#6ec6ff" />
		<circle cx="64" cy="36" r="5" fill="#6ec6ff" />
		<circle cx="34" cy="34" r="1.6" fill="#fff" />
		<circle cx="62" cy="34" r="1.6" fill="#fff" />
	{:else if wear === 'dive'}
		<rect x="0" y="50" width="100" height="50" fill="#2a9fd6" />
		<path d="M0 52q8-6 16 0t17 0t17 0t17 0t17 0t16 0" fill="none" stroke="#bfeaff" stroke-width="3" />
		<circle cx="18" cy="70" r="3" fill="#bfeaff" opacity="0.7" />
		<circle cx="82" cy="78" r="4" fill="#bfeaff" opacity="0.7" />
		<circle cx="76" cy="66" r="2" fill="#bfeaff" opacity="0.7" />
	{/if}
</g>
