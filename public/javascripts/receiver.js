function Player(room) {
  this.room = room;
  this.playing = false;
  this.shouldPlay = false;
  this.audioContext = new AudioContext();
  this.audioSource = null;
  this.ended_ = _.bind(this.ended_, this);
  this.worker = new Worker("/javascripts/receiverWorker.js");
  this.worker.postMessage({
    command: 'init',
    room: room
  });
  this.worker.onmessage = _.bind(this.onWorkerMessage_, this);
}

Player.prototype.onWorkerMessage_ = function(e) {
  if (this.currWorkerCallback) {
    this.currWorkerCallback.call(this, e.data);
  }
}

Player.prototype.getBuffer = function(callback) {
  this.currWorkerCallback = callback;
  this.worker.postMessage({
    command: 'getBuffer'
  })
}

Player.prototype.getRemainingBuffers = function(callback) {
  this.currWorkerCallback = callback;
  this.worker.postMessage({
    command: 'getRemainingBuffers'
  });
}

Player.prototype.play = function() {
  this.shouldPlay = true;
  this.startPlaying_();
  return true;
}

Player.prototype.startPlaying_ = function() {
  if (this.playing) {
    return true;
  }

  if (this.audioSource) {
    this.playCurrentBuffer_();
  }
  else {
    this.playNextBuffer_();
  }

  return false;
}

Player.prototype.stop = function() {
  this.shouldPlay = false;

  if (!this.playing) {
    return true;
  }

  if (this.audioSource) {
    this.audioSource.stop();
    this.audioSource.startOffset += this.audioContext.currentTime - this.audioSource.startTime;
    this.playing = false;
    return true;
  }

  return false;
}

Player.prototype.playCurrentBuffer_ = function() {
  var source = this.audioContext.createBufferSource();
  source.onended = this.ended_;
  source.connect(this.audioContext.destination);
  source.buffer = this.audioSource.buffer;
  source.startOffset = 0;

  source.startTime = this.audioContext.currentTime;
  source.startOffset = this.audioSource.startOffset;
  source.start(0, source.startOffset);

  this.audioSource.disconnect();
  this.audioSource = source;
  this.playing = true;
}

Player.prototype.playNextBuffer_ = function(callback) {
  var self = this;
  this.getBuffer(function(nextBuffer) {
    var result;
    if (nextBuffer) {
      var source = self.audioContext.createBufferSource();
      source.onended = self.ended_;
      source.connect(self.audioContext.destination);

      var sourceBuffer = self.audioContext.createBuffer(2, nextBuffer[0].length, self.audioContext.sampleRate);
      sourceBuffer.getChannelData(0).set(nextBuffer[0]);
      sourceBuffer.getChannelData(1).set(nextBuffer[1]);
      source.buffer = sourceBuffer;
      source.startTime = self.audioContext.currentTime;
      source.startOffset = 0;
      source.start(0, source.startOffset);

      if (self.audioSource) {
        self.audioSource.disconnect();
      }
      self.audioSource = source;
      this.playing = true;
    }
    else {
      self.audioSource = null;
      this.playing = false;
      this.currWorkerCallback = this.startPlaying_;
      this.worker.postMessage({
        command: 'notifyWhenReady'
      })
    }

    if (callback) {
      callback(result);
    }
  });
}

Player.prototype.ended_ = function() {
  if (this.playing) {
    this.playNextBuffer_();
  }
}

$(function() {
  var room = $('#room').val();
  var player = new Player(room);

  var $controls = $('#controls');
  $('button#play').click(function() {
    if (player.play()) {
      $controls.addClass('playing');
    }
  })

  $('button#pause').click(function() {
    if (player.stop()) {
      $controls.removeClass('playing');
    }
  });
});
