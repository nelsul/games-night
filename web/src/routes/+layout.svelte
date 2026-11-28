<script lang="ts">
	import '../app.css';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { loadMe, session } from '$lib/session.svelte';
	import { refreshInbox, startInbox, stopInbox } from '$lib/inbox.svelte';
	import TabBar from '$lib/components/TabBar.svelte';
	import { onMount } from 'svelte';
	import { tapFeedback } from '$lib/feel.svelte';

	let { children } = $props();

	const publicPaths = ['/entrar'];
	const tabPaths = ['/', '/salas', '/amigos', '/perfil'];

	let ready = $state(false);

	// Every tap on a button or button-like link gets a light haptic tick and a soft click.
	onMount(() => {
		const onTap = (e: MouseEvent) => {
			const el = (e.target as Element | null)?.closest('button, a.btn, [role="switch"], [role="radio"], [role="tab"]');
			if (el && !(el as HTMLButtonElement).disabled) tapFeedback();
		};
		document.addEventListener('click', onTap, true);
		return () => document.removeEventListener('click', onTap, true);
	});

	// Friend requests and invites follow the person around the site while signed in.
	$effect(() => {
		if (session.me?.onboarded) {
			startInbox();
			refreshInbox();
		} else {
			stopInbox();
		}
	});

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

<div class="statusbar" aria-hidden="true"></div>
{#if ready}
	{@render children()}
	{#if session.me?.onboarded && tabPaths.includes(page.url.pathname)}
		<TabBar />
	{/if}
{/if}
