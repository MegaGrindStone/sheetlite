<div class="formula-bar" aria-label="Formula bar">
	<!-- Name Box showing A1 -->
	<div class="name-box" aria-label="Current cell reference" title="Current cell (A1)">A1</div>

	<!-- Split Divider -->
	<div class="divider" aria-hidden="true"></div>

	<!-- fx marker -->
	<div class="fx-marker" aria-hidden="true" title="Insert function (inactive)">fx</div>

	<!-- Inactive formula display/input-like area -->
	<div
		class="formula-display"
		aria-label="Formula bar input (inactive)"
		title="Formula editing is currently inactive"
		aria-disabled="true"
	>
		<!-- Static placeholder to honestly convey the bar is inactive -->
	</div>
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
		font-style: italic;
		font-weight: 600;
		font-size: 14px;
		color: var(--color-text-muted);
		user-select: none;
		flex-shrink: 0;
	}

	.formula-display {
		flex: 1;
		height: 22px;
		background-color: var(--color-disabled-bg);
		border: 1px solid var(--color-border);
		cursor: not-allowed;
		box-sizing: border-box;
	}
</style>
