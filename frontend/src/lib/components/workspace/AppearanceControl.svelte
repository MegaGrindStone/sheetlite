<script lang="ts">
	import { appearanceState } from '$lib/appearance.svelte';
</script>

<div class="appearance-control">
	<span class="control-label" id="appearance-group-label">Appearance</span>
	<div class="segmented-control" role="group" aria-labelledby="appearance-group-label">
		<button
			type="button"
			class="control-btn {appearanceState.mode === 'system' ? 'active' : ''}"
			aria-pressed={appearanceState.mode === 'system'}
			onclick={() => appearanceState.setMode('system')}
		>
			System
		</button>
		<button
			type="button"
			class="control-btn {appearanceState.mode === 'light' ? 'active' : ''}"
			aria-pressed={appearanceState.mode === 'light'}
			onclick={() => appearanceState.setMode('light')}
		>
			Light
		</button>
		<button
			type="button"
			class="control-btn {appearanceState.mode === 'dark' ? 'active' : ''}"
			aria-pressed={appearanceState.mode === 'dark'}
			onclick={() => appearanceState.setMode('dark')}
		>
			Dark
		</button>
	</div>
</div>

<style>
	.appearance-control {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.control-label {
		font-size: 11px;
		font-weight: 500;
		color: var(--color-text-muted);
		user-select: none;
	}

	.segmented-control {
		display: inline-flex;
		background-color: var(--color-surface);
		border: 1px solid var(--color-border);
		padding: 1px;
		gap: 1px;
		border-radius: 2px; /* Subtle rounding aligns with DESIGN.md and gives a polished feel */
	}

	.control-btn {
		font-size: 11px;
		font-weight: 500;
		color: var(--color-text-muted);
		background: transparent;
		border: none;
		padding: 3px 8px;
		cursor: pointer;
		user-select: none;
		line-height: 1.2;
		border-radius: 1px;
		transition:
			background-color 0.1s ease,
			color 0.1s ease;
	}

	.control-btn:hover:not(.active) {
		background-color: var(--color-surface-hover);
		color: var(--color-text);
	}

	.control-btn.active {
		background-color: var(--color-accent);
		color: var(--color-surface);
		font-weight: 600;
	}

	.control-btn:focus-visible {
		outline: 2px solid var(--color-focus-ring);
		outline-offset: -1px;
		z-index: 1; /* Keep focus ring on top */
	}
</style>
