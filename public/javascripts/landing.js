$(function() {
  var $go = $('button#go');
  var $roomName = $('input#room-name');
  $roomName.keypress(function(e) {
    if (e.keyCode == 13) {
      $go.click();
    }
  });
  $go.click(function() {
    var room = $roomName.val();
    if (_.isEmpty(room)) {
      alert('You need to enter a room name, any room name.');
    }
    else {
      window.location = 'http://' + location.host + "/" + $roomName.val();
    }
  });
});