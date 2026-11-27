<script lang="ts">
	// Send feedback, a bug report or a game idea, with your name or anonymously.
	// Only admins read them (/admin/feedbacks).
	import { page } from '$app/state';
	import { scale, fly } from 'svelte/transition';
	import { backOut } from 'svelte/easing';
	import { api, ApiError, type FeedbackKind } from '$lib/api';
	import { session } from '$lib/session.svelte';
	import { haptic, ms, sound } from '$lib/feel.svelte';
	import Avatar from '$lib/components/Avatar.svelte';
	import Icon from '$lib/components/Icon.svelte';

	const kinds: { id: FeedbackKind; name: string; icon: 'chat' | 'bug' | 'bulb'; hint: string; placeholder: string }[] = [
		{
			id: 'feedback',
			name: 'Feedback',
			icon: 'chat',
			hint: 'O que achou do app?',
			placeholder: 'Conta o que você curtiu, o que te irritou ou o que faria diferente.'
		},
		{
			id: 'bug',
			name: 'Bug',
			icon: 'bug',
			hint: 'Algo quebrou?',
			placeholder: 'O que aconteceu, onde (qual tela ou jogo) e o que você esperava. Celular e navegador ajudam.'
		},
		{
			id: 'idea',
			name: 'Ideia de jogo',
			icon: 'bulb',
			hint: 'Que jogo a gente faz?',
			placeholder: 'Como seria o jogo? Quantas pessoas, qual o objetivo, qual a graça?'
		}
	];

	const asked = page.url.searchParams.get('tipo');
	let kind = $state<FeedbackKind>(kinds.some((k) => k.id === asked) ? (asked as FeedbackKind) : 'feedback');
	let body = $state('');
	let anonymous = $state(false);
	let busy = $state(false);
	let error = $state('');
	let sent = $state(false);

	const current = $derived(kinds.find((k) => k.id === kind)!);
	const length = $derived(body.trim().length);

	async function send(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		busy = true;
		try {
			await api.post('/feedback', { kind, body, anonymous });
			sent = true;
			sound('deploy');
			haptic('success');
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
		} finally {
			busy = false;
		}
	}

	function again() {
		body = '';
		sent = false;
	}
</script>

<svelte:head><title>Feedback · Games Night</title></svelte:head>

<main class="page fb">
	<header>
		<a href="/perfil" aria-label="Voltar"><Icon name="back" /></a>
	</header>

	{#if sent}
		<section class="thanks" aria-live="polite">
			<span class="badge" in:scale={{ start: 0.3, duration: ms(500), easing: backOut }}><Icon name="check" size={44} /></span>
			<h1 in:fly={{ y: 12, duration: ms(300), delay: ms(150) }}>Valeu!</h1>
			<p class="muted" in:fly={{ y: 12, duration: ms(300), delay: ms(250) }}>
				{anonymous ? 'Recebemos, sem o seu nome.' : 'Recebemos. Se precisar, a gente te chama.'}
			</p>
			<button class="btn btn-primary btn-block" onclick={again}>Enviar outro</button>
			<a class="btn btn-ghost btn-block" href="/">Voltar ao início</a>
		</section>
	{:else}
		<h1>Fale com a gente</h1>
		<p class="muted">Feedback, bugs e ideias de jogos novos. Quem lê é o time do Games Night.</p>

		<form onsubmit={send}>
			<div class="kinds" role="radiogroup" aria-label="Tipo">
				{#each kinds as k (k.id)}
					<button type="button" role="radio" aria-checked={kind === k.id} class="kind {k.id}" onclick={() => (kind = k.id)}>
						<Icon name={k.icon} size={26} />
						<span>{k.name}</span>
					</button>
				{/each}
			</div>

			<div class="field">
				<label for="body">{current.hint}</label>
				<textarea id="body" class="input" rows="7" maxlength="2000" placeholder={current.placeholder} bind:value={body} required></textarea>
				<span class="count muted" class:near={length > 1800}>{length}/2000</span>
			</div>

			<div class="who" role="radiogroup" aria-label="Enviar como">
				<button type="button" role="radio" aria-checked={!anonymous} class="as" onclick={() => (anonymous = false)}>
					{#if session.me}<Avatar user={session.me} size={36} radius={18} />{/if}
					<span class="as-text">
						<strong>{session.me?.name}</strong>
						<span class="muted">@{session.me?.handle}</span>
					</span>
					<span class="tick" aria-hidden="true">{#if !anonymous}<Icon name="check" size={18} />{/if}</span>
				</button>
				<button type="button" role="radio" aria-checked={anonymous} class="as" onclick={() => (anonymous = true)}>
					<span class="mask"><Icon name="mask" size={22} /></span>
					<span class="as-text">
						<strong>Anônimo</strong>
						<span class="muted">Não guardamos quem enviou.</span>
					</span>
					<span class="tick" aria-hidden="true">{#if anonymous}<Icon name="check" size={18} />{/if}</span>
				</button>
			</div>

			{#if error}<p class="error" role="alert">{error}</p>{/if}
			<button class="btn btn-primary btn-block" disabled={busy || length < 3}>
				<Icon name="send" size={18} />{busy ? 'Enviando…' : anonymous ? 'Enviar anônimo' : 'Enviar'}
			</button>
		</form>
	{/if}
</main>

<style>
	.fb {
		display: flex;
		flex-direction: column;
		gap: 14px;
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
		justify-content: center;
		margin-left: -10px;
	}
	h1 {
		margin: 4px 0 0;
		font-size: 28px;
	}
	p {
		margin: 0;
		line-height: 1.45;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 16px;
		margin-top: 6px;
	}
	.kinds {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 8px;
	}
	.kind {
		--tone: #4493f8;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 6px;
		min-height: 84px;
		padding: 10px 4px;
		border-radius: var(--radius);
		border: 2px solid transparent;
		background: var(--surface);
		color: var(--muted);
		font: inherit;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		transition:
			border-color 0.15s,
			color 0.15s,
			transform 0.15s;
	}
	.kind.bug {
		--tone: #ff7b72;
	}
	.kind.idea {
		--tone: #e3b341;
	}
	.kind[aria-checked='true'] {
		border-color: var(--tone);
		color: #fff;
		background: color-mix(in srgb, var(--tone) 14%, var(--surface));
	}
	.kind[aria-checked='true'] :global(svg) {
		color: var(--tone);
		animation: hop 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
	}
	textarea {
		height: auto;
		min-height: 150px;
		padding: 12px 14px;
		resize: vertical;
		font: inherit;
		line-height: 1.45;
	}
	.count {
		align-self: flex-end;
		font-size: 12px;
	}
	.count.near {
		color: #e3b341;
	}
	.who {
		display: flex;
		flex-direction: column;
		background: var(--surface);
		border-radius: var(--radius);
		overflow: hidden;
	}
	.as {
		display: flex;
		align-items: center;
		gap: 12px;
		min-height: 60px;
		padding: 8px 12px;
		background: none;
		border: none;
		color: #fff;
		font: inherit;
		text-align: left;
		cursor: pointer;
	}
	.as + .as {
		border-top: 1px solid var(--line);
	}
	.as-text {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 1px;
		font-size: 13px;
	}
	.as-text strong {
		font-size: 15px;
	}
	.mask {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		background: var(--surface-2);
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.tick {
		width: 24px;
		height: 24px;
		border-radius: 50%;
		border: 2px solid var(--line-strong);
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.as[aria-checked='true'] .tick {
		background: var(--brand);
		border-color: var(--brand);
	}
	.thanks {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		text-align: center;
		margin-top: 12vh;
	}
	.badge {
		width: 88px;
		height: 88px;
		border-radius: 50%;
		background: #238636;
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 0 40px rgba(63, 185, 80, 0.45);
	}
	.thanks h1 {
		margin: 8px 0 0;
	}
	.thanks .btn-primary {
		margin-top: 16px;
	}
	@keyframes hop {
		50% {
			transform: translateY(-4px) scale(1.12);
		}
	}
</style>
