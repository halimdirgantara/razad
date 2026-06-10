<script lang="ts">
	import Card from '$lib/components/Card.svelte';
</script>

<svelte:head><title>Domains — Razad</title></svelte:head>

<h1>Domains</h1>

<Card title="Domain & SSL Management">
	<div class="empty-state">
		<div class="empty-icon">◈</div>
		<p class="empty-title">No domains configured</p>
		<p class="text-muted meta">Bind custom domains to your applications. Razad will generate Nginx config and provision SSL certificates automatically.</p>
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
