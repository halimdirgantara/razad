<script lang="ts">
	import '$lib/styles/theme.css';
	import '$lib/styles/base.css';
	import '$lib/styles/typography.css';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import { page } from '$app/stores';
	import { currentUser, isAuthenticated, checkSession } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let { children }: { children: import('svelte').Snippet } = $props();

	let authChecked = $state(false);
	let userMenuOpen = $state(false);
	let mobileNavOpen = $state(false);
	let menuRef = $state<HTMLDivElement>();

	onMount(() => {
		window.addEventListener('resize', handleViewportChange);

		const initializeAuth = async () => {
			if ($page.url.pathname === '/login') {
				authChecked = true;
				return;
			}

			const token = localStorage.getItem('razad_token');
			if (!token) {
				goto('/login');
				return;
			}

			try {
				const res = await fetch('/api/v1/auth/me', {
					headers: { 'Authorization': `Bearer ${token}` }
				});
				if (!res.ok) {
					localStorage.removeItem('razad_token');
					goto('/login');
					return;
				}
				const user = await res.json();
				currentUser.set(user);
				isAuthenticated.set(true);
			} catch {
				// daemon may be down — keep going
			}
			authChecked = true;
		};

		initializeAuth();
		return () => window.removeEventListener('resize', handleViewportChange);
	});

	function toggleMenu(e: Event) {
		e.stopPropagation();
		userMenuOpen = !userMenuOpen;
	}

	function closeMenu(e: Event) {
		if (menuRef && !menuRef.contains(e.target as Node)) {
			userMenuOpen = false;
		}
	}

	function toggleMobileNav() {
		mobileNavOpen = !mobileNavOpen;
	}

	function closeMobileNav() {
		mobileNavOpen = false;
	}

	function handleViewportChange() {
		if (window.innerWidth > 900) {
			mobileNavOpen = false;
		}
	}

	function handleLogout() {
		userMenuOpen = false;
		isAuthenticated.set(false);
		currentUser.set(null);
		localStorage.removeItem('razad_token');
		fetch('/api/v1/auth/logout', { method: 'POST' }).catch(() => {});
		goto('/login');
	}
</script>

<svelte:window onclick={closeMenu} />

{#if $page.url.pathname === '/login'}
	{@render children()}
{:else if !authChecked}
	<div class="loading-screen">
		<p class="text-muted">Loading...</p>
	</div>
{:else}
	<div class="app-shell">
		<button class:mobile-backdrop={mobileNavOpen} class="backdrop" type="button" aria-label="Close navigation" onclick={closeMobileNav}></button>
		<Sidebar mobileOpen={mobileNavOpen} onNavigate={closeMobileNav} />
		<div class="main-area">
			<header class="topbar">
				<div class="topbar-left">
					<button class="nav-toggle" type="button" aria-label="Open navigation" aria-expanded={mobileNavOpen} onclick={toggleMobileNav}>☰</button>
					<span class="topbar-title">Razad</span>
				</div>
				<div class="topbar-right" bind:this={menuRef}>
					<div class="user-menu" role="button" tabindex="0"
						onclick={toggleMenu}
						onkeydown={(e) => { if (e.key === 'Enter') toggleMenu(e); }}
					>
						<div class="user-avatar-sm"></div>
						<span class="user-name-sm">{$currentUser?.name ?? 'admin'}</span>
					</div>
					{#if userMenuOpen}
						<div class="user-dropdown">
							<div class="dropdown-header">
								<span class="dropdown-name">{$currentUser?.name ?? 'admin'}</span>
								<span class="dropdown-email">{$currentUser?.email ?? 'admin@razad.local'}</span>
							</div>
							<hr class="dropdown-divider" />
							<button class="dropdown-item danger" onclick={handleLogout}>
								Sign Out
							</button>
						</div>
					{/if}
				</div>
			</header>
			<main class="workspace">
				{@render children()}
			</main>
		</div>
	</div>
{/if}

<style>
	.loading-screen {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 100vh;
		background: var(--bg);
		color: var(--text-muted);
	}
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
		min-width: 0;
	}
	.topbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		height: var(--topbar-height);
		padding: 0 var(--space-6);
		background: var(--bg-alt);
		border-bottom: 1px solid var(--border);
		flex-shrink: 0;
		position: relative;
		z-index: 100;
	}
	.topbar-left {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		min-width: 0;
	}
	.topbar-title {
		font-weight: var(--font-weight-bold);
		font-size: var(--font-size-base);
		color: var(--text);
	}
	.topbar-right {
		display: flex;
		align-items: center;
		gap: var(--space-4);
		position: relative;
		min-width: 0;
	}
	.nav-toggle {
		display: none;
		align-items: center;
		justify-content: center;
		width: 2.25rem;
		height: 2.25rem;
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		background: transparent;
		color: var(--text);
		cursor: pointer;
		flex-shrink: 0;
	}
	.user-menu {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		cursor: pointer;
		padding: var(--space-1) var(--space-2);
		border-radius: var(--radius-sm);
		transition: background 0.12s;
	}
	.user-menu:hover { background: var(--surface); }
	.user-avatar-sm {
		width: 26px; height: 26px;
		border-radius: 50%;
		background: var(--surface-3);
		border: 1px solid var(--border);
		flex-shrink: 0;
	}
	.user-name-sm { font-size: var(--font-size-sm); color: var(--text); }
	.backdrop {
		display: none;
		border: 0;
		padding: 0;
	}
	.user-dropdown {
		position: absolute;
		top: calc(100% + 4px);
		right: 0;
		width: 220px;
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
		box-shadow: var(--shadow-lg);
		z-index: 200;
		overflow: hidden;
	}
	.dropdown-header {
		padding: var(--space-3);
		display: flex;
		flex-direction: column;
		gap: 0.125rem;
	}
	.dropdown-name { font-size: var(--font-size-sm); font-weight: var(--font-weight-medium); color: var(--text); }
	.dropdown-email { font-size: var(--font-size-xs); color: var(--text-muted); }
	.dropdown-divider { border: none; border-top: 1px solid var(--border); margin: 0; }
	.dropdown-item {
		display: block;
		width: 100%;
		padding: var(--space-2) var(--space-3);
		font-size: var(--font-size-sm);
		text-align: left;
		background: none;
		border: none;
		cursor: pointer;
		color: var(--text);
		transition: background 0.12s;
	}
	.dropdown-item:hover { background: var(--surface-3); }
	.dropdown-item.danger { color: var(--danger); }
	.workspace {
		flex: 1;
		overflow-y: auto;
		padding: var(--space-6);
	}

	@media (max-width: 900px) {
		.topbar {
			padding: 0 var(--space-4);
		}

		.nav-toggle {
			display: inline-flex;
		}

		.user-name-sm {
			display: none;
		}

		.backdrop.mobile-backdrop {
			display: block;
			position: fixed;
			inset: 0;
			background: rgba(4, 22, 29, 0.72);
			z-index: 210;
		}

		.workspace {
			padding: var(--space-4);
		}
	}

	@media (max-width: 640px) {
		.topbar {
			padding: 0 var(--space-3);
		}

		.topbar-title {
			font-size: var(--font-size-sm);
		}

		.workspace {
			padding: var(--space-3);
		}
	}
</style>
