<script lang="ts">
	import Card from '$lib/components/Card.svelte';
</script>

<svelte:head><title>Audit Log — Razad</title></svelte:head>

<h1>Audit Log</h1>

<Card title="Audit Trail">
	<div class="empty-state">
		<div class="empty-icon">◷</div>
		<p class="empty-title">No audit entries</p>
		<p class="text-muted meta">Every privileged action — deploys, config changes, logins — will be recorded here with an immutable trail.</p>
	</div>
</Card>

<style>
	h1 { margin-bottom: var(--space-4); }
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-2);
		padding: var(--space-10) var(--space-4);
		text-align: center;
	}
	.empty-icon {
		font-size: 2rem;
		opacity: 0.3;
		margin-bottom: var(--space-2);
	}
	.empty-title {
		font-size: var(--font-size-base);
		font-weight: var(--font-weight-medium);
		color: var(--text);
	}
</style>
