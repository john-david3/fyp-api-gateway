const editor = document.getElementById("editor");
const uploadBtn = document.getElementById("saveBtn");

const FILE_URL = "http://localhost:80/file/gateway";
const SAVE_URL = "http://localhost:80/file/upload";
const FINDINGS_URL = "http://localhost:80/file/retrieve";
const ACCEPT_URL = "http://localhost:80/file/accept";

window.addEventListener("DOMContentLoaded", async () => {
    try {
        // retrieve the users config file from the database
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
        // Send new config to backend
        const response = await fetch(SAVE_URL, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ content: content })
        });

        if (!response.ok) throw new Error(`Save failed: ${response.statusText}`);

        // Wait for findings to appear (poll until non-empty)
        const findings = await pollFindings();

        if (!findings) {
            alert("Could not retrieve findings from backend.");
            return;
        }

        displayFindings(findings);

    } catch (err) {
        alert("Error saving: " + err.message);
    }
});

async function pollFindings(retries = 10, delayMs = 500) {
    for (let i = 0; i < retries; i++) {
        try {
            const res = await fetch(FINDINGS_URL);
            if (res.ok) {
                const text = await res.text();
                if (text.trim() !== "") {
                    return JSON.parse(text);
                }
            }
        } catch (err) {
            console.error("Polling error:", err);
        }
        await new Promise(r => setTimeout(r, delayMs));
    }
    return null;
}

// Display findings in the DOM
function displayFindings(findings) {
    const container = document.getElementById("findings");
    container.innerHTML = "";

    let hasErrors = false;

    // Check each key in findings
    for (const key in findings) {
        const section = document.createElement("div");
        section.innerHTML = `<h3>${key}</h3>`;
        const list = document.createElement("ul");

        const items = findings[key] || [];

        items.forEach(item => {
            const li = document.createElement("li");
            li.textContent = item;

            if (key.toLowerCase() === "errors") li.classList.add("error");
            if (key.toLowerCase() === "warnings") li.classList.add("warning");

            list.appendChild(li);
        });

        section.appendChild(list);
        container.appendChild(section);

        // Only mark as hasErrors if errors array is non-empty
        if (key.toLowerCase() === "errors" && items.length > 0) {
            hasErrors = true;
        }
    }

    // Show Accept button if there are no errors, or if findings map is empty
    const shouldShowAccept = !hasErrors;

    if (shouldShowAccept) {
        const acceptBtn = document.createElement("button");
        acceptBtn.textContent = "Accept Changes";
        acceptBtn.id = "acceptBtn";

        acceptBtn.addEventListener("click", async () => {
            const content = editor.value;

            try {
                console.log(content);
                const response = await fetch(ACCEPT_URL, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ content: content })
                });
                if (!response.ok) {
                    throw new Error("Accept failed");
                }
                alert("Configuration applied successfully!");
                acceptBtn.remove();
            } catch (err) {
                alert("Error applying config: " + err.message);
            }
        });

        container.appendChild(acceptBtn);
    }
}

