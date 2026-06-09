const BASE_URL = '/api/v1';

interface ApiError {
	code: string;
	message: string;
}

interface ApiResponse<T> {
	data?: T;
	error?: ApiError;
}

async function request<T>(
	method: string,
	path: string,
	body?: Record<string, unknown>
): Promise<ApiResponse<T>> {
	const headers: Record<string, string> = {
		'Content-Type': 'application/json'
	};

	// Attach auth token if available
	const token = localStorage.getItem('razad_token');
	if (token) {
		headers['Authorization'] = `Bearer ${token}`;
	}

	try {
		const res = await fetch(`${BASE_URL}${path}`, {
			method,
			headers,
			body: body ? JSON.stringify(body) : undefined
		});

		const json = await res.json();

		if (!res.ok) {
			return { error: json.error ?? { code: 'unknown', message: res.statusText } };
		}

		return { data: json };
	} catch (err) {
		return {
			error: {
				code: 'network_error',
				message: err instanceof Error ? err.message : 'Network error'
			}
		};
	}
}

export const api = {
	get: <T>(path: string) => request<T>('GET', path),
	post: <T>(path: string, body?: Record<string, unknown>) => request<T>('POST', path, body),
	put: <T>(path: string, body?: Record<string, unknown>) => request<T>('PUT', path, body),
	delete: <T>(path: string) => request<T>('DELETE', path)
};
