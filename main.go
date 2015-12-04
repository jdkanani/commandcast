package main

import (
	"bytes"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// Host name color
var hostColor = color.New(color.FgCyan).SprintFunc()

func ReadHostsFromFile(file string) []string {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	var results []string
	for _, host := range strings.Split(strings.Trim(string(buffer), "\n"), "\n") {
		if strings.TrimSpace(host) != "" {
			results = append(results, strings.TrimSpace(host))
		}
	}
	return results
}

func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func GetAuthKeys(keys []string) []ssh.AuthMethod {
	methods := []ssh.AuthMethod{}

	for _, keyname := range keys {
		pkey := PublicKeyFile(keyname)
		if pkey != nil {
			methods = append(methods, pkey)
		}
	}

	return methods
}

func ExecuteCmd(cmd string, hostname string, config *ssh.ClientConfig) string {
	conn, err := ssh.Dial("tcp", hostname+":22", config)
	if err != nil {
		return hostColor(hostname+": ") + "\n" + err.Error()
	}

	session, err := conn.NewSession()
	if err != nil {
		return hostColor(hostname+": ") + "\n" + err.Error()
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Run(cmd)

	return hostColor(hostname+":") + " \n" + stdoutBuf.String()
}

func main() {
	app := cli.NewApp()
	app.Name = "commandcast"
	app.Usage = "Run command on multiple hosts over SSH"
	app.Version = "1.0.0"
	app.Author = "Jaynti Kanani"
	app.Email = "jdkanani@gmail.com"

	var hostString, user, keyString, hostFileName string
	var to int
	app.Commands = []cli.Command{
		{
			Name:    "exec",
			Aliases: []string{"e"},
			Usage:   "Execute command to all hosts",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "hosts",
					Value:       "localhost",
					Usage:       "Multiple hosts (comma separated)",
					Destination: &hostString,
				},
				cli.StringFlag{
					Name:        "hostfile",
					Usage:       "File containing host names",
					Destination: &hostFileName,
				},
				cli.StringFlag{
					Name:        "user",
					Usage:       "SSH auth user",
					EnvVar:      "USER",
					Destination: &user,
				},
				cli.IntFlag{
					Name:        "timeout",
					Usage:       "SSH timeout (seconds)",
					Value:       15,
					Destination: &to,
				},
				cli.StringFlag{
					Name: "keys",
					Value: strings.Join([]string{
						os.Getenv("HOME") + "/.ssh/id_dsa",
						os.Getenv("HOME") + "/.ssh/id_rsa",
					}, ","),
					Usage:       "SSH auth keys (comma separated)",
					Destination: &keyString,
				},
			},
			Action: func(c *cli.Context) {
				cmd := c.Args().First()
				keys := strings.Split(keyString, ",")

				var hosts []string = nil
				if hostFileName != "" {
					hosts = ReadHostsFromFile(hostFileName)
				}

				if hosts == nil && hostString != "" {
					hosts = strings.Split(hostString, ",")
				}

				fmt.Printf("%s %s\n", color.GreenString("Command: "), cmd)
				fmt.Printf("%s %s\n", color.GreenString("Hosts: "), hosts)
				fmt.Printf("%s %s\n", color.GreenString("User: "), user)
				fmt.Printf("%s %s\n", color.GreenString("Keys: "), keys)
				fmt.Printf("%s \n\n", color.GreenString("Output: "))

				results := make(chan string, 10)
				timeout := time.After(time.Duration(to) * time.Second)

				authKeys := GetAuthKeys(keys)
				if len(authKeys) < 1 {
					color.Red("Key(s) doesn't exist.")
					return
				}

				config := &ssh.ClientConfig{
					User: user,
					Auth: authKeys,
				}

				// Execute command on hosts
				for _, hostname := range hosts {
					go func(hostname string) {
						results <- ExecuteCmd(cmd, hostname, config)
					}(hostname)
				}

				for i := 0; i < len(hosts); i++ {
					select {
					case res := <-results:
						fmt.Println(res)
					case <-timeout:
						color.Red("Timed out!")
						return
					}
				}
			},
		},
	}

	// Run app
	app.Run(os.Args)
}
