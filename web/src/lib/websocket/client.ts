type MessageHandler = (data: unknown) => void;

class WebSocketClient {
	private ws: WebSocket | null = null;
	private handlers = new Map<string, Set<MessageHandler>>();
	private token: string | null = null;

	connect(token: string): void {
		this.token = token;
		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const url = `${protocol}//${window.location.host}/ws`;

		this.ws = new WebSocket(url);
		this.ws.onmessage = (event) => {
			try {
				const msg = JSON.parse(event.data);
				const handlers = this.handlers.get(msg.type);
				if (handlers) {
					handlers.forEach((h) => h(msg.payload));
				}
			} catch {
				// ignore malformed messages
			}
		};
		this.ws.onclose = () => {
			// auto-reconnect after 3 seconds
			setTimeout(() => {
				if (this.token) this.connect(this.token);
			}, 3000);
		};
	}

	disconnect(): void {
		this.token = null;
		if (this.ws) {
			this.ws.close();
			this.ws = null;
		}
	}

	on(type: string, handler: MessageHandler): () => void {
		if (!this.handlers.has(type)) {
			this.handlers.set(type, new Set());
		}
		this.handlers.get(type)!.add(handler);
		return () => this.handlers.get(type)?.delete(handler);
	}

	send(type: string, payload: unknown): void {
		if (this.ws?.readyState === WebSocket.OPEN) {
			this.ws.send(JSON.stringify({ type, payload }));
		}
	}
}

export const wsClient = new WebSocketClient();
