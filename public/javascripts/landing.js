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
      window.location = location.protocol + '//' + location.host + "/" + $roomName.val();
    }
  });

  if (_.isEmpty(rooms)) {
    $('div#room-list-container').hide();
  } else {
    for (var i = 0; i < 10 && i < rooms.length; i++) {
      var r = rooms[i];
      $('<a/>', {href: location.protocol + '//' + location.host + "/" + r, text: r}).appendTo($('div#room-list'));
    }
  }
});