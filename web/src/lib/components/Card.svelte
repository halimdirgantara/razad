<script lang="ts">
	interface Props {
		title?: string;
		actions?: unknown; // for slot
		padding?: 'tight' | 'normal' | 'loose';
		class?: string;
	}
	let {
		title = '',
		actions = null,
		padding = 'normal',
		class: className = ''
	}: Props = $props();
</script>

<div class="card {padding} {className}">
	{#if title || actions}
		<div class="card-header">
			<h3 class="card-title">{title}</h3>
			{#if actions}
				<div class="card-actions">
					{@render actions()}
				</div>
			{/if}
		</div>
	{/if}
	<div class="card-body">
		{@render children()}
	</div>
</div>

<style>
	.card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
	}
	.card.tight  { padding: var(--space-3); }
	.card.normal { padding: var(--space-4); }
	.card.loose  { padding: var(--space-6); }
	.card-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--space-3);
		gap: var(--space-3);
	}
	.card-title {
		font-size: var(--font-size-sm);
		font-weight: var(--font-weight-semibold);
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.04em;
		margin: 0;
	}
	.card-actions {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		margin-left: auto;
	}
	.card-body {
		font-size: var(--font-size-base);
		color: var(--text);
	}
</style>
