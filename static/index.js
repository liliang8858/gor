var wsuri = "ws://127.0.0.1:3001/ws";
var token = "aaa";
var e = "{{.}}";
var insert = "<p>Start connecting " + wsuri + "</p>";
document.getElementsByClassName("content")[0].innerHTML += insert;

var sock = new WebSocket(wsuri);
sock.onopen = function () {
    console.log("connected to " + wsuri);
    var insert = "<p>Connected to " + wsuri + "</p>";
    document.getElementsByClassName("content")[0].innerHTML += insert;
    sock.send("{\"token\": \"" + token + "\", \"event\": \"" + e + "\"}")
    insert = "<p>Register message sent</p>";
    document.getElementsByClassName("content")[0].innerHTML += insert;
};
sock.onclose = function (e) {
    console.log("connection closed (" + e.code + ")");
    var insert = "<p>Connection be closed</p>";
    document.getElementsByClassName("content")[0].innerHTML += insert;
};
sock.onmessage = function (e) {
    console.log("Receive message: " + e.data);
    var insert = "<p>Receive: " + e.data + "</p>";
    document.getElementsByClassName("content")[0].innerHTML += insert;
};