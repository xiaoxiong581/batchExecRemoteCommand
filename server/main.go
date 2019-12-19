package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "github.com/xiaoxiong581/batchExecRemoteCommand/server/domain"
    "golang.org/x/crypto/ssh"
    "io"
    "io/ioutil"
    "net"
    "os"
    "path/filepath"
    "strings"
)

func main() {
    rootPath, err := os.Getwd()
    if err != nil {
        fmt.Printf("error get root path error, err: %s\n", err.Error())
        return
    }

    commandPath := strings.Join([]string{rootPath, "command", "command.json"}, string(filepath.Separator))
    data, err := ioutil.ReadFile(commandPath)
    if err != nil {
        fmt.Printf("read command file error, err: %s\n", err.Error())
        return
    }

    commandBodys := &[]domain.CommandBody{}
    err = json.Unmarshal(data, commandBodys)
    if err != nil {
        fmt.Printf("error deserialize command file error, err: %s\n", err.Error())
        return
    }

    for _, commandBody := range *commandBodys {
        execute(commandBody)
    }
}

func execute(commandBody domain.CommandBody) {
    host := commandBody.Host
    user := commandBody.User
    password := commandBody.Password
    commands := commandBody.Commands

    if host == "" || len(commands) == 0 {
        fmt.Printf("info param is invalid, command: %+v\n", commandBody)
        return
    }

    fmt.Printf("begin to process, host: %s\n", host)
    config := &ssh.ClientConfig{
        User: user,
        Auth: []ssh.AuthMethod{
            ssh.Password(password),
        },
        HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
    }
    client, err := ssh.Dial("tcp", host, config)
    if err != nil {
        fmt.Printf("error create ssh client failed, host: %s, err: %s\n", host, err.Error())
        return
    }
    defer client.Close()
    session, err := client.NewSession()
    if err != nil {
        fmt.Printf("error create ssh session failed, host: %s, err: %s\n", host, err.Error())
        return
    }
    defer session.Close()

    modes := ssh.TerminalModes{
        ssh.ECHO:          0,     // disable echoing
        ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
        ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
    }
    err = session.RequestPty("xterm", 80, 40, modes)
    if err != nil {
        fmt.Printf("ssh session request failed, host: %s, err: %s\n", host, err.Error())
        return
    }

    output := new(bytes.Buffer)
    session.Stdout = output
    input, _ := session.StdinPipe()

    go func(input io.Writer, output *bytes.Buffer) {
        for {
            if strings.Contains(string(output.Bytes()), "[sudo] password for ") {
                _, err = input.Write([]byte(password + "\n"))
                if err != nil {
                    fmt.Printf("error password is error, host: %s, password: %s, err: %s\n", host, password, err.Error())
                    break
                }
                fmt.Printf("put password, host: %s, password: %s\n", host, password)
                break
            }
        }
    }(input, output)

    if err = session.Run(strings.Join(commands, "; ")); err != nil {
        fmt.Printf("execute command error, host: %s, command: %s, err: %s\n", host, commands, err.Error())
        return
    }
    fmt.Printf("host: %s, command: %s, execute result:\n %s\n\n", host, commands, output.String())
}
