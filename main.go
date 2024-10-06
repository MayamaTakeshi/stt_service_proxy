package main

import (
    "context"
    "flag"
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"

    "encoding/base64"
    "net"
    "net/http"
    "net/url"

    "google.golang.org/api/option"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    speech "cloud.google.com/go/speech/apiv1"
    "cloud.google.com/go/speech/apiv1/speechpb"

    "google.golang.org/grpc/grpclog"

)

// Helper function for basic authentication
func basicAuth(username, password string) string {
    return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

// Custom dialer that routes gRPC through an HTTP proxy
func dialerWithProxy(ctx context.Context, addr string) (net.Conn, error) {
    proxyURL, err := http.ProxyFromEnvironment(&http.Request{
        URL: &url.URL{
            Scheme: "https",
            Host:   addr,
        },
    })
    if err != nil {
        return nil, err
    }

    if proxyURL == nil {
        // No proxy, dial directly
        return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
    }

    // Dial through proxy
    conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", proxyURL.Host)
    if err != nil {
        return nil, err
    }

    // Send a CONNECT request to the proxy to establish a tunnel to the gRPC server
    connectReq := &http.Request{
        Method: "CONNECT",
        URL:    &url.URL{Host: addr},
        Host:   addr,
        Header: make(http.Header),
    }
    if proxyAuth := proxyURL.User; proxyAuth != nil {
        password, _ := proxyAuth.Password()
        connectReq.Header.Set("Proxy-Authorization", "Basic "+basicAuth(proxyAuth.Username(), password))
    }

    connectReq.Write(conn)
    return conn, nil
}

func main() {
    grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stdout, os.Stderr, os.Stderr))

    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s <AUDIOFILE>\n", filepath.Base(os.Args[0]))
        fmt.Fprintf(os.Stderr, "<AUDIOFILE> must be a path to a local audio file. Audio file must be a 16-bit signed little-endian encoded with a sample rate of 16000.\n")

    }
    flag.Parse()
    if len(flag.Args()) != 1 {
        log.Fatal("Please pass path to your local audio file as a command line argument")
    }
    audioFile := flag.Arg(0)

    ctx := context.Background()

    // Set up gRPC connection with custom dialer
    conn, err := grpc.DialContext(
        ctx,
        "speech.googleapis.com:443",
        grpc.WithContextDialer(dialerWithProxy),
        grpc.WithTransportCredentials(insecure.NewCredentials()), // Use `insecure.NewCredentials()` for plaintext. Replace with `credentials.NewTLS()` if TLS is needed.
    )
    if err != nil {
        log.Fatalf("Failed to create gRPC connection: %v", err)
    }
    defer conn.Close()

    client, err := speech.NewClient(ctx, option.WithGRPCConn(conn))
    if err != nil {
        log.Fatalf("Failed to create Speech client: %v", err)
    }
    defer client.Close()

    log.Println("creating stream")
    stream, err := client.StreamingRecognize(ctx)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("stream created")
    // Send the initial configuration message.
    if err := stream.Send(&speechpb.StreamingRecognizeRequest{
        StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
            StreamingConfig: &speechpb.StreamingRecognitionConfig{
                Config: &speechpb.RecognitionConfig{
                    Encoding:        speechpb.RecognitionConfig_LINEAR16,
                    SampleRateHertz: 16000,
                    LanguageCode:    "en-US",
                },
            },
        },
    }); err != nil {
        log.Fatal(err)
    }

    f, err := os.Open(audioFile)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    go func() {
        buf := make([]byte, 1024)
        for {
            n, err := f.Read(buf)
            if n > 0 {
                if err := stream.Send(&speechpb.StreamingRecognizeRequest{
                    StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
                        AudioContent: buf[:n],
                    },
                }); err != nil {
                    log.Printf("Could not send audio: %v", err)
                }
            }
            if err == io.EOF {
                // Nothing else to pipe, close the stream.
                if err := stream.CloseSend(); err != nil {
                    log.Fatalf("Could not close stream: %v", err)
                }
                return
            }
            if err != nil {
                log.Printf("Could not read from %s: %v", audioFile, err)
                continue
            }
        }
    }()

    for {
        resp, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatalf("Cannot stream results: %v", err)
        }
        if err := resp.Error; err != nil {
            log.Fatalf("Could not recognize: %v", err)
        }
        for _, result := range resp.Results {
            fmt.Printf("Result: %+v\n", result)
        }
    }
}
