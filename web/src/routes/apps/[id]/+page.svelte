<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import Card from '$lib/components/Card.svelte';
	import Button from '$lib/components/Button.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';

	type App = {
		id: string;
		project_id: string;
		name: string;
		git_url?: string;
		runtime?: string;
		start_cmd?: string;
		status: string;
		created_at?: string;
		updated_at?: string;
	};

	type Deployment = {
		id: string;
		app_id: string;
		version: string;
		status: string;
		log?: string;
		created_at?: string;
		updated_at?: string;
	};

	type EnvVar = {
		id?: string;
		key: string;
		value: string;
	};

	type LiveLog = {
		app: string;
		message: string;
		timestamp: string;
	};

	let app = $state<App | null>(null);
	let deployments = $state<Deployment[]>([]);
	let envVars = $state<EnvVar[]>([]);
	let envDraft = $state<EnvVar[]>([{ key: '', value: '' }]);
	let liveLogs = $state<LiveLog[]>([]);
	let liveConnected = $state(false);
	let livePaused = $state(false);
	let liveError = $state('');
	let liveWs: WebSocket | null = null;
	let liveReconnectTimer: ReturnType<typeof setTimeout> | null = null;
	let loading = $state(true);
	let actionLoading = $state('');
	let savingEnv = $state(false);
	let error = $state('');
	let success = $state('');

	const appId = $derived($page.params.id ?? '');

	function authHeaders(extra: Record<string, string> = {}) {
		return {
			Authorization: `Bearer ${localStorage.getItem('razad_token') ?? ''}`,
			Accept: 'application/json',
			...extra,
		};
	}

	function badgeVariant(status: string) {
		if (status === 'running' || status === 'success') return 'success';
		if (status === 'failed' || status === 'deleted') return 'danger';
		if (status === 'deploying' || status === 'pending') return 'info';
		if (status === 'stopped') return 'warning';
		return 'neutral';
	}

	function liveWsUrl() {
		const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
		const host = location.port === '4174' || location.port === '5173'
			? `${location.hostname}:8080`
			: location.host;
		const url = new URL(`${protocol}//${host}/ws`);
		const token = localStorage.getItem('razad_token');
		if (token) url.searchParams.set('token', token);
		url.searchParams.set('app', appId);
		return url.toString();
	}

	function connectLiveLogs() {
		if (liveWs) {
			liveWs.onclose = null;
			liveWs.close();
		}
		const ws = new WebSocket(liveWsUrl());
		liveWs = ws;
		liveError = '';

		ws.onopen = () => {
			liveConnected = true;
		};

		ws.onmessage = (event) => {
			if (livePaused) return;
			try {
				const msg = JSON.parse(event.data);
				if (msg.type === 'log' && msg.payload?.app === appId) {
					liveLogs = [...liveLogs.slice(-199), msg.payload as LiveLog];
					scrollLiveLogsToBottom();
				}
			} catch {
				liveError = 'Received malformed log message.';
			}
		};

		ws.onclose = () => {
			liveConnected = false;
			if (liveReconnectTimer !== null) clearTimeout(liveReconnectTimer);
			liveReconnectTimer = setTimeout(connectLiveLogs, 3000);
		};

		ws.onerror = () => {
			liveError = 'Live log stream disconnected.';
			ws.close();
		};
	}

	function clearLiveLogs() {
		liveLogs = [];
	}

	function toggleLivePause() {
		livePaused = !livePaused;
	}

	function stopLiveLogs() {
		if (liveReconnectTimer !== null) {
			clearTimeout(liveReconnectTimer);
			liveReconnectTimer = null;
		}
		if (liveWs) {
			liveWs.onclose = null;
			liveWs.close();
			liveWs = null;
		}
	}

	function scrollLiveLogsToBottom() {
		const el = document.getElementById('live-log-stream');
		if (el) el.scrollTop = el.scrollHeight;
	}

	function updateAppAndStream() {
		void loadAppDetail();
	}

	async function loadAppDetail() {
		loading = true;
		error = '';
		success = '';

		try {
			const [appRes, deploymentsRes, envRes] = await Promise.all([
				fetch(`/api/v1/apps/${appId}`, { headers: authHeaders() }),
				fetch(`/api/v1/apps/${appId}/deployments`, { headers: authHeaders() }),
				fetch(`/api/v1/apps/${appId}/env`, { headers: authHeaders() }),
			]);

			if (appRes.status === 401 || deploymentsRes.status === 401 || envRes.status === 401) {
				error = 'Please log in first to view this app.';
				app = null;
				deployments = [];
				envVars = [];
				return;
			}

			if (!appRes.ok) {
				error = 'App not found.';
				return;
			}

			app = (await appRes.json()) as App;
			deployments = deploymentsRes.ok ? (((await deploymentsRes.json()) as Deployment[]) ?? []) : [];
			envVars = envRes.ok ? (((await envRes.json()) as EnvVar[]) ?? []) : [];
			envDraft = envVars.length > 0
				? envVars.map((item) => ({ key: item.key, value: '' }))
				: [{ key: '', value: '' }];
		} catch {
			error = 'Failed to load app details.';
		} finally {
			loading = false;
		}
	}

	async function runAction(endpoint: '/deploy' | '/restart' | '/stop', method = 'POST') {
		actionLoading = endpoint;
		error = '';
		success = '';
		try {
			const res = await fetch(`/api/v1/apps/${appId}${endpoint}`, {
				method,
				headers: authHeaders(),
			});
			const body = await res.json().catch(() => null);
			if (!res.ok) {
				error = body?.error?.message ?? 'Action failed.';
				return;
			}
			app = body as App;
			success = endpoint === '/deploy'
				? 'Deployment started.'
				: endpoint === '/restart'
					? 'App restart requested.'
					: 'App stopped.';
			await loadAppDetail();
		} catch {
			error = 'Failed to connect to daemon.';
		} finally {
			actionLoading = '';
		}
	}

	async function deleteApp() {
		if (!confirm(`Delete ${app?.name ?? 'this app'}? This cannot be undone.`)) return;
		actionLoading = 'delete';
		error = '';
		success = '';
		try {
			const res = await fetch(`/api/v1/apps/${appId}`, {
				method: 'DELETE',
				headers: authHeaders(),
			});
			const body = await res.json().catch(() => null);
			if (!res.ok) {
				error = body?.error?.message ?? 'Delete failed.';
				return;
			}
			success = 'App deleted.';
			await goto('/apps');
		} catch {
			error = 'Failed to connect to daemon.';
		} finally {
			actionLoading = '';
		}
	}

	function addEnvRow() {
		envDraft = [...envDraft, { key: '', value: '' }];
	}

	function removeEnvRow(index: number) {
		if (envDraft.length === 1) {
			envDraft = [{ key: '', value: '' }];
			return;
		}
		envDraft = envDraft.filter((_, i) => i !== index);
	}

	async function saveEnvVars() {
		savingEnv = true;
		error = '';
		success = '';
		try {
			const payload = envDraft
				.filter((row) => row.key.trim() && row.value.length > 0)
				.map((row) => ({ key: row.key.trim(), value: row.value }));

			if (payload.length === 0) {
				error = 'Add at least one environment variable key and value first.';
				return;
			}

			const res = await fetch(`/api/v1/apps/${appId}/env`, {
				method: 'PUT',
				headers: authHeaders({ 'Content-Type': 'application/json' }),
				body: JSON.stringify(payload),
			});
			const body = await res.json().catch(() => null);
			if (!res.ok) {
				error = body?.error?.message ?? 'Saving environment variables failed.';
				return;
			}
			success = 'Environment variables saved.';
			envDraft = [{ key: '', value: '' }];
			await loadAppDetail();
		} catch {
			error = 'Failed to connect to daemon.';
		} finally {
			savingEnv = false;
		}
	}

	onMount(() => {
		connectLiveLogs();
		void loadAppDetail();
		return () => stopLiveLogs();
	});

	onDestroy(() => {
		stopLiveLogs();
	});
</script>

<svelte:head><title>{app?.name ?? 'App'} — Razad</title></svelte:head>

{#if loading}
	<p class="text-muted">Loading app details...</p>
{:else if error && !app}
	<div class="empty-state">
		<p class="err">{error}</p>
		<a href="/apps"><Button variant="secondary" size="sm">Back to apps</Button></a>
	</div>
{:else if app}
	<div class="page-shell">
		<div class="page-header">
			<div>
				<p class="eyebrow">Application detail</p>
				<h1>{app.name}</h1>
				<p class="text-muted mono">{app.id}</p>
			</div>
			<div class="header-actions">
				<a href="/apps"><Button variant="ghost" size="sm">Back</Button></a>
				<Button variant="primary" size="sm" disabled={actionLoading === '/deploy'} onclick={() => runAction('/deploy')}>Deploy</Button>
				<Button variant="secondary" size="sm" disabled={actionLoading === '/restart'} onclick={() => runAction('/restart')}>Restart</Button>
				<Button variant="ghost" size="sm" disabled={actionLoading === '/stop'} onclick={() => runAction('/stop')}>Stop</Button>
				<Button variant="danger" size="sm" disabled={actionLoading === 'delete'} onclick={deleteApp}>Delete</Button>
			</div>
		</div>

		{#if success}
			<div class="flash success">{success}</div>
		{/if}
		{#if error && app}
			<div class="flash error">{error}</div>
		{/if}

		<div class="grid">
			<Card title="Summary" padding="loose">
				<div class="summary-grid">
					<div>
						<div class="label">Status</div>
						<StatusBadge status={app.status} variant={badgeVariant(app.status)} />
					</div>
					<div>
						<div class="label">Runtime</div>
						<div>{app.runtime ?? 'unknown'}</div>
					</div>
					<div>
						<div class="label">Project</div>
						<div class="mono">{app.project_id}</div>
					</div>
					<div>
						<div class="label">Start command</div>
						<div class="mono break">{app.start_cmd ?? 'No start command configured'}</div>
					</div>
					<div>
						<div class="label">Git repository</div>
						<div class="break">{app.git_url ?? 'No repository linked'}</div>
					</div>
				</div>
			</Card>

			<Card title="Environment variables" padding="loose">
				{#if envVars.length === 0}
					<p class="text-muted">No environment variables saved yet.</p>
				{:else}
					<ul class="list">
						{#each envVars as item}
							<li>
								<strong>{item.key}</strong>
								<span class="text-muted">value hidden by backend</span>
							</li>
						{/each}
					</ul>
				{/if}
				<div class="divider"></div>
				<div class="editor-head">
					<h3>Update / add variables</h3>
					<Button variant="secondary" size="sm" onclick={addEnvRow}>Add row</Button>
				</div>
				<div class="env-editor">
					{#each envDraft as row, index}
						<div class="env-row">
							<input class="input" placeholder="KEY" bind:value={row.key} />
							<input class="input" type="password" placeholder="value" bind:value={row.value} />
							<Button variant="ghost" size="sm" onclick={() => removeEnvRow(index)}>Remove</Button>
						</div>
					{/each}
				</div>
				<div class="editor-actions">
					<Button variant="primary" size="sm" disabled={savingEnv} onclick={saveEnvVars}>Save env vars</Button>
				</div>
			</Card>

			<Card title="Deployments" padding="loose">
				{#if deployments.length === 0}
					<p class="text-muted">No deployments yet.</p>
				{:else}
					<div class="deployment-list">
						{#each deployments as deployment}
							<div class="deployment-item">
								<div class="deployment-top">
									<div>
										<strong>{deployment.version}</strong>
										<div class="text-muted mono">{deployment.id}</div>
									</div>
									<StatusBadge status={deployment.status} variant={badgeVariant(deployment.status)} />
								</div>
								{#if deployment.log}
									<pre class="log">{deployment.log}</pre>
								{/if}
							</div>
						{/each}
					</div>
				{/if}
			</Card>

			<div class="live-card">
				<Card title="Live logs" padding="loose">
					<div class="live-toolbar">
						<div class="log-status">
							<span class="status-dot" class:ok={liveConnected} class:err={!liveConnected}></span>
							<span class="text-muted meta">{liveConnected ? 'Connected' : 'Disconnected — reconnecting...'}</span>
						</div>
						<div class="log-actions">
							<Button variant="ghost" size="sm" onclick={toggleLivePause}>
								{livePaused ? 'Resume' : 'Pause'}
							</Button>
							<Button variant="ghost" size="sm" onclick={clearLiveLogs}>Clear</Button>
						</div>
					</div>
					{#if liveError}
						<p class="text-muted">{liveError}</p>
					{/if}
					<div id="live-log-stream" class="live-log-viewer">
						{#if liveLogs.length === 0}
							<div class="log-empty">
								<span class="text-muted meta">Waiting for live log data...</span>
							</div>
						{:else}
							{#each liveLogs as log}
								<div class="log-line">
									<span class="log-time">{log.timestamp?.slice(11, 19) ?? ''}</span>
									<span class="log-app mono">{log.app}</span>
									<span class="log-msg mono">{log.message}</span>
								</div>
							{/each}
						{/if}
					</div>
				</Card>
			</div>
		</div>
	</div>
{/if}

<style>
	.page-shell {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: var(--space-4);
		flex-wrap: wrap;
	}
	.page-header h1 {
		margin: 0 0 var(--space-1);
	}
	.eyebrow {
		margin: 0 0 var(--space-1);
		text-transform: uppercase;
		letter-spacing: 0.08em;
		font-size: var(--font-size-xs);
		color: var(--text-muted);
	}
	.header-actions {
		display: flex;
		gap: var(--space-2);
		flex-wrap: wrap;
	}
	.flash {
		padding: var(--space-3) var(--space-4);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
	}
	.flash.success {
		background: var(--success-bg);
		color: var(--success);
	}
	.flash.error {
		background: var(--danger-bg);
		color: var(--danger);
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: var(--space-4);
	}
	.summary-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: var(--space-4);
	}
	.label {
		font-size: var(--font-size-xs);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-muted);
		margin-bottom: var(--space-1);
	}
	.list {
		list-style: none;
		padding: 0;
		margin: 0;
		display: grid;
		gap: var(--space-2);
	}
	.list li {
		display: flex;
		justify-content: space-between;
		gap: var(--space-3);
		padding: var(--space-2) 0;
		border-bottom: 1px solid var(--border);
	}
	.list li:last-child {
		border-bottom: 0;
	}
	.divider {
		height: 1px;
		background: var(--border);
		margin: var(--space-4) 0;
	}
	.editor-head,
	.deployment-top {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: var(--space-3);
		flex-wrap: wrap;
	}
	.env-editor {
		display: grid;
		gap: var(--space-2);
		margin-top: var(--space-3);
	}
	.env-row {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
		gap: var(--space-2);
	}
	.input {
		width: 100%;
		padding: 0.6rem 0.75rem;
		border-radius: var(--radius-sm);
		border: 1px solid var(--border);
		background: var(--surface-2);
		color: var(--text);
	}
	.editor-actions {
		display: flex;
		justify-content: flex-end;
		margin-top: var(--space-3);
	}
	.deployment-list {
		display: grid;
		gap: var(--space-3);
	}
	.deployment-item {
		padding: var(--space-3);
		border-radius: var(--radius-sm);
		border: 1px solid var(--border);
		background: var(--surface-2);
	}
	.live-card {
		grid-column: 1 / -1;
	}
	.live-toolbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		margin-bottom: var(--space-3);
		flex-wrap: wrap;
	}
	.log-status {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}
	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--danger);
	}
	.status-dot.ok { background: var(--success); }
	.status-dot.err { background: var(--danger); }
	.log-actions {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}
	.live-log-viewer {
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		max-height: 320px;
		overflow-y: auto;
		padding: var(--space-2) 0;
		font-family: var(--font-mono);
		font-size: var(--font-size-xs);
		line-height: 1.6;
	}
	.log-line {
		display: flex;
		gap: var(--space-3);
		padding: 0.125rem var(--space-3);
		transition: background 0.1s;
	}
	.log-line:hover { background: var(--surface); }
	.log-time {
		color: var(--text-muted);
		flex-shrink: 0;
		width: 5rem;
	}
	.log-app {
		color: var(--primary);
		flex-shrink: 0;
		width: 7rem;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.log-msg {
		color: var(--text);
		white-space: pre-wrap;
		word-break: break-all;
		flex: 1;
	}
	.log-empty {
		padding: var(--space-6) var(--space-3);
		text-align: center;
	}
	.log,
	.break,
	.mono {
		white-space: pre-wrap;
		word-break: break-word;
	}
	.err { color: var(--danger); }
	.text-muted { color: var(--text-muted); }
	@media (max-width: 900px) {
		.grid,
		.summary-grid {
			grid-template-columns: 1fr;
		}
		.env-row {
			grid-template-columns: 1fr;
		}
		.live-card {
			grid-column: auto;
		}
	}

	@media (max-width: 720px) {
		.list li,
		.log-line {
			flex-direction: column;
			align-items: flex-start;
		}

		.log-time,
		.log-app {
			width: auto;
		}

		.log-actions,
		.header-actions,
		.editor-actions {
			width: 100%;
		}

		.log-actions,
		.header-actions {
			justify-content: flex-start;
		}
	}
</style>
