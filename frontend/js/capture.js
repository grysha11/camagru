import { api } from './api.js';

let selectedOverlayId = null;

async function guardAndInit() {
    try {
        await api.me();
    } catch {
        try {
            await api.refresh();
            await api.me();
        } catch {
            window.location.href = "/index.html";
            return;
        }
    }

    init();
}

function init() {
    const statusMessage = document.getElementById("status-message");
    const overlayList = document.getElementById("overlay-list");
    const video = document.getElementById("camera-preview");
    const overlayPreview = document.getElementById("overlay-preview");
    const captureBtn = document.getElementById("capture-btn");
    const fileFallback = document.getElementById("file-fallback");
    const canvas = document.getElementById("capture-canvas");
    const myPostsEl = document.getElementById("my-posts");

    const showMessage = (msg, isError = false) => {
        statusMessage.textContent = msg;
        statusMessage.style.color = isError ? "red" : "green";
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

            const img = document.createElement("img");
            img.src = overlay.url;
            img.alt = overlay.id;
            img.style.height = "60px";
            button.appendChild(img);

            const label = document.createElement("span");
            label.textContent = overlay.id;
            button.appendChild(label);

            button.addEventListener("click", () => selectOverlay(overlay));
            overlayList.appendChild(button);
        }
    }

    function selectOverlay(overlay) {
        selectedOverlayId = overlay.id;

        const probe = new Image();
        probe.onload = () => {
            canvas.width = probe.naturalWidth;
            canvas.height = probe.naturalHeight;
        };
        probe.src = overlay.url;

        overlayPreview.src = overlay.url;
        overlayPreview.style.display = "block";

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

    function captureFrom(sourceElement) {
        if (!selectedOverlayId || canvas.width === 0 || canvas.height === 0) {
            showMessage("Choose an overlay first.", true);
            return;
        }

        const ctx = canvas.getContext("2d");
        ctx.drawImage(sourceElement, 0, 0, canvas.width, canvas.height);
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
            const img = document.createElement("img");
            img.src = post.image_path;
            img.alt = "Post";
            img.style.height = "120px";
            img.style.marginRight = "8px";
            myPostsEl.appendChild(img);
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
