import { api } from './api.js';

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const PASSWORD_REGEX = /^(?=.*[A-Z])(?=.*[a-z])(?=.*\d)(?=.*[^A-Za-z0-9]).{8,}$/;

const isValidEmail = (email) => EMAIL_REGEX.test(email);
const isValidPassword = (password) => PASSWORD_REGEX.test(password);

const OAUTH_ERROR_MESSAGES = {
    access_denied: "GitHub sign-in was cancelled.",
    state_mismatch: "GitHub sign-in failed, please try again.",
    missing_code: "GitHub sign-in failed, please try again.",
    exchange_failed: "GitHub sign-in failed, please try again.",
    no_email: "Your GitHub account has no verified email address.",
    email_conflict: "An account with that email already exists but isn't confirmed yet. Please confirm it before using GitHub sign-in.",
    server_error: "Something went wrong signing in with GitHub. Please try again.",
};

document.addEventListener("DOMContentLoaded", () => {
    const loginSection = document.getElementById("login-section");
    const registerSection = document.getElementById("register-section");
    const loginForm = document.getElementById("login-form");
    const registerForm = document.getElementById("register-form");
    const statusMessage = document.getElementById("status-message");

    const showMessage = (msg, isError = false) => {
        statusMessage.textContent = msg;
        statusMessage.classList.toggle("error", isError);
        statusMessage.classList.toggle("success", !isError && !!msg);
    };

    const clearMessage = () => {
        statusMessage.textContent = "";
        statusMessage.classList.remove("error", "success");
    };

    const oauthError = new URLSearchParams(window.location.search).get("oauth_error");
    if (oauthError) {
        showMessage(OAUTH_ERROR_MESSAGES[oauthError] || "GitHub sign-in failed, please try again.", true);
        window.history.replaceState({}, "", window.location.pathname);
    }

    document.getElementById("show-register").addEventListener("click", () => {
        loginSection.classList.add("hidden");
        registerSection.classList.remove("hidden");
        clearMessage();
    });

    document.getElementById("show-login").addEventListener("click", () => {
        registerSection.classList.add("hidden");
        loginSection.classList.remove("hidden");
        clearMessage();
    });

    registerForm.addEventListener("submit", async (e) => {
        e.preventDefault();
        const username = document.getElementById("register-username").value;
        const email = document.getElementById("register-email").value;
        const password = document.getElementById("register-password").value;

        if (!isValidEmail(email)) {
            return showMessage("Please enter valid email address.", true);
        }
        if (!isValidPassword(password)) {
            return showMessage("Password must be at least 8 characters and include uppercase, lowercase, number, and special character.", true);
        }

        try {
            const res = await api.register(username, email, password);
            showMessage(res.message);
            registerForm.reset();
        } catch (error) {
            showMessage(error.message, true);
        }
    });

    loginForm.addEventListener("submit", async (e) => {
        e.preventDefault();
        const email = document.getElementById("login-email").value;
        const password = document.getElementById("login-password").value;

        try {
            const res = await api.login(email, password);
            showMessage(res.message);
            window.location.href = "/gallery.html";
        } catch (error) {
            showMessage(error.message, true);
        }
    });
})
