package main

import (
    "context"
    "fmt"
    "io"
    "log"

    speech "cloud.google.com/go/speech/apiv1"
    "github.com/gordonklaus/portaudio"
    speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

const sampleRate = 16000 // 16kHz is a good rate for speech recognition

func main() {
    // Initialize PortAudio for microphone input
    err := portaudio.Initialize()
    if err != nil {
        log.Fatalf("Error initializing PortAudio: %v", err)
    }
    defer portaudio.Terminate()

    // Set up Google Cloud Speech-to-Text client
    ctx := context.Background()
    client, err := speech.NewClient(ctx)
    if err != nil {
        log.Fatalf("Error creating speech client: %v", err)
    }

    stream, err := client.StreamingRecognize(ctx)
    if err != nil {
        log.Fatalf("Error creating streaming client: %v", err)
    }

    go func() {
        for {
            resp, err := stream.Recv()
            if err == io.EOF {
                return
            }
            if err != nil {
                log.Fatalf("Error receiving recognition response: %v", err)
            }
            for _, result := range resp.Results {
                if result.IsFinal {
                    for _, alt := range result.Alternatives {
                        fmt.Printf("Transcription: %v\n", alt.Transcript)
                    }
                }
            }
        }
    }()

    // Configure audio stream format
    in := make([]int16, sampleRate/10) // Buffer for 100ms of audio (1600 samples)
    streamParams := &speechpb.StreamingRecognitionConfig{
        Config: &speechpb.RecognitionConfig{
            Encoding:        speechpb.RecognitionConfig_LINEAR16,
            SampleRateHertz: sampleRate,
            LanguageCode:    "en-US",
        },
        InterimResults: true,
    }

    if err := stream.Send(&speechpb.StreamingRecognizeRequest{
        StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
            StreamingConfig: streamParams,
        },
    }); err != nil {
        log.Fatalf("Error sending config: %v", err)
    }

    // Start microphone input stream
    audioStream, err := portaudio.OpenDefaultStream(1, 0, float64(sampleRate), len(in), &in)
    if err != nil {
        log.Fatalf("Error opening PortAudio stream: %v", err)
    }
    defer audioStream.Close()

    err = audioStream.Start()
    if err != nil {
        log.Fatalf("Error starting audio stream: %v", err)
    }

    // Stream audio to Google Cloud Speech-to-Text
    for {
        err = audioStream.Read()
        if err != nil {
            log.Fatalf("Error reading audio: %v", err)
        }

        if err := stream.Send(&speechpb.StreamingRecognizeRequest{
            StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
                AudioContent: convertToBytes(in),
            },
        }); err != nil {
            log.Fatalf("Error sending audio: %v", err)
        }
    }
}

func convertToBytes(input []int16) []byte {
    buf := make([]byte, len(input)*2)
    for i, v := range input {
        buf[2*i] = byte(v)
        buf[2*i+1] = byte(v >> 8)
    }
    return buf
}

