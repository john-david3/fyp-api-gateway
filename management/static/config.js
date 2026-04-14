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
            displayResultTab("Error", "Could not retrieve findings from the backend.", "error");
            return;
        }

        displayFindingsTab(findings, content);

    } catch (err) {
        displayResultTab("Error", "Error saving: " + err.message, "error");
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

function openFindingsWindow() {
    const popup = document.createElement("div");
    popup.className = "popup";

    const findingsWindow = document.createElement("div");
    findingsWindow.className = "window";

    const body = document.createElement("div");
    body.className = "window-body";

    const footer = document.createElement("div");
    footer.className = "window-footer";

    findingsWindow.appendChild(body);
    findingsWindow.appendChild(footer);
    popup.appendChild(findingsWindow);
    document.body.appendChild(popup);

    popup.addEventListener("click", (e) => {
        if (e.target === popup) closeFindingsWindow(popup);
    });

    const escHandler = (e) => {
        if (e.key === "Escape") {
            closeFindingsWindow(popup);
            document.removeEventListener("keydown", escHandler);
        }
    };
    document.addEventListener("keydown", escHandler);

    requestAnimationFrame(() => popup.classList.add("popup--open"));

    return { popup, body, footer };
}

function closeFindingsWindow(popup) {
    popup.classList.remove("popup--open");
    popup.addEventListener("transitionend", () => popup.remove(), { once: true });
}

function displayResultTab(title, message, type = "info") {
    const { popup, body, footer } = openFindingsWindow();

    const msg = document.createElement("p");
    msg.className = `message message--${type}`;
    msg.textContent = message;
    body.appendChild(msg);

    const okBtn = document.createElement("button");
    okBtn.textContent = "OK";
    okBtn.addEventListener("click", () => closeFindingsWindow(popup));
    footer.appendChild(okBtn);
}

function displayFindingsTab(findings, content) {
    const { popup, body, footer } = openFindingsWindow();
    let hasErrors = false;

    for (const key in findings) {
        const items = findings[key] || [];

        const section = document.createElement("div");
        section.className = "section";

        const sectionTitle = document.createElement("h3");
        sectionTitle.className = "section-title";
        sectionTitle.textContent = key;
        section.appendChild(sectionTitle);

        if (items.length === 0) {
            const none = document.createElement("p");
            none.className = "no-item";
            none.textContent = "None";
            section.appendChild(none);
        } else {
            const list = document.createElement("ul");
            list.className = "list-item";
            items.forEach(item => {
                const li = document.createElement("li");
                li.textContent = item;
                if (key.toLowerCase() === "errors") li.classList.add("error");
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
    closeBtn.addEventListener("click", () => closeFindingsWindow(popup));
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
                closeFindingsWindow(popup);
                displayResultTab("Success", "Configuration applied successfully!", "success");
            } catch (err) {
                closeFindingsWindow(popup);
                displayResultTab("Error", "Error applying config: " + err.message, "error");
            }
        });
        footer.appendChild(acceptBtn);
    }
}