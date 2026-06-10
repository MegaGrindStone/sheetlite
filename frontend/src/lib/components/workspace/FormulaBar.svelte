<script lang="ts">
	import type { main } from '$lib/wailsjs/go/models';

	type Props = {
		view?: main.WorkbookViewState;
		activeCell?: main.CellData;
	};

	let { view, activeCell }: Props = $props();

	const activeCellRef = $derived(view?.activeCell?.ref ?? '');
	const formulaText = $derived(
		activeCell?.hasFormula && activeCell.formula ? activeCell.formula : (activeCell?.value ?? '')
	);
	const activeCellTitle = $derived(
		activeCellRef ? `Current cell (${activeCellRef})` : 'Current cell'
	);
	const formulaTitle = $derived(
		activeCellRef ? `Formula bar for ${activeCellRef} is inactive.` : 'Formula bar is inactive.'
	);
</script>

<div class="formula-bar" aria-label="Formula bar">
	<!-- Name box showing the Go-owned active cell reference -->
	<div class="name-box" aria-label="Current cell reference" title={activeCellTitle}>
		{activeCellRef}
	</div>

	<!-- Split Divider -->
	<div class="divider" aria-hidden="true"></div>

	<!-- Static fx marker for the future formula controls -->
	<div class="fx-marker" aria-hidden="true">fx</div>

	<!-- Disabled formula input: visible shell only, not editable yet -->
	<input
		class="formula-display"
		type="text"
		disabled
		aria-label={`Formula bar input for ${activeCellRef || 'selected cell'} (inactive)`}
		title={formulaTitle}
		value={formulaText}
	/>
</div>

<style>
	.formula-bar {
		display: flex;
		align-items: center;
		width: 100%;
		height: 100%;
		background-color: var(--color-chrome);
		padding: 0 12px;
		gap: 8px;
		user-select: none;
		box-sizing: border-box;
	}

	.name-box {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 60px;
		height: 22px;
		background-color: var(--color-surface);
		border: 1px solid var(--color-selection-border);
		color: var(--color-text);
		font-family: SFMono-Regular, Consolas, 'Liberation Mono', Menlo, Courier, monospace;
		font-size: 11px;
		font-weight: 500;
		text-align: center;
		user-select: none;
		box-sizing: border-box;
	}

	.divider {
		width: 1px;
		height: 16px;
		background-color: var(--color-border);
		flex-shrink: 0;
	}

	.fx-marker {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 20px;
		height: 22px;
		font-family: 'Times New Roman', Georgia, serif;
		font-size: 14px;
		font-style: italic;
		font-weight: 600;
		color: var(--color-disabled-text);
		user-select: none;
		flex-shrink: 0;
	}

	.formula-display {
		flex: 1;
		height: 22px;
		background-color: var(--color-disabled-bg);
		border: 1px solid var(--color-border);
		color: var(--color-disabled-text);
		cursor: default;
		box-sizing: border-box;
	}
</style>
