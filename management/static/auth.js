async function sendLoginData(name, password, url) {
    try {
        const response = await fetch(url, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name: name, password: password })
        });

        if (!response.ok) {
            throw new Error("Request failed with status " + response.status);
        }

    } catch (error) {
        console.error("Error sending login data:", error);
    }
}

document.getElementById("loginForm").addEventListener("submit", function (e) {
    e.preventDefault();

    const name = document.getElementById("name").value;
    const password = document.getElementById("password").value;

    sendLoginData(name, password, "http://54.75.125.2:80/api/login");
});

document.getElementById("signupForm").addEventListener("submit", function(e){
    e.preventDefault();

    const name = document.getElementById("signupName").value;
    const password = document.getElementById("signupPassword").value;

    sendLoginData(name, password, "http://54.75.125.2:80/api/signup");
});