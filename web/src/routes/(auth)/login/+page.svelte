<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import { login } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleLogin() {
		error = '';
		if (!email || !password) {
			error = 'Email and password are required.';
			return;
		}
		loading = true;

		try {
			const res = await fetch('/api/v1/auth/login', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email, password })
			});

			const data = await res.json();

			if (!res.ok) {
				error = data.error?.message ?? 'Login failed.';
				return;
			}

			login(data.user, data.token);
			goto('/');
		} catch (e) {
			error = 'Failed to connect to daemon.';
		} finally {
			loading = false;
		}
	}
</script>

<div class="login-page">
	<div class="login-card">
		<div class="login-header">
			<span class="login-icon">◈</span>
			<h1>Razad</h1>
			<p class="login-tagline">Your server, guided.</p>
		</div>

		<form onsubmit={(e) => { e.preventDefault(); handleLogin(); }}>
			<div class="field">
				<label for="email">Email</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					placeholder="admin@razad.local"
					autocomplete="email"
					required
				/>
			</div>
			<div class="field">
				<label for="password">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					placeholder="password"
					autocomplete="current-password"
					required
				/>
			</div>

			{#if error}
				<p class="error-msg">{error}</p>
			{/if}

			<Button type="submit" variant="primary" size="lg" disabled={loading}>
				{loading ? 'Signing in...' : 'Sign In'}
			</Button>
		</form>

		<div class="login-footer">
			<span class="meta text-muted">Default: admin@razad.local / razadadmin</span>
		</div>
	</div>
</div>

<style>
	.login-page {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 100vh;
		background: var(--bg);
		padding: var(--space-4);
	}
	.login-card {
		width: 100%;
		max-width: 360px;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-8);
	}
	.login-header {
		text-align: center;
		margin-bottom: var(--space-6);
	}
	.login-icon {
		font-size: 2rem;
		color: var(--primary);
	}
	.login-header h1 {
		font-size: var(--font-size-2xl);
		margin: var(--space-2) 0 var(--space-1);
	}
	.login-tagline {
		color: var(--text-muted);
		font-size: var(--font-size-sm);
		margin: 0;
	}

	/* ---- Form ---- */
	.field {
		margin-bottom: var(--space-4);
	}
	label {
		display: block;
		font-size: var(--font-size-sm);
		font-weight: var(--font-weight-medium);
		color: var(--text-secondary);
		margin-bottom: var(--space-1);
	}
	input {
		width: 100%;
		padding: var(--space-2) var(--space-3);
		background: var(--bg-alt);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		color: var(--text);
		font-size: var(--font-size-base);
		outline: none;
		transition: border-color 0.15s;
	}
	input:focus {
		border-color: var(--primary);
	}
	.error-msg {
		color: var(--danger);
		font-size: var(--font-size-sm);
		margin-bottom: var(--space-3);
	}
	.login-footer {
		margin-top: var(--space-6);
		text-align: center;
	}
</style>
