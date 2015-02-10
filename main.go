package main

import (
    "flag"
    "github.com/go-fsnotify/fsnotify"
    "log"
    "os"
    "os/exec"
    "strings"
    "time"
)

var (
    directory         = flag.String("d", ".", "This is the directory you want to scan for changes")
    build_command     = flag.String("build", "", "This is your application's build command")
    program_to_launch = flag.String("webserver", "", "This is the go webserver that you plan on launching")
    pid               = 0
)

func main() {
    watcher, err := fsnotify.NewWatcher()

    if err != nil {
        log.Fatal(err)
    }

    defer watcher.Close()

    flag.Parse()
    if *program_to_launch == "" || *build_command == "" {
        flag.Usage()
        os.Exit(-1)
    }

    done := make(chan bool)

    go func() {
        for {
            select {
            case event := <-watcher.Events:
                if event.Op&fsnotify.Write == fsnotify.Write {
                    log.Println("modified file:", event.Name)
                    if strings.HasSuffix(event.Name, ".html") {
                        go killGoProgram()
                        time.Sleep(time.Millisecond * 500)
                    }
                    if strings.HasSuffix(event.Name, ".go") {
                        rebuildGoProgram()
                    }
                }
            case err := <-watcher.Errors:
                log.Println("error:", err)
            }
        }
    }()

    err = watcher.Add(*directory)

    if err != nil {
        log.Fatal(err)
    }
    <-done
}

func rebuildGoProgram() {
    launchGoProgram()

}

func killGoProgram() {
    cmd := exec.Command("./dev_server")
    err := cmd.Start()
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Waiting for command to finish...")
    err = cmd.Wait()
    log.Printf("Command finished with error: %v", err)
    go launchGoProgram()

}

func launchGoProgram() {
    if pid != 0 {
        process, _ := os.FindProcess(pid)
        process.Signal(os.Interrupt) // This should kill the process
    }

    cmd := exec.Command(*program_to_launch)

    if err := cmd.Start(); err != nil {
        log.Fatal(err)
    }
    pid = cmd.Process.Pid

}
