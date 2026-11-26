<script lang="ts">
	import { page } from '$app/state';
	import { ApiError } from '$lib/api';
	import { createRoom, games } from '$lib/games';
	import BackdoorTitle from '$lib/components/BackdoorTitle.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import KeyArt from '$lib/art/KeyArt.svelte';
	import Dev from '$lib/art/Dev.svelte';
	import Security from '$lib/art/Security.svelte';
	import Hacker from '$lib/art/Hacker.svelte';
	import Social from '$lib/art/Social.svelte';
	import ProductOwner from '$lib/art/ProductOwner.svelte';
	import ZeroDay from '$lib/art/ZeroDay.svelte';
	import Freelancer from '$lib/art/Freelancer.svelte';

	const game = $derived(games[page.params.id ?? '']);
	const art = { dev: Dev, security: Security, hacker: Hacker, social: Social, po: ProductOwner, zeroday: ZeroDay, freelancer: Freelancer };

	let tab = $state<'roles' | 'how'>('roles');
	let error = $state('');
	let creating = $state(false);

	async function create() {
		if (!game) return;
		error = '';
		creating = true;
		try {
			await createRoom(game.id);
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Algo deu errado.';
			creating = false;
		}
	}
</script>

<svelte:head><title>{game?.name ?? 'Jogo'} · Games Night</title></svelte:head>

{#if !game}
	<main class="page">
		<p>Esse jogo não existe.</p>
		<a class="btn btn-primary" href="/">Voltar ao início</a>
	</main>
{:else}
	<main class="detail">
		<div class="hero">
			<KeyArt height={240} />
			<a href="/" class="close" aria-label="Fechar"><Icon name="close" size={18} /></a>
		</div>

		<div class="body">
			<span class="original">GAMES NIGHT ORIGINAL</span>
			<h1><BackdoorTitle size={52} /><span class="sr-only">{game.name}</span></h1>
			<div class="meta">
				<span class="new">Novo</span><span>{game.players}</span><span>{game.minutes}</span><span class="lang">PT-BR</span>
			</div>
			<button class="btn btn-primary btn-block" onclick={create} disabled={creating}>
				<Icon name="play" size={18} />{creating ? 'Criando…' : 'Criar sala'}
			</button>
			<a class="btn btn-dark btn-block" href="/codigo"><Icon name="qr" size={18} />Entrar com código</a>
			{#if error}<p class="error" role="alert">{error}</p>{/if}
			<p class="synopsis">{game.synopsis}</p>

			<div class="tabs" role="tablist">
				<button role="tab" aria-selected={tab === 'roles'} onclick={() => (tab = 'roles')}>Personagens</button>
				<button role="tab" aria-selected={tab === 'how'} onclick={() => (tab = 'how')}>Como jogar</button>
			</div>

			{#if tab === 'roles'}
				<div role="tabpanel"><ul class="roles">
					{#each game.roles as role (role.name)}
						{@const Art = art[role.art]}
						<li>
							<span class="role-art"><Art size={88} /></span>
							<span class="role-text">
								<strong>{role.name}</strong>
								<span class="team {role.team}">{role.team === 'devs' ? 'Time Devs' : 'Time Hackers'}</span>
								<span class="muted">{role.text}</span>
							</span>
						</li>
					{/each}
				</ul></div>
			{:else}
				<div role="tabpanel"><ol class="how">
					{#each game.howTo as step, i (i)}
						<li><span class="num">{i + 1}</span><span>{step}</span></li>
					{/each}
				</ol></div>
			{/if}
		</div>
	</main>
{/if}

<style>
	.detail {
		max-width: 560px;
		margin: 0 auto;
		padding-bottom: calc(32px + var(--safe-bottom));
	}
	.hero {
		position: relative;
	}
	.close {
		position: absolute;
		top: calc(12px + var(--safe-top));
		right: 12px;
		width: 44px;
		height: 44px;
		border-radius: 50%;
		background: rgba(20, 20, 20, 0.75);
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.body {
		padding: 0 16px;
		margin-top: -24px;
		position: relative;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.original {
		font-size: 12px;
		font-weight: 700;
		letter-spacing: 2px;
		color: var(--brand);
	}
	h1 {
		margin: 0;
		font-weight: 400;
	}
	.meta {
		display: flex;
		flex-wrap: wrap;
		gap: 10px;
		font-size: 13px;
		color: var(--muted);
		align-items: center;
	}
	.new {
		color: var(--ok);
		font-weight: 700;
	}
	.lang {
		border: 1px solid #8c8c8c;
		padding: 0 5px;
		border-radius: 3px;
		font-size: 11px;
	}
	.btn-dark {
		background: var(--surface-2);
		color: #fff;
	}
	.synopsis {
		margin: 4px 0;
		font-size: 15px;
		line-height: 1.5;
		color: var(--text-2);
	}
	.tabs {
		display: flex;
		gap: 24px;
		border-top: 1px solid var(--line);
		margin: 4px -16px 0;
		padding: 0 16px;
	}
	.tabs button {
		background: none;
		border: none;
		border-top: 4px solid transparent;
		margin-top: -1px;
		padding: 12px 0;
		font-weight: 600;
		font-size: 15px;
		color: var(--muted);
		cursor: pointer;
		min-height: 44px;
	}
	.tabs button[aria-selected='true'] {
		color: #fff;
		border-top-color: var(--brand);
		font-weight: 700;
	}
	.roles,
	.how {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.roles li {
		display: flex;
		gap: 14px;
		align-items: center;
	}
	.role-art {
		flex-shrink: 0;
		display: flex;
	}
	.role-text {
		display: flex;
		flex-direction: column;
		gap: 4px;
		font-size: 14px;
		line-height: 1.4;
	}
	.role-text strong {
		font-size: 16px;
	}
	.team {
		font-size: 12px;
		font-weight: 600;
	}
	.team.devs {
		color: #3fb950;
	}
	.team.hackers {
		color: #ff7b72;
	}
	.how li {
		display: flex;
		gap: 12px;
		line-height: 1.5;
		font-size: 15px;
		color: var(--text-2);
	}
	.num {
		flex-shrink: 0;
		width: 28px;
		height: 28px;
		border-radius: 50%;
		background: var(--surface-2);
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
		font-size: 14px;
	}
</style>
