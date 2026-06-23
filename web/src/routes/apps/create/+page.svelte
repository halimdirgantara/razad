<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import Card from '$lib/components/Card.svelte';
	import { goto } from '$app/navigation';

	let name = $state('');
	let gitUrl = $state('');
	let runtime = $state('');
	let startCmd = $state('');
	let error = $state('');
	let loading = $state(false);

	const runtimes = ['node', 'go', 'python', 'php', 'ruby', 'bun', ''];

	async function handleSubmit() {
		error = '';
		if (!name) {
			error = 'App name is required.';
			return;
		}
		loading = true;

		try {
			const res = await fetch('/api/v1/apps', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${localStorage.getItem('razad_token')}`
				},
				body: JSON.stringify({
					name,
					git_url: gitUrl || undefined,
					runtime: runtime || undefined,
					start_cmd: startCmd || undefined,
				})
			});

			if (!res.ok) {
				const data = await res.json();
				error = data.error?.message ?? 'Failed to create app.';
				return;
			}

			const app = await res.json();
			goto(`/apps/${app.id}`);
		} catch {
			error = 'Failed to connect to daemon.';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head><title>Create App — Razad</title></svelte:head>

<h1>Create App</h1>

<Card title="New Application">
	<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
		<div class="field">
			<label for="name">App Name</label>
			<input id="name" type="text" bind:value={name} placeholder="my-app" required />
		</div>

		<div class="field">
			<label for="git">Git URL</label>
			<input id="git" type="text" bind:value={gitUrl} placeholder="https://github.com/user/repo.git" />
			<span class="hint">Optional. Upload manually if left empty.</span>
		</div>

		<div class="field-row">
			<div class="field">
				<label for="runtime">Runtime</label>
				<select id="runtime" bind:value={runtime}>
					<option value="">Auto-detect</option>
					{#each runtimes as r}
						{#if r}
							<option value={r}>{r}</option>
						{/if}
					{/each}
				</select>
			</div>
			<div class="field">
				<label for="cmd">Start Command</label>
				<input id="cmd" type="text" bind:value={startCmd} placeholder="npm start" />
			</div>
		</div>

		{#if error}
			<p class="error-msg">{error}</p>
		{/if}

		<div class="actions">
			<Button type="submit" variant="primary" size="lg" disabled={loading}>
				{loading ? 'Creating...' : 'Create App'}
			</Button>
		</div>
	</form>
</Card>

<style>
	h1 { margin-bottom: var(--space-4); }
	.field { margin-bottom: var(--space-4); flex: 1; }
	.field-row { display: flex; gap: var(--space-4); }
	label {
		display: block;
		font-size: var(--font-size-sm);
		font-weight: var(--font-weight-medium);
		color: var(--text-secondary);
		margin-bottom: var(--space-1);
	}
	input, select {
		width: 100%;
		padding: var(--space-2) var(--space-3);
		background: var(--bg-alt);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		color: var(--text);
		font-size: var(--font-size-base);
		outline: none;
	}
	input:focus, select:focus { border-color: var(--primary); }
	.hint {
		font-size: var(--font-size-xs);
		color: var(--text-muted);
		margin-top: var(--space-1);
		display: block;
	}
	.error-msg {
		color: var(--danger);
		font-size: var(--font-size-sm);
		margin-bottom: var(--space-3);
	}
	.actions { margin-top: var(--space-2); }

	@media (max-width: 720px) {
		.field-row {
			flex-direction: column;
			gap: 0;
		}
	}
</style>
