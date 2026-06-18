<script lang="ts">
	import Card from '$lib/components/Card.svelte';
	import Button from '$lib/components/Button.svelte';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';

	type ActionCapability = {
		name: string;
		label: string;
		description: string;
		allowed: boolean;
	};

	type Provider = {
		name: string;
		label: string;
		supported: boolean;
	};

	type AICapabilities = {
		providers: Provider[];
		allowed_actions: ActionCapability[];
		blocked_actions: ActionCapability[];
		safety_notes: string[];
	};

	type AIActionResult = {
		status: string;
		action: string;
		target?: string;
		message: string;
	};

	let capabilities: AICapabilities | null = null;
	let loading = true;
	let error = '';
	let action = 'restart_app';
	let target = '';
	let reason = '';
	let submitting = false;
	let result: AIActionResult | null = null;

	onMount(async () => {
		loading = true;
		const res = await api.get<AICapabilities>('/ai');
		if (res.data) {
			capabilities = res.data;
		} else {
			error = res.error?.message ?? 'Failed to load AI capabilities';
		}
		loading = false;
	});

	async function submitAction() {
		submitting = true;
		error = '';
		result = null;
		const res = await api.post<AIActionResult>('/ai/actions', {
			action,
			target,
			reason
		});
		if (res.data) {
			result = res.data;
		} else {
			error = res.error?.message ?? 'Failed to submit AI action';
		}
		submitting = false;
	}
</script>

<svelte:head><title>Razad AI — Razad</title></svelte:head>

<h1>Razad AI</h1>

<div class="grid">
	<Card title="AI Advisor">
		{#if loading}
			<p class="text-muted">Loading AI capabilities…</p>
		{:else if error}
			<p class="text-danger">{error}</p>
		{:else if capabilities}
			<div class="stack">
				<p class="text-muted meta">Razad AI can explain server status, watch for anomalies, and request only approved actions. Every request is written to the audit log.</p>
				<div>
					<h3>Supported providers</h3>
					<ul>
						{#each capabilities.providers as provider}
							<li>{provider.label} <span class="muted">({provider.supported ? 'BYOK supported' : 'coming soon'})</span></li>
						{/each}
					</ul>
				</div>
				<div>
					<h3>Safety notes</h3>
					<ul>
						{#each capabilities.safety_notes as note}
							<li>{note}</li>
						{/each}
					</ul>
				</div>
			</div>
		{/if}
	</Card>

	<Card title="Approved actions">
		{#if capabilities}
			<div class="two-col">
				<div>
					<h3>Allowed</h3>
					<ul>
						{#each capabilities.allowed_actions as item}
							<li><strong>{item.label}</strong> — {item.description}</li>
						{/each}
					</ul>
				</div>
				<div>
					<h3>Blocked</h3>
					<ul>
						{#each capabilities.blocked_actions as item}
							<li><strong>{item.label}</strong> — {item.description}</li>
						{/each}
					</ul>
				</div>
			</div>
		{/if}
	</Card>

	<Card title="Request a safe AI action">
		<div class="form">
			<label>
				<span>Action</span>
				<select bind:value={action}>
					<option value="restart_app">Restart app</option>
					<option value="reload_nginx">Reload Nginx</option>
					<option value="clear_app_cache">Clear app cache</option>
					<option value="restart_database_service">Restart database service</option>
					<option value="scale_worker_count">Scale worker count</option>
					<option value="send_alert_notification">Send alert notification</option>
					<option value="run_predefined_healthcheck">Run predefined healthcheck</option>
				</select>
			</label>
			<label>
				<span>Target</span>
				<input bind:value={target} placeholder="app ID, service name, or notification target" />
			</label>
			<label>
				<span>Reason</span>
				<textarea bind:value={reason} rows="3" placeholder="Why should AI request this action?"></textarea>
			</label>
			<Button onclick={submitAction} disabled={submitting}>
				{submitting ? 'Submitting…' : 'Request action'}
			</Button>
			{#if result}
				<div class="result">
					<p><strong>Status:</strong> {result.status}</p>
					<p>{result.message}</p>
				</div>
			{/if}
		</div>
	</Card>
</div>

<style>
	h1 { margin-bottom: var(--space-4); }
	.grid { display: grid; gap: var(--space-4); }
	.stack { display: grid; gap: var(--space-4); }
	.two-col { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--space-4); }
	.form { display: grid; gap: var(--space-3); }
	label { display: grid; gap: var(--space-2); }
	select, input, textarea { width: 100%; border: 1px solid var(--border); border-radius: var(--radius-md); padding: 0.75rem; background: var(--surface); color: var(--text); }
	ul { margin: 0; padding-left: 1.25rem; display: grid; gap: var(--space-2); }
	.muted { color: var(--text-muted); }
	.result { border: 1px solid var(--border); border-radius: var(--radius-md); padding: var(--space-3); background: var(--surface); }
	@media (max-width: 860px) { .two-col { grid-template-columns: 1fr; } }
</style>
