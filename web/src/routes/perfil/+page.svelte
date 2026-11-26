<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, ApiError, type Me } from '$lib/api';
	import { session } from '$lib/session.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import PhotoPicker from '$lib/components/PhotoPicker.svelte';
	import HandleField from '$lib/components/HandleField.svelte';

	let name = $state(session.me?.name ?? '');
	let handle = $state(session.me?.handle ?? '');
	let handleOk = $state(false);
	const handleChanged = $derived(handle !== (session.me?.handle ?? ''));
	const nameChanged = $derived(name.trim() !== (session.me?.name ?? ''));
	let error = $state('');
	let saved = $state(false);
	let busy = $state(false);

	async function save(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		saved = false;
		busy = true;
		try {
			session.me = await api.patch<Me>('/me', handleChanged ? { name, handle } : { name });
			saved = true;
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
		} finally {
			busy = false;
		}
	}

	async function removePhoto() {
		try {
			session.me = await api.del<Me>('/me/avatar');
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
		}
	}

	async function logout() {
		await api.post('/auth/logout').catch(() => {});
		session.me = null;
		await goto('/entrar', { replaceState: true });
	}
</script>

<svelte:head><title>Perfil · Games Night</title></svelte:head>

<main class="page profile">
	<h1>Perfil</h1>
	<PhotoPicker />
	{#if session.me?.avatar_url}
		<button type="button" class="link" onclick={removePhoto}>Remover foto</button>
	{/if}
	<form onsubmit={save}>
		<div class="field">
			<label for="name">Nome</label>
			<input id="name" class="input" maxlength="24" required bind:value={name} oninput={() => (saved = false)} />
		</div>
		<HandleField bind:handle bind:ok={handleOk} current={session.me?.handle ?? ''} />
		<p class="muted email">{session.me?.email}</p>
		{#if error}<p class="error" role="alert">{error}</p>{/if}
		<button class="btn btn-primary btn-block" disabled={busy || !name.trim() || (!nameChanged && !handleChanged) || (handleChanged && !handleOk)}>
			{saved ? 'Salvo' : 'Salvar'}
		</button>
	</form>
	<button type="button" class="btn btn-ghost btn-block" onclick={logout}><Icon name="logout" size={20} />Sair</button>
</main>

<style>
	.profile {
		display: flex;
		flex-direction: column;
		gap: 22px;
	}
	h1 {
		margin: 8px 0 0;
		font-size: 28px;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.email {
		margin: 0;
		font-size: 14px;
		word-break: break-all;
	}
	.link {
		align-self: center;
		background: none;
		border: none;
		color: var(--muted);
		text-decoration: underline;
		min-height: 44px;
		margin-top: -12px;
		cursor: pointer;
	}
</style>
