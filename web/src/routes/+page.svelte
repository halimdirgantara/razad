<script lang="ts">
	import Card from '$lib/components/Card.svelte';
	import Button from '$lib/components/Button.svelte';
	import { onMount } from 'svelte';

	interface HealthStats {
		cpu_usage: number;
		ram_usage: number;
		ram_total: number;
		ram_used: number;
		disk_usage: number;
		disk_total: number;
		disk_free: number;
		load_avg_1: number;
		load_avg_5: number;
		load_avg_15: number;
		uptime_seconds: number;
		hostname: string;
		process_count: number;
		recorded_at: string;
	}

	let stats = $state<HealthStats | null>(null);
	let error = $state('');

	async function fetchStats() {
		try {
			const token = localStorage.getItem('razad_token');
			const res = await fetch('/api/v1/health/stats', {
				headers: token ? { Authorization: `Bearer ${token}` } : {}
			});
			if (res.ok) {
				stats = await res.json();
				error = '';
			} else if (res.status === 401) {
				stats = null;
				error = '';
			} else {
				error = 'Failed to load system metrics';
			}
		} catch {
			error = 'Cannot reach daemon';
		}
	}

	function fmtBytes(b: number): string {
		if (b >= 1073741824) return (b / 1073741824).toFixed(1) + ' GB';
		if (b >= 1048576) return (b / 1048576).toFixed(1) + ' MB';
		if (b >= 1024) return (b / 1024).toFixed(0) + ' KB';
		return b + ' B';
	}

	function fmtUptime(sec: number): string {
		if (sec < 60) return Math.floor(sec) + 's';
		if (sec < 3600) return Math.floor(sec / 60) + 'm';
		if (sec < 86400) return Math.floor(sec / 3600) + 'h ' + Math.floor((sec % 3600) / 60) + 'm';
		return Math.floor(sec / 86400) + 'd ' + Math.floor((sec % 86400) / 3600) + 'h';
	}

	onMount(() => {
		fetchStats();
		const interval = setInterval(fetchStats, 10000);
		return () => clearInterval(interval);
	});
</script>

<svelte:head><title>Operations Center — Razad</title></svelte:head>

<h1>Operations Center</h1>

{#if error}
	<div class="error-banner">{error}</div>
{/if}

<!-- Health Strip -->
<Card title="System Health" padding="tight">
	<div class="health-strip">
		<div class="metric">
			<span class="metric-label">CPU</span>
			<span class="metric-value">{stats?.cpu_usage?.toFixed(1) ?? '—'}%</span>
		</div>
		<div class="metric">
			<span class="metric-label">RAM</span>
			<span class="metric-value">{stats?.ram_usage?.toFixed(1) ?? '—'}%</span>
			<span class="metric-max">{stats ? fmtBytes(stats.ram_used) + ' / ' + fmtBytes(stats.ram_total) : ''}</span>
		</div>
		<div class="metric">
			<span class="metric-label">Disk</span>
			<span class="metric-value">{stats?.disk_usage?.toFixed(1) ?? '—'}%</span>
			<span class="metric-max">{stats ? fmtBytes(stats.disk_free) + ' free' : ''}</span>
		</div>
		<div class="metric">
			<span class="metric-label">Load</span>
			<span class="metric-value">{stats?.load_avg_1?.toFixed(2) ?? '—'}</span>
			<span class="metric-max">{stats ? (stats.load_avg_5?.toFixed(1)) + ' / ' + (stats.load_avg_15?.toFixed(1)) : ''}</span>
		</div>
		<div class="metric">
			<span class="metric-label">Uptime</span>
			<span class="metric-value">{stats ? fmtUptime(stats.uptime_seconds) : '—'}</span>
		</div>
		<div class="metric">
			<span class="metric-label">Processes</span>
			<span class="metric-value">{stats?.process_count ?? '—'}</span>
		</div>
		<div class="metric">
			<span class="metric-label">Host</span>
			<span class="metric-value mono">{stats?.hostname ?? '—'}</span>
		</div>
	</div>
</Card>

<!-- Two-column layout -->
<div class="dashboard-grid">
	<div class="col">
		<Card title="Running Workloads">
			<div class="empty-hint">
				<p class="text-muted meta">No applications deployed yet.</p>
				<a href="/apps"><Button variant="primary" size="sm">View Apps</Button></a>
			</div>
		</Card>

		<Card title="Recent Deployments">
			<div class="empty-hint">
				<span class="text-muted meta">No recent deployments.</span>
			</div>
		</Card>
	</div>
	<div class="col">
		<Card title="Node Identity">
			{#if stats}
				<div class="identity-detail">
					<div class="id-row"><span class="label">Hostname</span><span class="value mono">{stats.hostname}</span></div>
					<div class="id-row"><span class="label">Processes</span><span class="value">{stats.process_count}</span></div>
					<div class="id-row"><span class="label">Disk Free</span><span class="value">{fmtBytes(stats.disk_free)}</span></div>
				</div>
			{/if}
		</Card>

		<Card title="Razad Advisor">
			<div class="empty-hint">
				<span class="text-muted meta">{stats ? 'System metrics available. AI monitoring coming soon.' : 'Start the daemon to enable monitoring.'}</span>
			</div>
		</Card>

		<Card title="System Logs">
			<div class="empty-hint">
				<span class="text-muted meta">Log streaming coming in Phase 7.</span>
			</div>
		</Card>
	</div>
</div>

<style>
	h1 { margin-bottom: var(--space-4); }
	h1,
	.mono,
	.metric-value,
	.metric-max,
	.id-row .value {
		overflow-wrap: anywhere;
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
		min-width: min(180px, 100%);
		padding: var(--space-2) var(--space-3);
		background: var(--bg-alt);
		border-radius: var(--radius-sm);
		flex: 1 1 180px;
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
	.mono { font-family: var(--font-mono); font-size: var(--font-size-sm); }
	.dashboard-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-4);
		margin-top: var(--space-4);
	}
	.col { display: flex; flex-direction: column; gap: var(--space-4); }
	.empty-hint {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
		align-items: flex-start;
		padding: var(--space-4) 0;
	}
	.identity-detail { display: flex; flex-direction: column; gap: var(--space-2); }
	.id-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-1) 0;
	}
	.label { font-size: var(--font-size-sm); color: var(--text-secondary); }
	.value { font-size: var(--font-size-sm); color: var(--text); }
	.error-banner {
		padding: var(--space-2) var(--space-3);
		background: color-mix(in srgb, var(--danger) 15%, transparent);
		border: 1px solid var(--danger);
		border-radius: var(--radius-sm);
		color: var(--danger);
		font-size: var(--font-size-sm);
		margin-bottom: var(--space-3);
	}

	@media (max-width: 900px) {
		.dashboard-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 640px) {
		h1 {
			font-size: var(--font-size-xl);
		}

		.health-strip {
			gap: var(--space-3);
		}

		.metric {
			min-width: 100%;
		}

		.metric-value {
			font-size: 1.125rem;
		}

		.id-row {
			flex-direction: column;
			align-items: flex-start;
			gap: var(--space-1);
		}
	}
</style>
