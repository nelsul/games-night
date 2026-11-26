<script lang="ts">
	import { page } from '$app/state';
	import { inbox } from '$lib/inbox.svelte';
	import { session } from '$lib/session.svelte';
	import Icon from './Icon.svelte';
	import Avatar from './Avatar.svelte';

	const path = $derived(page.url.pathname);
</script>

<nav aria-label="Navegação">
	<a href="/" aria-current={path === '/' ? 'page' : undefined}>
		<span class="ico"><Icon name="home" />{#if inbox.invites.length}<span class="dot" aria-hidden="true"></span>{/if}</span>
		Início{#if inbox.invites.length}<span class="sr-only">, {inbox.invites.length} convites</span>{/if}
	</a>
	<a href="/salas" aria-current={path === '/salas' ? 'page' : undefined}>
		<Icon name="rooms" />Salas
	</a>
	<a href="/amigos" aria-current={path === '/amigos' ? 'page' : undefined}>
		<span class="ico"
			><Icon name="friends" />{#if inbox.requests}<span class="count" aria-hidden="true">{inbox.requests}</span>{/if}</span
		>
		Amigos{#if inbox.requests}<span class="sr-only">, {inbox.requests} pedidos</span>{/if}
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
		grid-template-columns: repeat(4, minmax(0, 1fr));
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
	.ico {
		position: relative;
		display: flex;
	}
	.count {
		position: absolute;
		top: -4px;
		right: -10px;
		min-width: 18px;
		height: 18px;
		padding: 0 5px;
		border-radius: 9px;
		background: var(--brand);
		color: #fff;
		font-size: 11px;
		font-weight: 700;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.dot {
		position: absolute;
		top: -2px;
		right: -4px;
		width: 10px;
		height: 10px;
		border-radius: 50%;
		background: var(--brand);
		border: 2px solid #141414;
	}
</style>
