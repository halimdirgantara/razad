<script lang="ts">
	import '$lib/styles/theme.css';
	import '$lib/styles/base.css';
	import '$lib/styles/typography.css';
	import Sidebar from '$lib/components/Sidebar.svelte';

	let { children }: { children: import('svelte').Snippet } = $props();
</script>

<div class="app-shell">
	<Sidebar />
	<div class="main-area">
		<header class="topbar">
			<span class="topbar-title">Razad</span>
			<div class="topbar-right">
				<span class="topbar-status ok">● System Healthy</span>
			</div>
		</header>
		<main class="workspace">
			{@render children()}
		</main>
	</div>
</div>

<style>
	.app-shell {
		display: flex;
		height: 100vh;
		overflow: hidden;
	}
	.main-area {
		flex: 1;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		background: var(--bg);
	}
	.topbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		height: var(--topbar-height);
		padding: 0 var(--space-6);
		background: var(--bg-alt);
		border-bottom: 1px solid var(--border);
		flex-shrink: 0;
	}
	.topbar-title {
		font-weight: var(--font-weight-bold);
		font-size: var(--font-size-base);
		color: var(--text);
		display: none; /* shown on small screens instead of sidebar brand */
	}
	.topbar-right {
		display: flex;
		align-items: center;
		gap: var(--space-4);
	}
	.topbar-status {
		font-size: var(--font-size-xs);
		color: var(--text-muted);
	}
	.topbar-status.ok {
		color: var(--success);
	}
	.workspace {
		flex: 1;
		overflow-y: auto;
		padding: var(--space-6);
	}
</style>
