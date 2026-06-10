<script lang="ts">
	import TopChrome from './TopChrome.svelte';
	import FormulaBar from './FormulaBar.svelte';
	import SideRail from './SideRail.svelte';
	import BottomBar from './BottomBar.svelte';
	import SpreadsheetGrid from './SpreadsheetGrid.svelte';
</script>

<div class="workspace-shell">
	<!-- Top Chrome -->
	<header class="top-chrome" aria-label="Top chrome">
		<TopChrome />
	</header>

	<!-- Formula/Control Strip -->
	<section class="formula-strip" aria-label="Formula bar">
		<FormulaBar />
	</section>

	<!-- Left Rail -->
	<div class="left-rail-container">
		<SideRail side="left" />
	</div>

	<!-- Grid Canvas Region -->
	<main class="grid-canvas" aria-label="Grid canvas">
		<SpreadsheetGrid />
	</main>

	<!-- Right Rail -->
	<div class="right-rail-container">
		<SideRail side="right" />
	</div>

	<!-- Bottom Bar -->
	<footer class="bottom-bar" aria-label="Bottom bar">
		<BottomBar />
	</footer>
</div>

<style>
	/* Full-window viewport locked CSS Grid Layout */
	.workspace-shell {
		display: grid;
		grid-template-areas:
			'top-chrome top-chrome top-chrome'
			'formula-strip formula-strip formula-strip'
			'left-rail grid-canvas right-rail'
			'bottom-bar bottom-bar bottom-bar';
		grid-template-rows: 64px 32px minmax(0, 1fr) 36px;
		grid-template-columns: 48px 1fr 48px;
		width: 100vw;
		height: 100vh;
		overflow: hidden;
		background-color: var(--color-bg);
		color: var(--color-text);
		user-select: none;
	}

	/* Top Chrome Region */
	.top-chrome {
		grid-area: top-chrome;
		background-color: var(--color-chrome);
		border-bottom: 1px solid var(--color-border);
	}

	/* Formula/Control Strip Region */
	.formula-strip {
		grid-area: formula-strip;
		background-color: var(--color-chrome);
		border-bottom: 1px solid var(--color-border);
		display: flex;
		align-items: center;
	}

	/* Left Rail Region */
	.left-rail-container {
		grid-area: left-rail;
	}

	/* Spreadsheet Grid Canvas Region */
	.grid-canvas {
		grid-area: grid-canvas;
		background-color: var(--color-surface);
		overflow: hidden;
	}

	/* Right Rail Region */
	.right-rail-container {
		grid-area: right-rail;
	}

	/* Bottom Bar Region */
	.bottom-bar {
		grid-area: bottom-bar;
		background-color: var(--color-chrome);
		border-top: 1px solid var(--color-border);
		display: flex;
		align-items: stretch;
		padding: 0 12px;
	}
</style>
