<script lang="ts">
	// A person's square tile: their photo, or their initial on a color picked from their id.
	let {
		user,
		size = 40,
		radius = 4
	}: { user: { id: string; name: string; avatar_url: string }; size?: number; radius?: number } = $props();

	const colors = ['#2F80ED', '#7B4FD1', '#E8A317', '#2E9E5B', '#D94F8A', '#1FA3A3', '#C25B2E', '#5B6CF0'];
	const color = $derived(colors[[...user.id].reduce((n, c) => n + c.charCodeAt(0), 0) % colors.length]);
	const dark = $derived(color === '#E8A317');
	const initial = $derived((user.name || '?').trim().charAt(0).toUpperCase());
</script>

{#if user.avatar_url}
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
