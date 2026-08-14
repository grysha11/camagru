import { api } from './api.js';

let currentUser = null;

const statusMessage = document.getElementById("status-message");
const showMessage = (msg, isError = false) => {
    statusMessage.textContent = msg;
    statusMessage.style.color = isError ? "red" : "green";
};

function requireLogin() {
    showMessage("Please log in to do that.", true);
}

async function renderComments(postId, listEl) {
    let comments;
    try {
        comments = await api.listComments(postId);
    } catch (error) {
        return;
    }

    listEl.textContent = "";
    for (const comment of comments) {
        listEl.appendChild(renderCommentItem(postId, comment));
    }
}

function renderCommentItem(postId, comment) {
    const li = document.createElement("li");

    const author = document.createElement("strong");
    author.textContent = comment.username;
    li.appendChild(author);

    const text = document.createElement("span");
    text.textContent = ": " + comment.content;
    li.appendChild(text);

    if (currentUser && currentUser.id === comment.user_id) {
        const deleteBtn = document.createElement("button");
        deleteBtn.type = "button";
        deleteBtn.textContent = "Delete";
        deleteBtn.addEventListener("click", async () => {
            deleteBtn.disabled = true;
            try {
                await api.deleteComment(postId, comment.id);
                li.remove();
            } catch (error) {
                if (error.status === 401) {
                    requireLogin();
                } else {
                    showMessage(error.message, true);
                }
                deleteBtn.disabled = false;
            }
        });
        li.appendChild(deleteBtn);
    }

    return li;
}

function renderPostCard(post) {
    const article = document.createElement("article");
    article.style.marginBottom = "24px";
    article.style.borderBottom = "1px solid #ccc";
    article.style.paddingBottom = "12px";

    const username = document.createElement("p");
    username.textContent = post.username;
    article.appendChild(username);

    const img = document.createElement("img");
    img.src = post.image_path;
    img.alt = "Post by " + post.username;
    img.style.maxWidth = "400px";
    article.appendChild(img);

    const timestamp = document.createElement("p");
    timestamp.textContent = new Date(post.created_at).toLocaleString();
    article.appendChild(timestamp);

    let liked = post.liked_by_me;
    let likeCount = post.like_count;

    const likeBtn = document.createElement("button");
    likeBtn.type = "button";
    const updateLikeLabel = () => {
        likeBtn.textContent = `${liked ? "Unlike" : "Like"} (${likeCount})`;
    };
    updateLikeLabel();
    likeBtn.addEventListener("click", async () => {
        if (!currentUser) {
            requireLogin();
            return;
        }
        likeBtn.disabled = true;
        try {
            if (liked) {
                await api.unlikePost(post.id);
                likeCount--;
            } else {
                await api.likePost(post.id);
                likeCount++;
            }
            liked = !liked;
            updateLikeLabel();
        } catch (error) {
            if (error.status === 401) {
                requireLogin();
            } else {
                showMessage(error.message, true);
            }
        } finally {
            likeBtn.disabled = false;
        }
    });
    article.appendChild(likeBtn);

    if (currentUser && currentUser.id === post.user_id) {
        const deleteBtn = document.createElement("button");
        deleteBtn.type = "button";
        deleteBtn.textContent = "Delete post";
        deleteBtn.addEventListener("click", async () => {
            deleteBtn.disabled = true;
            try {
                await api.deletePost(post.id);
                article.remove();
            } catch (error) {
                if (error.status === 401) {
                    requireLogin();
                } else {
                    showMessage(error.message, true);
                }
                deleteBtn.disabled = false;
            }
        });
        article.appendChild(deleteBtn);
    }

    const commentsSection = document.createElement("div");
    const commentList = document.createElement("ul");
    commentsSection.appendChild(commentList);

    if (currentUser) {
        const form = document.createElement("form");

        const input = document.createElement("input");
        input.type = "text";
        input.required = true;
        input.maxLength = 1000;
        input.placeholder = "Add a comment";
        form.appendChild(input);

        const submitBtn = document.createElement("button");
        submitBtn.type = "submit";
        submitBtn.textContent = "Comment";
        form.appendChild(submitBtn);

        form.addEventListener("submit", async (e) => {
            e.preventDefault();
            const content = input.value.trim();
            if (!content) return;

            submitBtn.disabled = true;
            try {
                const comment = await api.createComment(post.id, content);
                commentList.appendChild(renderCommentItem(post.id, comment));
                input.value = "";
            } catch (error) {
                if (error.status === 401) {
                    requireLogin();
                } else {
                    showMessage(error.message, true);
                }
            } finally {
                submitBtn.disabled = false;
            }
        });

        commentsSection.appendChild(form);
    } else {
        const loginPrompt = document.createElement("p");
        loginPrompt.textContent = "Log in to comment.";
        commentsSection.appendChild(loginPrompt);
    }

    article.appendChild(commentsSection);

    renderComments(post.id, commentList);

    return article;
}

async function loadPage(page) {
    let data;
    try {
        data = await api.listPosts(page);
    } catch (error) {
        showMessage(error.message, true);
        return;
    }

    const postList = document.getElementById("post-list");
    postList.textContent = "";

    if (data.posts.length === 0) {
        const empty = document.createElement("p");
        empty.textContent = "No posts yet.";
        postList.appendChild(empty);
    } else {
        for (const post of data.posts) {
            postList.appendChild(renderPostCard(post));
        }
    }

    const pageIndicator = document.getElementById("page-indicator");
    pageIndicator.textContent = data.total_posts > 0
        ? `Page ${data.page} of ${data.total_pages}`
        : "";

    const prevBtn = document.getElementById("prev-page-btn");
    const nextBtn = document.getElementById("next-page-btn");
    prevBtn.disabled = !data.has_prev;
    nextBtn.disabled = !data.has_next;

    prevBtn.onclick = () => loadPage(data.page - 1);
    nextBtn.onclick = () => loadPage(data.page + 1);

    window.history.replaceState({}, "", `?page=${data.page}`);
}

async function init() {
    try {
        currentUser = await api.me();
    } catch {
        currentUser = null;
    }

    const accountLink = document.getElementById("account-link");
    const loginLink = document.getElementById("login-link");
    if (currentUser) {
        accountLink.style.display = "";
        loginLink.style.display = "none";
    }

    const requestedPage = parseInt(new URLSearchParams(window.location.search).get("page"), 10);
    loadPage(requestedPage > 0 ? requestedPage : 1);
}

document.addEventListener("DOMContentLoaded", init);
