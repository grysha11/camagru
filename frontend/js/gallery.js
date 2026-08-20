import { api } from './api.js';
import { initNav } from './nav.js';

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
    li.className = "comment-item";

    const text = document.createElement("span");
    const author = document.createElement("strong");
    author.className = "comment-author";
    author.textContent = comment.username;
    text.appendChild(author);
    const body = document.createElement("span");
    body.className = "comment-text";
    body.textContent = ": " + comment.content;
    text.appendChild(body);
    li.appendChild(text);

    if (currentUser && currentUser.id === comment.user_id) {
        const deleteBtn = document.createElement("button");
        deleteBtn.type = "button";
        deleteBtn.className = "link-btn danger";
        deleteBtn.textContent = "Delete";
        deleteBtn.addEventListener("click", async () => {
            if (!confirm("Are you sure you want to delete this comment?")) {
                return;
            }
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
            if (!confirm("Are you sure you want to delete this post?")) {
                return;
            }
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
        actions.appendChild(deleteBtn);
    }

    body.appendChild(actions);

    const commentsSection = document.createElement("div");
    commentsSection.className = "post-comments";
    const commentList = document.createElement("ul");
    commentList.className = "comment-list";
    commentsSection.appendChild(commentList);

    if (currentUser) {
        const form = document.createElement("form");
        form.className = "comment-form";

        const input = document.createElement("input");
        input.type = "text";
        input.className = "comment-input";
        input.required = true;
        input.maxLength = 1000;
        input.placeholder = "Add a comment";
        form.appendChild(input);

        const submitBtn = document.createElement("button");
        submitBtn.type = "submit";
        submitBtn.className = "btn-plain";
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
        loginPrompt.className = "hint";
        loginPrompt.textContent = "Log in to comment.";
        commentsSection.appendChild(loginPrompt);
    }

    body.appendChild(commentsSection);
    article.appendChild(body);

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
        empty.className = "hint";
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

    initNav("wall", currentUser);

    const requestedPage = parseInt(new URLSearchParams(window.location.search).get("page"), 10);
    loadPage(requestedPage > 0 ? requestedPage : 1);
}

document.addEventListener("DOMContentLoaded", init);
