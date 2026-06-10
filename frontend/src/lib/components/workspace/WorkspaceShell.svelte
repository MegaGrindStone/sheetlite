<script lang="ts">
	import TopChrome from './TopChrome.svelte';
</script>

<div class="workspace-shell">
	<!-- Top Chrome -->
	<header class="top-chrome" aria-label="Top chrome">
		<TopChrome />
	</header>

	<!-- Formula/Control Strip Placeholder -->
	<section class="formula-strip" aria-label="Formula/control strip placeholder">
		<span class="label">Formula/control strip</span>
	</section>

	<!-- Left Rail Placeholder -->
	<aside class="left-rail" aria-label="Left rail placeholder">
		<span class="label vertical">Left rail</span>
	</aside>

	<!-- Grid Canvas Placeholder -->
	<main class="grid-canvas" aria-label="Grid canvas placeholder">
		<span class="label">Grid canvas</span>
	</main>

	<!-- Right Rail Placeholder -->
	<aside class="right-rail" aria-label="Right rail placeholder">
		<span class="label vertical">Right rail</span>
	</aside>

	<!-- Bottom Bar Placeholder -->
	<footer class="bottom-bar" aria-label="Bottom bar placeholder">
		<span class="label">Bottom bar</span>
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
		grid-template-rows: 64px 32px minmax(0, 1fr) 32px;
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
		padding: 0 12px;
	}

	/* Left Rail Region */
	.left-rail {
		grid-area: left-rail;
		background-color: var(--color-chrome);
		border-right: 1px solid var(--color-border);
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
	}

	/* Spreadsheet Grid Canvas Region */
	.grid-canvas {
		grid-area: grid-canvas;
		background-color: var(--color-surface);
		display: flex;
		align-items: center;
		justify-content: center;
	}

	/* Right Rail Region */
	.right-rail {
		grid-area: right-rail;
		background-color: var(--color-chrome);
		border-left: 1px solid var(--color-border);
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
	}

	/* Bottom Bar Region */
	.bottom-bar {
		grid-area: bottom-bar;
		background-color: var(--color-chrome);
		border-top: 1px solid var(--color-border);
		display: flex;
		align-items: center;
		padding: 0 12px;
	}

	/* Shared Region labels */
	.label {
		font-size: 11px;
		font-weight: 500;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	/* Vertical labels for narrow vertical rails */
	.vertical {
		writing-mode: vertical-rl;
		transform: rotate(180deg);
		white-space: nowrap;
	}
</style>
