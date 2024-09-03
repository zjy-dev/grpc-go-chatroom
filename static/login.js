// "use strict";
const SERVER_IP = "127.0.0.1";
const SERVER_PORT = "8082";
function login() {
  const username = document.getElementById("username").value.trim();
  const password = document.getElementById("password").value.trim();
  if (
    username.length < 2 ||
    username.length > 24 ||
    password.length < 3 ||
    password.length > 25
  ) {
    alert("Invalid username or password");
    return;
  }
  var myHeaders = new Headers();
  myHeaders.append("Content-Type", "application/json");

  var raw = JSON.stringify({
    username,
    password,
  });

  var requestOptions = {
    method: "POST",
    headers: myHeaders,
    body: raw,
    // redirect: "follow",
  };

  fetch(`http://${SERVER_IP}:${SERVER_PORT}/login-or-register`, requestOptions)
    .then((response) => {
      if (!response.ok) {
        alert("Login failed");
        console.log(response);
        throw new Error("Login failed");
      }
      return response.text();
    })
    .then((result) => {
      token = JSON.parse(result)["token"];
      if (token === null || token === undefined || token === "") {
        alert("Login failed");
        return;
      }

      localStorage.setItem("grpc_go_chatroom_token", token);
      alert("Login success!!!");
      window.location.href = "/static/chat.html";
    })
    .catch((error) => console.log("error", error));
}
