<script lang="ts">
	import { onMount } from 'svelte';
	import { initializeAppearance } from '$lib/appearance.svelte';
	import favicon from '$lib/assets/favicon.svg';
	import '../app.css';

	let { children } = $props();

	onMount(() => {
		const cleanup = initializeAppearance();
		return cleanup;
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}
