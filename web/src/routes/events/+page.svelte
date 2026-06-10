<script lang="ts">
	import Card from '$lib/components/Card.svelte';

	interface Event {
		id: string;
		kind: string;
		source: string;
		message: string;
		severity: 'info' | 'warn' | 'error';
		occurred_at: string;
	}

	const mockEvents: Event[] = [];
</script>

<svelte:head><title>Events — Razad</title></svelte:head>

<h1>Events</h1>

<Card title="System Events">
	<div class="empty-state">
		<div class="empty-icon">⚡</div>
		<p class="empty-title">No events yet</p>
		<p class="text-muted meta">System and application events will appear here once services are running.</p>
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
