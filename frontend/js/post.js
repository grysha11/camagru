import { api } from './api.js';
import { initNav } from './nav.js';
import { confirmDialog } from './modal.js';
import { renderComments, renderCommentForm } from './comments.js';

let currentUser = null;

const statusMessage = document.getElementById("status-message");
const showMessage = (msg, isError = false) => {
    statusMessage.textContent = msg;
    statusMessage.classList.toggle("error", isError);
    statusMessage.classList.toggle("success", !isError && !!msg);
};

function requireLogin() {
    showMessage("Please log in to do that.", true);
}

const commentCallbacks = { showMessage, requireLogin };

function renderPost(post) {
    const article = document.createElement("article");
    article.className = "post-card";

    const img = document.createElement("img");
    img.className = "post-image";
    img.src = post.image_path;
    img.alt = "Post by " + post.username;
    article.appendChild(img);

    const body = document.createElement("div");
    body.className = "post-body";

    const meta = document.createElement("div");
    meta.className = "post-meta";

    const username = document.createElement("span");
    username.className = "post-username";
    username.textContent = post.username;
    meta.appendChild(username);

    const timestamp = document.createElement("span");
    timestamp.className = "post-timestamp";
    timestamp.textContent = new Date(post.created_at).toLocaleString();
    meta.appendChild(timestamp);

    body.appendChild(meta);

    let liked = post.liked_by_me;
    let likeCount = post.like_count;

    const actions = document.createElement("div");
    actions.className = "post-actions";

    const likeBtn = document.createElement("button");
    likeBtn.type = "button";
    likeBtn.className = "like-btn";
    const updateLikeLabel = () => {
        likeBtn.textContent = `${liked ? "Unlike" : "Like"} (${likeCount})`;
        likeBtn.classList.toggle("liked", liked);
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
    actions.appendChild(likeBtn);

    const commentCount = document.createElement("span");
    commentCount.className = "comment-count";
    commentCount.textContent = `${post.comment_count} comments`;
    actions.appendChild(commentCount);

    if (currentUser && currentUser.id === post.user_id) {
        const deleteBtn = document.createElement("button");
        deleteBtn.type = "button";
        deleteBtn.className = "post-delete-btn";
        deleteBtn.textContent = "Delete post";
        deleteBtn.addEventListener("click", async () => {
            if (!(await confirmDialog("Are you sure you want to delete this post?"))) {
                return;
            }
            deleteBtn.disabled = true;
            try {
                await api.deletePost(post.id);
                window.location.href = "/gallery.html";
            } catch (error) {
                if (error.status === 401) {
                    requireLogin();
                } else {
                    showMessage(error.message, true);
                }
                deleteBtn.disabled = false;
            }
        });
        actions.appendChild(deleteBtn);
    }

    body.appendChild(actions);

    const commentsSection = document.createElement("div");
    commentsSection.className = "post-comments";
    const commentList = document.createElement("ul");
    commentList.className = "comment-list";
    commentsSection.appendChild(commentList);
    commentsSection.appendChild(renderCommentForm(post.id, commentList, currentUser, commentCallbacks));

    body.appendChild(commentsSection);
    article.appendChild(body);

    renderComments(post.id, commentList, currentUser, commentCallbacks);

    return article;
}

function renderNotFound() {
    const container = document.getElementById("post-container");
    container.textContent = "";
    const msg = document.createElement("p");
    msg.className = "hint";
    msg.textContent = "This post does not exist. It may have been deleted.";
    container.appendChild(msg);
}

async function init() {
    try {
        currentUser = await api.me();
    } catch {
        currentUser = null;
    }

    initNav(null, currentUser);

    const postId = new URLSearchParams(window.location.search).get("id");
    const container = document.getElementById("post-container");

    if (!postId) {
        renderNotFound();
        return;
    }

    let post;
    try {
        post = await api.getPost(postId);
    } catch (error) {
        renderNotFound();
        return;
    }

    container.textContent = "";
    container.appendChild(renderPost(post));
}

document.addEventListener("DOMContentLoaded", init);
