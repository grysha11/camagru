import { api } from './api.js';

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
    renderUser(user);
}

function renderUser(user) {
    const info = document.getElementById("user-info");
    if (info) {
        info.textContent = `Logged in as ${user.username}`;
    }

    const statusMessage = document.getElementById("status-message");
    const showMessage = (msg, isError = false) => {
        statusMessage.textContent = msg;
        statusMessage.style.color = isError ? "red" : "green";
    };

    const changePasswordBtn = document.getElementById("change-password-btn");
    if (changePasswordBtn) {
        changePasswordBtn.addEventListener("click", async () => {
            changePasswordBtn.disabled = true;
            try {
                const { token } = await api.requestPasswordChange();
                window.location.href = `/reset-password.html?token=${encodeURIComponent(token)}`;
            } catch (error) {
                showMessage(error.message, true);
                changePasswordBtn.disabled = false;
            }
        });
    }

    const logoutBtn = document.getElementById("logout-btn");
    if (logoutBtn) {
        logoutBtn.addEventListener("click", async () => {
            await api.logout();
            window.location.href = "/index.html";
        });
    }
}

document.addEventListener("DOMContentLoaded", guardAndInit);
