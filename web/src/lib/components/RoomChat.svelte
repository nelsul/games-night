<script lang="ts">
	// The room chat, as a sheet over the lobby or the match.
	import { tick } from 'svelte';
	import { api, ApiError, type ChatMessage } from '$lib/api';
	import Avatar from './Avatar.svelte';
	import Icon from './Icon.svelte';

	let { code, me, messages, onclose }: { code: string; me: string; messages: ChatMessage[]; onclose: () => void } =
		$props();

	let text = $state('');
	let error = $state('');
	let sending = $state(false);
	let list = $state<HTMLOListElement>();
	let input = $state<HTMLTextAreaElement>();

	const time = (iso: string) => new Date(iso).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' });

	// Stick to the newest message whenever one arrives.
	$effect(() => {
		void messages.length;
		tick().then(() => list?.scrollTo({ top: list.scrollHeight }));
	});

	async function send(e?: SubmitEvent) {
		e?.preventDefault();
		const body = text.trim();
		if (!body || sending) return;
		error = '';
		sending = true;
		try {
			await api.post(`/rooms/${code}/messages`, { body });
			text = ''; // the message itself arrives over the room socket
			input?.focus();
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Não deu para enviar.';
		} finally {
			sending = false;
		}
	}

	function onKey(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
			e.preventDefault();
			send();
		}
	}
</script>

<div class="backdrop" role="presentation" onclick={(e) => e.target === e.currentTarget && onclose()}>
	<div class="sheet" role="dialog" aria-modal="true" aria-labelledby="chat-title" tabindex="-1">
		<header>
			<h2 id="chat-title">Chat da sala</h2>
			<button class="close" onclick={onclose} aria-label="Fechar o chat"><Icon name="close" size={20} /></button>
		</header>
		<ol bind:this={list} aria-live="polite">
			{#each messages as m, i (m.id)}
				{@const mine = m.user.id === me}
				{@const grouped = i > 0 && messages[i - 1].user.id === m.user.id}
				<li class:mine class:grouped>
					{#if !mine && !grouped}<Avatar user={m.user} size={28} />{:else if !mine}<span class="spacer"></span>{/if}
					<div class="bubble">
						{#if !mine && !grouped}<span class="who">{m.user.name}</span>{/if}
						<span class="body">{m.body}</span>
						<span class="time">{time(m.created_at)}</span>
					</div>
				</li>
			{:else}
				<li class="empty">Ninguém falou nada ainda. Puxe o assunto!</li>
			{/each}
		</ol>
		{#if error}<p class="error" role="alert">{error}</p>{/if}
		<form onsubmit={send}>
			<label for="chat-input" class="sr-only">Mensagem</label>
			<textarea
				id="chat-input"
				bind:this={input}
				bind:value={text}
				onkeydown={onKey}
				rows="1"
				maxlength="300"
				placeholder="Mensagem"
				autocomplete="off"
			></textarea>
			<button type="submit" class="send" disabled={!text.trim() || sending} aria-label="Enviar"><Icon name="send" size={20} /></button>
		</form>
	</div>
</div>

<style>
	.backdrop {
		position: fixed;
		inset: 0;
		z-index: 70;
		background: rgba(0, 0, 0, 0.6);
		display: flex;
		align-items: flex-end;
		justify-content: center;
	}
	.sheet {
		width: 100%;
		max-width: 560px;
		height: min(82dvh, 720px);
		background: #181818;
		border-radius: 16px 16px 0 0;
		display: flex;
		flex-direction: column;
		padding-bottom: var(--safe-bottom);
		color: #fff;
	}
	header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px 8px 6px 16px;
		border-bottom: 1px solid var(--line);
	}
	h2 {
		margin: 0;
		font-size: 17px;
	}
	.close {
		width: 44px;
		height: 44px;
		border: none;
		background: none;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
	}
	ol {
		flex: 1;
		overflow-y: auto;
		list-style: none;
		margin: 0;
		padding: 12px 12px 8px;
		display: flex;
		flex-direction: column;
		gap: 6px;
		overscroll-behavior: contain;
	}
	li {
		display: flex;
		align-items: flex-end;
		gap: 8px;
		max-width: 85%;
	}
	li.grouped {
		margin-top: -3px;
	}
	li.mine {
		align-self: flex-end;
	}
	.spacer {
		width: 28px;
		flex-shrink: 0;
	}
	.bubble {
		background: #2a2a2a;
		border-radius: 16px 16px 16px 4px;
		padding: 7px 10px 5px;
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
	}
	.mine .bubble {
		background: var(--brand);
		border-radius: 16px 16px 4px 16px;
	}
	.who {
		font-size: 12px;
		font-weight: 700;
		color: var(--muted);
	}
	.body {
		font-size: 15px;
		line-height: 1.35;
		white-space: pre-wrap;
		overflow-wrap: anywhere;
	}
	.time {
		align-self: flex-end;
		font-size: 11px;
		color: rgba(255, 255, 255, 0.6);
	}
	.empty {
		margin: auto;
		color: var(--muted);
		font-size: 14px;
		max-width: none;
	}
	.error {
		padding: 0 16px;
	}
	form {
		display: flex;
		gap: 8px;
		align-items: flex-end;
		padding: 8px 12px 10px;
		border-top: 1px solid var(--line);
	}
	textarea {
		flex: 1;
		resize: none;
		min-height: 44px;
		max-height: 120px;
		field-sizing: content;
		border-radius: 22px;
		border: 1px solid var(--line-strong);
		background: #2a2a2a;
		color: #fff;
		padding: 11px 16px;
		font: inherit;
		font-size: 16px;
		line-height: 1.3;
	}
	textarea:focus {
		outline: none;
		border-color: #fff;
	}
	.send {
		width: 44px;
		height: 44px;
		flex-shrink: 0;
		border-radius: 50%;
		border: none;
		background: var(--brand);
		color: #fff;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
	}
	.send:disabled {
		opacity: 0.4;
	}
</style>
