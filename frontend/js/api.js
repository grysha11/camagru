async function handleResponse(response) {
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
        const error = new Error(data.error || "An unexpected error ocurred");
        error.status = response.status;
        throw error;
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

    async login(username, password) {
        const response = await fetch("/api/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            credentials: "include",
            body: JSON.stringify({ username, password }),
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

    async updateProfile(body) {
        const response = await fetch("/api/me", {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            credentials: "include",
            body: JSON.stringify(body),
        });
        return handleResponse(response);
    },

    async confirmEmailChange(token) {
        const response = await fetch(`/api/confirm-email-change?token=${encodeURIComponent(token)}`);
        return handleResponse(response);
    },

    async uploadAvatar(formData) {
        const response = await fetch("/api/me/avatar", {
            method: "POST",
            credentials: "include",
            body: formData,
        });
        return handleResponse(response);
    },

    async deleteAccount() {
        const response = await fetch("/api/me", {
            method: "DELETE",
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
    },

    async listOverlays() {
        const response = await fetch("/api/overlays", { credentials: "include" });
        return handleResponse(response);
    },

    async myPosts() {
        const response = await fetch("/api/posts/mine", { credentials: "include" });
        return handleResponse(response);
    },

    async uploadPost(formData) {
        const response = await fetch("/api/posts", {
            method: "POST",
            credentials: "include",
            body: formData,
        });
        return handleResponse(response);
    },

    async listPosts(page = 1) {
        const response = await fetch(`/api/posts?page=${encodeURIComponent(page)}`, { credentials: "include" });
        return handleResponse(response);
    },

    async getPost(id) {
        const response = await fetch(`/api/posts/${encodeURIComponent(id)}`, { credentials: "include" });
        return handleResponse(response);
    },

    async deletePost(id) {
        const response = await fetch(`/api/posts/${encodeURIComponent(id)}`, {
            method: "DELETE",
            credentials: "include",
        });
        return handleResponse(response);
    },

    async likePost(id) {
        const response = await fetch(`/api/posts/${encodeURIComponent(id)}/like`, {
            method: "POST",
            credentials: "include",
        });
        return handleResponse(response);
    },

    async unlikePost(id) {
        const response = await fetch(`/api/posts/${encodeURIComponent(id)}/like`, {
            method: "DELETE",
            credentials: "include",
        });
        return handleResponse(response);
    },

    async listComments(id) {
        const response = await fetch(`/api/posts/${encodeURIComponent(id)}/comments`, { credentials: "include" });
        return handleResponse(response);
    },

    async createComment(id, content) {
        const response = await fetch(`/api/posts/${encodeURIComponent(id)}/comments`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            credentials: "include",
            body: JSON.stringify({ content }),
        });
        return handleResponse(response);
    },

    async deleteComment(postId, commentId) {
        const response = await fetch(`/api/posts/${encodeURIComponent(postId)}/comments/${encodeURIComponent(commentId)}`, {
            method: "DELETE",
            credentials: "include",
        });
        return handleResponse(response);
    }
};
