import { api } from './api.js';

document.addEventListener("DOMContentLoaded", () => {
    const loginSection = document.getElementById("login-section");
    const registerSection = document.getElementById("register-section");
    const loginForm = document.getElementById("login-form");
    const registerForm = document.getElementById("register-form");
    const statusMessage = document.getElementById("status-message");

    const showMessage = (msg, isError = false) => {
        statusMessage.textContent = msg;
        statusMessage.style.color = isError ? "red" : "green";
    };

    const isValidEmail = (email) => {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    };

    const isValidPassword = (password) => {
        const passwordRegex = /^(?=.*[A-Z])(?=.*\d).{8,}$/;
        return passwordRegex.test(password);
    };

    document.getElementById("show-register").addEventListener("click", () => {
        loginSection.style.display = "none";
        registerSection.style.display = "block";
        statusMessage.textContent = "";
    });

    document.getElementById("show-login").addEventListener("click", () => {
        registerSection.style.display = "none";
        loginSection.style.display = "block";
        statusMessage.textContent = "";
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
            return showMessage("Password must be at least 8 characters, contain 1 uppercase letter and 1 number.", true);
        }

        try {
            const res = await api.register(username, email, password);
            showMessage(res.message + " You can now log in.");
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