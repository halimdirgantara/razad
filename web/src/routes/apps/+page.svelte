<script lang="ts">
	import Card from '$lib/components/Card.svelte';
	import DataTable from '$lib/components/DataTable.svelte';
	import Button from '$lib/components/Button.svelte';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let apps = $state<Array<Record<string, unknown>>>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		const token = localStorage.getItem('razad_token');
		const headers = { 'Authorization': `Bearer ${token}` };

		try {
			const res = await fetch('/api/v1/apps', { headers });
			if (res.ok) {
				apps = await res.json();
			} else if (res.status === 401) {
				error = 'Please log in first.';
			} else {
				apps = [];
			}
		} catch {
			error = 'Could not connect to daemon.';
		} finally {
			loading = false;
		}
	});

	function viewApp(row: Record<string, unknown>) {
		goto(`/apps/${row.id}`);
	}

	const columns = [
		{ key: 'name', label: 'Name' },
		{ key: 'runtime', label: 'Runtime' },
		{ key: 'status', label: 'Status' },
		{ key: 'updated_at', label: 'Updated' },
	];
</script>

<svelte:head><title>Applications — Razad</title></svelte:head>

<h1>Applications</h1>

<Card title="All Applications">
	{#if loading}
		<p class="text-muted">Loading...</p>
	{:else if error}
		<div class="empty">
			<p class="err">{error}</p>
			<a href="/apps/create"><Button variant="primary" size="md">Deploy App</Button></a>
		</div>
	{:else if apps.length === 0}
		<div class="empty">
			<p class="text-muted">No applications yet. Deploy your first app!</p>
			<a href="/apps/create"><Button variant="primary" size="md">Deploy App</Button></a>
		</div>
	{:else}
		<DataTable
			{columns}
			rows={apps}
			emptyMessage="No applications found."
			onemptyclick={() => goto('/apps/create')}
			onselect={viewApp}
		/>
	{/if}
</Card>

<style>
	h1 { margin-bottom: var(--space-4); }
	.empty {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
		padding: var(--space-6) 0;
	}
	.err { color: var(--danger); font-size: var(--font-size-sm); }
</style>
