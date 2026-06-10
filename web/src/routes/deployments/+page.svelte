<script lang="ts">
	import Card from '$lib/components/Card.svelte';
	import Button from '$lib/components/Button.svelte';
</script>

<svelte:head><title>Deployments — Razad</title></svelte:head>

<h1>Deployments</h1>

<Card title="Deployment History">
	<div class="empty-state">
		<div class="empty-icon">↻</div>
		<p class="empty-title">No deployments yet</p>
		<p class="text-muted meta">Deploy an application to see its history here.</p>
		<a href="/apps"><Button variant="primary" size="sm">Go to Applications</Button></a>
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
