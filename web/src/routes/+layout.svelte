<nav>
	<div class="brand">Razad</div>
	<div class="nav-links">
		<a href="/">Dashboard</a>
		<a href="/apps">Apps</a>
		<a href="/servers">Servers</a>
		<a href="/settings">Settings</a>
	</div>
</nav>

<main>
	{@render children()}
</main>

<style>
	nav {
		display: flex;
		align-items: center;
		padding: 0 1.5rem;
		height: 3rem;
		background: #1a1a2e;
		color: #e0e0e0;
		gap: 2rem;
	}
	.brand {
		font-weight: 700;
		font-size: 1.1rem;
	}
	.nav-links {
		display: flex;
		gap: 1rem;
	}
	.nav-links a {
		color: #a0a0b0;
		text-decoration: none;
	}
	.nav-links a:hover {
		color: #ffffff;
	}
	main {
		padding: 1.5rem;
	}
</style>
