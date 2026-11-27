<script lang="ts">
	import RoleArt from './RoleArt.svelte';
	let { size = 88 }: { size?: number } = $props();
</script>

<RoleArt kind="orange" {size} />
