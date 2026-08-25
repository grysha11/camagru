import { api } from './api.js';
import { confirmDialog } from './modal.js';

export function renderCommentItem(postId, comment, currentUser, callbacks) {
    const li = document.createElement("li");
    li.className = "comment-item";

    const text = document.createElement("span");
    const author = document.createElement("a");
    author.className = "comment-author";
    author.href = `/profile.html?username=${encodeURIComponent(comment.username)}`;
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
            if (!(await confirmDialog("Are you sure you want to delete this comment?"))) {
                return;
            }
            deleteBtn.disabled = true;
            try {
                await api.deleteComment(postId, comment.id);
                li.remove();
            } catch (error) {
                if (error.status === 401) {
                    callbacks.requireLogin();
                } else {
                    callbacks.showMessage(error.message, true);
                }
                deleteBtn.disabled = false;
            }
        });
        li.appendChild(deleteBtn);
    }

    return li;
}

export async function renderComments(postId, listEl, currentUser, callbacks, opts = {}) {
    let comments;
    try {
        comments = await api.listComments(postId);
    } catch (error) {
        return [];
    }

    const shown = typeof opts.limit === "number" ? comments.slice(-opts.limit) : comments;

    listEl.textContent = "";
    for (const comment of shown) {
        listEl.appendChild(renderCommentItem(postId, comment, currentUser, callbacks));
    }

    return comments;
}

export function renderCommentForm(postId, commentListEl, currentUser, callbacks) {
    if (!currentUser) {
        const loginPrompt = document.createElement("p");
        loginPrompt.className = "hint";
        loginPrompt.textContent = "Log in to comment.";
        return loginPrompt;
    }

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
            const comment = await api.createComment(postId, content);
            commentListEl.appendChild(renderCommentItem(postId, comment, currentUser, callbacks));
            input.value = "";
        } catch (error) {
            if (error.status === 401) {
                callbacks.requireLogin();
            } else {
                callbacks.showMessage(error.message, true);
            }
        } finally {
            submitBtn.disabled = false;
        }
    });

    return form;
}
