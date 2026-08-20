import { api } from './api.js';

const PASSWORD_REGEX = /^(?=.*[A-Z])(?=.*[a-z])(?=.*\d)(?=.*[^A-Za-z0-9]).{8,}$/;
const isValidPassword = (password) => PASSWORD_REGEX.test(password);

document.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("reset-password-form");
    const statusMessage = document.getElementById("status-message");
    const submitBtn = document.getElementById("reset-password-submit");

    const showMessage = (msg, isError = false) => {
        statusMessage.textContent = msg;
        statusMessage.classList.toggle("error", isError);
        statusMessage.classList.toggle("success", !isError && !!msg);
    };

    const token = new URLSearchParams(window.location.search).get("token");
    if (!token) {
        showMessage("Missing or invalid reset link.", true);
        submitBtn.disabled = true;
        return;
    }

    form.addEventListener("submit", async (e) => {
        e.preventDefault();
        const password = document.getElementById("reset-password").value;

        if (!isValidPassword(password)) {
            return showMessage("Password must be at least 8 characters and include uppercase, lowercase, number, and special character.", true);
        }

        try {
            const res = await api.resetPassword(token, password);
            showMessage(res.message + " Redirecting to login...");
            setTimeout(() => { window.location.href = "/index.html"; }, 2000);
        } catch (error) {
            showMessage(error.message, true);
        }
    });
});
