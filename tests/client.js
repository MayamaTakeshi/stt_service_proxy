const WebSocket = require('ws');
const mic = require('mic');
const fs = require('fs');

// WebSocket server URL (replace with your Go server address)
const wsUrl = 'ws://192.168.88.136:9090/ws';
const ws = new WebSocket(wsUrl);

// File stream to save the audio locally for testing
const outputFile = fs.createWriteStream('test_audio.raw');

// Configuration for the mic to capture raw audio in 16-bit PCM format
const micInstance = mic({
  rate: '16000',
  channels: '1',
  bitwidth: '16',
  encoding: 'signed-integer', // L16 audio format
  fileType: 'raw',
  device: 'default',
});

// Start capturing audio from the mic
const micInputStream = micInstance.getAudioStream();

ws.on('open', () => {
  console.log('WebSocket connection established.');

  // Send command to start speech-to-text with language and timeout
  const command = JSON.stringify({
    type: 'start_speech_to_text',
    language: 'en-US',           // Set the language code
    voiceActivityTimeout: 5      // Timeout for speech detection (in seconds)
  });

  ws.send(command);
  console.log('Sent start_speech_to_text command to the server.');

  // Start streaming audio after the command
  micInputStream.on('data', (audioData) => {
    // Write audio data to the local file
    outputFile.write(audioData);

    // Send raw audio data to the WebSocket server
    ws.send(audioData, { binary: true }, (err) => {
      if (err) {
        console.error('Error sending audio data to server:', err);
      }
    });
  });

  micInstance.start(); // Start the mic recording
});

ws.on('message', (data) => {
  const result = JSON.parse(data);
  if (result.transcript) {
    console.log(`Transcript: ${result.transcript}`);
  } else if (result.interim_transcript) {
    console.log(`Interim Transcript: ${result.interim_transcript}`);
  }
});

ws.on('error', (error) => {
  console.error('WebSocket error:', error);
});

micInputStream.on('error', (err) => {
  console.error('Error in Audio Input Stream:', err);
});

ws.on('close', () => {
  console.log('WebSocket connection closed.');
  micInstance.stop(); // Stop the mic when the WebSocket connection closes
  outputFile.end(); // Close the audio file
});


