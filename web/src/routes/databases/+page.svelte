<script lang="ts">
	import Card from '$lib/components/Card.svelte';
	import Button from '$lib/components/Button.svelte';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';

	type DatabaseInstance = {
		id: string;
		owner_user_id: string;
		name: string;
		engine: string;
		version: string;
		host: string;
		port: number;
		username: string;
		password: string;
		database_name: string;
		status: string;
		connection_string: string;
		created_at: string;
		updated_at: string;
	};

	let databases = $state<DatabaseInstance[]>([]);
	let loading = $state(true);
	let error = $state('');
	let success = $state('');
	let creating = $state(false);

	let name = $state('');
	let engine = $state('postgresql');
	let version = $state('');

	async function loadDatabases() {
		loading = true;
		try {
			const res = await api.get<DatabaseInstance[]>('/databases');
			if (res.data) {
				databases = res.data;
				error = '';
			} else {
				error = res.error?.message ?? 'Failed to load databases';
			}
		} catch {
			error = 'Cannot reach daemon';
		} finally {
			loading = false;
		}
	}

	async function createDatabase() {
		creating = true;
		error = '';
		success = '';
		try {
			const res = await api.post<DatabaseInstance>('/databases', {
				name,
				engine,
				version: version || undefined
			});
			if (!res.data) {
				error = res.error?.message ?? 'Failed to create database';
				return;
			}
			const created = res.data;
			databases = [created, ...databases];
			name = '';
			version = '';
			success = `${created.name} provisioned successfully.`;
		} catch {
			error = 'Cannot reach daemon';
		} finally {
			creating = false;
		}
	}

	onMount(loadDatabases);
</script>

<svelte:head><title>Databases — Razad</title></svelte:head>

<h1>Databases</h1>

{#if error}
	<div class="banner error">{error}</div>
{/if}
{#if success}
	<div class="banner success">{success}</div>
{/if}

<div class="grid">
	<Card title="Provision Database">
		<div class="form">
			<label>
				<span>Name</span>
				<input bind:value={name} placeholder="Primary PostgreSQL" />
			</label>
			<label>
				<span>Engine</span>
				<select bind:value={engine}>
					<option value="postgresql">PostgreSQL</option>
					<option value="mysql">MySQL</option>
					<option value="redis">Redis</option>
				</select>
			</label>
			<label>
				<span>Version <small>(optional)</small></span>
				<input bind:value={version} placeholder="Auto-detect default" />
			</label>
			<div class="actions">
				<Button variant="primary" onclick={createDatabase} disabled={creating || !name.trim()}>
					{creating ? 'Provisioning…' : 'Provision database'}
				</Button>
			</div>
		</div>
		<p class="text-muted meta">Creates the database record, credentials, and connection details for the selected engine.</p>
	</Card>

	<Card title="Database Instances">
		{#if loading}
			<div class="empty-state">
				<p class="text-muted meta">Loading databases…</p>
			</div>
		{:else if databases.length === 0}
			<div class="empty-state">
				<div class="empty-icon">⛁</div>
				<p class="empty-title">No databases provisioned</p>
				<p class="text-muted meta">Provision PostgreSQL, MySQL, or Redis instances from this page.</p>
			</div>
		{:else}
			<div class="instances">
				{#each databases as db}
					<article class="instance">
						<div class="instance-head">
							<div>
								<h3>{db.name}</h3>
								<p class="text-muted meta">{db.engine} · v{db.version} · {db.status}</p>
							</div>
							<span class="badge">{db.database_name}</span>
						</div>
						<div class="info-grid">
							<div><span class="label">Host</span><code>{db.host}:{db.port}</code></div>
							<div><span class="label">Username</span><code>{db.username}</code></div>
							<div><span class="label">Password</span><code>{db.password}</code></div>
							<div><span class="label">Connection</span><code>{db.connection_string}</code></div>
						</div>
					</article>
				{/each}
			</div>
		{/if}
	</Card>
</div>

<style>
	h1 { margin-bottom: var(--space-4); }
	.grid { display: grid; gap: var(--space-4); }
	.form {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: var(--space-3);
		align-items: end;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.375rem;
	}
	span { font-size: var(--font-size-sm); color: var(--text-secondary); }
	input, select {
		padding: var(--space-2) var(--space-3);
		border-radius: var(--radius-sm);
		border: 1px solid var(--border);
		background: var(--surface);
		color: var(--text);
	}
	.actions { display: flex; justify-content: flex-start; }
	.banner {
		padding: var(--space-2) var(--space-3);
		border-radius: var(--radius-sm);
		margin-bottom: var(--space-3);
		font-size: var(--font-size-sm);
	}
	.banner.error { border: 1px solid var(--danger); color: var(--danger); background: color-mix(in srgb, var(--danger) 14%, transparent); }
	.banner.success { border: 1px solid var(--success, #2ea043); color: var(--success, #2ea043); background: color-mix(in srgb, var(--success, #2ea043) 12%, transparent); }
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-2);
		padding: var(--space-10) var(--space-4);
		text-align: center;
	}
	.empty-icon { font-size: 2rem; opacity: 0.3; margin-bottom: var(--space-2); }
	.empty-title { font-size: var(--font-size-base); font-weight: var(--font-weight-medium); color: var(--text); }
	.instances { display: grid; gap: var(--space-3); }
	.instance { padding: var(--space-4); border: 1px solid var(--border); border-radius: var(--radius-md); background: var(--surface); }
	.instance-head { display: flex; align-items: flex-start; justify-content: space-between; gap: var(--space-3); margin-bottom: var(--space-3); }
	h3 { margin: 0 0 0.25rem 0; font-size: var(--font-size-base); }
	.badge {
		padding: 0.25rem 0.5rem;
		border-radius: var(--radius-sm);
		background: var(--surface-3);
		color: var(--text-secondary);
		font-size: var(--font-size-xs);
		font-family: var(--font-mono);
	}
	.info-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--space-3); }
	.info-grid > div { display: flex; flex-direction: column; gap: 0.25rem; }
	.label { font-size: var(--font-size-xs); text-transform: uppercase; letter-spacing: 0.04em; }
	code {
		padding: 0.5rem 0.625rem;
		border-radius: var(--radius-sm);
		background: var(--bg-alt);
		border: 1px solid var(--border);
		white-space: break-spaces;
		word-break: break-word;
	}
	@media (max-width: 900px) {
		.form, .info-grid { grid-template-columns: 1fr; }
	}
</style>
