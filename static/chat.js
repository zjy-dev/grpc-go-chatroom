"use strict";

let token = localStorage.getItem("grpc_go_chatroom_token");
if (!token) {
  alert("No token found");
  return;
}

const SERVER_IP = "127.0.0.1";
const SERVER_PORT = "8082";
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

function sendMessage() {
  const messageInput = document.getElementById("user-input");
  const message = { text: messageInput.value.trim(), timestamp: Date.now() };
  wsocket.send(JSON.stringify(message));
  messageInput.value = "";
}
