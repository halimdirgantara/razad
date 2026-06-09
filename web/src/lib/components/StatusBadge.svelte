<script lang="ts">
	interface Props {
		status: string;
		variant?: 'success' | 'warning' | 'danger' | 'info' | 'neutral';
		label?: string;
	}
	let { status, variant = 'neutral', label }: Props = $props();
</script>

<span class="badge {variant}" title={status}>
	<span class="dot"></span>
	<span class="label">{label ?? status}</span>
</span>

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		font-size: var(--font-size-xs);
		font-weight: var(--font-weight-medium);
		padding: 0.125rem 0.5rem;
		border-radius: 999px;
		white-space: nowrap;
	}
	.dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		flex-shrink: 0;
	}
	.label {
		text-transform: capitalize;
	}

	/* ---- Variants ---- */
	.success {
		color: var(--success);
		background: var(--success-bg);
	}
	.success .dot { background: var(--success); }

	.warning {
		color: var(--warning);
		background: var(--warning-bg);
	}
	.warning .dot { background: var(--warning); }

	.danger {
		color: var(--danger);
		background: var(--danger-bg);
	}
	.danger .dot { background: var(--danger); }

	.info {
		color: var(--info);
		background: var(--info-bg);
	}
	.info .dot { background: var(--info); }

	.neutral {
		color: var(--text-muted);
		background: var(--surface);
		border: 1px solid var(--border);
	}
	.neutral .dot {
		background: var(--text-muted);
	}
</style>
