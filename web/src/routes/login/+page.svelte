<script lang="ts">
	import '$lib/styles/theme.css';
	import '$lib/styles/base.css';
	import '$lib/styles/typography.css';
	import Button from '$lib/components/Button.svelte';
	import { login } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	let email = $state('');
	let password = $state('');
	let showPassword = $state(false);
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
			<img class="login-logo" src="/razad-logo.png" alt="Razad" />
			<h1>Razad</h1>
			<p class="tagline">Your server, guided.</p>
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
				<div class="password-wrap">
					<input
						id="password"
						type={showPassword ? 'text' : 'password'}
						bind:value={password}
						placeholder="password"
						autocomplete="current-password"
						required
					/>
					<button
						type="button"
						class="toggle-pw"
						onclick={() => showPassword = !showPassword}
						aria-label={showPassword ? 'Hide password' : 'Show password'}
					>
						{#if showPassword}
							<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								<path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94" />
								<path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19" />
								<line x1="1" y1="1" x2="23" y2="23" />
								<path d="M14.12 14.12a3 3 0 1 1-4.24-4.24" />
							</svg>
						{:else}
							<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
								<circle cx="12" cy="12" r="3" />
							</svg>
						{/if}
					</button>
				</div>
			</div>

			{#if error}
				<p class="error-msg">{error}</p>
			{/if}

			<Button type="submit" variant="primary" size="lg" disabled={loading}>
				{loading ? 'Signing in...' : 'Sign In'}
			</Button>
		</form>
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
		max-width: 380px;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-8);
	}
	.login-header {
		text-align: center;
		margin-bottom: var(--space-6);
	}
	.login-logo {
		width: 96px;
		height: 96px;
		border-radius: var(--radius-md);
		display: block;
		margin: 0 auto var(--space-3);
	}
	.login-header h1 {
		font-size: var(--font-size-2xl);
		margin: 0 0 var(--space-1);
	}
	.tagline {
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

	/* ---- Password toggle ---- */
	.password-wrap {
		position: relative;
	}
	.password-wrap input {
		padding-right: 2.5rem;
	}
	.toggle-pw {
		position: absolute;
		right: 0.5rem;
		top: 50%;
		transform: translateY(-50%);
		background: none;
		border: none;
		color: var(--text-muted);
		cursor: pointer;
		padding: 0.25rem;
		display: flex;
		align-items: center;
		transition: color 0.15s;
	}
	.toggle-pw:hover {
		color: var(--text-secondary);
	}

	.error-msg {
		color: var(--danger);
		font-size: var(--font-size-sm);
		margin-bottom: var(--space-3);
	}
</style>
