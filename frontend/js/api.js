async function handleResponse(response) {
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
        throw new Error(data.error || "An unexpected error ocurred");
    }
    return data;
}

export const api = {
    async register(username, email, password) {
        const response = await fetch("/api/register", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username, email, password }),
        });
        return handleResponse(response);
    },

    async login(email, password) {
        const response = await fetch("/api/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            credentials: "include",
            body: JSON.stringify({ email, password }),
        });
        return handleResponse(response);
    },

    async me() {
        const response = await fetch("/api/me", { credentials: "include" });
        return handleResponse(response);
    },

    async refresh() {
        const response = await fetch("/api/refresh", {
            method: "POST",
            credentials: "include",
        });
        return handleResponse(response);
    },

    async logout() {
        const response = await fetch("/api/logout", {
            method: "POST",
            credentials: "include",
        });
        return handleResponse(response);
    },

    async confirmEmail(token) {
        const response = await fetch(`/api/confirm-email?token=${encodeURIComponent(token)}`);
        return handleResponse(response);
    },

    async forgotPassword(email) {
        const response = await fetch("/api/forgot-password", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ email }),
        });
        return handleResponse(response);
    },

    async requestPasswordChange() {
        const response = await fetch("/api/request-password-change", {
            method: "POST",
            credentials: "include",
        });
        return handleResponse(response);
    },

    async resetPassword(token, password) {
        const response = await fetch("/api/reset-password", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ token, password }),
        });
        return handleResponse(response);
    }
};
