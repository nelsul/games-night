<script lang="ts">
	import { page } from '$app/state';
	import { session } from '$lib/session.svelte';
	import Icon from './Icon.svelte';
	import Avatar from './Avatar.svelte';

	const path = $derived(page.url.pathname);
</script>

<nav aria-label="Navegação">
	<a href="/" aria-current={path === '/' ? 'page' : undefined}>
		<Icon name="home" />Início
	</a>
	<a href="/salas" aria-current={path === '/salas' ? 'page' : undefined}>
		<Icon name="rooms" />Salas
	</a>
	<a href="/perfil" aria-current={path === '/perfil' ? 'page' : undefined}>
		{#if session.me}<Avatar user={session.me} size={24} />{/if}Perfil
	</a>
</nav>

<style>
	nav {
		position: fixed;
		left: 0;
		right: 0;
		bottom: 0;
		z-index: 20;
		background: rgba(20, 20, 20, 0.96);
		backdrop-filter: blur(8px);
		-webkit-backdrop-filter: blur(8px);
		border-top: 1px solid #262626;
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		padding-bottom: var(--safe-bottom);
	}
	a {
		height: 62px;
		color: var(--muted);
		text-decoration: none;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 4px;
		font-size: 11px;
		font-weight: 500;
	}
	a[aria-current='page'] {
		color: #fff;
		font-weight: 600;
	}
</style>
