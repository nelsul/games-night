<script lang="ts">
	// The profile photo: tap to pick one, then crop, rotate or mirror it before upload.
	import { api, ApiError, type Me } from '$lib/api';
	import AvatarCropper from './AvatarCropper.svelte';
	import { session } from '$lib/session.svelte';
	import Avatar from './Avatar.svelte';
	import Icon from './Icon.svelte';

	let { size = 112 }: { size?: number } = $props();
	let input = $state<HTMLInputElement>();
	let busy = $state(false);
	let error = $state('');
	let picked = $state<File | null>(null);

	function onPick() {
		if (!input) return;
		const file = input.files?.[0];
		input.value = '';
		if (!file) return;
		error = '';
		picked = file;
	}

	async function upload(blob: Blob) {
		busy = true;
		try {
			const form = new FormData();
			form.append('file', blob, 'avatar.jpg');
			session.me = await api.post<Me>('/me/avatar', form);
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Não deu para usar essa foto. Tente outra.';
		} finally {
			busy = false;
			picked = null; // close the cropper so a problem shows under the photo
		}
	}
</script>

{#if session.me}
	<div class="picker">
		<button
			type="button"
			class="photo"
			onclick={() => input?.click()}
			disabled={busy || !session.me.photos_enabled}
			aria-label={session.me.avatar_url ? 'Trocar foto' : 'Escolher foto'}
		>
			<Avatar user={session.me} {size} radius={8} />
			{#if session.me.photos_enabled}
				<span class="badge" aria-hidden="true"><Icon name="camera" size={18} /></span>
			{/if}
		</button>
		<input bind:this={input} type="file" accept="image/*" class="sr-only" tabindex="-1" onchange={onPick} />
		{#if busy}<p class="muted small">Enviando foto…</p>{/if}
		{#if !session.me.photos_enabled}<p class="muted small">Fotos de perfil chegam em breve.</p>{/if}
		{#if error}<p class="error" role="alert">{error}</p>{/if}
	</div>
{/if}

{#if picked}
	<AvatarCropper file={picked} onsave={upload} oncancel={() => (picked = null)} />
{/if}

<style>
	.picker {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 10px;
	}
	.photo {
		position: relative;
		border: none;
		background: none;
		padding: 0;
		cursor: pointer;
	}
	.photo:disabled {
		cursor: default;
	}
	.badge {
		position: absolute;
		right: -8px;
		bottom: -8px;
		width: 36px;
		height: 36px;
		border-radius: 50%;
		background: #fff;
		color: #141414;
		display: flex;
		align-items: center;
		justify-content: center;
		border: 3px solid var(--bg);
	}
	.small {
		margin: 0;
		font-size: 13px;
	}
</style>
