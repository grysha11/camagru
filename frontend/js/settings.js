import { api } from './api.js';
import { initNav } from './nav.js';
import { confirmDialog } from './modal.js';

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

function setAvatarPreview(avatarPath) {
    const preview = document.getElementById("avatar-preview");
    preview.textContent = "";
    if (avatarPath) {
        const img = document.createElement("img");
        img.src = avatarPath;
        img.alt = "Your avatar";
        preview.appendChild(img);
    } else {
        preview.textContent = "[ avatar ]";
    }
}

function init(user) {
    initNav("settings", user);

    document.getElementById("account-username").value = user.username;
    document.getElementById("account-email").value = user.email;
    document.getElementById("notify-comment-checkbox").checked = user.notify_on_comment !== false;
    setAvatarPreview(user.avatar_path);

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
            const res = await api.requestPasswordChange();
            showMessage(res.message);
        } catch (error) {
            showMessage(error.message, true);
        } finally {
            changePasswordBtn.disabled = false;
        }
    });

    const accountForm = document.getElementById("account-form");
    accountForm.addEventListener("submit", async (e) => {
        e.preventDefault();
        const submitBtn = accountForm.querySelector("button[type=submit]");
        submitBtn.disabled = true;
        try {
            const res = await api.updateProfile({
                username: document.getElementById("account-username").value,
                email: document.getElementById("account-email").value,
            });
            document.getElementById("account-username").value = res.user.username;
            document.getElementById("account-email").value = res.user.email;
            showMessage(res.message || "Profile updated.");
        } catch (error) {
            showMessage(error.message, true);
        } finally {
            submitBtn.disabled = false;
        }
    });

    const notifyCheckbox = document.getElementById("notify-comment-checkbox");
    notifyCheckbox.addEventListener("change", async () => {
        notifyCheckbox.disabled = true;
        try {
            await api.updateProfile({ notify_on_comment: notifyCheckbox.checked });
        } catch (error) {
            notifyCheckbox.checked = !notifyCheckbox.checked;
            showMessage(error.message, true);
        } finally {
            notifyCheckbox.disabled = false;
        }
    });

    const avatarInput = document.getElementById("avatar-input");
    const avatarUploadBtn = document.getElementById("avatar-upload-btn");
    avatarUploadBtn.addEventListener("click", async () => {
        const file = avatarInput.files[0];
        if (!file) {
            return showMessage("Choose an image file first.", true);
        }
        const formData = new FormData();
        formData.append("image", file);
        avatarUploadBtn.disabled = true;
        try {
            const updated = await api.uploadAvatar(formData);
            setAvatarPreview(updated.avatar_path);
            avatarInput.value = "";
            showMessage("Avatar updated.");
        } catch (error) {
            showMessage(error.message, true);
        } finally {
            avatarUploadBtn.disabled = false;
        }
    });

    const deleteAccountBtn = document.getElementById("delete-account-btn");
    deleteAccountBtn.addEventListener("click", async () => {
        if (!(await confirmDialog("Are you sure? This will permanently delete your account and all of your pictures."))) {
            return;
        }
        deleteAccountBtn.disabled = true;
        try {
            await api.deleteAccount();
            window.location.href = "/index.html";
        } catch (error) {
            showMessage(error.message, true);
            deleteAccountBtn.disabled = false;
        }
    });
}

document.addEventListener("DOMContentLoaded", guardAndInit);
