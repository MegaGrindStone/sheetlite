<script lang="ts">
	import { onMount } from 'svelte';
	import type { Attachment } from 'svelte/attachments';
	import {
		applyAppearanceAttributes,
		detectSystemTheme,
		readPersistedAppearanceMode,
		subscribeToSystemThemeChanges,
		writePersistedAppearanceMode,
		type AppearanceMode
	} from '$lib/appearance.svelte';
	import {
		InitializeAppearance,
		OpenDroppedFiles,
		OpenWorkbook,
		SaveWorkbook,
		SaveWorkbookAs,
		SelectCell,
		SetActiveSheet,
		SetAppearanceMode,
		SetCellValue,
		SetScrollPosition,
		SetSystemTheme,
		SetZoom
	} from '$lib/wailsjs/go/main/App';
	import type { main } from '$lib/wailsjs/go/models';
	import { OnFileDrop, OnFileDropOff, Quit } from '$lib/wailsjs/runtime/runtime';
	import TopChrome from './TopChrome.svelte';
	import FormulaBar from './FormulaBar.svelte';
	import SideRail from './SideRail.svelte';
	import BottomBar from './BottomBar.svelte';
	import SpreadsheetGrid from './SpreadsheetGrid.svelte';
	import type { CellEditSession, CellEditSource } from './cellEditSession';

	type StateCommand = () => Promise<main.AppState>;

	let appState = $state<main.AppState | null>(null);
	let isDragOver = $state(false);
	let cellEditSession = $state<CellEditSession | null>(null);
	let cellEditCommitting = $state(false);
	let isMounted = false;
	let dragDepth = 0;

	const workbook = $derived(appState?.workbook);
	const view = $derived(appState?.view);
	const status = $derived(appState?.status);
	const appearance = $derived(appState?.appearance);
	const activeSheet = $derived(
		workbook?.sheets?.find((sheet) => sheet.name === view?.activeSheetName) ?? workbook?.sheets?.[0]
	);
	const activeCell = $derived(
		activeSheet?.cells?.find((cell) => cell.ref === view?.activeCell?.ref)
	);

	function acceptSnapshot(nextState: main.AppState): void {
		if (!isMounted) {
			return;
		}

		appState = nextState;
		applyAppearanceAttributes(nextState.appearance);
	}

	async function runStateCommand(command: StateCommand): Promise<main.AppState | null> {
		try {
			const nextState = await command();
			acceptSnapshot(nextState);
			return nextState;
		} catch (error) {
			console.warn('Wails state command failed.', error);
			return null;
		}
	}

	async function updateSnapshot(command: StateCommand): Promise<void> {
		await runStateCommand(command);
	}

	function openWorkbook(): Promise<void> {
		return updateSnapshot(() => OpenWorkbook());
	}

	function saveWorkbook(): Promise<void> {
		return updateSnapshot(() => SaveWorkbook());
	}

	function saveWorkbookAs(): Promise<void> {
		return updateSnapshot(() => SaveWorkbookAs());
	}

	function selectCell(cellRef: string): Promise<void> {
		return updateSnapshot(() => SelectCell(cellRef));
	}

	function setCellValue(sheetName: string, cellRef: string, value: string): Promise<void> {
		return updateSnapshot(() => SetCellValue(sheetName, cellRef, value));
	}

	function isCurrentEditSession(sheetName: string, cellRef: string): boolean {
		return cellEditSession?.sheetName === sheetName && cellEditSession.cellRef === cellRef;
	}

	function beginCellEdit(
		source: CellEditSource,
		sheetName: string,
		cellRef: string,
		value: string
	): void {
		if (cellEditCommitting || !sheetName || !cellRef) {
			return;
		}

		cellEditSession = { source, sheetName, cellRef, value };
	}

	function updateCellEdit(
		source: CellEditSource,
		sheetName: string,
		cellRef: string,
		value: string
	): void {
		if (cellEditCommitting || !sheetName || !cellRef) {
			return;
		}

		if (isCurrentEditSession(sheetName, cellRef) && cellEditSession) {
			cellEditSession = { ...cellEditSession, value };
			return;
		}

		beginCellEdit(source, sheetName, cellRef, value);
	}

	function cancelCellEdit(sheetName?: string, cellRef?: string): void {
		if (!sheetName || !cellRef || isCurrentEditSession(sheetName, cellRef)) {
			cellEditSession = null;
		}
	}

	async function commitCellEdit(sheetName: string, cellRef: string, value: string): Promise<void> {
		if (cellEditCommitting || !sheetName || !cellRef) {
			return;
		}

		cellEditCommitting = true;
		try {
			await setCellValue(sheetName, cellRef, value);
			cancelCellEdit(sheetName, cellRef);
		} finally {
			cellEditCommitting = false;
		}
	}

	function setActiveSheet(sheetName: string): Promise<void> {
		return updateSnapshot(() => SetActiveSheet(sheetName));
	}

	function setScrollPosition(topRow: number, leftColumn: number): Promise<void> {
		return updateSnapshot(() => SetScrollPosition(topRow, leftColumn));
	}

	function setZoom(percent: number): Promise<void> {
		return updateSnapshot(() => SetZoom(percent));
	}

	async function setAppearanceMode(mode: AppearanceMode): Promise<void> {
		const nextState = await runStateCommand(() => SetAppearanceMode(mode));
		if (nextState) {
			writePersistedAppearanceMode(mode);
		}
	}

	function resetDragState(): void {
		dragDepth = 0;
		isDragOver = false;
	}

	function handleDragEnter(event: DragEvent): void {
		event.preventDefault();
		dragDepth += 1;
		isDragOver = true;
	}

	function handleDragOver(event: DragEvent): void {
		event.preventDefault();
		isDragOver = true;
	}

	function handleDragLeave(event: DragEvent): void {
		event.preventDefault();
		dragDepth = Math.max(0, dragDepth - 1);
		if (dragDepth === 0) {
			isDragOver = false;
		}
	}

	function handleDomDrop(event: DragEvent): void {
		event.preventDefault();
		resetDragState();
	}

	function handleDroppedPaths(paths: string[]): void {
		resetDragState();
		void updateSnapshot(() => OpenDroppedFiles(paths));
	}

	const dragAffordance: Attachment<HTMLDivElement> = (node) => {
		node.addEventListener('dragenter', handleDragEnter);
		node.addEventListener('dragover', handleDragOver);
		node.addEventListener('dragleave', handleDragLeave);
		node.addEventListener('drop', handleDomDrop);

		return () => {
			node.removeEventListener('dragenter', handleDragEnter);
			node.removeEventListener('dragover', handleDragOver);
			node.removeEventListener('dragleave', handleDragLeave);
			node.removeEventListener('drop', handleDomDrop);
		};
	};

	onMount(() => {
		isMounted = true;
		void updateSnapshot(() =>
			InitializeAppearance(readPersistedAppearanceMode(), detectSystemTheme())
		);
		const unsubscribeSystemTheme = subscribeToSystemThemeChanges((theme) => {
			void updateSnapshot(() => SetSystemTheme(theme));
		});

		let fileDropRegistered = false;
		try {
			OnFileDrop((_x, _y, paths) => handleDroppedPaths(paths), true);
			fileDropRegistered = true;
		} catch (error) {
			console.warn('Wails file-drop handler is not available.', error);
		}

		return () => {
			isMounted = false;
			unsubscribeSystemTheme();
			resetDragState();

			if (fileDropRegistered) {
				try {
					OnFileDropOff();
				} catch (error) {
					console.warn('Wails file-drop cleanup failed.', error);
				}
			}
		};
	});
</script>

<div
	{@attach dragAffordance}
	class="workspace-shell --wails-drop-target"
	class:drag-over={isDragOver}
>
	<!-- Top Chrome -->
	<header class="top-chrome" aria-label="Top chrome">
		<TopChrome
			{workbook}
			{status}
			{appearance}
			onOpenWorkbook={openWorkbook}
			onSaveWorkbook={saveWorkbook}
			onSaveWorkbookAs={saveWorkbookAs}
			onExitApp={Quit}
			onSetAppearanceMode={setAppearanceMode}
		/>
	</header>

	<!-- Formula/Control Strip -->
	<section class="formula-strip" aria-label="Formula bar">
		<FormulaBar
			{view}
			{activeCell}
			editSession={cellEditSession}
			editCommitting={cellEditCommitting}
			onBeginCellEdit={beginCellEdit}
			onUpdateCellEdit={updateCellEdit}
			onCancelCellEdit={cancelCellEdit}
			onCommitCellEdit={commitCellEdit}
		/>
	</section>

	<!-- Left Rail -->
	<div class="left-rail-container">
		<SideRail side="left" />
	</div>

	<!-- Grid Canvas Region -->
	<main class="grid-canvas" aria-label="Grid canvas">
		<SpreadsheetGrid
			{activeSheet}
			{view}
			styles={workbook?.styles ?? []}
			dragActive={isDragOver}
			editSession={cellEditSession}
			editCommitting={cellEditCommitting}
			onSelectCell={selectCell}
			onBeginCellEdit={beginCellEdit}
			onUpdateCellEdit={updateCellEdit}
			onCancelCellEdit={cancelCellEdit}
			onCommitCellEdit={commitCellEdit}
			onSetScrollPosition={setScrollPosition}
		/>
	</main>

	<!-- Right Rail -->
	<div class="right-rail-container">
		<SideRail side="right" />
	</div>

	<!-- Bottom Bar -->
	<footer class="bottom-bar" aria-label="Bottom bar">
		<BottomBar {workbook} {view} {status} onSetActiveSheet={setActiveSheet} onSetZoom={setZoom} />
	</footer>
</div>

<style>
	/* Full-window viewport locked CSS Grid Layout */
	.workspace-shell {
		--wails-drop-target: drop;
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
		position: relative;
		z-index: 20;
		overflow: visible;
		background-color: var(--color-chrome);
		border-bottom: 1px solid var(--color-border);
	}

	/* Formula/Control Strip Region */
	.formula-strip {
		grid-area: formula-strip;
		position: relative;
		z-index: 10;
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
