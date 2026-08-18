import { api } from "./api.js";

export function initNav(activePage, user) {
    const navUser = document.getElementById("nav-user");
    const navLogout = document.getElementById("nav-logout");
    const navLoginLink = document.getElementById("nav-login-link");
    const navRegisterLink = document.getElementById("nav-register-link");
    const siteNav = document.getElementById("site-nav");

    if (user) {
        navUser.textContent = user.username;
        navUser.classList.remove("hidden");
        navLogout.classList.remove("hidden");
        navLoginLink.classList.add("hidden");
        navRegisterLink.classList.add("hidden");
        siteNav.classList.remove("hidden");
    } else {
        navUser.classList.add("hidden");
        navLogout.classList.add("hidden");
        navLoginLink.classList.remove("hidden");
        navRegisterLink.classList.remove("hidden");
        siteNav.classList.add("hidden");
    }

    for (const link of siteNav.querySelectorAll(".nav-link")) {
        link.classList.toggle("active", link.dataset.nav === activePage);
    }

    navLogout.addEventListener("click", async (e) => {
        e.preventDefault();
        try {
            await api.logout();
        } finally {
            window.location.href = "/index.html";
        }
    });
}
