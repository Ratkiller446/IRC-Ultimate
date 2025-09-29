package main

import (
        "bufio"
        "context"
        "flag"
        "fmt"
        "log"
        "math/rand"
        "os"
        "os/signal"
        "os/user"
        "strings"
        "sync"
        "time"

        "irc-client/asciiart"
        "irc-client/asciiart/art"
        "irc-client/commands"
        "irc-client/conn"
        "irc-client/parser"
)

func prompt(label, def string) string {
        fmt.Printf("%s [%s]: ", label, def)
        scanner := bufio.NewScanner(os.Stdin)
        scanner.Scan()
        input := strings.TrimSpace(scanner.Text())
        if input == "" {
                return def
        }
        return input
}

// displayKawaiiArt displays a random kawaii art with a message
func displayKawaiiArt(message string) {
        // Seed the random number generator
        rand.Seed(time.Now().UnixNano())

        // Create a collection for our art
        kawaiiCollection := asciiart.NewCollection("Kawaii IRC")

        // Add some cat art
        for i, cat := range art.CatFaces {
                kawaiiCollection.AddArt(fmt.Sprintf("cat_%d", i), cat)
        }

        // Add some kawaii faces
        for i, face := range art.KawaiiFaces {
                kawaiiCollection.AddArt(fmt.Sprintf("face_%d", i), face)
        }

        // Get terminal width
        width := asciiart.TerminalWidth()

        // Display the message in a box
        fmt.Println(asciiart.CenterText("⋆⋅☆⋅⋆ "+message+" ⋆⋅☆⋅⋆", width))

        // Get and display a random piece of art
        art, err := kawaiiCollection.GetRandom()
        if err == nil {
                fmt.Println(asciiart.CenterText(art.Content, width))
        }
}

var outputMutex sync.Mutex

// safePrintf is a thread-safe wrapper around fmt.Printf
func safePrintf(format string, a ...interface{}) {
        outputMutex.Lock()
        defer outputMutex.Unlock()
        fmt.Printf(format, a...)
}

// safePrintln is a thread-safe wrapper around fmt.Println
func safePrintln(a ...interface{}) {
        outputMutex.Lock()
        defer outputMutex.Unlock()
        fmt.Println(a...)
}

func main() {
        // Display welcome message with kawaii art
        displayKawaiiArt("Welcome to Kawaii IRC! (◕‿◕✿)")

        // Defaults
        defaultPort := 6697
        defaultTLS := true
        defaultNick := ""
        if u, err := user.Current(); err == nil {
                defaultNick = u.Username
        }
        nick := flag.String("nick", "", "Nickname")
        server := flag.String("server", os.Getenv("IRC_SERVER"), "IRC server address")
        port := flag.Int("port", defaultPort, "IRC server port")
        tls := flag.Bool("tls", defaultTLS, "Enable TLS (SSL)")
        insecure := flag.Bool("insecure", false, "Disable TLS certificate verification (not recommended)")
        verbose := flag.Bool("verbose", false, "Enable verbose logging to stderr")
        flag.Parse()

        *nick = prompt("Nickname", defaultNick)
        if *server == "" {
                *server = prompt("Server", "irc.libera.chat")
        }

        logger := log.New(os.Stderr, "[LOG] ", log.LstdFlags)
        if *verbose {
                logger.Printf("Connecting to %s:%d as %s (TLS: %v)", *server, *port, *nick, *tls)
        }
        cfg := conn.Config{
                Server:   *server,
                Port:     *port,
                TLS:      *tls,
                Timeout:  10 * time.Second,
                Insecure: *insecure,
        }
        c, err := conn.Connect(cfg)
        if err != nil {
                logger.Printf("Connection error: %v", err)
                fmt.Fprintf(os.Stderr, "Connection error: %v\n", err)
                os.Exit(1)
        }
        defer c.Close()
        if *verbose {
                logger.Println("Connected!")
        } else {
                displayKawaiiArt("Connected to server! (ﾉ◕ヮ◕)ﾉ*:･ﾟ✧")
        }
        // Send NICK and USER immediately after connecting
        fmt.Fprintf(c, "NICK %s\r\n", *nick)
        fmt.Fprintf(c, "USER %s 0 * :%s\r\n", *nick, *nick)

        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()

        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, os.Interrupt)
        go func() {
                <-sigCh
                fmt.Fprintln(os.Stderr, "\nReceived interrupt, shutting down...")
                if *verbose {
                        logger.Println("Interrupt received, shutting down...")
                }
                cancel()
        }()

        userInput := make(chan string)
        go func() {
                scanner := bufio.NewScanner(os.Stdin)
                for scanner.Scan() {
                        userInput <- scanner.Text()
                }
                if err := scanner.Err(); err != nil {
                        fmt.Fprintf(os.Stderr, "[ERROR] Stdin read: %v\n", err)
                }
                close(userInput)
        }()

        serverDone := make(chan struct{})
        go func() {
                scanner := bufio.NewScanner(c)
                for scanner.Scan() {
                        line := scanner.Text()
                        msg := parser.ParseMessage(line)
                        if msg.Command == "PING" && len(msg.Params) > 0 {
                                fmt.Fprintf(c, "PONG :%s\r\n", msg.Params[0])
                                if *verbose {
                                        safePrintf("Replied to PING with PONG :%s\n", msg.Params[0])
                                }
                                continue
                        }
                        timestamp := time.Now().Format("2006-01-02 15:04:05")
                        source := msg.Prefix
                        if source == "" {
                                source = "SERVER"
                        }
                        if msg.Command == "PRIVMSG" || msg.Command == "NOTICE" {
                                if len(msg.Params) >= 2 {
                                        safePrintf("[%s] %s %s\n", timestamp, source, msg.Params[1])
                                }
                        } else {
                                safePrintf("[%s] %s %s %v\n", timestamp, source, msg.Command, msg.Params)
                        }
                }
                if err := scanner.Err(); err != nil {
                        fmt.Fprintf(os.Stderr, "[ERROR] Server read: %v\n", err)
                }
                close(serverDone)
        }()

        writer := bufio.NewWriter(c)
        currentChannel := ""
        
        for {
                select {
                case <-ctx.Done():
                        return
                case line, ok := <-userInput:
                        if !ok {
                                return
                        }
                        
                        // Create current client state
                        clientState := commands.ClientState{
                                CurrentChannel: currentChannel,
                                Nick:          *nick,
                        }
                        
                        // Build IRC commands using extracted function
                        ircCommands, err := commands.BuildIRCCommand(clientState, line)
                        if err != nil {
                                fmt.Fprintf(os.Stderr, "[ERROR] Build command: %v\n", err)
                                continue
                        }
                        
                        // Process each IRC command
                        for _, ircCmd := range ircCommands {
                                // Handle special cases for state management
                                if strings.HasPrefix(ircCmd, "JOIN ") {
                                        // Extract channel name for state tracking
                                        parts := strings.Fields(ircCmd)
                                        if len(parts) >= 2 {
                                                currentChannel = parts[1]
                                        }
                                } else if strings.HasPrefix(ircCmd, "PART ") {
                                        // Clear current channel on part
                                        currentChannel = ""
                                } else if strings.HasPrefix(ircCmd, "PRIVMSG ") && !strings.HasPrefix(ircCmd, "PRIVMSG "+currentChannel) {
                                        // For private messages, display locally
                                        parts := strings.SplitN(ircCmd, " :", 2)
                                        if len(parts) >= 2 {
                                                timestamp := time.Now().Format("2006-01-02 15:04:05")
                                                safePrintf("[%s] %s %s\n", timestamp, *nick, parts[1])
                                        }
                                } else if strings.HasPrefix(ircCmd, "PRIVMSG "+currentChannel) {
                                        // For channel messages, display locally
                                        parts := strings.SplitN(ircCmd, " :", 2)
                                        if len(parts) >= 2 {
                                                timestamp := time.Now().Format("2006-01-02 15:04:05")
                                                safePrintf("[%s] %s %s\n", timestamp, *nick, parts[1])
                                        }
                                }
                                
                                // Send command to server
                                if _, err := writer.WriteString(ircCmd + "\r\n"); err != nil {
                                        fmt.Fprintf(os.Stderr, "[ERROR] Write %s: %v\n", strings.Fields(ircCmd)[0], err)
                                        continue
                                }
                                if err := writer.Flush(); err != nil {
                                        fmt.Fprintf(os.Stderr, "[ERROR] Flush %s: %v\n", strings.Fields(ircCmd)[0], err)
                                        continue
                                }
                                
                                // Handle QUIT command by returning
                                if strings.HasPrefix(ircCmd, "QUIT") {
                                        return
                                }
                        }
                        
                case <-serverDone:
                        return
                }
        }
}
