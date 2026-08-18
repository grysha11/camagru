import { api } from './api.js';
import { initNav } from './nav.js';

async function guardAndInit() {
    let user;
    try {
        user = await api.me();
    } catch {
        try {
            await api.refresh();
            user = await api.me();
        } catch {
            window.location.href = "/index.html";
            return;
        }
    }

    init(user);
}

function init(user) {
    initNav("settings", user);

    document.getElementById("account-username").value = user.username;
    document.getElementById("account-email").value = user.email;

    const statusMessage = document.getElementById("status-message");
    const showMessage = (msg, isError = false) => {
        statusMessage.textContent = msg;
        statusMessage.classList.toggle("error", isError);
        statusMessage.classList.toggle("success", !isError && !!msg);
    };

    const changePasswordBtn = document.getElementById("change-password-btn");
    changePasswordBtn.addEventListener("click", async () => {
        changePasswordBtn.disabled = true;
        try {
            const { token } = await api.requestPasswordChange();
            sessionStorage.setItem("resetToken", token);
            window.location.href = "/reset-password.html";
        } catch (error) {
            showMessage(error.message, true);
            changePasswordBtn.disabled = false;
        }
    });
}

document.addEventListener("DOMContentLoaded", guardAndInit);
