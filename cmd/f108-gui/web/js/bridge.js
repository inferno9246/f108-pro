/**
 * bridge.js - HTTP bridge to the local Go server.
 *
 * Derived from f108-pro-gui (MIT, see THIRD_PARTY.md). The original talked to a
 * Python/WebKit host; this version sends each action to /api/invoke.
 */

class HostBridge {
    constructor() {
        this.token = new URLSearchParams(location.search).get('t');
        this.eventListeners = new Map();
        if (!this.token) {
            console.error('Missing token: open the full URL printed by f108-gui (with ?t=...).');
        }
    }

    async invoke(action, payload = {}) {
        if (!this.token) {
            throw new Error('Falta el token: abre la URL completa que imprime el programa (con ?t=...).');
        }
        const res = await fetch('/api/invoke', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', 'X-Token': this.token },
            body: JSON.stringify({ action, payload }),
        });
        if (!res.ok) {
            throw new Error(`${res.status} ${res.statusText}`);
        }
        const body = await res.json();
        if (!body.success) {
            throw new Error(body.error || 'Unknown error');
        }
        return body.data;
    }

    on(eventName, callback) {
        if (!this.eventListeners.has(eventName)) {
            this.eventListeners.set(eventName, []);
        }
        this.eventListeners.get(eventName).push(callback);
    }

    off(eventName, callback) {
        if (this.eventListeners.has(eventName)) {
            this.eventListeners.set(eventName, this.eventListeners.get(eventName).filter(cb => cb !== callback));
        }
    }
}

window.hostBridge = new HostBridge();
