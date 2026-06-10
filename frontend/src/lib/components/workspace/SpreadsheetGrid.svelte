<script lang="ts">
	import { SvelteMap, SvelteSet } from 'svelte/reactivity';
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

	type MergedInfo = {
		range: main.CellRange;
		value: string;
		rowSpan: number;
		colSpan: number;
	};

	function createMergedLookups(sheet: main.WorkbookSheet | undefined) {
		const mergedCellLookup = new SvelteMap<string, MergedInfo>();
		const coveredCells = new SvelteSet<string>();

		if (!sheet || !sheet.mergedCells) {
			return { mergedCellLookup, coveredCells };
		}

		for (const m of sheet.mergedCells) {
			const start = m.range.start;
			const end = m.range.end;
			if (!start || !end) continue;

			const rowSpan = end.row - start.row + 1;
			const colSpan = end.column - start.column + 1;

			const tlRef = start.ref;
			mergedCellLookup.set(tlRef, {
				range: m.range,
				value: m.value,
				rowSpan,
				colSpan
			});

			for (let r = start.row; r <= end.row; r++) {
				for (let c = start.column; c <= end.column; c++) {
					const colLabel = getColLabel(c - 1);
					const ref = `${colLabel}${r}`;
					if (ref !== tlRef) {
						coveredCells.add(ref);
					}
				}
			}
		}

		return { mergedCellLookup, coveredCells };
	}

	const stylesMap = $derived(new Map(styles.map((s) => [s.id, s])));

	function getCellStyles(cell: main.CellData | undefined): string {
		if (!cell) return '';
		const style = stylesMap.get(cell.styleId);
		if (!style) return '';

		const cssParts: string[] = [];

		if (style.font) {
			if (style.font.family) {
				cssParts.push(`font-family: "${style.font.family}", sans-serif`);
			}
			if (style.font.size) {
				cssParts.push(`font-size: ${style.font.size}pt`);
			}
			if (style.font.bold) {
				cssParts.push('font-weight: bold');
			}
			if (style.font.italic) {
				cssParts.push('font-style: italic');
			}

			let textDecoration = '';
			if (style.font.underline) {
				textDecoration += 'underline ';
			}
			if (style.font.strikethrough) {
				textDecoration += 'line-through ';
			}
			if (textDecoration) {
				cssParts.push(`text-decoration: ${textDecoration.trim()}`);
			}
			if (style.font.color) {
				cssParts.push(`color: ${style.font.color}`);
			}
		}

		if (style.fill) {
			if (style.fill.color) {
				cssParts.push(`background-color: ${style.fill.color}`);
			}
		}

		if (style.alignment) {
			const hAlign = style.alignment.horizontal;
			if (hAlign && hAlign !== 'general') {
				let flexAlign = '';
				let textAlign = '';
				if (hAlign === 'left') {
					flexAlign = 'flex-start';
					textAlign = 'left';
				} else if (hAlign === 'right') {
					flexAlign = 'flex-end';
					textAlign = 'right';
				} else if (hAlign === 'center' || hAlign === 'centerContinuous') {
					flexAlign = 'center';
					textAlign = 'center';
				} else if (hAlign === 'justify') {
					flexAlign = 'space-between';
					textAlign = 'justify';
				}
				if (flexAlign) {
					cssParts.push(`justify-content: ${flexAlign}`);
				}
				if (textAlign) {
					cssParts.push(`text-align: ${textAlign}`);
				}
			}

			const vAlign = style.alignment.vertical;
			if (vAlign && vAlign !== 'general') {
				let flexVAlign = '';
				if (vAlign === 'top') {
					flexVAlign = 'flex-start';
				} else if (vAlign === 'center') {
					flexVAlign = 'center';
				} else if (vAlign === 'bottom') {
					flexVAlign = 'flex-end';
				}
				if (flexVAlign) {
					cssParts.push(`align-items: ${flexVAlign}`);
				}
			}

			if (style.alignment.wrapText) {
				cssParts.push('white-space: normal');
				cssParts.push('word-break: break-all');
			}
		}

		if (style.borders) {
			for (const border of style.borders) {
				const side = border.side;
				const bStyle = border.style;
				if (bStyle > 0) {
					const color = border.color || 'currentColor';
					const width = '1px';
					if (side === 'left') cssParts.push(`border-left: ${width} solid ${color}`);
					if (side === 'right') cssParts.push(`border-right: ${width} solid ${color}`);
					if (side === 'top') cssParts.push(`border-top: ${width} solid ${color}`);
					if (side === 'bottom') cssParts.push(`border-bottom: ${width} solid ${color}`);
				}
			}
		}

		return cssParts.join('; ');
	}

	function isCellSelected(row: number, col: number): boolean {
		if (!view?.selection?.start || !view?.selection?.end) return false;
		const start = view.selection.start;
		const end = view.selection.end;

		const minRow = Math.min(start.row, end.row);
		const maxRow = Math.max(start.row, end.row);
		const minCol = Math.min(start.column, end.column);
		const maxCol = Math.max(start.column, end.column);

		return row >= minRow && row <= maxRow && col >= minCol && col <= maxCol;
	}

	function getGridCellTitle(
		ref: string,
		cell: main.CellData | undefined,
		mergedInfo: MergedInfo | undefined
	): string {
		if (cell) {
			if (cell.value) {
				return `${ref}: ${cell.value}`;
			}
			if (cell.hasFormula) {
				return `${ref}: formula cell`;
			}
		}
		if (mergedInfo) {
			if (mergedInfo.value) {
				return `${ref}: ${mergedInfo.value} (merged)`;
			}
			return `${ref}: merged cells`;
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

	const mergedLookups = $derived(createMergedLookups(activeSheet));
	const mergedCellLookup = $derived(mergedLookups.mergedCellLookup);
	const coveredCells = $derived(mergedLookups.coveredCells);

	const hiddenColumns = $derived(
		new Set((activeSheet?.columns ?? []).filter((col) => col.hidden).map((col) => col.index))
	);

	const hiddenRows = $derived(
		new Set((activeSheet?.rows ?? []).filter((row) => row.hidden).map((row) => row.index))
	);

	const colWidthsCss = $derived(
		columns
			.map((col) => {
				const layout = activeSheet?.columns?.find((c) => c.index === col.index);
				if (layout) {
					if (layout.hidden) return '0px';
					return `${layout.width * 8}px`;
				}
				const defWidth = activeSheet?.defaultColumnWidth || 8.43;
				return `${defWidth * 8}px`;
			})
			.join(' ')
	);

	const rowHeightsCss = $derived(
		rows
			.map((row) => {
				const layout = activeSheet?.rows?.find((r) => r.index === row);
				if (layout) {
					if (layout.hidden) return '0px';
					return `${layout.height * 1.3}px`;
				}
				const defHeight = activeSheet?.defaultRowHeight || 15;
				return `${defHeight * 1.3}px`;
			})
			.join(' ')
	);

	const metadataLabel = $derived(
		`${activeSheet?.cells?.length ?? 0} loaded cells rendered across ${rowCount} rows and ${columnCount} columns. ${styles.length} styles are available for later grid rendering.`
	);
	const viewportLabel = $derived(
		`Spreadsheet grid for ${activeSheetName}; active cell ${activeCellRef}; displayed workbook values are rendered when loaded.`
	);

	function handleScroll(event: Event) {
		if (!onSetScrollPosition) return;
		const target = event.currentTarget as HTMLDivElement;
		if (!target) return;

		const scrollTop = target.scrollTop;
		const scrollLeft = target.scrollLeft;

		let accumulatedWidth = 0;
		let leftColumn = 1;
		for (const col of columns) {
			const layout = activeSheet?.columns?.find((c) => c.index === col.index);
			const width = layout
				? layout.hidden
					? 0
					: layout.width * 8
				: (activeSheet?.defaultColumnWidth || 8.43) * 8;

			if (accumulatedWidth + width > scrollLeft) {
				leftColumn = col.index;
				break;
			}
			accumulatedWidth += width;
		}

		let accumulatedHeight = 0;
		let topRow = 1;
		for (const row of rows) {
			const layout = activeSheet?.rows?.find((r) => r.index === row);
			const height = layout
				? layout.hidden
					? 0
					: layout.height * 1.3
				: (activeSheet?.defaultRowHeight || 15) * 1.3;

			if (accumulatedHeight + height > scrollTop) {
				topRow = row;
				break;
			}
			accumulatedHeight += height;
		}

		if (view?.scroll?.topRow !== topRow || view?.scroll?.leftColumn !== leftColumn) {
			onSetScrollPosition(topRow, leftColumn);
		}
	}
</script>

<div
	class="spreadsheet-viewport"
	class:drag-active={dragActive}
	role="region"
	aria-label={viewportLabel}
	onscroll={handleScroll}
>
	<div
		class="grid-table"
		role="grid"
		aria-rowcount={rowCount + 1}
		aria-colcount={columnCount + 1}
		title={metadataLabel}
		data-select-command-wired={selectCommandWired}
		data-scroll-command-wired={scrollCommandWired}
		style="
			grid-template-columns: var(--row-header-width, 40px) {colWidthsCss};
			grid-template-rows: var(--header-row-height, 24px) {rowHeightsCss};
		"
	>
		<!-- Top-left corner selector block -->
		<div class="corner-header" aria-hidden="true" style="grid-row: 1; grid-column: 1;"></div>

		<!-- Column Headers -->
		{#each columns as column (column.index)}
			{#if !hiddenColumns.has(column.index)}
				<div
					class="column-header"
					class:active={column.label === activeColumnLabel}
					style="grid-row: 1; grid-column: {column.index + 1};"
					role="columnheader"
				>
					{column.label}
				</div>
			{/if}
		{/each}

		<!-- Row Loop -->
		{#each rows as row (row)}
			<!-- Row Header -->
			{#if !hiddenRows.has(row)}
				<div
					class="row-header"
					class:active={row === activeRowIndex}
					style="grid-row: {row + 1}; grid-column: 1;"
					role="rowheader"
				>
					{row}
				</div>
			{/if}

			<!-- Grid Cells for this row -->
			{#each columns as column (column.index)}
				{@const ref = `${column.label}${row}`}
				{#if !coveredCells.has(ref) && !hiddenRows.has(row) && !hiddenColumns.has(column.index)}
					{@const cell = cellsByRef[ref]}
					{@const mergedInfo = mergedCellLookup.get(ref)}
					<div
						class="grid-cell"
						class:active-cell={ref === activeCellRef}
						class:selected-cell={isCellSelected(row, column.index) && ref !== activeCellRef}
						class:has-value={Boolean(cell?.value || mergedInfo?.value)}
						data-cell-ref={ref}
						data-cell-kind={cell?.kind}
						title={getGridCellTitle(ref, cell, mergedInfo)}
						onclick={() => onSelectCell?.(ref)}
						onkeydown={(e) => {
							if (e.key === 'Enter' || e.key === ' ') {
								e.preventDefault();
								onSelectCell?.(ref);
							}
						}}
						role="gridcell"
						tabindex="0"
						aria-selected={isCellSelected(row, column.index)}
						aria-rowindex={row + 1}
						aria-colindex={column.index + 1}
						style="
							grid-row: {mergedInfo ? `${row + 1} / span ${mergedInfo.rowSpan}` : `${row + 1}`};
							grid-column: {mergedInfo
							? `${column.index + 1} / span ${mergedInfo.colSpan}`
							: `${column.index + 1}`};
							{getCellStyles(cell)}
						"
					>
						{#if cell?.value}
							<span class="cell-value">{cell.value}</span>
						{:else if mergedInfo?.value}
							<span class="cell-value">{mergedInfo.value}</span>
						{/if}
					</div>
				{/if}
			{/each}
		{/each}

		<!-- Active Selection range outline -->
		{#if view?.selection?.start && view?.selection?.end}
			{@const start = view.selection.start}
			{@const end = view.selection.end}
			{@const minRow = Math.min(start.row, end.row)}
			{@const maxRow = Math.max(start.row, end.row)}
			{@const minCol = Math.min(start.column, end.column)}
			{@const maxCol = Math.max(start.column, end.column)}
			<div
				class="selection-outline"
				aria-hidden="true"
				style="
					grid-row: {minRow + 1} / {maxRow + 2};
					grid-column: {minCol + 1} / {maxCol + 2};
				"
			></div>
		{/if}
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
		--header-row-height: 24px;

		display: grid;
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
		background-color: var(--color-surface);
		box-shadow: inset 0 0 0 2px var(--color-selection-border);
		/* Keep selection outline visible above surrounding cells */
		z-index: 1;
		position: relative;
	}

	.grid-cell.selected-cell {
		background-color: var(--color-selection-bg);
	}

	.selection-outline {
		border: 2px solid var(--color-selection-border);
		pointer-events: none;
		z-index: 6;
		box-sizing: border-box;
	}
</style>
