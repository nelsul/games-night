<script lang="ts">
	// A person's square tile: their photo, or their initial on a color picked from their id.
	let {
		user,
		size = 40,
		radius = 4,
		online
	}: {
		user: { id: string; name: string; avatar_url: string; is_bot?: boolean };
		size?: number;
		radius?: number;
		online?: boolean; // when set, shows a presence dot
	} = $props();

	const colors = ['#2F80ED', '#7B4FD1', '#E8A317', '#2E9E5B', '#D94F8A', '#1FA3A3', '#C25B2E', '#5B6CF0'];
	const color = $derived(colors[[...user.id].reduce((n, c) => n + c.charCodeAt(0), 0) % colors.length]);
	const dark = $derived(color === '#E8A317');
	const initial = $derived((user.name || '?').trim().charAt(0).toUpperCase());
</script>

<span class="av" style="width: {size}px; height: {size}px">
{#if user.is_bot}
	<svg
		viewBox="0 0 48 48"
		width={size}
		height={size}
		aria-hidden="true"
		style="border-radius: {radius}px; flex-shrink: 0; display: block; background: {color}"
	>
		<path d="M24 7v5" stroke="#fff" stroke-width="2.5" stroke-linecap="round" />
		<circle cx="24" cy="6" r="2.5" fill="#fff" />
		<rect x="10" y="13" width="28" height="22" rx="7" fill="#fff" />
		<rect x="15" y="19" width="18" height="9" rx="4.5" fill="#141414" />
		<circle cx="20" cy="23.5" r="2" fill={color} />
		<circle cx="28" cy="23.5" r="2" fill={color} />
		<path d="M6 22v6M42 22v6" stroke="#fff" stroke-width="3" stroke-linecap="round" />
		<path d="M18 39h12" stroke="#fff" stroke-width="3" stroke-linecap="round" />
	</svg>
{:else if user.avatar_url}
	<img
		src={user.avatar_url}
		alt=""
		width={size}
		height={size}
		style="width: {size}px; height: {size}px; border-radius: {radius}px; object-fit: cover; display: block; flex-shrink: 0"
	/>
{:else}
	<span
		aria-hidden="true"
		style="width: {size}px; height: {size}px; border-radius: {radius}px; background: {color}; color: {dark
			? '#141414'
			: '#fff'}; display: inline-flex; align-items: center; justify-content: center; font-weight: 700; font-size: {Math.round(
			size * 0.4
		)}px; flex-shrink: 0">{initial}</span
	>
{/if}
{#if online !== undefined}
	<span class="dot" class:on={online} style="width: {Math.max(8, Math.round(size * 0.26))}px; height: {Math.max(8, Math.round(size * 0.26))}px"
		><span class="sr-only">{online ? 'online' : 'offline'}</span></span
	>
{/if}
</span>

<style>
	.av {
		position: relative;
		display: inline-flex;
		flex-shrink: 0;
	}
	.dot {
		position: absolute;
		right: -2px;
		bottom: -2px;
		border-radius: 50%;
		background: #6e7681;
		border: 2px solid #141414;
		box-sizing: content-box;
	}
	.dot.on {
		background: #3fb950;
		animation: gn-glow 2.4s ease-in-out infinite;
	}
</style>
