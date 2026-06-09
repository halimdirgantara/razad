<script lang="ts">
	import { page } from '$app/stores';

	const navItems = [
		{ href: '/', label: 'Dashboard', icon: '◉' },
		{ href: '/apps', label: 'Applications', icon: '⊞' },
		{ href: '/deployments', label: 'Deployments', icon: '↻' },
		{ href: '/domains', label: 'Domains', icon: '◈' },
		{ href: '/databases', label: 'Databases', icon: '⛁' },
		{ href: '/services', label: 'Services', icon: '⚙' },
		{ href: '/logs', label: 'Logs', icon: '☰' },
		{ href: '/events', label: 'Events', icon: '⚡' },
		{ href: '/audit', label: 'Audit', icon: '◷' },
		{ href: '/ai', label: 'Razad AI', icon: '◆' },
		{ href: '/policies', label: 'Policies', icon: '⊡' },
		{ href: '/settings', label: 'Settings', icon: '⚙' },
	];

	const sections = [
		{ name: 'Overview', items: [navItems[0], navItems[10]] },
		{ name: 'Infrastructure', items: navItems.slice(1, 6) },
		{ name: 'Observability', items: navItems.slice(6, 9) },
		{ name: 'Automation', items: navItems.slice(9, 11) },
		{ name: 'System', items: navItems.slice(11) },
	];
</script>

<aside class="sidebar">
	<div class="sidebar-brand">
		<img class="brand-logo" src="/razad-logo.png" alt="Razad" />
		<span class="brand-name">Razad</span>
		<span class="brand-version">v0.1</span>
	</div>
	<nav class="sidebar-nav">
		{#each sections as section}
			<div class="nav-section">
				<span class="nav-section-title">{section.name}</span>
				<ul>
					{#each section.items as item}
						<li>
							<a
								href={item.href}
								class="nav-link"
								class:active={$page.url.pathname === item.href}
							>
								<span class="nav-icon">{item.icon}</span>
								<span class="nav-label">{item.label}</span>
							</a>
						</li>
					{/each}
				</ul>
			</div>
		{/each}
	</nav>
</aside>

<style>
	.sidebar {
		width: var(--sidebar-width);
		background: var(--surface);
		border-right: 1px solid var(--border);
		display: flex;
		flex-direction: column;
		flex-shrink: 0;
		overflow-y: auto;
	}
	.sidebar-brand {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: var(--space-5) var(--space-4) var(--space-3);
		border-bottom: 1px solid var(--border);
		gap: var(--space-2);
	}
	.brand-logo {
		width: 40px;
		height: 40px;
		border-radius: var(--radius-md);
	}
	.brand-name {
		font-weight: var(--font-weight-bold);
		font-size: var(--font-size-base);
		color: var(--text);
	}
	.brand-version {
		font-size: var(--font-size-xs);
		color: var(--text-muted);
	}
	.sidebar-nav {
		flex: 1;
		padding: var(--space-3) var(--space-2);
		overflow-y: auto;
	}
	.nav-section {
		margin-bottom: var(--space-4);
	}
	.nav-section-title {
		display: block;
		font-size: var(--font-size-xs);
		font-weight: var(--font-weight-semibold);
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding: 0 var(--space-2);
		margin-bottom: var(--space-1);
	}
	.nav-link {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		padding: var(--space-1) var(--space-2);
		border-radius: var(--radius-sm);
		color: var(--text-secondary);
		text-decoration: none;
		font-size: var(--font-size-sm);
		transition: background 0.12s, color 0.12s;
	}
	.nav-link:hover {
		background: var(--surface-2);
		color: var(--text);
		text-decoration: none;
	}
	.nav-link.active {
		background: var(--primary-bg);
		color: var(--primary);
	}
	.nav-icon {
		font-size: 0.875rem;
		width: 1.25rem;
		text-align: center;
		flex-shrink: 0;
		opacity: 0.7;
	}
	.nav-label {
		white-space: nowrap;
	}
</style>
