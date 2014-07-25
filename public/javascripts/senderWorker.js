this.onmessage = function(e){
  switch(e.data.command){
    case 'init':
      init(e.data.room);
      break;
    case 'sendBuffer':
      sendBuffer(e.data.left, e.data.right);
      break;
  }
};

var s = null;

function init(room) {
  s = new WebSocket("ws://" + location.host + "/sock/" + room + '/s');
  s.binaryType = 'arraybuffer';
}

function sendBuffer(left, right) {
  s.send(interleave(left, right));
}

function interleave(left, right) {
  var output = new Float32Array(left.length * 2);
  for (var inIdx = 0, outIdx = 0; inIdx < left.length; inIdx++) {
    output[outIdx++] = left[inIdx];
    output[outIdx++] = right[inIdx];
  }
  return output;
}
