const editor = document.getElementById("editor");
const uploadBtn = document.getElementById("saveBtn");

const SAVE_URL = "http://localhost:80/file/upload";
const FILE_URL = "gateway.yaml";

window.addEventListener("DOMContentLoaded", async () => {
    try {
        const response = await fetch(FILE_URL);
        if (!response.ok) {
            throw new Error(`Failed to load ${FILE_URL}: ${response.statusText}`);
        }
        const text = await response.text();
        editor.value = text;
    } catch (err) {
        editor.value = `Error loading ${FILE_URL}:\n${err.message}`;
    }
});

// Upload button
uploadBtn.addEventListener("click", async () => {
    const content = editor.value;

    try {
        const response = await fetch(SAVE_URL, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                filename: "gateway.yaml",
                content: content
            }),
        });

        if (!response.ok) {
            throw new Error(`Save failed: ${response.statusText}`);
        }

        alert("Saved successfully!");
    } catch (err) {
        alert("Error saving: " + err.message);
    }
});
