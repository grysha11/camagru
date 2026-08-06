import { api } from './api.js';

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const isValidEmail = (email) => EMAIL_REGEX.test(email);

document.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("forgot-password-form");
    const statusMessage = document.getElementById("status-message");

    const showMessage = (msg, isError = false) => {
        statusMessage.textContent = msg;
        statusMessage.style.color = isError ? "red" : "green";
    };

    form.addEventListener("submit", async (e) => {
        e.preventDefault();
        const email = document.getElementById("forgot-email").value;

        if (!isValidEmail(email)) {
            return showMessage("Please enter a valid email address.", true);
        }

        try {
            const res = await api.forgotPassword(email);
            showMessage(res.message);
            form.reset();
        } catch (error) {
            showMessage(error.message, true);
        }
    });
});
