var buffers = [];
var notifyWhenReady = false;
var self = this;

this.onmessage = function(e){
  switch(e.data.command){
    case 'init':
      init(e.data.room);
      break;
    case 'getRemainingBuffers':
      getRemainingBuffers();
      break;
    case 'getBuffer':
      getBuffer();
      break;
    case 'notifyWhenReady':
      notifyWhenReady = true;
      break;
  }
};

function init(room) {
  var s = new WebSocket("ws://" + location.host + "/sock/" + room);
  s.binaryType = 'arraybuffer';
  s.onmessage = onSocketMessage;
}

function onSocketMessage(e) {
  var buffer = new Float32Array(e.data);
  buffers.push(deinterleave(buffer));
  if (notifyWhenReady) {
    notifyWhenReady = false;
    self.postMessage();
  }
}

function getBuffer() {
  this.postMessage(buffers.shift());
}

function getRemainingBuffers() {
  var numBuffers = buffers.length;

  if (numBuffers == 1) {
    this.postMessage(buffers.shift());
  }
  else if (numBuffers > 1) {
    var buffersToSend = [];
    for (var i = 0; i < numBuffers; i++) {
      var buffer = buffers.shift();
      if (buffer) {
        buffersToSend.push(buffer);
      }
      else {
        break;
      }
    }
    this.postMessage(mergeBuffers.apply(this, buffersToSend));
  }
  else {
    this.postMessage(null);
  } 
}

function deinterleave(input) {
  var outLength = input.length / 2;
  var left = new Float32Array(outLength);
  var right = new Float32Array(outLength);
  for (var outIdx = 0, inIdx = 0; outIdx < outLength; outIdx++) {
    left[outIdx] = input[inIdx++];
    right[outIdx] = input[inIdx++];
  }
  return [left, right];
};

function copyBuffer(dst, dstOffset, src, srcOffset, length) {
  for (var i = 0; i < length; i++) {
    dst[dstOffset + i] = src[srcOffset + i]
  }
};

function mergeBuffers() {
  var newLength = 0;
  for (var i = 0; i < arguments.length; i++) {
    newLength += arguments[i][0].length;
  }
  var outBuffer = [new Float32Array(newLength), new Float32Array(newLength)];
  var outIdx = 0;
  for (var bufIdx = 0; bufIdx < arguments.length; bufIdx++) {
    var buffer = arguments[bufIdx];
    var bufferLength = buffer[0].length;
    copyBuffer(outBuffer[0], outIdx, buffer[0], 0, bufferLength);
    copyBuffer(outBuffer[1], outIdx, buffer[1], 0, bufferLength);
    outIdx += bufferLength;    
  }
  return outBuffer;
}
