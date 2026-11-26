<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, ApiError, type Me } from '$lib/api';
	import { session } from '$lib/session.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import PhotoPicker from '$lib/components/PhotoPicker.svelte';
	import HandleField from '$lib/components/HandleField.svelte';
	import { feel, haptic, saveFeel, sound } from '$lib/feel.svelte';

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
	<section class="prefs" aria-labelledby="prefs-title">
		<h2 id="prefs-title">Nos jogos</h2>
		<label class="pref">
			<span><strong>Sons</strong><span class="muted">Efeitos de votos, deploys, rollbacks e chat.</span></span>
			<input
				type="checkbox"
				role="switch"
				bind:checked={feel.sound}
				onchange={() => {
					saveFeel();
					if (feel.sound) sound('select');
				}}
			/>
		</label>
		<label class="pref">
			<span><strong>Vibração</strong><span class="muted">Ao tocar nos botões e nos momentos importantes.</span></span>
			<input
				type="checkbox"
				role="switch"
				bind:checked={feel.haptics}
				onchange={() => {
					saveFeel();
					if (feel.haptics) haptic('select');
				}}
			/>
		</label>
	</section>

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
	.prefs {
		display: flex;
		flex-direction: column;
		gap: 2px;
		background: var(--surface);
		border-radius: var(--radius);
		padding: 4px 12px;
	}
	.prefs h2 {
		margin: 8px 0 4px;
		font-size: 16px;
	}
	.pref {
		display: flex;
		align-items: center;
		gap: 12px;
		min-height: 60px;
		cursor: pointer;
	}
	.pref > span {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 2px;
		font-size: 13px;
	}
	.pref strong {
		font-size: 15px;
	}
	.pref input {
		appearance: none;
		width: 52px;
		height: 32px;
		border-radius: 16px;
		background: #4d4d4d;
		position: relative;
		cursor: pointer;
		flex-shrink: 0;
		transition: background-color 0.15s;
	}
	.pref input::after {
		content: '';
		position: absolute;
		top: 3px;
		left: 3px;
		width: 26px;
		height: 26px;
		border-radius: 50%;
		background: #fff;
		transition: transform 0.15s;
	}
	.pref input:checked {
		background: var(--brand);
	}
	.pref input:checked::after {
		transform: translateX(20px);
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
