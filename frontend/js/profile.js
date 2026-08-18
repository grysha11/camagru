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

function formatDate(dateString) {
    return new Date(dateString).toLocaleDateString(undefined, {
        year: "numeric",
        month: "long",
        day: "numeric",
    });
}

function renderPictureCard(post) {
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

    const deleteBtn = document.createElement("button");
    deleteBtn.type = "button";
    deleteBtn.className = "link-btn danger";
    deleteBtn.textContent = "delete";
    deleteBtn.addEventListener("click", async () => {
        deleteBtn.disabled = true;
        try {
            await api.deletePost(post.id);
            card.remove();
        } catch (error) {
            deleteBtn.disabled = false;
        }
    });
    meta.appendChild(deleteBtn);

    card.appendChild(meta);
    return card;
}

async function init(user) {
    initNav("profile", user);

    document.getElementById("profile-username").textContent = user.username;
    if (user.created_at) {
        document.getElementById("profile-member-since").textContent = `Member since ${formatDate(user.created_at)}`;
    }

    const grid = document.getElementById("my-pictures-grid");
    const meta = document.getElementById("my-pictures-meta");

    let posts;
    try {
        posts = await api.myPosts();
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
            grid.appendChild(renderPictureCard(post));
        }
    }
}

document.addEventListener("DOMContentLoaded", guardAndInit);
