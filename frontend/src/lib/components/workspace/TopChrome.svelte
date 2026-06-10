<script lang="ts">
	import AppearanceControl from './AppearanceControl.svelte';
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
				<span class="doc-title">Untitled</span>
				<!-- Muted Status Affordance -->
				<div class="status-affordance" title="Local draft">
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
					<span>Local draft</span>
				</div>
			</div>
		</div>

		<!-- Right Side Controls -->
		<div class="right-section">
			<AppearanceControl />
		</div>
	</div>

	<!-- Bottom Row: Inactive Menus & Compact Disabled Toolbar Groups -->
	<div class="bottom-row">
		<!-- Inactive Menus -->
		<nav class="menu-bar" aria-label="Inactive main menu">
			{#each ['File', 'Edit', 'View', 'Insert', 'Format', 'Data', 'Help'] as label (label)}
				<span class="menu-item stub-disabled" aria-disabled="true" title="{label} menu is inactive">
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
		gap: 12px;
	}

	.menu-item {
		color: var(--color-disabled-text);
		font-size: 12px;
		font-weight: 400;
		line-height: 1;
		cursor: default;
		user-select: none;
		padding: 2px 4px;
		border-radius: 2px;
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
