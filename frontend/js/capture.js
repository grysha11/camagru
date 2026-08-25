import { api } from './api.js';
import { confirmDialog } from './modal.js';
import { initNav } from './nav.js';

let selectedOverlayId = null;

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

function init(user) {
    const statusMessage = document.getElementById("status-message");
    const overlayList = document.getElementById("overlay-list");
    const video = document.getElementById("camera-preview");
    const cameraWrap = document.getElementById("camera-wrap");
    const overlayPreview = document.getElementById("overlay-preview");
    const captureBtn = document.getElementById("capture-btn");
    const fileFallback = document.getElementById("file-fallback");
    const canvas = document.getElementById("capture-canvas");
    const myPostsEl = document.getElementById("my-posts");

    initNav("capture", user);

    const showMessage = (msg, isError = false) => {
        statusMessage.textContent = msg;
        statusMessage.classList.toggle("error", isError);
        statusMessage.classList.toggle("success", !isError && !!msg);
    };

    async function loadOverlays() {
        let overlays;
        try {
            overlays = await api.listOverlays();
        } catch (error) {
            showMessage(error.message, true);
            return;
        }

        overlayList.textContent = "";
        for (const overlay of overlays) {
            const button = document.createElement("button");
            button.type = "button";
            button.className = "overlay-option";

            const img = document.createElement("img");
            img.className = "overlay-thumb";
            img.src = overlay.url;
            img.alt = overlay.id;
            button.appendChild(img);

            const label = document.createElement("span");
            label.className = "overlay-label";
            label.textContent = overlay.id;
            button.appendChild(label);

            button.addEventListener("click", () => selectOverlay(overlay, button));
            overlayList.appendChild(button);
        }
    }

    function selectOverlay(overlay, buttonEl) {
        selectedOverlayId = overlay.id;

        for (const btn of overlayList.querySelectorAll(".overlay-option")) {
            btn.classList.toggle("selected", btn === buttonEl);
        }

        const probe = new Image();
        probe.onload = () => {
            canvas.width = probe.naturalWidth;
            canvas.height = probe.naturalHeight;
            cameraWrap.style.aspectRatio = `${probe.naturalWidth} / ${probe.naturalHeight}`;
        };
        probe.src = overlay.url;

        overlayPreview.src = overlay.url;
        overlayPreview.classList.remove("hidden");

        captureBtn.disabled = false;
        fileFallback.disabled = false;
    }

    async function initCamera() {
        try {
            const stream = await navigator.mediaDevices.getUserMedia({ video: true });
            video.srcObject = stream;
        } catch {
            showMessage("Camera unavailable — use the file upload option instead.", true);
        }
    }

    function getSourceDimensions(el) {
        return {
            width: el.videoWidth ?? el.naturalWidth,
            height: el.videoHeight ?? el.naturalHeight,
        };
    }

    function captureFrom(sourceElement) {
        if (!selectedOverlayId || canvas.width === 0 || canvas.height === 0) {
            showMessage("Choose an overlay first.", true);
            return;
        }

        const { width: srcW, height: srcH } = getSourceDimensions(sourceElement);
        if (!srcW || !srcH) {
            showMessage("Camera isn't ready yet — try again in a moment.", true);
            return;
        }

        const sourceAspect = srcW / srcH;
        const targetAspect = canvas.width / canvas.height;
        let sx = 0, sy = 0, cropW = srcW, cropH = srcH;
        if (sourceAspect > targetAspect) {
            cropW = srcH * targetAspect;
            sx = (srcW - cropW) / 2;
        } else {
            cropH = srcW / targetAspect;
            sy = (srcH - cropH) / 2;
        }

        const ctx = canvas.getContext("2d");
        ctx.drawImage(sourceElement, sx, sy, cropW, cropH, 0, 0, canvas.width, canvas.height);
        canvas.toBlob((blob) => uploadCapture(blob), "image/png");
    }

    async function uploadCapture(blob) {
        const formData = new FormData();
        formData.append("image", blob, "capture.png");
        formData.append("overlay_id", selectedOverlayId);

        captureBtn.disabled = true;
        try {
            await api.uploadPost(formData);
            showMessage("Post created!");
            await loadMyPosts();
        } catch (error) {
            showMessage(error.message, true);
        } finally {
            captureBtn.disabled = false;
        }
    }

    async function loadMyPosts() {
        let posts;
        try {
            posts = await api.myPosts();
        } catch (error) {
            showMessage(error.message, true);
            return;
        }

        myPostsEl.textContent = "";
        for (const post of posts) {
            const wrap = document.createElement("div");
            wrap.className = "my-post-thumb-wrap";

            const img = document.createElement("img");
            img.className = "my-post-thumb";
            img.src = post.image_path;
            img.alt = "Post";
            wrap.appendChild(img);

            const deleteBtn = document.createElement("button");
            deleteBtn.type = "button";
            deleteBtn.className = "thumb-delete-btn";
            deleteBtn.textContent = "×";
            deleteBtn.setAttribute("aria-label", "Delete post");
            deleteBtn.addEventListener("click", async () => {
                if (!(await confirmDialog("Are you sure you want to delete this post?"))) {
                    return;
                }
                deleteBtn.disabled = true;
                try {
                    await api.deletePost(post.id);
                    wrap.remove();
                } catch (error) {
                    showMessage(error.message, true);
                    deleteBtn.disabled = false;
                }
            });
            wrap.appendChild(deleteBtn);

            myPostsEl.appendChild(wrap);
        }
    }

    captureBtn.addEventListener("click", () => captureFrom(video));

    fileFallback.addEventListener("change", () => {
        const file = fileFallback.files[0];
        if (!file) return;

        const objectUrl = URL.createObjectURL(file);
        const tempImg = new Image();
        tempImg.onload = () => {
            captureFrom(tempImg);
            URL.revokeObjectURL(objectUrl);
        };
        tempImg.src = objectUrl;
    });

    loadOverlays();
    loadMyPosts();
    initCamera();
}

document.addEventListener("DOMContentLoaded", guardAndInit);
