<script lang="ts">
	// The wordmark, with the same lit-ribbon finish as the app icon.
	let { size = 32 }: { size?: number } = $props();
</script>

<span class="logo" style="font-size: {size}px">GAMES NIGHT</span>

<style>
	.logo {
		font-family: var(--display);
		letter-spacing: 1px;
		line-height: 1;
		color: var(--brand);
		background: linear-gradient(180deg, #ff7173 0%, #e5383b 48%, #a91f23 100%);
		-webkit-background-clip: text;
		background-clip: text;
		-webkit-text-fill-color: transparent;
		filter: drop-shadow(0 2px 3px rgba(0, 0, 0, 0.7));
		padding-top: 0.08em; /* Bebas sits low; keeps the clip from cutting the tops */
	}
</style>
