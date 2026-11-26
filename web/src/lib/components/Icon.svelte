<script lang="ts">
	// Stroke icons used across the hub. Decorative unless a label is given.
	type Name =
		| 'home'
		| 'rooms'
		| 'search'
		| 'play'
		| 'plus'
		| 'back'
		| 'close'
		| 'qr'
		| 'share'
		| 'book'
		| 'logout'
		| 'camera'
		| 'check'
		| 'door'
		| 'friends'
		| 'userplus'
		| 'trash'
		| 'rotate'
		| 'flip'
		| 'chat'
		| 'send'
		| 'sound'
		| 'mute'
		| 'smile';
	let { name, size = 24, label }: { name: Name; size?: number; label?: string } = $props();
</script>

<svg
	width={size}
	height={size}
	viewBox="0 0 24 24"
	fill={name === 'play' || name === 'home' ? 'currentColor' : 'none'}
	stroke={name === 'play' || name === 'home' ? 'none' : 'currentColor'}
	stroke-width="2"
	stroke-linecap="round"
	stroke-linejoin="round"
	role={label ? 'img' : undefined}
	aria-label={label}
	aria-hidden={label ? undefined : 'true'}
>
	{#if name === 'home'}<path d="M3 11l9-8 9 8v10h-6v-6H9v6H3z" />
	{:else if name === 'rooms'}<circle cx="9" cy="8" r="3.5" /><path d="M2.5 20c0-3.6 2.9-6 6.5-6s6.5 2.4 6.5 6" /><circle cx="17" cy="9" r="2.5" /><path d="M17 14c2.8 0 4.5 2 4.5 5" />
	{:else if name === 'search'}<circle cx="11" cy="11" r="7" /><path d="M20 20l-3.5-3.5" />
	{:else if name === 'play'}<path d="M7 4l13 8-13 8z" />
	{:else if name === 'plus'}<path d="M12 5v14M5 12h14" />
	{:else if name === 'back'}<path d="M15 18l-6-6 6-6" />
	{:else if name === 'close'}<path d="M6 6l12 12M18 6L6 18" />
	{:else if name === 'qr'}<rect x="3" y="3" width="7" height="7" /><rect x="14" y="3" width="7" height="7" /><rect x="3" y="14" width="7" height="7" /><path d="M14 14h3v3h-3zM18 18h3v3" />
	{:else if name === 'share'}<path d="M4 12v7a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-7M16 6l-4-4-4 4M12 2v13" />
	{:else if name === 'book'}<path d="M4 19V5a2 2 0 0 1 2-2h12v16H6a2 2 0 0 0-2 2zM8 7h6" />
	{:else if name === 'logout'}<path d="M15 4h3a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-3M10 17l-5-5 5-5M5 12h11" />
	{:else if name === 'camera'}<path d="M4 8h3l2-3h6l2 3h3v11H4z" /><circle cx="12" cy="13" r="3.5" />
	{:else if name === 'check'}<path d="M5 12l5 5L20 7" />
	{:else if name === 'friends'}<circle cx="12" cy="8" r="4" /><path d="M4 21c0-4.4 3.6-7 8-7s8 2.6 8 7" />
	{:else if name === 'userplus'}<circle cx="10" cy="8" r="4" /><path d="M3 21c0-4 3.1-7 7-7 1.8 0 3.4.6 4.6 1.6M19 14v6M16 17h6" />
	{:else if name === 'trash'}<path d="M4 7h16M10 11v6M14 11v6M6 7l1 13h10l1-13M9 7V4h6v3" />
	{:else if name === 'rotate'}<path d="M20 11a8 8 0 1 0-2.3 5.7M20 4v7h-7" />
	{:else if name === 'flip'}<path d="M12 3v18M8 7l-5 10h5zM16 7l5 10h-5z" />
	{:else if name === 'chat'}<path d="M4 5h16v11H9l-5 4z" />
	{:else if name === 'send'}<path d="M4 12l16-8-6 16-3-7z" />
	{:else if name === 'sound'}<path d="M4 9h4l5-4v14l-5-4H4zM16 9a4 4 0 0 1 0 6M18.5 6.5a7.5 7.5 0 0 1 0 11" />
	{:else if name === 'mute'}<path d="M4 9h4l5-4v14l-5-4H4zM16 9l5 6M21 9l-5 6" />
	{:else if name === 'smile'}<circle cx="12" cy="12" r="9" /><path d="M8 14c1 1.5 2.4 2.2 4 2.2s3-.7 4-2.2M9 9.5h.01M15 9.5h.01" />
	{:else if name === 'door'}<path d="M6 21V4a1 1 0 0 1 1-1h10a1 1 0 0 1 1 1v17M3 21h18M14 12h.01" />
	{/if}
</svg>
