async function sendLoginData(name, password) {
    try {
        const response = await fetch("http://localhost:80/api/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name: name, password: password })
        });

        if (!response.ok) {
            throw new Error("Request failed with status " + response.status);
        }

        const result = await response.json();
        console.log("Server response:", result);
    } catch (error) {
        console.error("Error sending login data:", error);
    }
}

document.getElementById("loginForm").addEventListener("submit", function (e) {
    e.preventDefault();

    const name = document.getElementById("name").value;
    const password = document.getElementById("password").value;

    sendLoginData(name, password);
});