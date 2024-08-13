// "user strict";
let wsocket;
let token = "";
const SERVER_IP = "139.196.93.196";
const SERVER_PORT = "8080";
function login() {
  const username = document.getElementById("username").value;
  if (!username) {
    alert("Please enter a username");
    return;
  }
  var myHeaders = new Headers();
  myHeaders.append("Content-Type", "application/json");

  var raw = JSON.stringify({
    username,
  });

  var requestOptions = {
    method: "POST",
    headers: myHeaders,
    body: raw,
    redirect: "follow",
  };

  fetch("http://" + SERVER_IP + ":" + SERVER_PORT + "/login", requestOptions)
    .then((response) => response.text())
    .then((result) => {
      token = JSON.parse(result)["token"];
      if (token === null || token === undefined || token === "") {
        alert("Login failed");
        return;
      }
      wsocket = new WebSocket(
        "ws://" + SERVER_IP + ":" + SERVER_PORT + "/ws?token=" + token
      );

      wsocket.onopen = function () {
        document.getElementById("messageInput").disabled = false;
        document.getElementById("sendButton").disabled = false;
        document.getElementById("username").disabled = true;
      };

      wsocket.onmessage = function (event) {
        const message = JSON.parse(event.data);
        const messagesDiv = document.getElementById("messages");
        const messageElem = document.createElement("div");
        let date = new Date(parseInt(message.timestamp));
        messageElem.textContent = `${date.toLocaleString()}: ${message.text}`;
        messageElem.className = "message";
        messagesDiv.appendChild(messageElem);
        messagesDiv.scrollTop = messagesDiv.scrollHeight;
      };

      wsocket.onclose = function () {
        console.log("WebSocket is closed now.");
      };

      wsocket.onerror = function (error) {
        console.log("WebSocket Error: ", error);
      };
    })
    .catch((error) => console.log("error", error));
}

function sendMessage() {
  const messageInput = document.getElementById("messageInput");
  const message = { text: messageInput.value, timestamp: Date.now() };
  wsocket.send(JSON.stringify(message));
  messageInput.value = "";
}
