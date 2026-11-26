<script lang="ts">
	let { size = 32 }: { size?: number } = $props();
</script>

<span class="logo" style="font-size: {size}px">GAMES NIGHT</span>

<style>
	.logo {
		font-family: var(--display);
		letter-spacing: 1px;
		color: var(--brand);
		line-height: 1;
	}
</style>
