<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api, ApiError, type Me } from '$lib/api';
	import { session } from '$lib/session.svelte';
	import Logo from '$lib/components/Logo.svelte';

	let step = $state<'email' | 'code'>('email');
	let email = $state('');
	let code = $state('');
	let error = $state('');
	let busy = $state(false);

	async function sendCode(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		busy = true;
		try {
			await api.post('/auth/request-code', { email });
			step = 'code';
			code = '';
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
		} finally {
			busy = false;
		}
	}

	async function verify(e?: SubmitEvent) {
		e?.preventDefault();
		if (busy) return;
		error = '';
		busy = true;
		try {
			session.me = await api.post<Me>('/auth/verify', { email, code });
			session.loaded = true;
			const next = page.url.searchParams.get('next');
			const safe = next && next.startsWith('/') && !next.startsWith('//') ? next : '/';
			await goto(session.me.onboarded ? safe : `/bem-vindo?next=${encodeURIComponent(safe)}`, { replaceState: true });
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
			busy = false;
		}
	}

	function onCodeInput() {
		code = code.replace(/\D/g, '').slice(0, 6);
		if (code.length === 6) verify();
	}
</script>

<svelte:head><title>Entrar · Games Night</title></svelte:head>

<main class="page login">
	<header><Logo size={34} /></header>

	<div class="card">
		{#if step === 'email'}
			<h1>Entrar</h1>
			<p class="muted">Mandamos um código para o seu e-mail. Sem senha.</p>
			<form onsubmit={sendCode}>
				<div class="field">
					<label for="email">E-mail</label>
					<input
						id="email"
						class="input"
						type="email"
						autocomplete="email"
						inputmode="email"
						required
						bind:value={email}
					/>
				</div>
				{#if error}<p class="error" role="alert">{error}</p>{/if}
				<button class="btn btn-brand btn-block" disabled={busy}>{busy ? 'Enviando…' : 'Receber código'}</button>
			</form>
		{:else}
			<h1>Digite o código</h1>
			<p class="muted">Enviamos 6 números para <strong class="email">{email}</strong>.</p>
			<form onsubmit={verify}>
				<div class="field">
					<label for="code">Código</label>
					<input
						id="code"
						class="input code"
						type="text"
						inputmode="numeric"
						autocomplete="one-time-code"
						maxlength="6"
						required
						bind:value={code}
						oninput={onCodeInput}
					/>
				</div>
				{#if error}<p class="error" role="alert">{error}</p>{/if}
				<button class="btn btn-brand btn-block" disabled={busy || code.length !== 6}>{busy ? 'Entrando…' : 'Entrar'}</button>
				<button type="button" class="link" onclick={() => ((step = 'email'), (error = ''))}>Usar outro e-mail</button>
			</form>
		{/if}
	</div>
</main>

<style>
	.login {
		display: flex;
		flex-direction: column;
		gap: 40px;
		padding-bottom: 40px;
	}
	header {
		height: 48px;
		display: flex;
		align-items: center;
	}
	.card {
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding-top: 20px;
	}
	h1 {
		margin: 0;
		font-size: 32px;
		font-weight: 700;
	}
	p {
		margin: 0 0 12px;
		line-height: 1.45;
	}
	.email {
		color: #fff;
		overflow-wrap: anywhere;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.code {
		font-family: var(--mono);
		font-size: 26px;
		letter-spacing: 10px;
		text-align: center;
	}
	.btn-brand {
		background: var(--brand);
		color: #fff;
	}
	.link {
		background: none;
		border: none;
		color: var(--muted);
		text-decoration: underline;
		min-height: 44px;
		cursor: pointer;
	}
</style>
