<script lang="ts">
	import { goto } from '$app/navigation';
	import Icon from '$lib/components/Icon.svelte';

	const alphabet = 'ABCDEFGHJKMNPQRSTUVWXYZ23456789';
	let code = $state('');
	let error = $state('');

	function onInput() {
		code = code
			.toUpperCase()
			.replace(/[^A-Z0-9]/g, '')
			.slice(0, 4);
		error = '';
	}

	async function go(e: SubmitEvent) {
		e.preventDefault();
		if ([...code].some((c) => !alphabet.includes(c))) {
			error = 'Esse código não existe. Confira as letras com quem criou a sala.';
			return;
		}
		await goto(`/sala/${code}`);
	}
</script>

<svelte:head><title>Entrar com código · Games Night</title></svelte:head>

<main class="page enter">
	<header>
		<a href="/" aria-label="Voltar"><Icon name="back" /></a>
	</header>
	<h1>Código da sala</h1>
	<p class="muted">Peça para quem criou a sala mostrar o código ou o QR.</p>
	<form onsubmit={go}>
		<label class="sr-only" for="code">Código da sala</label>
		<input
			id="code"
			class="input code"
			bind:value={code}
			oninput={onInput}
			autocomplete="off"
			autocapitalize="characters"
			spellcheck="false"
			maxlength="4"
			placeholder="KX7Q"
			required
		/>
		{#if error}<p class="error" role="alert">{error}</p>{/if}
		<button class="btn btn-primary btn-block" disabled={code.length !== 4}>Entrar na sala</button>
	</form>
</main>

<style>
	.enter {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	header {
		height: 44px;
		display: flex;
		align-items: center;
	}
	header a {
		width: 44px;
		height: 44px;
		display: flex;
		align-items: center;
		margin-left: -10px;
		justify-content: center;
	}
	h1 {
		margin: 12px 0 0;
		font-size: 28px;
	}
	p {
		margin: 0;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 16px;
		margin-top: 8px;
	}
	.code {
		height: 76px;
		font-family: var(--display);
		font-size: 48px;
		letter-spacing: 14px;
		text-align: center;
		padding-top: 6px;
	}
	.code::placeholder {
		color: #5a5a5a;
	}
</style>
