<script lang="ts">
	// The secret badge: hidden until you hold it, so nobody peeks at your phone.
	import Dev from '$lib/art/Dev.svelte';
	import Hacker from '$lib/art/Hacker.svelte';
	import Security from '$lib/art/Security.svelte';
	import Social from '$lib/art/Social.svelte';
	import ProductOwner from '$lib/art/ProductOwner.svelte';
	import ZeroDay from '$lib/art/ZeroDay.svelte';
	import Freelancer from '$lib/art/Freelancer.svelte';
	import type { RoomPlayer } from '$lib/api';
	import { roleInfo, type MatchView } from './types';
	import { haptic, sound } from '$lib/feel.svelte';

	let { match, people }: { match: MatchView; people: Record<string, RoomPlayer> } = $props();

	const art = { dev: Dev, security: Security, hacker: Hacker, social: Social, po: ProductOwner, zeroday: ZeroDay, freelancer: Freelancer };
	const info = $derived(roleInfo[match.me.role]);
	const Art = $derived(art[info.art]);
	let open = $state(false);

	const asLabel: Record<string, string> = {
		hacker: 'hacker',
		social_engineer: 'Engenharia Social',
		zero_day: 'Zero-day',
		analyst_candidate: 'pode ser o Analista'
	};
	const knowsTitle = $derived(
		match.me.role === 'product_owner'
			? 'Um destes é o Analista de Segurança'
			: match.me.role === 'analyst'
				? 'Hackers que você enxerga'
				: 'Seus comparsas'
	);
</script>

<div class="badge" class:hackers={info.team === 'hackers'} class:open>
	<div class="strip"><span>CUDIA · ACESSO</span><span>{open ? 'segure' : ''}</span></div>
	<button
		type="button"
		class="hold"
		onpointerdown={() => {
			open = true;
			sound('reveal');
			haptic('select');
		}}
		onpointerup={() => (open = false)}
		onpointerleave={() => (open = false)}
		onpointercancel={() => (open = false)}
		onkeydown={(e) => (e.key === ' ' || e.key === 'Enter') && (open = true)}
		onkeyup={() => (open = false)}
		oncontextmenu={(e) => e.preventDefault()}
		aria-label="Segure para ver seu crachá"
	>
		{#if open}
			<span class="role">
				<Art size={96} />
				<span class="who">
					<span class="muted small">Você é</span>
					<strong>{info.name}</strong>
					<span class="team">{info.team === 'devs' ? 'Time Devs' : 'Time Hackers'}</span>
				</span>
			</span>
			<span class="text">{info.text}</span>
			{#if match.me.knows.length}
				<span class="knows">
					<span class="muted small">{knowsTitle}</span>
					{#each match.me.knows as k (k.id)}
						<span class="chip"
							>{people[k.id]?.name ?? '?'}{#if match.me.role !== 'product_owner' && k.as !== 'hacker'}
								<em> · {asLabel[k.as]}</em>{/if}</span
						>
					{/each}
				</span>
			{:else if match.me.role === 'freelancer'}
				<span class="muted small">Você não sabe quem são os outros hackers.</span>
			{/if}
		{:else}
			<span class="closed">
				<span class="lock" aria-hidden="true">
					<svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"
						><rect x="5" y="11" width="14" height="9" rx="2" /><path d="M8 11V8a4 4 0 0 1 8 0v3" /></svg
					>
				</span>
				<strong>Segure para ver</strong>
				<span class="muted small">Esconda a tela dos outros. Solte e o crachá fecha.</span>
			</span>
		{/if}
	</button>
</div>

<style>
	.badge {
		background: var(--bd-surface);
		border: 1px solid var(--bd-line);
		border-radius: 24px;
		overflow: hidden;
	}
	.strip {
		display: flex;
		justify-content: space-between;
		padding: 10px 16px;
		background: #161b22;
		font-family: var(--mono);
		font-size: 12px;
		font-weight: 700;
		color: var(--bd-muted);
	}
	.open.hackers .strip {
		background: #2a1215;
		color: #ffa198;
	}
	.open:not(.hackers) .strip {
		background: #12261e;
		color: #8bd5a0;
	}
	.hold {
		width: 100%;
		min-height: 240px;
		border: none;
		background: none;
		padding: 18px 16px;
		display: flex;
		flex-direction: column;
		gap: 14px;
		text-align: left;
		cursor: pointer;
		touch-action: none;
		user-select: none;
		-webkit-user-select: none;
		-webkit-touch-callout: none;
	}
	.closed {
		margin: auto;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		text-align: center;
	}
	.lock {
		color: var(--bd-muted);
	}
	.role,
	.text,
	.knows {
		animation: bd-reveal 0.28s ease-out;
	}
	@keyframes bd-reveal {
		from {
			opacity: 0;
			transform: translateY(8px) scale(0.97);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}
	.role {
		display: flex;
		gap: 14px;
		align-items: center;
	}
	.who {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.who strong {
		font-size: 26px;
		line-height: 1.1;
	}
	.hackers .who strong {
		color: var(--bd-red);
	}
	.badge:not(.hackers) .who strong {
		color: var(--bd-green);
	}
	.team {
		font-size: 12px;
		font-weight: 600;
		color: var(--bd-muted);
	}
	.text {
		font-size: 15px;
		line-height: 1.45;
		color: #c9d1d9;
	}
	.knows {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		align-items: center;
		border-top: 1px solid var(--bd-line);
		padding-top: 12px;
	}
	.knows .small {
		width: 100%;
	}
	.chip {
		padding: 6px 12px;
		border-radius: 999px;
		background: #21262d;
		font-family: var(--mono);
		font-size: 14px;
	}
	.chip em {
		font-style: normal;
		color: var(--bd-muted);
	}
	.small {
		font-size: 13px;
	}
</style>
