<script lang="ts">
	import Card from '$lib/components/Card.svelte';
	import Button from '$lib/components/Button.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import DataTable from '$lib/components/DataTable.svelte';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	let app: Record<string, any> | null = $state(null);
	let loading = $state(true);
	let error = $state('');

	const appId = $derived($page.params.id);

	async function fetchApp() {
		try {
			const token = localStorage.getItem('razad_token');
			const headers = { 'Authorization': `Bearer ${token}` };

			const res = await fetch(`/api/v1/apps/${appId}`, { headers });
			if (res.ok) {
				app = await res.json();
			} else {
				error = 'App not found.';
			}
		} catch {
			error = 'Failed to load app.';
		} finally {
			loading = false;
		}
	}

	async function doAction(endpoint: string) {
		const res = await fetch(`/api/v1/apps/${appId}${endpoint}`, {
			method: 'POST',
			headers: { 'Authorization': `Bearer ${localStorage.getItem('razad_token')}` }
		});
		if (res.ok) {
			app = await res.json();
		}
	}

	onMount(fetchApp);

	function badgeVariant(s: string) {
		if (s === 'running' || s === 'success') return 'success';
		if (s === 'failed' || s === 'deleted') return 'danger';
		if (s === 'deploying' || s === 'pending') return 'info';
		if (s === 'stopped') return 'warning';
		return 'neutral';
	}
</script>

<svelte:head><title>{app?.name ?? 'App'} — Razad</title></svelte:head>

{#if loading}
	<p class="text-muted">Loading...</p>
{:else if error}
	<p class="err">{error}</p>
	<a href="/apps">← Back to apps</a>
{:else if app}
	<div class="app-header">
		<div>
			<h1>{app.name}</h1>
			<span class="mono text-muted">{app.id}</span>
		</div>
		<div class="actions">
			<Button variant="secondary" size="sm" onclick={() => doAction('/restart')}>Restart</Button>
			<Button variant="ghost" size="sm" onclick={() => doAction('/stop')}>Stop</Button>
			<Button variant="primary" size="sm" onclick={() => doAction('/deploy')}>Deploy</Button>
		</div>
	</div>

	<div class="detail-grid">
		<div class="col">
			<Card title="Status">
				<div class="status-row">
					<StatusBadge status={app.status} variant={badgeVariant(app.status)} />
					<span class="meta text-muted">Runtime: {app.runtime ?? 'unknown'}</span>
				</div>
			</Card>

			<Card title="Environment Variables">
				<p class="text-muted meta">Coming soon.</p>
			</Card>
		</div>
		<div class="col">
			<Card title="Deployments">
				<p class="text-muted meta">No deployments yet.</p>
			</Card>

			<Card title="Recent Logs">
				<p class="text-muted meta">Live log streaming coming soon.</p>
			</Card>
		</div>
	</div>
{/if}

<style>
	.app-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-4);
	}
	.app-header h1 { margin: 0 0 var(--space-1); }
	.actions { display: flex; gap: var(--space-2); }
	.detail-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-4);
	}
	.col { display: flex; flex-direction: column; gap: var(--space-4); }
	.status-row { display: flex; align-items: center; gap: var(--space-3); }
	.err { color: var(--danger); }
</style>
