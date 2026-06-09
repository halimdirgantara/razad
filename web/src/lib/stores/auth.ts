import { writable } from 'svelte/store';

export interface User {
	id: string;
	name: string;
	email: string;
}

export const currentUser = writable<User | null>(null);
export const isAuthenticated = writable<boolean>(false);

export function login(user: User, token: string): void {
	localStorage.setItem('razad_token', token);
	currentUser.set(user);
	isAuthenticated.set(true);
}

export function logout(): void {
	localStorage.removeItem('razad_token');
	currentUser.set(null);
	isAuthenticated.set(false);
}
