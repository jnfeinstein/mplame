window.AudioContext = window.AudioContext || window.webkitAudioContext;
navigator.getUserMedia = navigator.getUserMedia || navigator.webkitGetUserMedia || navigator.mozGetUserMedia;


function setupChat(room) {
  var $chatWindow = $('div#chat-window');
  var $chatName = $('input#chat-name');
  var $chatText = $('input#chat-text');
  var $chatSendButton = $('button#chat-send');
  var s = new WebSocket("ws://" + location.host + "/sock/" + room + "/c");

  var appendChat = function(chat) {
    $('<div/>', {'class': 'chat', text: chat}).appendTo($chatWindow);
  };

  s.onmessage = function(e) {
    appendChat(e.data);
  };
  $chatSendButton.click(function() {
    var name = $chatName.val();
    var text = $chatText.val();

    if (_.isEmpty(name)) {
      alert('You must pick a chat handle!');
      return;
    }
    if (_.isEmpty(text)) {
      alert("No really, your message can't be blank...");
      return;
    }

    var msg = name + ": " + text;
    s.send(msg);
    appendChat(msg);
    $chatText.val("");
  });

  $chatText.keypress(function(e) {
    if (e.keyCode == 13) {
      $chatSendButton.click();
    }
  });
}

$(function() {
  var room = $('#room').val();
  setupChat(room);
});