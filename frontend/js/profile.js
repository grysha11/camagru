import { api } from './api.js';
import { initNav } from './nav.js';
import { confirmDialog } from './modal.js';

async function guardAndInit() {
    const username = new URLSearchParams(window.location.search).get("username");

    let loggedInUser = null;
    try {
        loggedInUser = await api.me();
    } catch {
        try {
            await api.refresh();
            loggedInUser = await api.me();
        } catch {
            loggedInUser = null;
        }
    }

    if (!username) {
        if (!loggedInUser) {
            window.location.href = "/index.html";
            return;
        }
        init(loggedInUser, loggedInUser, true);
        return;
    }

    let profileUser;
    try {
        profileUser = await api.getUserProfile(username);
    } catch {
        renderNotFound();
        return;
    }

    const isOwner = loggedInUser?.username === username;
    init(profileUser, loggedInUser, isOwner);
}

function renderNotFound() {
    const grid = document.getElementById("my-pictures-grid");
    grid.textContent = "";
    const msg = document.createElement("p");
    msg.className = "hint";
    msg.textContent = "This user does not exist.";
    grid.appendChild(msg);
}

function formatDate(dateString) {
    return new Date(dateString).toLocaleDateString(undefined, {
        year: "numeric",
        month: "long",
        day: "numeric",
    });
}

function renderPictureCard(post, isOwner) {
    const card = document.createElement("div");
    card.className = "my-picture-card";

    const img = document.createElement("img");
    img.className = "my-picture-image";
    img.src = post.image_path;
    img.alt = "Post";
    card.appendChild(img);

    const meta = document.createElement("div");
    meta.className = "my-picture-meta";

    const likes = document.createElement("span");
    likes.className = "muted";
    likes.textContent = `${post.like_count} likes`;
    meta.appendChild(likes);

    if (isOwner) {
        const deleteBtn = document.createElement("button");
        deleteBtn.type = "button";
        deleteBtn.className = "link-btn danger";
        deleteBtn.textContent = "delete";
        deleteBtn.addEventListener("click", async () => {
            if (!(await confirmDialog("Are you sure you want to delete this post?"))) {
                return;
            }
            deleteBtn.disabled = true;
            try {
                await api.deletePost(post.id);
                card.remove();
            } catch (error) {
                deleteBtn.disabled = false;
            }
        });
        meta.appendChild(deleteBtn);
    }

    card.appendChild(meta);
    return card;
}

async function init(profileUser, navUser, isOwner) {
    initNav("profile", navUser);

    document.title = isOwner ? "Camagru - My Page" : `Camagru - ${profileUser.username}`;
    document.getElementById("profile-username").textContent = profileUser.username;
    if (profileUser.created_at) {
        document.getElementById("profile-member-since").textContent = `Member since ${formatDate(profileUser.created_at)}`;
    }
    document.getElementById("pictures-heading").textContent = isOwner ? "My pictures" : `${profileUser.username}'s pictures`;

    const avatarBox = document.getElementById("profile-avatar");
    if (profileUser.avatar_path) {
        avatarBox.textContent = "";
        const avatarImg = document.createElement("img");
        avatarImg.src = profileUser.avatar_path;
        avatarImg.alt = `${profileUser.username}'s avatar`;
        avatarBox.appendChild(avatarImg);
    }

    const grid = document.getElementById("my-pictures-grid");
    const meta = document.getElementById("my-pictures-meta");

    let posts;
    try {
        posts = isOwner ? await api.myPosts() : await api.listUserPosts(profileUser.username);
    } catch {
        posts = [];
    }

    const likesReceived = posts.reduce((sum, p) => sum + p.like_count, 0);
    const commentsReceived = posts.reduce((sum, p) => sum + p.comment_count, 0);

    document.getElementById("stat-pictures").textContent = posts.length;
    document.getElementById("stat-likes").textContent = likesReceived;
    document.getElementById("stat-comments").textContent = commentsReceived;
    meta.textContent = `${posts.length} total`;

    grid.textContent = "";
    if (posts.length === 0) {
        const empty = document.createElement("p");
        empty.className = "hint";
        empty.textContent = "No pictures yet.";
        grid.appendChild(empty);
    } else {
        for (const post of posts) {
            grid.appendChild(renderPictureCard(post, isOwner));
        }
    }
}

document.addEventListener("DOMContentLoaded", guardAndInit);
