<script lang="ts">
	import type { main } from '$lib/wailsjs/go/models';
	import type { CellEditSession, CellEditSource } from './cellEditSession';

	type CellEditCommitCallback = (
		sheetName: string,
		cellRef: string,
		value: string
	) => Promise<boolean | void> | boolean | void;

	type ReturnFocusCallback = () => Promise<void> | void;

	type Props = {
		view?: main.WorkbookViewState;
		activeCell?: main.CellData;
		editSession?: CellEditSession | null;
		editCommitting?: boolean;
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
		onCommitCellEdit?: CellEditCommitCallback;
		onReturnFocusToCell?: ReturnFocusCallback;
	};

	let {
		view,
		activeCell,
		editSession,
		editCommitting = false,
		onBeginCellEdit,
		onUpdateCellEdit,
		onCancelCellEdit,
		onCommitCellEdit,
		onReturnFocusToCell
	}: Props = $props();
	let skipNextBlurCommit = false;

	const activeCellRef = $derived(view?.activeCell?.ref ?? '');
	const activeSheetName = $derived(view?.activeSheetName ?? '');
	const formulaText = $derived(
		activeCell?.hasFormula && activeCell.formula ? activeCell.formula : (activeCell?.value ?? '')
	);
	const activeEditSession = $derived(
		editSession?.sheetName === activeSheetName && editSession.cellRef === activeCellRef
			? editSession
			: null
	);
	const activeCellTitle = $derived(
		activeCellRef ? `Current cell (${activeCellRef})` : 'Current cell'
	);
	const canEdit = $derived(
		Boolean(onBeginCellEdit && onUpdateCellEdit && onCommitCellEdit) &&
			Boolean(activeSheetName) &&
			Boolean(activeCellRef) &&
			!editCommitting
	);
	const inputText = $derived(activeEditSession ? activeEditSession.value : formulaText);
	const formulaTitle = $derived(
		canEdit
			? activeCellRef
				? `Formula bar for ${activeCellRef}`
				: 'Formula bar'
			: activeCellRef
				? `Formula bar for ${activeCellRef} is unavailable.`
				: 'Formula bar is unavailable.'
	);

	function beginEditing(): void {
		if (!canEdit) {
			return;
		}

		onBeginCellEdit?.('formula', activeSheetName, activeCellRef, inputText);
	}

	function cancelEdit(): void {
		skipNextBlurCommit = true;
		onCancelCellEdit?.(activeSheetName, activeCellRef);
	}

	async function commitEdit(): Promise<boolean> {
		if (editCommitting || !activeEditSession) {
			return false;
		}

		if (!onCommitCellEdit || !activeSheetName || !activeCellRef) {
			onCancelCellEdit?.(activeSheetName, activeCellRef);
			return false;
		}

		const nextValue = activeEditSession.value;

		if (nextValue === formulaText) {
			onCancelCellEdit?.(activeSheetName, activeCellRef);
			return true;
		}

		try {
			const commitResult = await onCommitCellEdit(activeSheetName, activeCellRef, nextValue);
			return commitResult !== false;
		} catch (error) {
			console.warn('Formula bar edit commit failed.', error);
			return false;
		}
	}

	function handleFormulaInput(event: Event): void {
		onUpdateCellEdit?.(
			'formula',
			activeSheetName,
			activeCellRef,
			(event.currentTarget as HTMLInputElement).value
		);
	}

	async function commitEditAndReturnFocus(input: HTMLInputElement): Promise<void> {
		skipNextBlurCommit = true;
		const committed = await commitEdit();
		if (!committed) {
			skipNextBlurCommit = false;
			return;
		}

		input.blur();
		await onReturnFocusToCell?.();
	}

	async function cancelEditAndReturnFocus(input: HTMLInputElement): Promise<void> {
		cancelEdit();
		input.blur();
		await onReturnFocusToCell?.();
	}

	function handleFormulaKeydown(event: KeyboardEvent): void {
		if (event.key === 'Enter') {
			event.preventDefault();
			void commitEditAndReturnFocus(event.currentTarget as HTMLInputElement);
			return;
		}

		if (event.key === 'Escape') {
			event.preventDefault();
			void cancelEditAndReturnFocus(event.currentTarget as HTMLInputElement);
		}
	}

	function handleFormulaBlur(): void {
		if (skipNextBlurCommit) {
			skipNextBlurCommit = false;
			return;
		}

		void commitEdit();
	}
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

	<input
		class="formula-display"
		type="text"
		disabled={!canEdit}
		aria-label={`Formula bar input for ${activeCellRef || 'selected cell'}`}
		title={formulaTitle}
		value={inputText}
		onfocus={beginEditing}
		oninput={handleFormulaInput}
		onkeydown={handleFormulaKeydown}
		onblur={handleFormulaBlur}
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
		color: var(--color-text-muted);
		user-select: none;
		flex-shrink: 0;
	}

	.formula-display {
		flex: 1;
		height: 22px;
		background-color: var(--color-surface);
		border: 1px solid var(--color-border);
		color: var(--color-text);
		cursor: text;
		box-sizing: border-box;
		padding: 0 6px;
		user-select: text;
	}

	.formula-display:focus {
		border-color: var(--color-selection-border);
	}

	.formula-display:disabled {
		background-color: var(--color-disabled-bg);
		color: var(--color-disabled-text);
		cursor: default;
	}
</style>
