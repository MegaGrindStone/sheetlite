<script lang="ts">
	import type { main } from '$lib/wailsjs/go/models';

	type Props = {
		activeSheet?: main.WorkbookSheet;
		view?: main.WorkbookViewState;
		styles?: main.CellStyle[];
		dragActive?: boolean;
		onSelectCell?: (cellRef: string) => Promise<void> | void;
		onSetScrollPosition?: (topRow: number, leftColumn: number) => Promise<void> | void;
	};

	type ColumnHeader = {
		index: number;
		label: string;
	};

	type CellLookup = Record<string, main.CellData>;

	const MIN_COLUMN_COUNT = 40;
	const MIN_ROW_COUNT = 100;

	let {
		activeSheet,
		view,
		styles = [],
		dragActive = false,
		onSelectCell,
		onSetScrollPosition
	}: Props = $props();

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

	function maxPositiveIndex(current: number, value: number | undefined): number {
		if (typeof value !== 'number' || !Number.isFinite(value) || value < 1) {
			return current;
		}

		return Math.max(current, Math.trunc(value));
	}

	function getColumnCount(
		sheet: main.WorkbookSheet | undefined,
		stateView: main.WorkbookViewState | undefined
	): number {
		let count = MIN_COLUMN_COUNT;
		count = maxPositiveIndex(count, sheet?.bounds?.end?.column);
		count = maxPositiveIndex(count, stateView?.activeCell?.column);
		count = maxPositiveIndex(count, stateView?.selection?.start?.column);
		count = maxPositiveIndex(count, stateView?.selection?.end?.column);

		for (const cell of sheet?.cells ?? []) {
			count = maxPositiveIndex(count, cell.column);
		}

		for (const column of sheet?.columns ?? []) {
			count = maxPositiveIndex(count, column.index);
		}

		return count;
	}

	function getRowCount(
		sheet: main.WorkbookSheet | undefined,
		stateView: main.WorkbookViewState | undefined
	): number {
		let count = MIN_ROW_COUNT;
		count = maxPositiveIndex(count, sheet?.bounds?.end?.row);
		count = maxPositiveIndex(count, stateView?.activeCell?.row);
		count = maxPositiveIndex(count, stateView?.selection?.start?.row);
		count = maxPositiveIndex(count, stateView?.selection?.end?.row);

		for (const cell of sheet?.cells ?? []) {
			count = maxPositiveIndex(count, cell.row);
		}

		for (const row of sheet?.rows ?? []) {
			count = maxPositiveIndex(count, row.index);
		}

		return count;
	}

	function createColumns(count: number): ColumnHeader[] {
		return Array.from({ length: count }, (_, index) => ({
			index: index + 1,
			label: getColLabel(index)
		}));
	}

	function createRows(count: number): number[] {
		return Array.from({ length: count }, (_, index) => index + 1);
	}

	function createCellLookup(sheet: main.WorkbookSheet | undefined): CellLookup {
		const lookup: CellLookup = {};

		for (const cell of sheet?.cells ?? []) {
			if (cell.ref) {
				lookup[cell.ref] = cell;
			}
		}

		return lookup;
	}

	function getCellTitle(ref: string, cell: main.CellData | undefined): string | undefined {
		if (!cell) {
			return undefined;
		}

		if (cell.value) {
			return `${ref}: ${cell.value}`;
		}

		if (cell.hasFormula) {
			return `${ref}: formula cell`;
		}

		return ref;
	}

	const activeCellRef = $derived(view?.activeCell?.ref ?? 'A1');
	const activeColumnLabel = $derived(getColLabel((view?.activeCell?.column ?? 1) - 1));
	const activeRowIndex = $derived(view?.activeCell?.row ?? 1);
	const activeSheetName = $derived(activeSheet?.name ?? 'Sheet 1');
	const columnCount = $derived(getColumnCount(activeSheet, view));
	const rowCount = $derived(getRowCount(activeSheet, view));
	const columns = $derived(createColumns(columnCount));
	const rows = $derived(createRows(rowCount));
	const cellsByRef = $derived(createCellLookup(activeSheet));
	const selectCommandWired = $derived(Boolean(onSelectCell));
	const scrollCommandWired = $derived(Boolean(onSetScrollPosition));
	const metadataLabel = $derived(
		`${activeSheet?.cells?.length ?? 0} loaded cells rendered across ${rowCount} rows and ${columnCount} columns. ${styles.length} styles are available for later grid rendering.`
	);
	const viewportLabel = $derived(
		`Spreadsheet grid for ${activeSheetName}; active cell ${activeCellRef}; displayed workbook values are rendered when loaded.`
	);
</script>

<div
	class="spreadsheet-viewport"
	class:drag-active={dragActive}
	role="img"
	aria-label={viewportLabel}
>
	<div
		class="grid-table"
		aria-hidden="true"
		title={metadataLabel}
		data-select-command-wired={selectCommandWired}
		data-scroll-command-wired={scrollCommandWired}
		style={`--column-count: ${columnCount};`}
	>
		<!-- Top-left corner selector block -->
		<div class="corner-header" aria-hidden="true"></div>

		<!-- Column Headers -->
		{#each columns as column (column.index)}
			<div class="column-header" class:active={column.label === activeColumnLabel}>
				{column.label}
			</div>
		{/each}

		<!-- Row Loop -->
		{#each rows as row (row)}
			<!-- Row Header -->
			<div class="row-header" class:active={row === activeRowIndex}>
				{row}
			</div>

			<!-- Grid Cells for this row -->
			{#each columns as column (column.index)}
				{@const ref = `${column.label}${row}`}
				{@const cell = cellsByRef[ref]}
				<div
					class="grid-cell"
					class:active-cell={ref === activeCellRef}
					class:has-value={Boolean(cell?.value)}
					data-cell-ref={ref}
					data-cell-kind={cell?.kind}
					title={getCellTitle(ref, cell)}
				>
					{#if cell?.value}
						<span class="cell-value">{cell.value}</span>
					{/if}
				</div>
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
		position: relative;
	}

	.spreadsheet-viewport.drag-active {
		outline: 2px solid var(--color-selection-border);
		outline-offset: -2px;
	}

	/* CSS Grid for perfect tabular spreadsheet alignment */
	.grid-table {
		--row-header-width: 40px;
		--column-width: 100px;
		--row-height: 24px;

		display: grid;
		grid-template-columns: var(--row-header-width) repeat(var(--column-count), var(--column-width));
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
		color: var(--color-text);
		display: flex;
		align-items: center;
		min-width: 0;
		overflow: hidden;
		padding: 0 6px;
		white-space: nowrap;
	}

	.grid-cell.has-value {
		user-select: text;
	}

	.cell-value {
		display: block;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
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
