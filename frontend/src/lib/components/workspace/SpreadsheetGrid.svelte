<script lang="ts">
	import { onMount, tick } from 'svelte';
	import type { Attachment } from 'svelte/attachments';
	import { SvelteMap, SvelteSet } from 'svelte/reactivity';
	import type { main } from '$lib/wailsjs/go/models';
	import type { CellEditSession, CellEditSource } from './cellEditSession';

	type ResizeAxis = 'column' | 'row';

	type ResizeCommitCallback = (
		sheetName: string,
		index: number,
		size: number
	) => Promise<void> | void;

	type ResizeSession = {
		axis: ResizeAxis;
		index: number;
		sheetName: string;
		pointerId: number;
		startPointerPosition: number;
		startSizePx: number;
		draftSizePx: number;
	};

	type Props = {
		activeSheet?: main.WorkbookSheet;
		view?: main.WorkbookViewState;
		styles?: main.CellStyle[];
		dragActive?: boolean;
		editSession?: CellEditSession | null;
		editCommitting?: boolean;
		onSelectCell?: (cellRef: string) => Promise<void> | void;
		onBeginCellEdit?: (
			source: CellEditSource,
			sheetName: string,
			cellRef: string,
			value: string
		) => void;
		onUpdateCellEdit?: (
			source: CellEditSource,
			sheetName: string,
			cellRef: string,
			value: string
		) => void;
		onCancelCellEdit?: (sheetName?: string, cellRef?: string) => void;
		onCommitCellEdit?: (sheetName: string, cellRef: string, value: string) => Promise<void> | void;
		onSetColumnWidth?: ResizeCommitCallback;
		onSetRowHeight?: ResizeCommitCallback;
		onSetScrollPosition?: (topRow: number, leftColumn: number) => Promise<void> | void;
	};

	type ColumnHeader = {
		index: number;
		label: string;
	};

	type CellLookup = Record<string, main.CellData>;

	const MIN_COLUMN_COUNT = 40;
	const MIN_ROW_COUNT = 100;
	const COLUMN_WIDTH_TO_PX = 8;
	const ROW_HEIGHT_TO_PX = 1.3;
	const MIN_COLUMN_WIDTH_PX = 24;
	const MIN_ROW_HEIGHT_PX = 18;
	const DEFAULT_COLUMN_WIDTH = 8.43;
	const DEFAULT_ROW_HEIGHT = 15;
	const AUTO_FIT_HORIZONTAL_BUFFER_PX = 12;
	const AUTO_FIT_VERTICAL_BUFFER_PX = 4;
	const RESIZE_COMMIT_EPSILON_PX = 0.5;

	let {
		activeSheet,
		view,
		styles = [],
		dragActive = false,
		editSession,
		editCommitting = false,
		onSelectCell,
		onBeginCellEdit,
		onUpdateCellEdit,
		onCancelCellEdit,
		onCommitCellEdit,
		onSetColumnWidth,
		onSetRowHeight,
		onSetScrollPosition
	}: Props = $props();

	let selectEditorTextOnFocus = true;
	let skipNextEditorBlurCommit = false;
	let resizeSession = $state<ResizeSession | null>(null);
	let resizeEditCommitInProgress = false;
	let gridTableElement: HTMLDivElement | undefined;

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

	function columnWidthToPx(width: number): number {
		return width * COLUMN_WIDTH_TO_PX;
	}

	function columnPxToWidth(widthPx: number): number {
		return widthPx / COLUMN_WIDTH_TO_PX;
	}

	function rowHeightToPx(height: number): number {
		return height * ROW_HEIGHT_TO_PX;
	}

	function rowPxToHeight(heightPx: number): number {
		return heightPx / ROW_HEIGHT_TO_PX;
	}

	function getEffectiveColumnWidth(index: number): number {
		const layout = columnLayoutsByIndex.get(index);
		if (layout) {
			return layout.width;
		}

		return activeSheet?.defaultColumnWidth || DEFAULT_COLUMN_WIDTH;
	}

	function getEffectiveRowHeight(index: number): number {
		const layout = rowLayoutsByIndex.get(index);
		if (layout) {
			return layout.height;
		}

		return activeSheet?.defaultRowHeight || DEFAULT_ROW_HEIGHT;
	}

	function getColumnWidthPx(index: number, includeDraft = true): number {
		const layout = columnLayoutsByIndex.get(index);
		if (layout?.hidden) {
			return 0;
		}

		if (includeDraft && resizeSession?.axis === 'column' && resizeSession.index === index) {
			return resizeSession.draftSizePx;
		}

		return columnWidthToPx(getEffectiveColumnWidth(index));
	}

	function getRowHeightPx(index: number, includeDraft = true): number {
		const layout = rowLayoutsByIndex.get(index);
		if (layout?.hidden) {
			return 0;
		}

		if (includeDraft && resizeSession?.axis === 'row' && resizeSession.index === index) {
			return resizeSession.draftSizePx;
		}

		return rowHeightToPx(getEffectiveRowHeight(index));
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

		const textColor = style.render?.textColor || style.font?.color;

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
		}

		if (textColor) {
			cssParts.push(`color: ${textColor}`);
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

	function getCellDisplayText(
		cell: main.CellData | undefined,
		mergedInfo: MergedInfo | undefined
	): string {
		if (cell?.value) {
			return cell.value;
		}

		return mergedInfo?.value ?? '';
	}

	function getCellEditText(
		cell: main.CellData | undefined,
		mergedInfo: MergedInfo | undefined
	): string {
		if (cell?.hasFormula && cell.formula) {
			return cell.formula;
		}

		return getCellDisplayText(cell, mergedInfo);
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
	const cellEditCommandWired = $derived(Boolean(onCommitCellEdit));
	const scrollCommandWired = $derived(Boolean(onSetScrollPosition));
	const resizeCommandWired = $derived(Boolean(onSetColumnWidth && onSetRowHeight));

	const mergedLookups = $derived(createMergedLookups(activeSheet));
	const mergedCellLookup = $derived(mergedLookups.mergedCellLookup);
	const coveredCells = $derived(mergedLookups.coveredCells);

	const hiddenColumns = $derived(
		new Set((activeSheet?.columns ?? []).filter((col) => col.hidden).map((col) => col.index))
	);

	const hiddenRows = $derived(
		new Set((activeSheet?.rows ?? []).filter((row) => row.hidden).map((row) => row.index))
	);

	const columnLayoutsByIndex = $derived(
		new Map((activeSheet?.columns ?? []).map((column) => [column.index, column]))
	);
	const rowLayoutsByIndex = $derived(
		new Map((activeSheet?.rows ?? []).map((row) => [row.index, row]))
	);

	const colWidthsCss = $derived(columns.map((col) => `${getColumnWidthPx(col.index)}px`).join(' '));

	const rowHeightsCss = $derived(rows.map((row) => `${getRowHeightPx(row)}px`).join(' '));

	const metadataLabel = $derived(
		`${activeSheet?.cells?.length ?? 0} loaded cells rendered across ${rowCount} rows and ${columnCount} columns. ${styles.length} styles are available for later grid rendering.`
	);
	const viewportLabel = $derived(
		`Spreadsheet grid for ${activeSheetName}; active cell ${activeCellRef}; displayed workbook values are rendered when loaded.`
	);

	function canEditCell(ref: string): boolean {
		return (
			Boolean(onBeginCellEdit && onUpdateCellEdit && onCommitCellEdit) &&
			Boolean(activeSheetName) &&
			Boolean(ref) &&
			!editCommitting
		);
	}

	function isDraftCell(ref: string): boolean {
		return editSession?.sheetName === activeSheetName && editSession.cellRef === ref;
	}

	function isInlineEditingCell(ref: string): boolean {
		return isDraftCell(ref) && editSession?.source === 'grid';
	}

	function getLiveCellDisplayText(
		ref: string,
		cell: main.CellData | undefined,
		mergedInfo: MergedInfo | undefined
	): string {
		if (isDraftCell(ref) && editSession) {
			return editSession.value;
		}

		return getCellDisplayText(cell, mergedInfo);
	}

	const focusInlineEditor: Attachment<HTMLInputElement> = (node) => {
		queueMicrotask(() => {
			node.focus();
			if (selectEditorTextOnFocus) {
				node.select();
				return;
			}

			const cursorPosition = node.value.length;
			node.setSelectionRange(cursorPosition, cursorPosition);
		});
	};

	const gridTableAttachment: Attachment<HTMLDivElement> = (node) => {
		gridTableElement = node;

		return () => {
			if (gridTableElement === node) {
				gridTableElement = undefined;
			}
		};
	};

	function beginInlineEdit(
		ref: string,
		cell: main.CellData | undefined,
		mergedInfo: MergedInfo | undefined,
		initialText?: string,
		selectText = true
	): void {
		if (!canEditCell(ref)) {
			return;
		}

		const existingDraft = isDraftCell(ref) && editSession ? editSession.value : undefined;
		selectEditorTextOnFocus = selectText;
		onBeginCellEdit?.(
			'grid',
			activeSheetName,
			ref,
			initialText ?? existingDraft ?? getCellEditText(cell, mergedInfo)
		);
		onSelectCell?.(ref);
	}

	function cancelInlineEdit(): void {
		skipNextEditorBlurCommit = true;
		onCancelCellEdit?.(editSession?.sheetName, editSession?.cellRef);
	}

	async function commitInlineEdit(): Promise<void> {
		if (editCommitting || !editSession || editSession.source !== 'grid') {
			return;
		}

		if (!onCommitCellEdit || !editSession.sheetName || !editSession.cellRef) {
			cancelInlineEdit();
			return;
		}

		const cellRef = editSession.cellRef;
		const sheetName = editSession.sheetName;
		const nextValue = editSession.value;
		const shouldSelectAfterCommit = activeCellRef !== cellRef;
		const originalValue = getCellEditText(cellsByRef[cellRef], mergedCellLookup.get(cellRef));

		if (nextValue === originalValue) {
			onCancelCellEdit?.(sheetName, cellRef);
			return;
		}

		await onCommitCellEdit(sheetName, cellRef, nextValue);
		if (shouldSelectAfterCommit) {
			await onSelectCell?.(cellRef);
		}
	}

	function shouldStartEditFromKey(event: KeyboardEvent): boolean {
		return (
			event.key.length === 1 &&
			event.key !== ' ' &&
			!event.altKey &&
			!event.ctrlKey &&
			!event.metaKey
		);
	}

	function handleCellKeydown(
		event: KeyboardEvent,
		ref: string,
		cell: main.CellData | undefined,
		mergedInfo: MergedInfo | undefined
	): void {
		if (event.key === 'F2') {
			event.preventDefault();
			beginInlineEdit(ref, cell, mergedInfo);
			return;
		}

		if (shouldStartEditFromKey(event)) {
			event.preventDefault();
			beginInlineEdit(ref, cell, mergedInfo, event.key, false);
			return;
		}

		if (event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			onSelectCell?.(ref);
		}
	}

	function handleEditorInput(event: Event): void {
		if (!editSession) {
			return;
		}

		onUpdateCellEdit?.(
			'grid',
			editSession.sheetName,
			editSession.cellRef,
			(event.currentTarget as HTMLInputElement).value
		);
	}

	function handleEditorKeydown(event: KeyboardEvent): void {
		event.stopPropagation();

		if (event.key === 'Enter') {
			event.preventDefault();
			void commitInlineEdit();
			return;
		}

		if (event.key === 'Escape') {
			event.preventDefault();
			cancelInlineEdit();
			(event.currentTarget as HTMLInputElement).blur();
		}
	}

	function handleEditorBlur(): void {
		if (skipNextEditorBlurCommit) {
			skipNextEditorBlurCommit = false;
			return;
		}

		if (resizeEditCommitInProgress) {
			return;
		}

		void commitInlineEdit();
	}

	async function focusInitialActiveCell(): Promise<void> {
		await tick();
		if (document.activeElement && document.activeElement !== document.body) {
			return;
		}

		document
			.querySelector<HTMLElement>('.spreadsheet-viewport .grid-cell.active-cell')
			?.focus({ preventScroll: true });
	}

	onMount(() => {
		void focusInitialActiveCell();
		window.addEventListener('keydown', handleResizeKeydown);

		return () => {
			window.removeEventListener('keydown', handleResizeKeydown);
			resizeSession = null;
		};
	});

	function handleScroll(event: Event) {
		if (!onSetScrollPosition) return;
		const target = event.currentTarget as HTMLDivElement;
		if (!target) return;

		const scrollTop = target.scrollTop;
		const scrollLeft = target.scrollLeft;

		let accumulatedWidth = 0;
		let leftColumn = 1;
		for (const col of columns) {
			const width = getColumnWidthPx(col.index);

			if (accumulatedWidth + width > scrollLeft) {
				leftColumn = col.index;
				break;
			}
			accumulatedWidth += width;
		}

		let accumulatedHeight = 0;
		let topRow = 1;
		for (const row of rows) {
			const height = getRowHeightPx(row);

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

	function stopResizeEvent(event: Event): void {
		event.preventDefault();
		event.stopPropagation();
	}

	function canResizeAxis(axis: ResizeAxis): boolean {
		return (
			Boolean(activeSheet?.name) && Boolean(axis === 'column' ? onSetColumnWidth : onSetRowHeight)
		);
	}

	function isActiveResize(axis: ResizeAxis, index: number): boolean {
		return resizeSession?.axis === axis && resizeSession.index === index;
	}

	function getResizePointerPosition(event: PointerEvent, axis: ResizeAxis): number {
		return axis === 'column' ? event.clientX : event.clientY;
	}

	function getResizeStartSizePx(axis: ResizeAxis, index: number): number {
		return axis === 'column' ? getColumnWidthPx(index, false) : getRowHeightPx(index, false);
	}

	function getResizeMinimumPx(axis: ResizeAxis): number {
		return axis === 'column' ? MIN_COLUMN_WIDTH_PX : MIN_ROW_HEIGHT_PX;
	}

	function getResizeCommitCallback(axis: ResizeAxis): ResizeCommitCallback | undefined {
		return axis === 'column' ? onSetColumnWidth : onSetRowHeight;
	}

	async function commitActiveEditBeforeResize(): Promise<boolean> {
		if (!editSession) {
			return true;
		}

		if (editCommitting || resizeEditCommitInProgress) {
			return false;
		}

		if (!onCommitCellEdit || !editSession.sheetName || !editSession.cellRef) {
			onCancelCellEdit?.(editSession.sheetName, editSession.cellRef);
			return true;
		}

		resizeEditCommitInProgress = true;
		try {
			await onCommitCellEdit(editSession.sheetName, editSession.cellRef, editSession.value);
			return true;
		} finally {
			resizeEditCommitInProgress = false;
		}
	}

	async function handleResizePointerDown(
		event: PointerEvent,
		axis: ResizeAxis,
		index: number
	): Promise<void> {
		stopResizeEvent(event);

		if (event.button !== 0 || event.detail > 1 || resizeSession || !canResizeAxis(axis)) {
			return;
		}

		const handle = event.currentTarget as HTMLElement;
		try {
			handle.setPointerCapture(event.pointerId);
		} catch (error) {
			console.warn('Resize pointer capture failed.', error);
			return;
		}

		const readyToResize = await commitActiveEditBeforeResize();
		if (!readyToResize || !handle.hasPointerCapture(event.pointerId)) {
			if (handle.hasPointerCapture(event.pointerId)) {
				handle.releasePointerCapture(event.pointerId);
			}
			return;
		}

		const startPointerPosition = getResizePointerPosition(event, axis);
		const startSizePx = Math.max(getResizeMinimumPx(axis), getResizeStartSizePx(axis, index));
		resizeSession = {
			axis,
			index,
			sheetName: activeSheet?.name ?? activeSheetName,
			pointerId: event.pointerId,
			startPointerPosition,
			startSizePx,
			draftSizePx: startSizePx
		};
	}

	function handleResizePointerMove(event: PointerEvent): void {
		if (!resizeSession || event.pointerId !== resizeSession.pointerId) {
			return;
		}

		stopResizeEvent(event);
		const pointerPosition = getResizePointerPosition(event, resizeSession.axis);
		const delta = pointerPosition - resizeSession.startPointerPosition;
		const draftSizePx = Math.max(
			getResizeMinimumPx(resizeSession.axis),
			resizeSession.startSizePx + delta
		);
		resizeSession = { ...resizeSession, draftSizePx };
	}

	async function handleResizePointerUp(event: PointerEvent): Promise<void> {
		if (!resizeSession || event.pointerId !== resizeSession.pointerId) {
			return;
		}

		stopResizeEvent(event);
		const session = resizeSession;
		resizeSession = null;

		const handle = event.currentTarget as HTMLElement;
		if (handle.hasPointerCapture(event.pointerId)) {
			handle.releasePointerCapture(event.pointerId);
		}

		const finalSizePx = Math.max(getResizeMinimumPx(session.axis), session.draftSizePx);
		if (Math.abs(finalSizePx - session.startSizePx) < RESIZE_COMMIT_EPSILON_PX) {
			return;
		}

		const commitResize = getResizeCommitCallback(session.axis);
		if (!commitResize) {
			return;
		}

		const workbookSize =
			session.axis === 'column' ? columnPxToWidth(finalSizePx) : rowPxToHeight(finalSizePx);
		await commitResize(session.sheetName, session.index, workbookSize);
	}

	function handleResizePointerCancel(event: PointerEvent): void {
		if (!resizeSession || event.pointerId !== resizeSession.pointerId) {
			return;
		}

		stopResizeEvent(event);
		resizeSession = null;
	}

	function handleResizeLostPointerCapture(event: PointerEvent): void {
		if (resizeSession?.pointerId === event.pointerId) {
			resizeSession = null;
		}
	}

	function handleResizeKeydown(event: KeyboardEvent): void {
		if (event.key !== 'Escape' || !resizeSession) {
			return;
		}

		event.preventDefault();
		resizeSession = null;
	}

	function parseCssPixels(value: string): number {
		const parsed = Number.parseFloat(value);
		return Number.isFinite(parsed) ? parsed : 0;
	}

	function getBoxExtraPixels(element: HTMLElement, axis: ResizeAxis): number {
		const styles = getComputedStyle(element);
		if (axis === 'column') {
			return (
				parseCssPixels(styles.paddingLeft) +
				parseCssPixels(styles.paddingRight) +
				parseCssPixels(styles.borderLeftWidth) +
				parseCssPixels(styles.borderRightWidth)
			);
		}

		return (
			parseCssPixels(styles.paddingTop) +
			parseCssPixels(styles.paddingBottom) +
			parseCssPixels(styles.borderTopWidth) +
			parseCssPixels(styles.borderBottomWidth)
		);
	}

	function measureRangeSize(element: HTMLElement): { width: number; height: number } {
		const range = document.createRange();
		range.selectNodeContents(element);
		const rect = range.getBoundingClientRect();
		range.detach();
		return { width: rect.width, height: rect.height };
	}

	function measureRenderedContent(element: HTMLElement): { width: number; height: number } {
		const content =
			element.querySelector<HTMLElement>('.header-label, .cell-value, .cell-editor') ?? element;
		const usesElementContent = content === element;
		const rangeSize =
			content instanceof HTMLInputElement ? { width: 0, height: 0 } : measureRangeSize(content);
		const contentWidth =
			content instanceof HTMLInputElement
				? content.scrollWidth
				: usesElementContent
					? rangeSize.width
					: Math.max(rangeSize.width, content.scrollWidth);
		const contentHeight =
			content instanceof HTMLInputElement
				? content.scrollHeight
				: usesElementContent
					? rangeSize.height
					: Math.max(rangeSize.height, content.scrollHeight);

		return {
			width: contentWidth + getBoxExtraPixels(element, 'column'),
			height: contentHeight + getBoxExtraPixels(element, 'row')
		};
	}

	function measureColumnAutoFitPx(columnIndex: number): number {
		let maxWidth = 0;
		const measureTargets = gridTableElement?.querySelectorAll<HTMLElement>(
			`.column-header[data-column-index="${columnIndex}"], .grid-cell[data-column-index="${columnIndex}"]`
		);

		for (const element of measureTargets ?? []) {
			maxWidth = Math.max(maxWidth, measureRenderedContent(element).width);
		}

		return Math.max(MIN_COLUMN_WIDTH_PX, Math.ceil(maxWidth + AUTO_FIT_HORIZONTAL_BUFFER_PX));
	}

	function measureRowAutoFitPx(rowIndex: number): number {
		let maxHeight = 0;
		const measureTargets = gridTableElement?.querySelectorAll<HTMLElement>(
			`.row-header[data-row-index="${rowIndex}"], .grid-cell[data-row-index="${rowIndex}"]`
		);

		for (const element of measureTargets ?? []) {
			maxHeight = Math.max(maxHeight, measureRenderedContent(element).height);
		}

		return Math.max(MIN_ROW_HEIGHT_PX, Math.ceil(maxHeight + AUTO_FIT_VERTICAL_BUFFER_PX));
	}

	async function handleResizeDoubleClick(
		event: MouseEvent,
		axis: ResizeAxis,
		index: number
	): Promise<void> {
		stopResizeEvent(event);
		resizeSession = null;

		if (!canResizeAxis(axis)) {
			return;
		}

		const readyToResize = await commitActiveEditBeforeResize();
		if (!readyToResize) {
			return;
		}

		const commitResize = getResizeCommitCallback(axis);
		if (!commitResize) {
			return;
		}

		const fittedSizePx =
			axis === 'column' ? measureColumnAutoFitPx(index) : measureRowAutoFitPx(index);
		const workbookSize =
			axis === 'column' ? columnPxToWidth(fittedSizePx) : rowPxToHeight(fittedSizePx);
		await commitResize(activeSheet?.name ?? activeSheetName, index, workbookSize);
	}
</script>

<div
	class="spreadsheet-viewport"
	class:drag-active={dragActive}
	class:resize-active={resizeSession !== null}
	class:resizing-column={resizeSession?.axis === 'column'}
	class:resizing-row={resizeSession?.axis === 'row'}
	role="region"
	aria-label={viewportLabel}
	onscroll={handleScroll}
>
	<div
		{@attach gridTableAttachment}
		class="grid-table"
		role="grid"
		aria-rowcount={rowCount + 1}
		aria-colcount={columnCount + 1}
		title={metadataLabel}
		data-select-command-wired={selectCommandWired}
		data-cell-edit-command-wired={cellEditCommandWired}
		data-scroll-command-wired={scrollCommandWired}
		data-resize-command-wired={resizeCommandWired}
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
					data-column-index={column.index}
				>
					<span class="header-label">{column.label}</span>
					<button
						type="button"
						class="resize-handle resize-handle--column"
						class:active={isActiveResize('column', column.index)}
						tabindex="-1"
						aria-hidden="true"
						onclick={stopResizeEvent}
						ondblclick={(event) => void handleResizeDoubleClick(event, 'column', column.index)}
						onpointerdown={(event) => void handleResizePointerDown(event, 'column', column.index)}
						onpointermove={handleResizePointerMove}
						onpointerup={(event) => void handleResizePointerUp(event)}
						onpointercancel={handleResizePointerCancel}
						onlostpointercapture={handleResizeLostPointerCapture}
					></button>
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
					data-row-index={row}
				>
					<span class="header-label">{row}</span>
					<button
						type="button"
						class="resize-handle resize-handle--row"
						class:active={isActiveResize('row', row)}
						tabindex="-1"
						aria-hidden="true"
						onclick={stopResizeEvent}
						ondblclick={(event) => void handleResizeDoubleClick(event, 'row', row)}
						onpointerdown={(event) => void handleResizePointerDown(event, 'row', row)}
						onpointermove={handleResizePointerMove}
						onpointerup={(event) => void handleResizePointerUp(event)}
						onpointercancel={handleResizePointerCancel}
						onlostpointercapture={handleResizeLostPointerCapture}
					></button>
				</div>
			{/if}

			<!-- Grid Cells for this row -->
			{#each columns as column (column.index)}
				{@const ref = `${column.label}${row}`}
				{#if !coveredCells.has(ref) && !hiddenRows.has(row) && !hiddenColumns.has(column.index)}
					{@const cell = cellsByRef[ref]}
					{@const mergedInfo = mergedCellLookup.get(ref)}
					{@const displayText = getLiveCellDisplayText(ref, cell, mergedInfo)}
					<div
						class="grid-cell"
						class:active-cell={ref === activeCellRef}
						class:selected-cell={isCellSelected(row, column.index) && ref !== activeCellRef}
						class:has-value={Boolean(displayText)}
						class:editing-cell={isInlineEditingCell(ref)}
						data-cell-ref={ref}
						data-cell-kind={cell?.kind}
						data-row-index={row}
						data-column-index={column.index}
						title={getGridCellTitle(ref, cell, mergedInfo)}
						onclick={() => onSelectCell?.(ref)}
						ondblclick={(event) => {
							event.preventDefault();
							beginInlineEdit(ref, cell, mergedInfo);
						}}
						onkeydown={(event) => handleCellKeydown(event, ref, cell, mergedInfo)}
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
						{#if isInlineEditingCell(ref)}
							<input
								{@attach focusInlineEditor}
								class="cell-editor"
								type="text"
								aria-label={`Edit ${ref} in ${activeSheetName}`}
								value={editSession?.value ?? ''}
								oninput={handleEditorInput}
								onkeydown={handleEditorKeydown}
								onclick={(event) => event.stopPropagation()}
								ondblclick={(event) => event.stopPropagation()}
								onblur={handleEditorBlur}
							/>
						{:else if displayText}
							<span class="cell-value">{displayText}</span>
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

	.spreadsheet-viewport.resize-active {
		user-select: none;
	}

	.spreadsheet-viewport.resizing-column,
	.spreadsheet-viewport.resizing-column * {
		cursor: col-resize;
	}

	.spreadsheet-viewport.resizing-row,
	.spreadsheet-viewport.resizing-row * {
		cursor: row-resize;
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

	.header-label {
		pointer-events: none;
	}

	.resize-handle {
		appearance: none;
		position: absolute;
		z-index: 7;
		margin: 0;
		padding: 0;
		border: 0;
		border-radius: 0;
		background: transparent;
		box-sizing: border-box;
		touch-action: none;
		user-select: none;
	}

	.resize-handle:focus {
		outline: none;
	}

	.resize-handle::after {
		content: '';
		position: absolute;
		background-color: var(--color-selection-border);
		opacity: 0;
		pointer-events: none;
		transition: opacity 0.1s ease;
	}

	.resize-handle--column {
		top: 0;
		right: -4px;
		width: 8px;
		height: 100%;
		cursor: col-resize;
	}

	.resize-handle--column::after {
		top: 3px;
		bottom: 3px;
		left: 3px;
		width: 2px;
	}

	.resize-handle--row {
		left: 0;
		right: 0;
		bottom: -4px;
		height: 8px;
		cursor: row-resize;
	}

	.resize-handle--row::after {
		top: 3px;
		left: 3px;
		right: 3px;
		height: 2px;
	}

	.column-header:hover .resize-handle--column::after,
	.row-header:hover .resize-handle--row::after,
	.resize-handle.active::after {
		opacity: 1;
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

	.grid-cell.editing-cell {
		padding: 0;
		position: relative;
		z-index: 8;
		box-shadow: inset 0 0 0 2px var(--color-selection-border);
	}

	.cell-editor {
		width: 100%;
		height: 100%;
		min-width: 0;
		padding: 0 6px;
		background-color: transparent;
		border: none;
		color: inherit;
		font: inherit;
		outline: none;
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
