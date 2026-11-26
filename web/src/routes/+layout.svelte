<script lang="ts">
	import '../app.css';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { loadMe, session } from '$lib/session.svelte';
	import TabBar from '$lib/components/TabBar.svelte';

	let { children } = $props();

	const publicPaths = ['/entrar'];
	const tabPaths = ['/', '/salas', '/perfil'];

	let ready = $state(false);

	$effect(() => {
		const path = page.url.pathname;
		(async () => {
			if (!session.loaded) await loadMe();
			const me = session.me;
			// Where to go after signing in: the page asked for, or the one already carried along.
			const carried = page.url.searchParams.get('next');
			const target = path === '/entrar' || path === '/bem-vindo' ? carried : path === '/' ? null : path + page.url.search;
			const next = target ? `?next=${encodeURIComponent(target)}` : '';
			if (!me && !publicPaths.includes(path)) {
				await goto('/entrar' + next, { replaceState: true });
			} else if (me && !me.onboarded && path !== '/bem-vindo') {
				await goto('/bem-vindo' + next, { replaceState: true });
			} else if (me && me.onboarded && (path === '/entrar' || path === '/bem-vindo')) {
				await goto('/', { replaceState: true });
			}
			ready = true;
		})();
	});
</script>

{#if ready}
	{@render children()}
	{#if session.me?.onboarded && tabPaths.includes(page.url.pathname)}
		<TabBar />
	{/if}
{/if}
