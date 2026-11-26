<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api, ApiError, type Me } from '$lib/api';
	import { session } from '$lib/session.svelte';
	import Logo from '$lib/components/Logo.svelte';
	import PhotoPicker from '$lib/components/PhotoPicker.svelte';
	import HandleField from '$lib/components/HandleField.svelte';

	let name = $state(session.me?.name ?? '');
	let handle = $state(session.me?.handle ?? '');
	let handleOk = $state(false);
	const handleReady = $derived(handleOk || (!!session.me?.handle && handle === session.me.handle));
	let error = $state('');
	let busy = $state(false);

	async function save(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		busy = true;
		try {
			session.me = await api.patch<Me>('/me', { name, handle });
			const next = page.url.searchParams.get('next');
			await goto(next && next.startsWith('/') && !next.startsWith('//') ? next : '/', { replaceState: true });
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
			busy = false;
		}
	}
</script>

<svelte:head><title>Bem-vindo · Games Night</title></svelte:head>

<main class="page welcome">
	<header><Logo size={30} /></header>
	<h1>Como a galera te chama?</h1>
	<PhotoPicker />
	<form onsubmit={save}>
		<div class="field">
			<label for="name">Seu nome</label>
			<input id="name" class="input" maxlength="24" autocomplete="nickname" required bind:value={name} />
		</div>
		<HandleField bind:handle bind:ok={handleOk} current={session.me?.handle ?? ''} />
		{#if error}<p class="error" role="alert">{error}</p>{/if}
		<button class="btn btn-primary btn-block" disabled={busy || !name.trim() || !handleReady}>Continuar</button>
	</form>
</main>

<style>
	.welcome {
		display: flex;
		flex-direction: column;
		gap: 28px;
	}
	header {
		height: 48px;
		display: flex;
		align-items: center;
	}
	h1 {
		margin: 0;
		font-size: 28px;
		font-weight: 600;
		text-align: center;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
</style>
