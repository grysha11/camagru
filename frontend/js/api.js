import { ENV } from './env.js';

const API_BASE = ENV.API_URL;

async function handleResponse(response) {
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
        throw new Error(data.error || "An unexpected error ocurred");
    }
    return data;
}

export const api = {
    async register(username, email, password) {
        const response = await fetch(`{API_BASE}/register`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username, email, password }),
        });
        return handleResponse(response);
    },

    async login(email, password) {
        const response = await fetch(`{API_BASE}/login`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            credentials: "include",
            body: JSON.stringify({ email, password }),
        });
        return handleResponse(response);
    }
};
