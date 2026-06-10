<script lang="ts">
	// Helper to generate spreadsheet column labels (A, B, ..., Z, AA, AB, ..., AN)
	function getColLabel(index: number): string {
		let label = '';
		let temp = index;
		while (temp >= 0) {
			label = String.fromCharCode((temp % 26) + 65) + label;
			temp = Math.floor(temp / 26) - 1;
		}
		return label;
	}

	// Generate 40 column headers (A through AN)
	const columns = Array.from({ length: 40 }, (_, i) => getColLabel(i));

	// Generate 100 row headers (1 through 100)
	const rows = Array.from({ length: 100 }, (_, i) => i + 1);
</script>

<div class="spreadsheet-viewport" tabindex="-1" aria-label="Spreadsheet grid">
	<div class="grid-table">
		<!-- Top-left corner selector block -->
		<div class="corner-header" aria-hidden="true"></div>

		<!-- Column Headers A..AN -->
		{#each columns as col (col)}
			<div class="column-header" class:active={col === 'A'} aria-label="Column {col}">
				{col}
			</div>
		{/each}

		<!-- Row Loop -->
		{#each rows as row (row)}
			<!-- Row Header -->
			<div class="row-header" class:active={row === 1} aria-label="Row {row}">
				{row}
			</div>

			<!-- Grid Cells for this row -->
			{#each columns as col (col)}
				{#if col === 'A' && row === 1}
					<!-- Active A1 Selection Cell -->
					<div
						class="grid-cell active-cell"
						data-cell-ref="A1"
						aria-label="Cell A1, selected"
					></div>
				{:else}
					<!-- Standard Blank Grid Cell -->
					<div
						class="grid-cell"
						data-cell-ref="{col}{row}"
						aria-label="Cell {col}{row}, empty"
					></div>
				{/if}
			{/each}
		{/each}
	</div>
</div>

<style>
	/* Spreadsheet viewport handles scrolling naturally without custom scroll logic */
	.spreadsheet-viewport {
		width: 100%;
		height: 100%;
		overflow: auto;
		background-color: var(--color-surface);
		outline: none;
		position: relative;
	}

	/* CSS Grid for perfect tabular spreadsheet alignment */
	.grid-table {
		--row-header-width: 40px;
		--column-width: 100px;
		--row-height: 24px;

		display: grid;
		grid-template-columns: var(--row-header-width) repeat(40, var(--column-width));
		grid-auto-rows: var(--row-height);
		width: max-content;
		position: relative;
	}

	/* Top-Left Corner Header Block */
	.corner-header {
		position: sticky;
		top: 0;
		left: 0;
		z-index: 10;
		background-color: var(--color-chrome);
		border-right: 1px solid var(--color-border);
		border-bottom: 1px solid var(--color-border);
	}

	/* Column Header cells sticky to top */
	.column-header {
		position: sticky;
		top: 0;
		z-index: 5;
		background-color: var(--color-chrome);
		color: var(--color-text-muted);
		font-family: SFMono-Regular, Consolas, 'Liberation Mono', Menlo, Courier, monospace;
		font-size: 11px;
		font-weight: 500;
		display: flex;
		align-items: center;
		justify-content: center;
		border-right: 1px solid var(--color-border);
		border-bottom: 1px solid var(--color-border);
		user-select: none;
		transition:
			background-color 0.1s ease,
			color 0.1s ease;
	}

	/* Subtle active column highlight for A */
	.column-header.active {
		color: var(--color-accent);
		font-weight: 600;
		background-color: var(--color-surface-hover);
		/* Subtle active indicator bar at the bottom */
		box-shadow: inset 0 -2px 0 0 var(--color-selection-border);
	}

	/* Row Header cells sticky to left */
	.row-header {
		position: sticky;
		left: 0;
		z-index: 5;
		background-color: var(--color-chrome);
		color: var(--color-text-muted);
		font-family: SFMono-Regular, Consolas, 'Liberation Mono', Menlo, Courier, monospace;
		font-size: 11px;
		font-weight: 500;
		display: flex;
		align-items: center;
		justify-content: center;
		border-right: 1px solid var(--color-border);
		border-bottom: 1px solid var(--color-border);
		user-select: none;
		transition:
			background-color 0.1s ease,
			color 0.1s ease;
	}

	/* Subtle active row highlight for row 1 */
	.row-header.active {
		color: var(--color-accent);
		font-weight: 600;
		background-color: var(--color-surface-hover);
		/* Subtle active indicator bar on the right edge */
		box-shadow: inset -2px 0 0 0 var(--color-selection-border);
	}

	/* Standard spreadsheet data cell with subtle gridline */
	.grid-cell {
		background-color: var(--color-surface);
		border-right: 1px solid var(--color-gridline);
		border-bottom: 1px solid var(--color-gridline);
		box-sizing: border-box;
	}

	/* Active selection styling (A1 cell) using inset box shadow for crisp non-shifting layout */
	.grid-cell.active-cell {
		background-color: var(--color-selection-bg);
		box-shadow: inset 0 0 0 2px var(--color-selection-border);
		/* Keep selection outline visible above surrounding cells */
		z-index: 1;
		position: relative;
	}
</style>
