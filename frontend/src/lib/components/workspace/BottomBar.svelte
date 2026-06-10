<script lang="ts">
	import type { main } from '$lib/wailsjs/go/models';

	type Props = {
		workbook?: main.WorkbookState;
		view?: main.WorkbookViewState;
		status?: main.AppStatus;
		onSetActiveSheet?: (sheetName: string) => Promise<void> | void;
		onSetZoom?: (percent: number) => Promise<void> | void;
	};

	let { workbook, view, status, onSetActiveSheet, onSetZoom }: Props = $props();

	const sheets = $derived(workbook?.sheets ?? []);
	const activeSheetName = $derived(view?.activeSheetName || sheets[0]?.name || '');
	const sheetCommandWired = $derived(Boolean(onSetActiveSheet));
	const zoomCommandWired = $derived(Boolean(onSetZoom));
	const statusKind = $derived(status?.kind ?? '');
	const statusText = $derived(status?.message?.trim() ?? '');
	const statusTitle = $derived(
		statusText ? (statusKind ? `${statusKind}: ${statusText}` : statusText) : 'Status'
	);
	const selectionRef = $derived(view?.selection?.ref || view?.activeCell?.ref || '');
	const zoomPercent = $derived(view?.zoomPercent);
	const zoomText = $derived(zoomPercent == null ? '' : `${zoomPercent}%`);
	const zoomTitle = $derived(
		zoomCommandWired
			? 'Zoom readout from app state; controls are inactive.'
			: 'Zoom controls are inactive.'
	);

	function isActiveSheet(sheetName: string): boolean {
		return sheetName === activeSheetName;
	}

	function sheetTitle(sheet: main.WorkbookSheet): string {
		if (isActiveSheet(sheet.name)) {
			return `Active sheet: ${sheet.name}`;
		}

		return sheetCommandWired ? `Switch to sheet: ${sheet.name}` : `Sheet: ${sheet.name}`;
	}

	async function handleSheetClick(sheetName: string): Promise<void> {
		if (!onSetActiveSheet || isActiveSheet(sheetName)) {
			return;
		}

		await onSetActiveSheet(sheetName);
	}
</script>

<div class="bottom-bar-inner">
	<!-- Left: Add sheet & Sheet navigation/tabs -->
	<div class="tabs-section">
		<button
			class="add-sheet-btn"
			disabled
			aria-label="Add sheet (inactive)"
			title="Add sheet (inactive)"
		>
			<svg
				width="14"
				height="14"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
				aria-hidden="true"
				focusable="false"
			>
				<line x1="12" y1="5" x2="12" y2="19" />
				<line x1="5" y1="12" x2="19" y2="12" />
			</svg>
		</button>

		<div class="vertical-divider" aria-hidden="true"></div>

		<!-- Sheet tab area -->
		<div class="sheet-tabs" aria-label="Workbook sheets">
			{#each sheets as sheet (sheet.name)}
				<button
					type="button"
					class="sheet-tab"
					class:active={isActiveSheet(sheet.name)}
					aria-current={isActiveSheet(sheet.name) ? 'page' : undefined}
					aria-label={`${isActiveSheet(sheet.name) ? 'Active sheet' : 'Sheet'}: ${sheet.name}`}
					title={sheetTitle(sheet)}
					disabled={!sheetCommandWired}
					onclick={() => void handleSheetClick(sheet.name)}
				>
					<span class="sheet-tab-text">{sheet.name}</span>
					{#if isActiveSheet(sheet.name)}
						<div class="active-indicator"></div>
					{/if}
				</button>
			{/each}
		</div>
	</div>

	<!-- Flexible Spacer -->
	<div class="spacer" aria-hidden="true"></div>

	<!-- Right: Status / metrics readouts -->
	<div class="status-section">
		<div
			class="status-item status-message"
			title={statusTitle}
			data-status-kind={statusKind}
			aria-live="polite"
		>
			<span class="status-text">{statusText}</span>
		</div>

		<div class="vertical-divider" aria-hidden="true"></div>

		<div class="status-item selection" title="Active selection coordinate">
			<span class="selection-text">{selectionRef}</span>
		</div>

		<div class="vertical-divider" aria-hidden="true"></div>

		<div class="status-item zoom" title={zoomTitle}>
			<span class="zoom-text">{zoomText}</span>
		</div>
	</div>
</div>

<style>
	.bottom-bar-inner {
		display: flex;
		align-items: stretch;
		justify-content: space-between;
		width: 100%;
		height: 100%;
		box-sizing: border-box;
		user-select: none;
	}

	.tabs-section {
		display: flex;
		align-items: flex-end;
		height: 100%;
		gap: 8px;
	}

	.add-sheet-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 24px;
		height: 24px;
		margin-bottom: 6px; /* Centered in the 36px height bar */
		color: var(--color-disabled-text);
		background-color: transparent;
		border: none;
		cursor: default;
		padding: 0;
		border-radius: 2px;
	}

	.vertical-divider {
		width: 1px;
		height: 16px;
		align-self: center;
		background-color: var(--color-border);
		flex-shrink: 0;
	}

	.sheet-tabs {
		display: flex;
		align-items: flex-end;
		height: 100%;
		min-width: 0;
		overflow-x: auto;
		overflow-y: hidden;
		scrollbar-width: none;
	}

	.sheet-tabs::-webkit-scrollbar {
		display: none;
	}

	.sheet-tab {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		height: 32px; /* Sits 4px below the top of the 36px height bar, flush with the bottom */
		padding: 0 16px;
		background-color: transparent;
		border: 1px solid transparent;
		color: var(--color-text-muted);
		font-size: 11px;
		font-weight: 500;
		cursor: pointer;
		box-sizing: border-box;
	}

	.sheet-tab:not(.active):hover {
		background-color: var(--color-surface-hover);
		color: var(--color-text);
	}

	.sheet-tab.active {
		background-color: var(--color-surface);
		border-left-color: var(--color-border);
		border-right-color: var(--color-border);
		border-top-color: var(--color-border);
		border-bottom-color: var(--color-surface);
		color: var(--color-accent);
		border-radius: 2px 2px 0 0;
		cursor: default;
	}

	.sheet-tab-text {
		max-width: 160px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.active-indicator {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 2px;
		background-color: var(--color-accent);
	}

	.spacer {
		flex: 1;
	}

	.status-section {
		display: flex;
		align-items: center;
		height: 100%;
		gap: 12px;
	}

	.status-item {
		display: flex;
		align-items: center;
		font-size: 11px;
		font-weight: 500;
	}

	.status-message .status-text {
		color: var(--color-text-muted);
	}

	.status-item[data-status-kind='loading'] .status-text {
		color: var(--color-accent);
	}

	.status-item[data-status-kind='error'] .status-text {
		color: var(--color-text);
	}

	.selection .selection-text {
		color: var(--color-text);
		font-family: SFMono-Regular, Consolas, 'Liberation Mono', Menlo, Courier, monospace;
	}

	.zoom .zoom-text {
		color: var(--color-text-muted);
	}
</style>
