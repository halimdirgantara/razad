<script lang="ts">
	import Card from '$lib/components/Card.svelte';
</script>

<svelte:head><title>Settings — Razad</title></svelte:head>

<h1>Settings</h1>

<div class="settings-grid">
	<Card title="Node">
		<div class="empty-state">
			<p class="text-muted meta">Node identity, enrollment token, and data directory configuration.</p>
		</div>
	</Card>

	<Card title="Authentication">
		<div class="empty-state">
			<p class="text-muted meta">Session TTL, password policy, and API token management.</p>
		</div>
	</Card>

	<Card title="Notifications">
		<div class="empty-state">
			<p class="text-muted meta">Alerts for offline nodes, SSL expiry, and deployment failures.</p>
		</div>
	</Card>

	<Card title="Danger Zone">
		<div class="empty-state">
			<p class="text-muted meta">Factory reset, data purge, and uninstall options.</p>
		</div>
	</Card>
</div>

<style>
	h1 { margin-bottom: var(--space-4); }
	.settings-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-4);
	}
	.empty-state {
		padding: var(--space-4) 0;
	}
</style>
