const editor = document.getElementById("editor");
const uploadBtn = document.getElementById("saveBtn");

const FILE_URL = "/file/gateway";
const SAVE_URL = "/file/upload";
const FINDINGS_URL = "/file/retrieve";
const ACCEPT_URL = "/file/accept";

window.addEventListener("DOMContentLoaded", async () => {
    try {
        const response = await fetch(FILE_URL);
        if (!response.ok) throw new Error(`Failed to load ${FILE_URL}: ${response.statusText}`);
        const text = await response.text();
        editor.value = text;
    } catch (err) {
        editor.value = `Error loading ${FILE_URL}:\n${err.message}`;
        console.log("Error loading Gateway Config file");
    }
});

uploadBtn.addEventListener("click", async () => {
    const content = editor.value;

    try {
        const response = await fetch(SAVE_URL, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ content: content })
        });

        if (!response.ok) throw new Error(`Save failed: ${response.statusText}`);

        const findings = await pollFindings();

        if (!findings) {
            showMessageModal("Error", "Could not retrieve findings from the backend.", "error");
            return;
        }

        showFindingsModal(findings, content);

    } catch (err) {
        showMessageModal("Error", "Error saving: " + err.message, "error");
    }
});

async function pollFindings(retries = 10, delayMs = 500) {
    for (let i = 0; i < retries; i++) {
        try {
            const res = await fetch(FINDINGS_URL);
            if (res.ok) {
                const text = await res.text();
                if (text.trim() !== "") return JSON.parse(text);
            }
        } catch (err) {
            console.error("Polling error:", err);
        }
        await new Promise(r => setTimeout(r, delayMs));
    }
    return null;
}

function openModal(title) {
    const backdrop = document.createElement("div");
    backdrop.className = "modal-backdrop";

    const dialog = document.createElement("div");
    dialog.className = "modal";
    dialog.setAttribute("role", "dialog");
    dialog.setAttribute("aria-modal", "true");

    const header = document.createElement("div");
    header.className = "modal-header";

    const titleEl = document.createElement("h2");
    titleEl.className = "modal-title";
    titleEl.textContent = title;

    const closeBtn = document.createElement("button");
    closeBtn.className = "modal-close";
    closeBtn.innerHTML = "&times;";
    closeBtn.setAttribute("aria-label", "Close");
    closeBtn.addEventListener("click", () => closeModal(backdrop));

    header.appendChild(titleEl);
    header.appendChild(closeBtn);

    const body = document.createElement("div");
    body.className = "modal-body";

    const footer = document.createElement("div");
    footer.className = "modal-footer";

    dialog.appendChild(header);
    dialog.appendChild(body);
    dialog.appendChild(footer);
    backdrop.appendChild(dialog);
    document.body.appendChild(backdrop);

    backdrop.addEventListener("click", (e) => {
        if (e.target === backdrop) closeModal(backdrop);
    });

    const escHandler = (e) => {
        if (e.key === "Escape") {
            closeModal(backdrop);
            document.removeEventListener("keydown", escHandler);
        }
    };
    document.addEventListener("keydown", escHandler);

    requestAnimationFrame(() => backdrop.classList.add("modal-backdrop--open"));

    return { backdrop, body, footer };
}

function closeModal(backdrop) {
    backdrop.classList.remove("modal-backdrop--open");
    backdrop.addEventListener("transitionend", () => backdrop.remove(), { once: true });
}

function showMessageModal(title, message, type = "info") {
    const { backdrop, body, footer } = openModal(title);

    const msg = document.createElement("p");
    msg.className = `modal-message modal-message--${type}`;
    msg.textContent = message;
    body.appendChild(msg);

    const okBtn = document.createElement("button");
    okBtn.textContent = "OK";
    okBtn.addEventListener("click", () => closeModal(backdrop));
    footer.appendChild(okBtn);
}

function showFindingsModal(findings, content) {
    const { backdrop, body, footer } = openModal("Config Validation Results");

    let hasErrors = false;

    for (const key in findings) {
        const items = findings[key] || [];

        const section = document.createElement("div");
        section.className = "modal-section";

        const sectionTitle = document.createElement("h3");
        sectionTitle.className = "modal-section-title";
        sectionTitle.textContent = key;
        section.appendChild(sectionTitle);

        if (items.length === 0) {
            const none = document.createElement("p");
            none.className = "modal-none";
            none.textContent = "None";
            section.appendChild(none);
        } else {
            const list = document.createElement("ul");
            list.className = "modal-list";
            items.forEach(item => {
                const li = document.createElement("li");
                li.textContent = item;
                if (key.toLowerCase() === "errors")   li.classList.add("error");
                if (key.toLowerCase() === "warnings") li.classList.add("warning");
                list.appendChild(li);
            });
            section.appendChild(list);
        }

        body.appendChild(section);

        if (key.toLowerCase() === "errors" && items.length > 0) hasErrors = true;
    }

    const closeBtn = document.createElement("button");
    closeBtn.textContent = "Close";
    closeBtn.className = "btn-secondary";
    closeBtn.addEventListener("click", () => closeModal(backdrop));
    footer.appendChild(closeBtn);

    if (!hasErrors) {
        const acceptBtn = document.createElement("button");
        acceptBtn.textContent = "Accept Changes";
        acceptBtn.addEventListener("click", async () => {
            try {
                const response = await fetch(ACCEPT_URL, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ content: content })
                });
                if (!response.ok) throw new Error("Accept failed");
                closeModal(backdrop);
                showMessageModal("Success", "Configuration applied successfully!", "success");
            } catch (err) {
                closeModal(backdrop);
                showMessageModal("Error", "Error applying config: " + err.message, "error");
            }
        });
        footer.appendChild(acceptBtn);
    }
}