<script lang="ts">
	import Card from '$lib/components/Card.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import Button from '$lib/components/Button.svelte';

	let health: { status: string } | null = $state(null);
	let error: string | null = $state(null);

	async function checkHealth() {
		try {
			const res = await fetch('/api/v1/health');
			health = await res.json();
			error = null;
		} catch (e) {
			error = 'Failed to connect to daemon';
			health = null;
		}
	}

	const systemMetrics = $state([
		{ label: 'CPU', value: '—', max: '100%' },
		{ label: 'RAM', value: '—', max: '—' },
		{ label: 'Disk', value: '—', max: '—' },
		{ label: 'Uptime', value: '—', max: '' },
	]);

	const recentApps = $state<Array<Record<string, unknown>>>([]);
</script>

<h1>Operations Center</h1>

<!-- Health Strip -->
<Card title="System Health" padding="tight">
	<div class="health-strip">
		{#each systemMetrics as metric}
			<div class="metric">
				<span class="metric-label">{metric.label}</span>
				<span class="metric-value">{metric.value}</span>
				{#if metric.max}
					<span class="metric-max">{metric.max}</span>
				{/if}
			</div>
		{/each}
		<div class="metric">
			<span class="metric-label">Daemon</span>
			{#if health}
				<StatusBadge status={health.status} variant={health.status === 'ok' ? 'success' : 'warning'} />
			{:else if error}
				<StatusBadge status="offline" variant="danger" />
			{:else}
				<Button variant="ghost" size="sm" onclick={checkHealth}>Check</Button>
			{/if}
		</div>
	</div>
</Card>

<!-- Two-column layout -->
<div class="dashboard-grid">
	<div class="col">
		<Card title="Running Workloads">
			{#if recentApps.length === 0}
				<div class="empty-hint">
					<p>No applications deployed yet.</p>
					<Button variant="primary" size="sm">Deploy App</Button>
				</div>
			{/if}
		</Card>

		<Card title="Recent Deployments" >
			<div class="empty-hint">
				<span class="text-muted meta">No recent deployments.</span>
			</div>
		</Card>
	</div>
	<div class="col">
		<Card title="Razad Advisor">
			{#if !health}
				<div class="empty-hint">
					<p class="text-muted meta">Start the daemon to enable AI monitoring.</p>
				</div>
			{/if}
		</Card>

		<Card title="Active Alerts" >
			<div class="empty-hint">
				<span class="text-muted meta">No active alerts.</span>
			</div>
		</Card>

		<Card title="System Logs" >
			<div class="empty-hint">
				<span class="text-muted meta">Logs will appear once applications are running.</span>
			</div>
		</Card>
	</div>
</div>

<style>
	h1 {
		margin-bottom: var(--space-4);
	}
	.health-strip {
		display: flex;
		gap: var(--space-4);
		flex-wrap: wrap;
	}
	.metric {
		display: flex;
		flex-direction: column;
		gap: 0.125rem;
		min-width: 80px;
		padding: var(--space-2) var(--space-3);
		background: var(--bg-alt);
		border-radius: var(--radius-sm);
	}
	.metric-label {
		font-size: var(--font-size-xs);
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}
	.metric-value {
		font-size: var(--font-size-xl);
		font-weight: var(--font-weight-bold);
		font-family: var(--font-mono);
		color: var(--text);
	}
	.metric-max {
		font-size: var(--font-size-xs);
		color: var(--text-muted);
	}
	.dashboard-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-4);
		margin-top: var(--space-4);
	}
	.col {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
.empty-hint {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
		align-items: flex-start;
		padding: var(--space-4) 0;
	}
</style>
