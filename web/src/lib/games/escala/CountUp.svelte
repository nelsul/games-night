<script lang="ts">
	// A number that counts up from 0 when it appears (instantly with reduced motion).
	import { reducedMotion } from '$lib/feel.svelte';

	let { value, duration = 650 }: { value: number; duration?: number } = $props();
	let shown = $state(0);

	$effect(() => {
		const target = value;
		if (reducedMotion) {
			shown = target;
			return;
		}
		const start = performance.now();
		let raf = 0;
		const tick = (now: number) => {
			const t = Math.min(1, (now - start) / duration);
			shown = Math.round(target * (1 - (1 - t) ** 3)); // ease out
			if (t < 1) raf = requestAnimationFrame(tick);
		};
		raf = requestAnimationFrame(tick);
		return () => cancelAnimationFrame(raf);
	});
</script>

{shown}
