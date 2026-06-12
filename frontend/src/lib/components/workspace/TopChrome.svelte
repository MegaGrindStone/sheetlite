<script lang="ts">
	import type { Attachment } from 'svelte/attachments';
	import type { AppearanceMode } from '$lib/appearance.svelte';
	import { main } from '$lib/wailsjs/go/models';
	import AppearanceControl from './AppearanceControl.svelte';

	type FileCommand = 'open' | 'save' | 'save-as' | 'exit';

	type Props = {
		workbook?: main.WorkbookState;
		status?: main.AppStatus;
		appearance?: main.AppearanceState;
		onOpenWorkbook?: () => Promise<void> | void;
		onSaveWorkbook?: () => Promise<void> | void;
		onSaveWorkbookAs?: () => Promise<void> | void;
		onExitApp?: () => Promise<void> | void;
		onSetAppearanceMode?: (mode: AppearanceMode) => Promise<void> | void;
	};

	let {
		workbook,
		status,
		appearance,
		onOpenWorkbook,
		onSaveWorkbook,
		onSaveWorkbookAs,
		onExitApp,
		onSetAppearanceMode
	}: Props = $props();
	let fileMenuOpen = $state(false);
	let activeFileCommand = $state<FileCommand | null>(null);

	const documentTitle = $derived(workbook?.title?.trim() || 'Untitled');
	const workbookReady = $derived(Boolean(workbook));
	const workbookDirty = $derived(Boolean(workbook?.dirty));
	const workbookUnsaved = $derived(workbookReady && !workbook?.filePath);
	const statusKind = $derived(status?.kind ?? main.AppStatusKind.Ready);
	const statusMessage = $derived(status?.message?.trim() || 'Ready');
	const statusLabel = $derived(status?.busy ? 'Working…' : statusMessage);
	const statusTitle = $derived(`${statusKind}: ${statusMessage}`);
	const commandBusy = $derived(Boolean(status?.busy) || activeFileCommand !== null);
	const hasFileActions = $derived(
		Boolean(onOpenWorkbook || onSaveWorkbook || onSaveWorkbookAs || onExitApp)
	);
	const canUseFileMenu = $derived(hasFileActions && !commandBusy);
	const canOpenWorkbook = $derived(Boolean(onOpenWorkbook) && !commandBusy);
	const canSaveWorkbook = $derived(
		Boolean(onSaveWorkbook) && workbookReady && !commandBusy && (workbookDirty || workbookUnsaved)
	);
	const canSaveWorkbookAs = $derived(Boolean(onSaveWorkbookAs) && workbookReady && !commandBusy);
	const canExitApp = $derived(Boolean(onExitApp) && !commandBusy);
	const showFileMenu = $derived(fileMenuOpen && canUseFileMenu);
	const fileMenuTitle = $derived(
		!hasFileActions
			? 'File menu is unavailable'
			: commandBusy
				? 'A file command is already in progress'
				: 'Open File menu'
	);
	const openWorkbookLabel = $derived(activeFileCommand === 'open' ? 'Opening…' : 'Open');
	const saveWorkbookLabel = $derived(activeFileCommand === 'save' ? 'Saving…' : 'Save');
	const saveWorkbookAsLabel = $derived(activeFileCommand === 'save-as' ? 'Saving…' : 'Save As…');
	const saveWorkbookMeta = $derived(
		workbookUnsaved ? 'Choose location' : workbookDirty ? 'Write changes' : 'No unsaved changes'
	);

	function closeFileMenu(): void {
		fileMenuOpen = false;
	}

	function toggleFileMenu(): void {
		if (!canUseFileMenu) {
			return;
		}

		fileMenuOpen = !showFileMenu;
	}

	async function runFileCommand(
		command: FileCommand,
		callback: (() => Promise<void> | void) | undefined
	): Promise<void> {
		if (!callback || commandBusy) {
			return;
		}

		closeFileMenu();
		activeFileCommand = command;

		try {
			await callback();
		} finally {
			activeFileCommand = null;
		}
	}

	function handleOpenWorkbook(): Promise<void> {
		return runFileCommand('open', canOpenWorkbook ? onOpenWorkbook : undefined);
	}

	function handleSaveWorkbook(): Promise<void> {
		return runFileCommand('save', canSaveWorkbook ? onSaveWorkbook : undefined);
	}

	function handleSaveWorkbookAs(): Promise<void> {
		return runFileCommand('save-as', canSaveWorkbookAs ? onSaveWorkbookAs : undefined);
	}

	function handleExitApp(): Promise<void> {
		return runFileCommand('exit', canExitApp ? onExitApp : undefined);
	}

	const fileMenuBehavior: Attachment<HTMLDivElement> = (node) => {
		function handleKeydown(event: KeyboardEvent): void {
			if (event.key !== 'Escape' || !showFileMenu) {
				return;
			}

			event.stopPropagation();
			closeFileMenu();
		}

		function handleDocumentPointerDown(event: PointerEvent): void {
			if (!showFileMenu || !(event.target instanceof Node) || node.contains(event.target)) {
				return;
			}

			closeFileMenu();
		}

		node.addEventListener('keydown', handleKeydown);
		document.addEventListener('pointerdown', handleDocumentPointerDown);

		return () => {
			node.removeEventListener('keydown', handleKeydown);
			document.removeEventListener('pointerdown', handleDocumentPointerDown);
		};
	};
</script>

<div class="top-chrome-container">
	<!-- Top Row: Identity, Document Title, Status, and Theme Selector -->
	<div class="top-row">
		<div class="left-section">
			<!-- Brand Block -->
			<div class="brand-block">
				<div class="brand-mark" aria-hidden="true">S</div>
				<span class="brand-name">Sheetlite</span>
			</div>

			<!-- Divider -->
			<div class="section-divider" aria-hidden="true"></div>

			<!-- Doc Title Block -->
			<div class="title-block">
				<span class="doc-title">{documentTitle}</span>
				{#if workbookDirty}
					<span class="dirty-indicator" title="Unsaved changes" aria-label="Unsaved changes">
						<span class="dirty-dot" aria-hidden="true"></span>
						<span>Unsaved</span>
					</span>
				{/if}
				<!-- Muted Status Affordance -->
				<div class="status-affordance" title={statusTitle} data-status-kind={statusKind}>
					<svg
						class="status-icon"
						width="6"
						height="6"
						viewBox="0 0 24 24"
						fill="currentColor"
						stroke="none"
						aria-hidden="true"
						focusable="false"
					>
						<circle cx="12" cy="12" r="12" />
					</svg>
					<span>{statusLabel}</span>
				</div>
			</div>
		</div>

		<!-- Right Side Controls -->
		<div class="right-section">
			<AppearanceControl {appearance} {onSetAppearanceMode} />
		</div>
	</div>

	<!-- Bottom Row: Inactive Menus & Compact Disabled Toolbar Groups -->
	<div class="bottom-row">
		<!-- Main Menus -->
		<nav class="menu-bar" aria-label="Main menu">
			<div {@attach fileMenuBehavior} class="file-menu-shell">
				<button
					type="button"
					class="menu-item file-menu-trigger"
					class:open={showFileMenu}
					aria-haspopup="true"
					aria-expanded={showFileMenu}
					aria-controls="file-menu-popover"
					disabled={!canUseFileMenu}
					title={fileMenuTitle}
					onclick={toggleFileMenu}
				>
					File
				</button>

				{#if showFileMenu}
					<div id="file-menu-popover" class="file-menu-popover" aria-label="File menu">
						<button
							type="button"
							class="file-menu-command"
							disabled={!canOpenWorkbook}
							onclick={handleOpenWorkbook}
						>
							<span class="file-menu-command-label">{openWorkbookLabel}</span>
							<span class="file-menu-command-meta">Excel workbook</span>
						</button>
						<div class="file-menu-separator" aria-hidden="true"></div>
						<button
							type="button"
							class="file-menu-command"
							disabled={!canSaveWorkbook}
							title={saveWorkbookMeta}
							onclick={handleSaveWorkbook}
						>
							<span class="file-menu-command-label">{saveWorkbookLabel}</span>
							<span class="file-menu-command-meta">{saveWorkbookMeta}</span>
						</button>
						<button
							type="button"
							class="file-menu-command"
							disabled={!canSaveWorkbookAs}
							onclick={handleSaveWorkbookAs}
						>
							<span class="file-menu-command-label">{saveWorkbookAsLabel}</span>
							<span class="file-menu-command-meta">Choose location</span>
						</button>
						<div class="file-menu-separator" aria-hidden="true"></div>
						<button
							type="button"
							class="file-menu-command"
							disabled={!canExitApp}
							onclick={handleExitApp}
						>
							<span class="file-menu-command-label">Exit</span>
							<span class="file-menu-command-meta">Close app</span>
						</button>
					</div>
				{/if}
			</div>

			{#each ['Edit', 'View', 'Insert', 'Format', 'Data', 'Help'] as label (label)}
				<span
					class="menu-item menu-stub stub-disabled"
					aria-disabled="true"
					tabindex="-1"
					title={`${label} menu is inactive`}
				>
					{label}
				</span>
			{/each}
		</nav>

		<!-- Divider between Menus and Toolbar -->
		<div class="toolbar-divider" aria-hidden="true"></div>

		<!-- Disabled Toolbar Groups -->
		<div class="toolbar" aria-label="Inactive toolbar controls">
			<!-- Undo/Redo Group -->
			<div class="toolbar-group">
				<button class="toolbar-btn" disabled aria-label="Undo">
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
						<path d="M3 7v6h6" />
						<path d="M21 17a9 9 0 0 0-9-9 9 9 0 0 0-6 2.3L3 13" />
					</svg>
				</button>
				<button class="toolbar-btn" disabled aria-label="Redo">
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
						<path d="M21 7v6h-6" />
						<path d="M3 17a9 9 0 0 1 9-9 9 9 0 0 1 6 2.3l3 2.7" />
					</svg>
				</button>
			</div>

			<div class="toolbar-divider" aria-hidden="true"></div>

			<!-- Text Styling Group -->
			<div class="toolbar-group">
				<button class="toolbar-btn bold-btn" disabled aria-label="Bold"> B </button>
				<button class="toolbar-btn italic-btn" disabled aria-label="Italic"> I </button>
				<button class="toolbar-btn strike-btn" disabled aria-label="Strikethrough"> S </button>
			</div>

			<div class="toolbar-divider" aria-hidden="true"></div>

			<!-- Grid/Cell Formatting Group -->
			<div class="toolbar-group">
				<button class="toolbar-btn" disabled aria-label="Borders">
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
						<rect x="3" y="3" width="18" height="18" rx="0" />
						<line x1="12" y1="3" x2="12" y2="21" />
						<line x1="3" y1="12" x2="21" y2="12" />
					</svg>
				</button>
				<button class="toolbar-btn" disabled aria-label="Merge Cells">
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
						<rect x="3" y="3" width="18" height="18" rx="0" />
						<line x1="3" y1="12" x2="9" y2="12" />
						<line x1="15" y1="12" x2="21" y2="12" />
						<polyline points="11 9 9 12 11 15" />
						<polyline points="13 9 15 12 13 15" />
					</svg>
				</button>
			</div>

			<div class="toolbar-divider" aria-hidden="true"></div>

			<!-- Alignment Group -->
			<div class="toolbar-group">
				<button class="toolbar-btn" disabled aria-label="Align Left">
					<svg
						width="14"
						height="14"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						aria-hidden="true"
						focusable="false"
					>
						<line x1="3" y1="6" x2="21" y2="6" />
						<line x1="3" y1="12" x2="17" y2="12" />
						<line x1="3" y1="18" x2="21" y2="18" />
					</svg>
				</button>
				<button class="toolbar-btn" disabled aria-label="Align Center">
					<svg
						width="14"
						height="14"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						aria-hidden="true"
						focusable="false"
					>
						<line x1="3" y1="6" x2="21" y2="6" />
						<line x1="6" y1="12" x2="18" y2="12" />
						<line x1="3" y1="18" x2="21" y2="18" />
					</svg>
				</button>
				<button class="toolbar-btn" disabled aria-label="Align Right">
					<svg
						width="14"
						height="14"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						aria-hidden="true"
						focusable="false"
					>
						<line x1="3" y1="6" x2="21" y2="6" />
						<line x1="7" y1="12" x2="21" y2="12" />
						<line x1="3" y1="18" x2="21" y2="18" />
					</svg>
				</button>
			</div>
		</div>
	</div>
</div>

<style>
	.top-chrome-container {
		position: relative;
		z-index: 2;
		display: flex;
		flex-direction: column;
		width: 100%;
		height: 100%;
		background-color: var(--color-chrome);
	}

	/* Top Row styling */
	.top-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		height: 32px;
		padding: 0 12px;
	}

	.left-section {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.brand-block {
		display: flex;
		align-items: center;
		gap: 6px;
	}

	.brand-mark {
		background-color: var(--color-accent);
		color: var(--color-surface);
		font-weight: 700;
		font-size: 11px;
		width: 18px;
		height: 18px;
		border-radius: 2px; /* sm rounding as per DESIGN.md */
		display: flex;
		align-items: center;
		justify-content: center;
		line-height: 1;
		user-select: none;
	}

	.brand-name {
		font-weight: 600;
		font-size: 13px;
		color: var(--color-text);
		line-height: 1;
	}

	.section-divider {
		width: 1px;
		height: 16px;
		background-color: var(--color-border);
	}

	.title-block {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.doc-title {
		font-weight: 500;
		font-size: 13px;
		color: var(--color-text);
		line-height: 1;
	}

	.dirty-indicator {
		display: flex;
		align-items: center;
		gap: 4px;
		color: var(--color-accent);
		font-size: 11px;
		font-weight: 500;
		line-height: 1;
		user-select: none;
	}

	.dirty-dot {
		width: 6px;
		height: 6px;
		border-radius: 9999px;
		background-color: var(--color-accent);
		flex-shrink: 0;
	}

	.status-affordance {
		display: flex;
		align-items: center;
		gap: 4px;
		color: var(--color-text-muted);
		font-size: 11px;
		line-height: 1;
		user-select: none;
	}

	.status-icon {
		color: var(--color-text-muted);
		opacity: 0.8;
	}

	.right-section {
		display: flex;
		align-items: center;
	}

	/* Bottom Row styling */
	.bottom-row {
		display: flex;
		align-items: center;
		height: 32px;
		padding: 0 12px;
		gap: 12px;
		border-top: 1px solid var(--color-border);
	}

	.menu-bar {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.file-menu-shell {
		position: relative;
		display: flex;
		align-items: center;
	}

	.menu-item {
		color: var(--color-text-muted);
		font-size: 12px;
		font-weight: 400;
		line-height: 1;
		user-select: none;
		padding: 2px 4px;
		border-radius: 2px;
	}

	.file-menu-trigger {
		cursor: pointer;
		background-color: transparent;
	}

	.file-menu-trigger:not(:disabled):hover,
	.file-menu-trigger.open {
		background-color: var(--color-surface-hover);
		color: var(--color-text);
	}

	.file-menu-trigger:focus-visible,
	.file-menu-command:focus-visible {
		outline: 2px solid var(--color-focus-ring);
		outline-offset: -1px;
	}

	.menu-stub {
		cursor: default;
	}

	.file-menu-popover {
		position: absolute;
		top: calc(100% + 6px);
		left: 0;
		z-index: 20;
		min-width: 208px;
		padding: 4px;
		border: 1px solid var(--color-border);
		border-radius: 2px;
		background-color: var(--color-surface);
		color: var(--color-text);
	}

	.file-menu-separator {
		height: 1px;
		margin: 4px;
		background-color: var(--color-border);
	}

	.file-menu-command {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		width: 100%;
		min-height: 28px;
		padding: 6px 8px;
		border-radius: 2px;
		background-color: transparent;
		color: var(--color-text);
		cursor: pointer;
		text-align: left;
	}

	.file-menu-command:not(:disabled):hover {
		background-color: var(--color-surface-hover);
	}

	.file-menu-command:disabled {
		background-color: transparent;
		color: var(--color-disabled-text);
		opacity: 0.72;
	}

	.file-menu-command:disabled .file-menu-command-meta {
		color: var(--color-disabled-text);
	}

	.file-menu-command-label {
		font-weight: 500;
		white-space: nowrap;
	}

	.file-menu-command-meta {
		color: var(--color-text-muted);
		font-size: 11px;
		white-space: nowrap;
	}

	.toolbar-divider {
		width: 1px;
		height: 14px;
		background-color: var(--color-border);
	}

	.toolbar {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.toolbar-group {
		display: flex;
		align-items: center;
		gap: 2px;
	}

	.toolbar-btn {
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 2px;
		border: none;
		background: transparent;
		color: var(--color-disabled-text);
		cursor: default;
		font-size: 11px;
		line-height: 1;
	}

	/* Text style indicators for bold, italic, and strikethrough */
	.bold-btn {
		font-weight: 700;
	}
	.italic-btn {
		font-style: italic;
	}
	.strike-btn {
		text-decoration: line-through;
	}
</style>
