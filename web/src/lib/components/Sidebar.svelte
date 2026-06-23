<script lang="ts">
	import { page } from '$app/stores';

	interface Props {
		mobileOpen?: boolean;
		onNavigate?: () => void;
	}

	let { mobileOpen = false, onNavigate }: Props = $props();

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
		{ name: 'Overview', items: [navItems[0]] },
		{ name: 'Infrastructure', items: navItems.slice(1, 6) },
		{ name: 'Observability', items: navItems.slice(6, 9) },
		{ name: 'Automation', items: [navItems[9], navItems[10]] },
		{ name: 'System', items: [navItems[11]] },
	];
</script>

<aside class="sidebar" class:mobile-open={mobileOpen}>
	<div class="sidebar-brand">
		<img class="brand-logo" src="/razad-logo.png" alt="Razad" />
		<span class="brand-name">Razad</span>
		<span class="brand-version">v0.1</span>
		<button class="sidebar-close" type="button" aria-label="Close navigation" onclick={onNavigate}>✕</button>
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
								onclick={onNavigate}
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
		transition: transform 0.18s ease, box-shadow 0.18s ease;
	}
	.sidebar-brand {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		padding: var(--space-3) var(--space-4);
		border-bottom: 1px solid var(--border);
	}
	.brand-logo {
		width: 28px;
		height: 28px;
		border-radius: var(--radius-sm);
		flex-shrink: 0;
	}
	.brand-name {
		font-weight: var(--font-weight-bold);
		font-size: var(--font-size-base);
		color: var(--text);
	}
	.brand-version {
		font-size: var(--font-size-xs);
		color: var(--text-muted);
		margin-left: auto;
	}
	.sidebar-close {
		display: none;
		align-items: center;
		justify-content: center;
		width: 2rem;
		height: 2rem;
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		background: transparent;
		color: var(--text);
		cursor: pointer;
		flex-shrink: 0;
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

	@media (max-width: 900px) {
		.sidebar {
			position: fixed;
			top: 0;
			left: 0;
			height: 100dvh;
			z-index: 220;
			box-shadow: var(--shadow-lg);
			transform: translateX(-100%);
		}

		.sidebar.mobile-open {
			transform: translateX(0);
		}

		.sidebar-close {
			display: inline-flex;
		}
	}
</style>
