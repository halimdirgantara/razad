import { writable } from 'svelte/store';

export interface User {
	id: string;
	name: string;
	email: string;
}

export const currentUser = writable<User | null>(null);
export const isAuthenticated = writable<boolean>(false);

let storedToken: string | null = null;

export function getToken(): string | null {
	return storedToken;
}

export function login(user: User, token: string): void {
	storedToken = token;
	localStorage.setItem('razad_token', token);
	currentUser.set(user);
	isAuthenticated.set(true);
}

export function logout(): void {
	storedToken = null;
	localStorage.removeItem('razad_token');
	currentUser.set(null);
	isAuthenticated.set(false);
}

export async function checkSession(): Promise<void> {
	const token = localStorage.getItem('razad_token');
	if (!token) {
		logout();
		return;
	}

	storedToken = token;

	try {
		const res = await fetch('/api/v1/auth/me', {
			headers: { 'Authorization': `Bearer ${token}` }
		});

		if (!res.ok) {
			logout();
			return;
		}

		const user: User = await res.json();
		currentUser.set(user);
		isAuthenticated.set(true);
	} catch {
		logout();
	}
}
