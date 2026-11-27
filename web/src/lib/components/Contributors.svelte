<script lang="ts">
	// The people who build Games Night: the repository's contributors on GitHub, as a
	// little solar system. The Games Night icon is the sun; people orbit it on rings,
	// the ones with the most commits closest. Rings turn at their own speed (alternating
	// direction), photos stay upright, and they pop in when the section scrolls into
	// view. Tapping someone opens their GitHub.
	import { onMount } from 'svelte';
	import { api, type Contributors } from '$lib/api';
	import { reducedMotion } from '$lib/feel.svelte';

	let data = $state<Contributors | null>(null);
	let failed = $state(false);
	let shown = $state(reducedMotion);
	let grid = $state<HTMLElement>();
	let width = $state(340);

	onMount(() => {
		api
			.get<Contributors>('/contributors')
			.then((d) => (data = d))
			.catch(() => (failed = true));
	});

	// Start the pop-in only once the grid is on screen.
	$effect(() => {
		if (!grid || shown) return;
		const io = new IntersectionObserver(
			(entries) => {
				if (entries.some((e) => e.isIntersecting)) {
					shown = true;
					io.disconnect();
				}
			},
			{ threshold: 0.2 }
		);
		io.observe(grid);
		return () => io.disconnect();
	});

	const repo = $derived(data?.repo ?? 'https://github.com/nelsul/games-night');
	const repoName = $derived(repo.replace('https://github.com/', ''));
	const people = $derived(data?.contributors ?? []);

	interface Body {
		login: string;
		avatar: string;
		url: string;
		commits: number;
		size: number;
		radius: number;
		angle: number;
		seconds: number;
		i: number;
	}

	// A steady random number in [0, 1) per person and purpose, so the layout looks
	// loose but doesn't jump around between visits.
	function wobble(login: string, salt: number) {
		let h = 2166136261 ^ salt;
		for (const c of login) h = Math.imul(h ^ c.charCodeAt(0), 16777619);
		h = Math.imul(h ^ (h >>> 15), 2246822507);
		return ((h ^ (h >>> 13)) >>> 0) / 4294967296;
	}

	// Bands from the inside out, the people with the most commits closest to the sun.
	// First work out how many bands are needed (each holds as many photos as fit around
	// it), then spread just those over the space so a small team still looks like a
	// solar system. Inside a band everyone gets their own orbit: a slightly different
	// distance, start and speed (farther is slower, as with planets), so it never looks
	// like a spinning wheel. Whoever doesn't fit is a "+N" next to the sun.
	const MAX_RINGS = 4;
	const GAP = 14;
	const layout = $derived.by(() => {
		const half = width / 2;
		const sun = Math.max(44, width * 0.17);
		const inner = sun / 2 + 30;
		const outer = half - 22;
		// Photos shrink a little per band, and never get wider than the space between bands.
		const spacingOf = (n: number) => (n === 1 ? outer - inner : (outer - inner) / (n - 1));
		const sizeOf = (k: number, n: number) =>
			Math.round(Math.max(22, Math.min((46 - k * 5) * Math.min(1, width / 340), spacingOf(n) - 8)));
		const radiusOf = (k: number, n: number) => (n === 1 ? inner + (outer - inner) * 0.55 : inner + ((outer - inner) * k) / (n - 1));
		const fits = (k: number, n: number) => Math.max(1, Math.floor((2 * Math.PI * radiusOf(k, n)) / (sizeOf(k, n) + GAP)));

		let n = 1;
		while (n < MAX_RINGS) {
			let room = 0;
			for (let k = 0; k < n; k++) room += fits(k, n);
			if (room >= people.length) break;
			n++;
		}
		let total = 0;
		for (let k = 0; k < n; k++) total += radiusOf(k, n);
		const bodies: Body[] = [];
		let next = 0;
		for (let k = 0; k < n && next < people.length; k++) {
			const share = k === n - 1 ? people.length - next : Math.round((people.length * radiusOf(k, n)) / total);
			const count = Math.min(fits(k, n), Math.max(1, share), people.length - next);
			const step = 360 / count;
			const drift = n === 1 ? Math.min(40, (outer - inner) * 0.34) : Math.min(10, spacingOf(n) * 0.22);
			people.slice(next, next + count).forEach((p, j) => {
				const radius = radiusOf(k, n) + (wobble(p.login, 1) * 2 - 1) * drift;
				bodies.push({
					login: p.login,
					avatar: p.avatar_url,
					url: p.url,
					commits: p.contributions,
					size: Math.round(sizeOf(k, n) * (0.88 + wobble(p.login, 2) * 0.2)),
					radius,
					angle: wobble('ring', k) * 360 + j * step + (wobble(p.login, 3) - 0.5) * step * 0.6,
					seconds: 55 * Math.pow(radius / 100, 1.5) * (0.85 + wobble(p.login, 4) * 0.3),
					i: next + j
				});
			});
			next += count;
		}
		return { sun, bodies, more: people.length - next, bands: n, ringRadii: Array.from({ length: n }, (_, k) => radiusOf(k, n)) };
	});
</script>

<section class="contrib" aria-labelledby="contrib-title">
	<h2 class="row-title" id="contrib-title">Feito pela galera</h2>

	<a class="repo" href={repo} target="_blank" rel="noopener noreferrer">
		<svg class="mark" viewBox="0 0 16 16" width="34" height="34" aria-hidden="true" fill="currentColor"
			><path
				d="M8 0c4.42 0 8 3.58 8 8a8.013 8.013 0 0 1-5.45 7.59c-.4.08-.55-.17-.55-.38 0-.27.01-1.13.01-2.2 0-.75-.25-1.23-.54-1.48 1.78-.2 3.65-.88 3.65-3.95 0-.88-.31-1.59-.82-2.15.08-.2.36-1.02-.08-2.12 0 0-.67-.22-2.2.82-.64-.18-1.32-.27-2-.27-.68 0-1.36.09-2 .27-1.53-1.03-2.2-.82-2.2-.82-.44 1.1-.16 1.92-.08 2.12-.51.56-.82 1.28-.82 2.15 0 3.06 1.86 3.75 3.64 3.95-.23.2-.44.55-.51 1.07-.46.21-1.61.55-2.33-.66-.15-.24-.6-.83-1.23-.82-.67.01-.27.38.01.53.34.19.73.9.82 1.13.16.45.68 1.31 2.69.94 0 .67.01 1.3.01 1.49 0 .21-.15.45-.55.38A7.995 7.995 0 0 1 0 8c0-4.42 3.58-8 8-8Z"
			/></svg
		>
		<span class="repo-text">
			<strong>Código aberto no GitHub</strong>
			<span class="muted">{repoName} · Tem ideia de jogo? Manda um PR.</span>
		</span>
		<span class="arrow" aria-hidden="true">→</span>
	</a>

	{#if !failed}
		<div class="head">
			<span>Contribuidores</span>
			{#if data}<span class="count">{people.length}</span>{/if}
		</div>
		<div class="system" class:shown bind:this={grid} bind:clientWidth={width} aria-busy={!data}>
			{#if layout.bodies.length && layout.bodies.length <= 24}
				{#each layout.bodies as b (b.login)}
					<span class="track" style="width: {b.radius * 2}px; height: {b.radius * 2}px" aria-hidden="true"></span>
				{/each}
			{:else}
				{#each layout.ringRadii as r, k (k)}
					<span class="track" style="width: {r * 2}px; height: {r * 2}px" aria-hidden="true"></span>
				{/each}
			{/if}
			<a class="sun" href={repo} target="_blank" rel="noopener noreferrer" style="--sun: {layout.sun}px" aria-label="Games Night no GitHub">
				<img src="/icon.svg" alt="" width={layout.sun} height={layout.sun} />
				{#if layout.more > 0}<span class="more">+{layout.more}</span>{/if}
			</a>
			<ul class="orbits">
				{#each layout.bodies as b (b.login)}
					<li class="orbit" style="--dur: {b.seconds.toFixed(1)}s; --size: {b.size}px">
						<span class="planet" style="--angle: {b.angle.toFixed(1)}deg; --r: {b.radius.toFixed(1)}px; --i: {b.i}">
							<span class="upright">
								<a
									href={b.url}
									target="_blank"
									rel="noopener noreferrer"
									aria-label="@{b.login}, {b.commits} {b.commits === 1 ? 'commit' : 'commits'}"
								>
									<span class="ring" aria-hidden="true"></span>
									<img src="{b.avatar}&s=96" alt="" width={b.size} height={b.size} loading="lazy" decoding="async" />
									<span class="tip" aria-hidden="true">@{b.login}</span>
								</a>
							</span>
						</span>
					</li>
				{/each}
			</ul>
		</div>
	{/if}
</section>

<style>
	.contrib {
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding-bottom: 8px;
	}
	.repo {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 14px 16px;
		border-radius: 12px;
		background: linear-gradient(135deg, #24292f, #161b22);
		border: 1px solid #30363d;
		color: #fff;
		text-decoration: none;
		transition: transform 0.15s ease;
	}
	.repo:active {
		transform: scale(0.98);
	}
	.mark {
		flex-shrink: 0;
	}
	.repo-text {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
		font-size: 13px;
	}
	.repo-text strong {
		font-size: 15px;
	}
	.muted {
		color: var(--muted);
	}
	.arrow {
		font-size: 20px;
		color: var(--muted);
		transition: transform 0.2s ease;
	}
	.repo:hover .arrow {
		transform: translateX(4px);
	}
	.head {
		display: flex;
		align-items: center;
		gap: 8px;
		font-weight: 700;
		font-size: 15px;
	}
	.count {
		min-width: 24px;
		height: 22px;
		padding: 0 7px;
		border-radius: 999px;
		background: var(--surface-2);
		color: var(--muted);
		font-size: 12px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}
	.system {
		position: relative;
		width: 100%;
		max-width: 420px;
		aspect-ratio: 1;
		margin: 4px auto 8px;
		border-radius: 50%;
		background: radial-gradient(circle, rgba(229, 56, 59, 0.16), transparent 62%);
	}
	.track {
		position: absolute;
		left: 50%;
		top: 50%;
		translate: -50% -50%;
		border-radius: 50%;
		border: 1px solid rgba(255, 255, 255, 0.06);
		pointer-events: none;
	}
	.sun {
		position: absolute;
		left: 50%;
		top: 50%;
		width: var(--sun);
		height: var(--sun);
		translate: -50% -50%;
		border-radius: 50%;
		z-index: 2;
	}
	.sun img {
		display: block;
		width: 100%;
		height: 100%;
		border-radius: 50%;
		box-shadow:
			0 0 18px 4px rgba(229, 56, 59, 0.55),
			0 0 50px 14px rgba(229, 56, 59, 0.25);
		animation: glow 3s ease-in-out infinite alternate;
	}
	.more {
		position: absolute;
		right: -10px;
		bottom: -6px;
		padding: 1px 7px;
		border-radius: 999px;
		background: #fff;
		color: #111;
		font-size: 11px;
		font-weight: 800;
	}
	.orbits {
		list-style: none;
		margin: 0;
		padding: 0;
	}
	.orbit {
		position: absolute;
		inset: 0;
		animation: spin var(--dur) linear infinite;
		pointer-events: none;
	}
	.planet {
		display: block;
		position: absolute;
		left: 50%;
		top: 50%;
		width: var(--size);
		height: var(--size);
		margin: calc(var(--size) / -2) 0 0 calc(var(--size) / -2);
		transform: rotate(var(--angle)) translateX(var(--r)) rotate(calc(-1 * var(--angle)));
	}
	/* Turns against the orbit, so the photo stays upright while it goes around. */
	.upright {
		display: block;
		width: 100%;
		height: 100%;
		animation: spin var(--dur) linear infinite reverse;
	}
	.upright a {
		position: relative;
		display: block;
		width: 100%;
		height: 100%;
		border-radius: 50%;
		text-decoration: none;
		pointer-events: auto;
		opacity: 0;
		transform: scale(0.2);
		-webkit-tap-highlight-color: transparent;
	}
	.system.shown .upright a {
		animation: pop 0.55s cubic-bezier(0.34, 1.56, 0.64, 1) forwards calc(min(var(--i), 30) * 60ms);
	}
	.ring {
		position: absolute;
		inset: -2px;
		border-radius: 50%;
		background: conic-gradient(from var(--ring, 0deg), var(--brand), #ff9f1c, #b5179e, var(--brand));
		animation: ring 5s linear infinite;
		animation-delay: calc(var(--i) * -400ms);
	}
	.upright img {
		position: relative;
		display: block;
		width: 100%;
		height: 100%;
		border-radius: 50%;
		border: 2px solid var(--bg, #141414);
		object-fit: cover;
		background: var(--surface-2);
		transition: transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
	}
	.upright a:hover img,
	.upright a:focus-visible img {
		transform: scale(1.18);
	}
	.upright a:hover,
	.upright a:focus-visible {
		z-index: 3;
	}
	.tip {
		position: absolute;
		left: 50%;
		top: calc(100% + 6px);
		transform: translateX(-50%) translateY(-4px);
		white-space: nowrap;
		font-size: 11px;
		font-weight: 600;
		padding: 2px 7px;
		border-radius: 999px;
		background: #fff;
		color: #111;
		opacity: 0;
		pointer-events: none;
		transition:
			opacity 0.15s ease,
			transform 0.15s ease;
		z-index: 1;
	}
	.upright a:hover .tip,
	.upright a:focus-visible .tip {
		opacity: 1;
		transform: translateX(-50%);
	}
	@property --ring {
		syntax: '<angle>';
		initial-value: 0deg;
		inherits: false;
	}
	@keyframes ring {
		to {
			--ring: 360deg;
		}
	}
	@keyframes spin {
		to {
			rotate: 360deg;
		}
	}
	@keyframes pop {
		to {
			opacity: 1;
			transform: scale(1);
		}
	}
	@keyframes glow {
		to {
			box-shadow:
				0 0 24px 6px rgba(229, 56, 59, 0.7),
				0 0 70px 20px rgba(229, 56, 59, 0.3);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.orbit,
		.upright,
		.ring,
		.sun img {
			animation: none;
		}
		.upright a,
		.system.shown .upright a {
			opacity: 1;
			transform: none;
			animation: none;
		}
	}
</style>
