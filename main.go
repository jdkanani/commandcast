package main

import (
	"bufio"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"
)

// Colors
var labelColor = color.New(color.FgMagenta).Add(color.Bold).SprintFunc()
var hostColor = color.New(color.FgCyan).SprintFunc()

// Clean command - trim space and new line
func CleanText(cmd string) string {
	return strings.TrimSpace(strings.Trim(cmd, "\n"))
}

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

func GetAuthPassword(password string) []ssh.AuthMethod {
	return []ssh.AuthMethod{ssh.Password(password)}
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

func Execute(cmd string, hosts []HostConfig, to int) {
	// Run parallel ssh session (max 10)
	results := make(chan string, 10)
	timeout := time.After(time.Duration(to) * time.Second)

	// Execute command on hosts
	for _, host := range hosts {
		go func(host HostConfig) {
			var result string

			if text, err := host.ExecuteCmd(cmd); err != nil {
				result = err.Error()
			} else {
				result = text
			}
			results <- fmt.Sprintf("%s > %s\n%s\n", hostColor(host), cmd, result)
		}(host)
	}

	for i := 0; i < len(hosts); i++ {
		select {
		case res := <-results:
			if res != "" {
				fmt.Println(res)
			}
		case <-timeout:
			color.Red("Timed out!")
			return
		}
	}
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
	var interactive bool = false
	app.Commands = []cli.Command{
		{
			Name:    "exec",
			Aliases: []string{"e"},
			Usage:   "Execute command to all hosts",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "interactive, i",
					Usage:       "Enable intereactive mode",
					Destination: &interactive,
				},
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
					Name:        "user, u",
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
				keys := strings.Split(keyString, ",")

				var hosts []string = nil
				if hostFileName != "" {
					hosts = ReadHostsFromFile(hostFileName)
				}

				if hosts == nil && hostString != "" {
					hosts = strings.Split(hostString, ",")
				}

				authKeys := GetAuthKeys(keys)
				if len(authKeys) < 1 {
					color.Red("Key(s) doesn't exist.")
					return
				}

				hostConfigs := make([]HostConfig, len(hosts))
				for i, hostName := range hosts {
					urlInfo, err := url.Parse("ssh://" + hostName)
					if err != nil || urlInfo.Host == "" {
						continue
					}

					// client config
					username := user
					keys := authKeys
					if urlInfo.User != nil {
						if urlInfo.User.Username() != "" {
							username = urlInfo.User.Username()
						}
						if password, ok := urlInfo.User.Password(); ok {
							keys = GetAuthPassword(password)
						}
					}

					// create new host config
					hostConfigs[i] = HostConfig{
						User:         username,
						Host:         urlInfo.Host,
						Timeout:      to,
						ClientConfig: &ssh.ClientConfig{User: username, Auth: keys},
					}
				}

				// Print host configs and keys
				fmt.Printf("%s %s\n", labelColor("Keys: "), keys)
				fmt.Printf("%s %+v\n", labelColor("Hosts: "), hostConfigs)

				// single command mode
				if !interactive {
					cmd := CleanText(c.Args().First())
					if cmd != "" {
						fmt.Printf(">>> %s\n", cmd)
						Execute(cmd, hostConfigs, to)
					}
				}

				// Interactive mode
				if interactive {
					for {
						reader := bufio.NewReader(os.Stdin)
						fmt.Print(">>> ")
						cmd, _ := reader.ReadString('\n')
						cmd = CleanText(cmd)

						if cmd == "exit" {
							break
						}

						if cmd != "" {
							Execute(cmd, hostConfigs, to)
						}
					}
				}

				// Stop host session
				for _, hostConfig := range hostConfigs {
					hostConfig.StopSession()
				}
			},
		},
	}

	// Run app
	app.Run(os.Args)
}
