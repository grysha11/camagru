async function handleResponse(response) {
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
        throw new Error(data.error || "An unexpected error ocurred");
    }
    return data;
}

export const api = {
    async register(username, email, password) {
        const response = await fetch("/register", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username, email, password }),
        });
        return handleResponse(response);
    },

    async login(email, password) {
        const response = await fetch("/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            credentials: "include",
            body: JSON.stringify({ email, password }),
        });
        return handleResponse(response);
    },

    async me() {
        const response = await fetch("/me", { credentials: "include" });
        return handleResponse(response);
    },

    async refresh() {
        const response = await fetch("/refresh", {
            method: "POST",
            credentials: "include",
        });
        return handleResponse(response);
    },

    async logout() {
        const response = await fetch("/logout", {
            method: "POST",
            credentials: "include",
        });
        return handleResponse(response);
    }
};
