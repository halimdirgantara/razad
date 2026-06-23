<script lang="ts">
	import Card from '$lib/components/Card.svelte';
	import Button from '$lib/components/Button.svelte';
	import { onMount, onDestroy } from 'svelte';

	let logs: Array<{ app: string; message: string; timestamp: string }> = $state([]);
	let connected = $state(false);
	let paused = $state(false);
	let filterText = $state('');
	let ws: WebSocket | null = null;
	let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

	const filteredLogs = $derived(
		filterText
			? logs.filter(l => l.message.toLowerCase().includes(filterText.toLowerCase()))
			: logs
	);

	function connect() {
		const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
		const url = `${protocol}//${location.host}/ws`;

		ws = new WebSocket(url);

		ws.onopen = () => { connected = true; };

		ws.onmessage = (event) => {
			if (paused) return;

			try {
				const msg = JSON.parse(event.data);
				if (msg.type === 'log') {
					logs = [...logs.slice(-499), msg.payload];
				}
			} catch { /* skip malformed */ }
		};

		ws.onclose = () => {
			connected = false;
			reconnectTimer = setTimeout(connect, 3000);
		};

		ws.onerror = () => ws?.close();
	}

	function clearLogs() { logs = []; }

	onMount(connect);

	onDestroy(() => {
		if (reconnectTimer !== null) clearTimeout(reconnectTimer);
		if (ws) {
			ws.onclose = null;
			ws.close();
		}
	});
</script>

<svelte:head><title>Logs — Razad</title></svelte:head>

<h1>Logs</h1>

<div class="log-toolbar">
	<div class="log-status">
		<span class="status-dot" class:ok={connected} class:err={!connected}></span>
		<span class="text-muted meta">{connected ? 'Connected' : 'Disconnected — reconnecting...'}</span>
	</div>
	<div class="log-actions">
		<input
			type="text"
			placeholder="Filter logs..."
			bind:value={filterText}
			class="filter-input"
		/>
		<Button variant="ghost" size="sm" onclick={() => paused = !paused}>
			{paused ? 'Resume' : 'Pause'}
		</Button>
		<Button variant="ghost" size="sm" onclick={clearLogs}>Clear</Button>
	</div>
</div>

<Card title="Live Log Stream" padding="tight">
	<div class="log-viewer">
		{#if filteredLogs.length === 0}
			<div class="log-empty">
				<span class="text-muted meta">Waiting for log data{filterText ? ' matching "' + filterText + '"' : '...'}</span>
			</div>
		{:else}
			{#each filteredLogs as log}
				<div class="log-line">
					<span class="log-time">{log.timestamp?.slice(11, 19) ?? ''}</span>
					<span class="log-app mono">{log.app}</span>
					<span class="log-msg mono">{log.message}</span>
				</div>
			{/each}
		{/if}
	</div>
</Card>

<style>
	h1 { margin-bottom: var(--space-3); }
	.log-toolbar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-3);
		gap: var(--space-4);
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
		flex-wrap: wrap;
	}
	.filter-input {
		padding: var(--space-1) var(--space-2);
		background: var(--bg-alt);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		color: var(--text);
		font-size: var(--font-size-sm);
		width: 180px;
		outline: none;
	}
	.filter-input:focus { border-color: var(--primary); }

	.log-viewer {
		background: var(--bg-alt);
		border-radius: var(--radius-sm);
		padding: var(--space-2) 0;
		max-height: 65vh;
		overflow-y: auto;
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

	@media (max-width: 720px) {
		.log-toolbar {
			flex-direction: column;
			align-items: stretch;
		}

		.log-actions {
			width: 100%;
		}

		.filter-input {
			width: 100%;
			flex: 1 1 100%;
		}
	}
</style>
