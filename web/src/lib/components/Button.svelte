<script lang="ts">
	interface Props {
		variant?: 'primary' | 'secondary' | 'ghost' | 'danger';
		size?: 'sm' | 'md' | 'lg';
		disabled?: boolean;
		type?: 'button' | 'submit';
		onclick?: (e: MouseEvent) => void;
		class?: string;
		children: import('svelte').Snippet;
	}
	let {
		variant = 'primary',
		size = 'md',
		disabled = false,
		type = 'button',
		onclick,
		class: className = '',
		children
	}: Props = $props();
</script>

<button
	{type}
	class="btn {variant} {size} {className}"
	{disabled}
	{onclick}
>
	{@render children()}
</button>

<style>
	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		font-weight: var(--font-weight-medium);
		border: 1px solid transparent;
		border-radius: var(--radius-sm);
		cursor: pointer;
		transition: background 0.15s, border-color 0.15s, color 0.15s, opacity 0.15s;
		white-space: nowrap;
		user-select: none;
	}
	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.btn.sm { font-size: var(--font-size-xs); padding: 0.25rem 0.5rem; }
	.btn.md { font-size: var(--font-size-sm); padding: 0.375rem 0.75rem; }
	.btn.lg { font-size: var(--font-size-base); padding: 0.5rem 1rem; }

	.primary {
		background: var(--primary);
		color: #04161D;
		border-color: var(--primary);
	}
	.primary:hover:not(:disabled) {
		background: var(--primary-hover);
		border-color: var(--primary-hover);
	}
	.secondary {
		background: var(--surface-2);
		color: var(--text);
		border-color: var(--border);
	}
	.secondary:hover:not(:disabled) {
		background: var(--surface-3);
		border-color: var(--border-strong);
	}
	.ghost {
		background: transparent;
		color: var(--text-secondary);
		border-color: transparent;
	}
	.ghost:hover:not(:disabled) {
		background: var(--surface);
		color: var(--text);
	}
	.danger {
		background: var(--danger-bg);
		color: var(--danger);
		border-color: var(--danger);
	}
	.danger:hover:not(:disabled) {
		background: var(--danger);
		color: #04161D;
	}
</style>
