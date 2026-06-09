<script lang="ts">
	let health: { status: string } | null = null;
	let error: string | null = null;

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
</script>

<h1>Razad</h1>
<p class="tagline">Your server, guided.</p>

<div class="card">
	<h2>Server Health</h2>
	<button onclick={checkHealth}>Check</button>
	{#if health}
		<p class="ok">Status: {health.status}</p>
	{/if}
	{#if error}
		<p class="err">{error}</p>
	{/if}
</div>

<style>
	.tagline {
		color: #666;
		margin-bottom: 2rem;
	}
	.card {
		background: #f5f5f5;
		border-radius: 8px;
		padding: 1.5rem;
		max-width: 400px;
	}
	.ok { color: #2e7d32; }
	.err { color: #c62828; }
	button {
		margin: 0.5rem 0;
		padding: 0.5rem 1rem;
	}
</style>
