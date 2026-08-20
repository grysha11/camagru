import { api } from './api.js';

document.addEventListener("DOMContentLoaded", async () => {
    const statusMessage = document.getElementById("status-message");

    const showMessage = (msg, isError = false) => {
        statusMessage.textContent = msg;
        statusMessage.classList.toggle("error", isError);
        statusMessage.classList.toggle("success", !isError && !!msg);
    };

    const params = new URLSearchParams(window.location.search);
    const token = params.get("token");
    const type = params.get("type");
    if (!token) {
        return showMessage("Missing or invalid confirmation link.", true);
    }

    try {
        const res = type === "change_email" ? await api.confirmEmailChange(token) : await api.confirmEmail(token);
        showMessage(res.message);
    } catch (error) {
        showMessage(error.message, true);
    }
});
