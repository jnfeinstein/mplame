function Recorder(room) {
  this.recording = false;
  this.audioContext = new AudioContext();
  this.sourceId = null;
  this.sourceInput = null;
  this.sourceProcessor = null;
  this.startUserMedia_ = _.bind(this.startUserMedia_, this);
  this.newAudio_ = _.bind(this.newAudio_, this);
  this.worker = new Worker("/javascripts/senderWorker.js");
  this.worker.postMessage({
    command: 'init',
    room: room
  });
}

Recorder.prototype.record = function(sourceId) {
  if (this.recording && this.sourceId == sourceId) {
    return true;
  }

  if (this.sourceId != sourceId) {
    this.changeSource(sourceId);
  }

  this.recording = true;

  return true;
}

Recorder.prototype.changeSource = function(sourceId) {
  if (this.sourceProcessor) {
    this.sourceProcessor.disconnect();
  }
  if (this.sourceInput) {
    this.sourceInput.disconnect();
  }
  this.prepareSource_(sourceId);
}

Recorder.prototype.prepareSource_ = function(sourceId) {
  var constraints = {audio: true};
  if (sourceId) {
    this.sourceId = sourceId;
    constraints.audio = {optional: [{sourceId: sourceId}]};
  }
  navigator.getUserMedia(constraints, this.startUserMedia_, function(e) {
    console.log('No audio input');
  });
}

Recorder.prototype.startUserMedia_ = function(stream) {
  this.sourceInput = this.audioContext.createMediaStreamSource(stream);
  this.sourceProcessor = (this.audioContext.createScriptProcessor ||
               this.audioContext.createJavaScriptNode).call(this.audioContext, 16384, 2, 2);
  this.sourceProcessor.onaudioprocess = this.newAudio_;
  this.sourceInput.connect(this.sourceProcessor);
  this.sourceProcessor.connect(this.audioContext.destination);
}

Recorder.prototype.newAudio_ = function(e) {
  if (this.recording) {
    this.worker.postMessage({
      command: 'sendBuffer',
      left: e.inputBuffer.getChannelData(0),
      right: e.inputBuffer.getChannelData(1)
    });
  }
}

Recorder.prototype.stop = function() {
  this.recording = false;
  return true;
}

$(function() {
  var $controls = $('#controls');
  var $record = $controls.find('button#record');
  var $pause = $controls.find('button#pause');
  var $audioSource = $('#audioSource');

  var room = $('#room').val();
  var recorder = new Recorder(room);

  var friendUrl = 'http://' + location.host + '/room/' + room;
  $('a#friend-url').attr('href', friendUrl).find('h4').text(friendUrl);

  $record.click(function() {
    if (recorder.record($audioSource.val())) {
      $controls.addClass('recording');
    }
  })
  $pause.click(function() {
    if (recorder.stop()) {
      $controls.removeClass('recording');
    }
  });
  $audioSource.change(function() {
    $pause.click();
    recorder.changeSource($audioSource.val());
  });

  if (typeof MediaStreamTrack === 'undefined'){
    alert('This browser does not support MediaStreamTrack.\n\nTry Chrome Canary.');
  } else {
    MediaStreamTrack.getSources(function(sourceInfos) {
      var $options = _.map(sourceInfos, function(sourceInfo, i) {
        if (sourceInfo.kind == 'audio')
          return $('<option/>', {value: sourceInfo.id, text: sourceInfo.label || 'microphone ' + i});
      });
      $audioSource.html(_.compact($options));
    });
  }
});
