package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"

    "github.com/gorilla/websocket"
    speech "cloud.google.com/go/speech/apiv1"
    speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func speechToText(conn *websocket.Conn, lang string, voiceTimeout int) error {
    ctx := context.Background()
    client, err := speech.NewClient(ctx)
    if err != nil {
        return fmt.Errorf("failed to create speech client: %v", err)
    }
    defer client.Close()

    stream, err := client.StreamingRecognize(ctx)
    if err != nil {
        return fmt.Errorf("failed to create stream: %v", err)
    }
    defer stream.CloseSend()

    // Send the initial request with configuration for the stream
    config := &speechpb.RecognitionConfig{
        Encoding:        speechpb.RecognitionConfig_LINEAR16,
        SampleRateHertz: 16000,
        LanguageCode:    lang,
        EnableAutomaticPunctuation: true,
    }
    streamingConfig := &speechpb.StreamingRecognitionConfig{
        Config:         config,
        InterimResults: true,
        SingleUtterance: false,
    }

    if err := stream.Send(&speechpb.StreamingRecognizeRequest{
        StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
            StreamingConfig: streamingConfig,
        },
    }); err != nil {
        return fmt.Errorf("failed to send config request: %v", err)
    }

    // Handle incoming audio and recognition responses
    go func() {
        for {
            resp, err := stream.Recv()
            if err == io.EOF {
                break
            }
            if err != nil {
                log.Printf("Error receiving response: %v", err)
                break
            }
            if resp.Error != nil {
                log.Printf("Speech API error: %v", resp.Error)
                continue
            }
            for _, result := range resp.Results {
                if result.IsFinal {
                    log.Printf("Final transcript: %s\n", result.Alternatives[0].Transcript)
                    conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"transcript": "%s"}`, result.Alternatives[0].Transcript)))
                } else {
                    log.Printf("Interim transcript: %s\n", result.Alternatives[0].Transcript)
                    conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"interim_transcript": "%s"}`, result.Alternatives[0].Transcript)))
                }
            }
        }
    }()

    // Read and forward the binary audio data to Google Speech API
    for {
        messageType, audioData, err := conn.ReadMessage()
        if err != nil {
            log.Println("Error reading WebSocket message:", err)
            return err
        }

        // Handle text (commands) vs binary (audio data)
        if messageType == websocket.TextMessage {
            log.Println("Received text message when expecting audio, ignoring.")
            continue
        }

        fmt.Println(audioData)
        if messageType == websocket.BinaryMessage {
            if err := stream.Send(&speechpb.StreamingRecognizeRequest{
                StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
                    AudioContent: audioData,
                },
            }); err != nil {
                return fmt.Errorf("failed to send audio content: %v", err)
            }
        }
    }
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Error upgrading to WebSocket:", err)
        return
    }
    defer conn.Close()

    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            log.Println("Error reading message:", err)
            break
        }

        if messageType == websocket.TextMessage {
            // Expecting JSON command, attempt to parse it
            var command map[string]interface{}
            if err := json.Unmarshal(message, &command); err != nil {
                log.Println("Error parsing JSON:", err)
                continue
            }

            if command["type"] == "start_speech_to_text" {
                language, _ := command["language"].(string)
                voiceActivityTimeout, _ := command["voiceActivityTimeout"].(float64)

                log.Printf("Starting speech-to-text with language: %s, timeout: %.0f seconds", language, voiceActivityTimeout)

                // Start speech-to-text with the given language and timeout
                err := speechToText(conn, language, int(voiceActivityTimeout))
                if err != nil {
                    log.Println("Error during speech-to-text:", err)
                }
            }
        } else {
            log.Println("Received non-text message when expecting command, ignoring.")
        }
    }
}

func main() {
    http.HandleFunc("/ws", wsHandler)
    listen := "192.168.88.136:9090"
    log.Printf("Server started on %s", listen)
    log.Fatal(http.ListenAndServe(listen, nil))
}

