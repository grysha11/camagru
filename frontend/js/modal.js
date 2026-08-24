export function confirmDialog(message) {
    return new Promise((resolve) => {
        const overlay = document.createElement("div");
        overlay.className = "modal-overlay";

        const box = document.createElement("div");
        box.className = "modal-box";

        const text = document.createElement("p");
        text.className = "modal-message";
        text.textContent = message;
        box.appendChild(text);

        const actions = document.createElement("div");
        actions.className = "modal-actions";

        const cancelBtn = document.createElement("button");
        cancelBtn.type = "button";
        cancelBtn.className = "btn-plain";
        cancelBtn.textContent = "Cancel";

        const confirmBtn = document.createElement("button");
        confirmBtn.type = "button";
        confirmBtn.className = "btn-danger";
        confirmBtn.textContent = "Confirm";

        const close = (result) => {
            document.removeEventListener("keydown", onKeydown);
            overlay.remove();
            resolve(result);
        };

        const onKeydown = (e) => {
            if (e.key === "Escape") close(false);
        };

        cancelBtn.addEventListener("click", () => close(false));
        confirmBtn.addEventListener("click", () => close(true));
        overlay.addEventListener("click", (e) => {
            if (e.target === overlay) close(false);
        });
        document.addEventListener("keydown", onKeydown);

        actions.appendChild(cancelBtn);
        actions.appendChild(confirmBtn);
        box.appendChild(actions);
        overlay.appendChild(box);
        document.body.appendChild(overlay);

        confirmBtn.focus();
    });
}
