<div class="bottom-bar-inner">
	<!-- Left: Add sheet & Sheet navigation/tabs -->
	<div class="tabs-section">
		<button class="add-sheet-btn" disabled tabindex="-1" aria-label="Add sheet" title="Add sheet">
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
		<div class="sheet-tabs" aria-label="Sheets">
			<div class="sheet-tab active" aria-current="true" title="Sheet 1">
				<span class="sheet-tab-text">Sheet 1</span>
				<div class="active-indicator"></div>
			</div>
		</div>
	</div>

	<!-- Flexible Spacer -->
	<div class="spacer" aria-hidden="true"></div>

	<!-- Right: Status / metrics readouts -->
	<div class="status-section">
		<div class="status-item ready" title="Ready status">
			<span class="status-text">Ready</span>
		</div>

		<div class="vertical-divider" aria-hidden="true"></div>

		<div class="status-item selection" title="Active selection coordinate">
			<span class="selection-text">A1</span>
		</div>

		<div class="vertical-divider" aria-hidden="true"></div>

		<div class="status-item zoom" title="Zoom level">
			<span class="zoom-text">100%</span>
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
	}

	.sheet-tab {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		height: 32px; /* Sits 4px below the top of the 36px height bar, flush with the bottom */
		padding: 0 16px;
		font-size: 11px;
		font-weight: 500;
		cursor: default;
		box-sizing: border-box;
	}

	.sheet-tab.active {
		background-color: var(--color-surface);
		border-left: 1px solid var(--color-border);
		border-right: 1px solid var(--color-border);
		border-top: 1px solid var(--color-border);
		color: var(--color-accent);
		border-radius: 2px 2px 0 0;
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

	.ready .status-text {
		color: var(--color-text-muted);
	}

	.selection .selection-text {
		color: var(--color-text);
		font-family: SFMono-Regular, Consolas, 'Liberation Mono', Menlo, Courier, monospace;
	}

	.zoom .zoom-text {
		color: var(--color-text-muted);
	}
</style>
