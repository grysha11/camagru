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

    const logoutBtn = document.getElementById("logout-btn");
    if (logoutBtn) {
        logoutBtn.addEventListener("click", async () => {
            await api.logout();
            window.location.href = "/index.html";
        });
    }
}

document.addEventListener("DOMContentLoaded", guardAndInit);
