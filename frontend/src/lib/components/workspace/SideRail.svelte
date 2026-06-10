<script lang="ts">
	let { side = 'left' } = $props<{ side?: 'left' | 'right' }>();
</script>

<aside
	class="side-rail {side}"
	aria-label={side === 'left' ? 'Workbook navigation rail' : 'Contextual utility rail'}
>
	<div class="rail-group">
		{#if side === 'left'}
			<!-- Left side navigation items -->
			<button
				class="rail-item"
				disabled
				tabindex="-1"
				title="Workbook Outline"
				aria-label="Workbook Outline"
			>
				<svg
					class="rail-icon"
					width="18"
					height="18"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<line x1="8" y1="6" x2="21" y2="6" />
					<line x1="8" y1="12" x2="21" y2="12" />
					<line x1="8" y1="18" x2="21" y2="18" />
					<line x1="3" y1="6" x2="3.01" y2="6" />
					<line x1="3" y1="12" x2="3.01" y2="12" />
					<line x1="3" y1="18" x2="3.01" y2="18" />
				</svg>
			</button>

			<button
				class="rail-item"
				disabled
				tabindex="-1"
				title="Named Ranges & Tables"
				aria-label="Named Ranges & Tables"
			>
				<svg
					class="rail-icon"
					width="18"
					height="18"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<rect x="3" y="3" width="18" height="18" rx="0" />
					<line x1="9" y1="3" x2="9" y2="21" />
					<line x1="3" y1="9" x2="21" y2="9" />
				</svg>
			</button>

			<button
				class="rail-item"
				disabled
				tabindex="-1"
				title="Imports & Files"
				aria-label="Imports & Files"
			>
				<svg
					class="rail-icon"
					width="18"
					height="18"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
					<polyline points="14 2 14 8 20 8" />
					<line x1="12" y1="18" x2="12" y2="12" />
					<polyline points="9 15 12 12 15 15" />
				</svg>
			</button>

			<button
				class="rail-item"
				disabled
				tabindex="-1"
				title="Templates & Examples"
				aria-label="Templates & Examples"
			>
				<svg
					class="rail-icon"
					width="18"
					height="18"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<rect x="3" y="3" width="7" height="9" />
					<rect x="14" y="3" width="7" height="5" />
					<rect x="14" y="12" width="7" height="9" />
					<rect x="3" y="16" width="7" height="5" />
				</svg>
			</button>
		{:else}
			<!-- Right side contextual items -->
			<button
				class="rail-item"
				disabled
				tabindex="-1"
				title="Format & Properties"
				aria-label="Format & Properties"
			>
				<svg
					class="rail-icon"
					width="18"
					height="18"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<path d="M12 20h9" />
					<path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
				</svg>
			</button>

			<button
				class="rail-item"
				disabled
				tabindex="-1"
				title="Comments & Notes"
				aria-label="Comments & Notes"
			>
				<svg
					class="rail-icon"
					width="18"
					height="18"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
				</svg>
			</button>

			<button
				class="rail-item"
				disabled
				tabindex="-1"
				title="Functions & Help"
				aria-label="Functions & Help"
			>
				<svg
					class="rail-icon"
					width="18"
					height="18"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<circle cx="12" cy="12" r="10" />
					<path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
					<line x1="12" y1="17" x2="12.01" y2="17" />
				</svg>
			</button>

			<button
				class="rail-item"
				disabled
				tabindex="-1"
				title="Activity, History & Diagnostics"
				aria-label="Activity, History & Diagnostics"
			>
				<svg
					class="rail-icon"
					width="18"
					height="18"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<circle cx="12" cy="12" r="10" />
					<polyline points="12 6 12 12 16 14" />
				</svg>
			</button>
		{/if}
	</div>
</aside>

<style>
	.side-rail {
		display: flex;
		flex-direction: column;
		width: 100%;
		height: 100%;
		background-color: var(--color-chrome);
		box-sizing: border-box;
		align-items: center;
		padding-top: 12px;
	}

	.left {
		border-right: 1px solid var(--color-border);
	}

	.right {
		border-left: 1px solid var(--color-border);
	}

	.rail-group {
		display: flex;
		flex-direction: column;
		gap: 16px;
		align-items: center;
		width: 100%;
	}

	.rail-item {
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 4px;
		border: none;
		background: transparent;
		color: var(--color-disabled-text);
		cursor: default;
		box-sizing: border-box;
		padding: 0;
	}

	.rail-icon {
		opacity: 0.8;
	}
</style>
