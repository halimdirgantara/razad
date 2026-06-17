<script lang="ts">
	import Card from '$lib/components/Card.svelte';
	import Button from '$lib/components/Button.svelte';
	import { onMount } from 'svelte';

	type ProxyMode = 'render' | 'apply' | 'rollback';
	type SSLMode = 'issue' | 'renew';

	let activeTab = $state<'proxy' | 'ssl'>('proxy');
	let loading = $state(false);
	let error = $state('');
	let success = $state('');
	let output = $state('');
	let outputLabel = $state('');

	let proxyName = $state('web');
	let proxyDomain = $state('app.example.com');
	let proxyUpstreamHost = $state('127.0.0.1');
	let proxyUpstreamPort = $state(3000);
	let proxyTls = $state(true);
	let proxyBodyLimitMb = $state(20);

	let sslDomain = $state('app.example.com');
	let sslEmail = $state('ops@example.com');
	let sslWebroot = $state('/var/www/html');

	const quickPresets = [
		{
			label: 'Node app',
			name: 'node-web',
			domain: 'web.example.com',
			upstreamHost: '127.0.0.1',
			upstreamPort: 3000,
			tls: true,
			bodyLimitMb: 20,
		},
		{
			label: 'Service app',
			name: 'service-api',
			domain: 'api.example.com',
			upstreamHost: '10.0.0.12',
			upstreamPort: 8080,
			tls: false,
			bodyLimitMb: 50,
		},
	];

	function authHeaders() {
		return {
			'Content-Type': 'application/json',
			'Authorization': `Bearer ${localStorage.getItem('razad_token')}`,
		};
	}

	function pretty(value: unknown): string {
		return JSON.stringify(value, null, 2);
	}

	function proxyPayload() {
		return {
			name: proxyName,
			domain: proxyDomain,
			upstream_host: proxyUpstreamHost,
			upstream_port: proxyUpstreamPort,
			tls: proxyTls,
			body_limit_mb: proxyBodyLimitMb,
		};
	}

	function sslIssuePayload() {
		return {
			domain: sslDomain,
			email: sslEmail,
			webroot: sslWebroot,
		};
	}

	async function submitProxy(mode: ProxyMode) {
		loading = true;
		error = '';
		success = '';
		output = '';
		outputLabel = '';
		try {
			const endpoint = mode === 'render'
				? '/api/v1/proxy/render'
				: mode === 'apply'
					? '/api/v1/proxy/apply'
					: '/api/v1/proxy/rollback';
			const res = await fetch(endpoint, {
				method: 'POST',
				headers: authHeaders(),
				body: JSON.stringify(proxyPayload()),
			});
			const data = await res.json().catch(() => null);
			if (!res.ok) {
				error = data?.error?.message ?? 'Proxy action failed.';
				return;
			}
			outputLabel = mode === 'render' ? 'Rendered Nginx config' : mode === 'apply' ? 'Apply result' : 'Rollback result';
			output = data?.config ? data.config : pretty(data);
			success = mode === 'render' ? 'Config rendered successfully.' : mode === 'apply' ? 'Proxy config applied.' : 'Proxy config rolled back.';
		} catch {
			error = 'Failed to connect to daemon.';
		} finally {
			loading = false;
		}
	}

	async function submitSSL(mode: SSLMode) {
		loading = true;
		error = '';
		success = '';
		output = '';
		outputLabel = '';
		try {
			const endpoint = mode === 'issue' ? '/api/v1/ssl/issue' : '/api/v1/ssl/renew';
			const body = mode === 'issue' ? sslIssuePayload() : { domain: sslDomain };
			const res = await fetch(endpoint, {
				method: 'POST',
				headers: authHeaders(),
				body: JSON.stringify(body),
			});
			const data = await res.json().catch(() => null);
			if (!res.ok) {
				error = data?.error?.message ?? 'SSL action failed.';
				return;
			}
			outputLabel = mode === 'issue' ? 'certbot issuance command' : 'certbot renewal command';
			output = data?.command ? data.command : pretty(data);
			success = mode === 'issue' ? 'Certificate issuance command ready.' : 'Renewal command ready.';
		} catch {
			error = 'Failed to connect to daemon.';
		} finally {
			loading = false;
		}
	}

	function applyPreset(preset: typeof quickPresets[number]) {
		proxyName = preset.name;
		proxyDomain = preset.domain;
		proxyUpstreamHost = preset.upstreamHost;
		proxyUpstreamPort = preset.upstreamPort;
		proxyTls = preset.tls;
		proxyBodyLimitMb = preset.bodyLimitMb;
		sslDomain = preset.domain;
	}

	onMount(() => {
		if (!sslDomain) sslDomain = proxyDomain;
	});
</script>

<svelte:head><title>Domains — Razad</title></svelte:head>

<div class="page-header">
	<div>
		<h1>Domains & SSL</h1>
		<p class="text-muted">Render, apply, and rollback Nginx bindings, then prepare certbot commands for TLS.</p>
	</div>
	<div class="header-actions">
		<Button variant={activeTab === 'proxy' ? 'primary' : 'secondary'} size="sm" onclick={() => activeTab = 'proxy'}>Proxy</Button>
		<Button variant={activeTab === 'ssl' ? 'primary' : 'secondary'} size="sm" onclick={() => activeTab = 'ssl'}>SSL</Button>
	</div>
</div>

<div class="grid">
	<div class="main-col">
		<Card title="Quick presets" padding="tight">
			<div class="preset-list">
				{#each quickPresets as preset}
					<button class="preset" onclick={() => applyPreset(preset)}>
						<div>
							<strong>{preset.label}</strong>
							<p class="text-muted">{preset.domain} → {preset.upstreamHost}:{preset.upstreamPort}</p>
						</div>
						<span class="preset-pill">Use</span>
					</button>
				{/each}
			</div>
		</Card>

		{#if activeTab === 'proxy'}
			<Card title="Proxy binding" padding="loose">
				<div class="form-grid">
					<div class="field">
						<label for="proxy-name">Binding name</label>
						<input id="proxy-name" bind:value={proxyName} placeholder="web" />
					</div>
					<div class="field">
						<label for="proxy-domain">Domain</label>
						<input id="proxy-domain" bind:value={proxyDomain} placeholder="app.example.com" />
					</div>
					<div class="field-row">
						<div class="field">
							<label for="proxy-host">Upstream host</label>
							<input id="proxy-host" bind:value={proxyUpstreamHost} placeholder="127.0.0.1" />
						</div>
						<div class="field narrow">
							<label for="proxy-port">Port</label>
							<input id="proxy-port" type="number" bind:value={proxyUpstreamPort} min="1" max="65535" />
						</div>
					</div>
					<div class="field-row">
						<div class="field narrow">
							<label for="proxy-limit">Body limit (MB)</label>
							<input id="proxy-limit" type="number" bind:value={proxyBodyLimitMb} min="0" />
						</div>
						<div class="field checkbox-field">
							<label for="proxy-tls">TLS enabled</label>
							<div class="toggle-row">
								<input id="proxy-tls" type="checkbox" bind:checked={proxyTls} />
								<span class="hint">Generate HTTPS server block and redirect HTTP to HTTPS.</span>
							</div>
						</div>
					</div>
				</div>

				<div class="actions">
					<Button variant="secondary" size="md" disabled={loading} onclick={() => submitProxy('render')}>Render config</Button>
					<Button variant="primary" size="md" disabled={loading} onclick={() => submitProxy('apply')}>Apply</Button>
					<Button variant="ghost" size="md" disabled={loading} onclick={() => submitProxy('rollback')}>Rollback</Button>
				</div>
			</Card>
		{:else}
			<Card title="SSL issuance" padding="loose">
				<div class="form-grid">
					<div class="field">
						<label for="ssl-domain">Domain</label>
						<input id="ssl-domain" bind:value={sslDomain} placeholder="app.example.com" />
					</div>
					<div class="field-row">
						<div class="field">
							<label for="ssl-email">Email</label>
							<input id="ssl-email" bind:value={sslEmail} placeholder="ops@example.com" />
						</div>
						<div class="field narrow">
							<label for="ssl-webroot">Webroot</label>
							<input id="ssl-webroot" bind:value={sslWebroot} placeholder="/var/www/html" />
						</div>
					</div>
				</div>
				<div class="actions">
					<Button variant="primary" size="md" disabled={loading} onclick={() => submitSSL('issue')}>Issue certificate</Button>
					<Button variant="secondary" size="md" disabled={loading} onclick={() => submitSSL('renew')}>Renew command</Button>
				</div>
			</Card>
		{/if}

		<Card title="Preview" padding="loose">
			{#if error}
				<p class="banner error">{error}</p>
			{/if}
			{#if success}
				<p class="banner success">{success}</p>
			{/if}
			{#if output}
				<div class="output-header">
					<span class="output-label">{outputLabel}</span>
					<span class="text-muted">Copy this into the daemon or use the UI action above.</span>
				</div>
				<pre class="output">{output}</pre>
			{:else}
				<p class="text-muted">Generate a proxy config or certbot command to see a preview here.</p>
			{/if}
		</Card>
	</div>

	<div class="side-col">
		<Card title="What this page does" padding="tight">
			<ul class="notes">
				<li>Compose Nginx server blocks for HTTP or HTTPS.</li>
				<li>Keep a rollback snapshot before promoting new config.</li>
				<li>Prepare certbot webroot issuance and renewal commands.</li>
				<li>Log every sensitive action to the audit trail.</li>
			</ul>
		</Card>

		<Card title="Suggested flow" padding="tight">
			<ol class="notes ordered">
				<li>Pick a preset or fill the proxy form.</li>
				<li>Render first to inspect the generated Nginx config.</li>
				<li>Apply when ready; rollback is one click away.</li>
				<li>Switch to SSL and issue the certificate after the DNS points here.</li>
			</ol>
		</Card>
	</div>
</div>

<style>
	h1 { margin: 0 0 var(--space-2); }
	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: var(--space-4);
		margin-bottom: var(--space-5);
	}
	.page-header p { max-width: 56rem; margin: 0; }
	.header-actions { display: flex; gap: var(--space-2); flex-wrap: wrap; }
	.grid {
		display: grid;
		grid-template-columns: minmax(0, 1.8fr) minmax(280px, 0.9fr);
		gap: var(--space-4);
		align-items: start;
	}
	.main-col, .side-col {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.preset-list {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: var(--space-3);
	}
	.preset {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: var(--space-3);
		padding: var(--space-3);
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
		color: inherit;
		text-align: left;
		cursor: pointer;
	}
	.preset:hover { background: var(--surface-3); }
	.preset strong { display: block; margin-bottom: 0.125rem; }
	.preset p { margin: 0; font-size: var(--font-size-xs); }
	.preset-pill {
		padding: 0.2rem 0.5rem;
		border-radius: 999px;
		font-size: var(--font-size-xs);
		border: 1px solid var(--border);
		color: var(--text-muted);
		flex-shrink: 0;
	}
	.form-grid { display: grid; gap: var(--space-4); }
	.field { display: flex; flex-direction: column; gap: var(--space-1); }
	.field-row { display: grid; grid-template-columns: minmax(0, 1fr) 220px; gap: var(--space-3); }
	.field.narrow { min-width: 0; }
	.checkbox-field { justify-content: flex-start; }
	.toggle-row { display: flex; align-items: center; gap: var(--space-2); min-height: 42px; }
	label {
		font-size: var(--font-size-sm);
		font-weight: var(--font-weight-medium);
		color: var(--text-secondary);
	}
	input {
		width: 100%;
		padding: var(--space-2) var(--space-3);
		background: var(--bg-alt);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		color: var(--text);
		font-size: var(--font-size-base);
		outline: none;
	}
	input:focus { border-color: var(--primary); }
	input[type='checkbox'] {
		width: 18px;
		height: 18px;
		padding: 0;
	}
	.hint {
		font-size: var(--font-size-xs);
		color: var(--text-muted);
	}
	.actions {
		display: flex;
		gap: var(--space-2);
		flex-wrap: wrap;
		margin-top: var(--space-1);
	}
	.banner {
		padding: var(--space-3);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
		margin-bottom: var(--space-3);
	}
	.banner.error {
		background: var(--danger-bg);
		color: var(--danger);
		border: 1px solid var(--danger);
	}
	.banner.success {
		background: var(--success-bg);
		color: var(--success);
		border: 1px solid var(--success);
	}
	.output-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: var(--space-3);
		margin-bottom: var(--space-2);
	}
	.output-label {
		font-size: var(--font-size-xs);
		font-weight: var(--font-weight-semibold);
		text-transform: uppercase;
		letter-spacing: 0.04em;
		color: var(--text-muted);
	}
	.output {
		margin: 0;
		padding: var(--space-4);
		background: var(--bg-alt);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		color: var(--text);
		font-size: var(--font-size-xs);
		line-height: 1.6;
		overflow-x: auto;
		white-space: pre-wrap;
		word-break: break-word;
	}
	.notes {
		margin: 0;
		padding-left: var(--space-4);
		color: var(--text-secondary);
		font-size: var(--font-size-sm);
		display: grid;
		gap: var(--space-2);
	}
	.notes.ordered {
		list-style: decimal;
	}
	@media (max-width: 1080px) {
		.grid { grid-template-columns: 1fr; }
		.preset-list { grid-template-columns: 1fr; }
		.field-row { grid-template-columns: 1fr; }
	}
	@media (max-width: 720px) {
		.page-header { flex-direction: column; }
		.header-actions { width: 100%; }
	}
</style>
