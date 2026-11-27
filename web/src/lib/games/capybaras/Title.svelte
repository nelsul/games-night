<script lang="ts">
	// The wordmark. It breaks into two lines (CRAZY / CAPYBARAS) where it doesn't fit.
	let { size = 60 }: { size?: number } = $props();
</script>

<span style="display: inline-block; font-family: var(--display); font-size: {size}px; line-height: 0.95; letter-spacing: 2px; text-align: center"
	><span style="color: #ff9f1c">CRAZY</span> <span style="color: #f6e7c8; white-space: nowrap">CAPYBARAS</span></span
>
