package main

import (
	"bytes"
	"golang.org/x/crypto/ssh"
	"strings"
)

// Host configure
type HostConfig struct {
	Host         string
	User         string
	Timeout      int
	ClientConfig *ssh.ClientConfig
	Session      *ssh.Session
}

// Start SSH session
func (this *HostConfig) StartSession() (*ssh.Session, error) {
	host := this.Host
	if !strings.ContainsAny(this.Host, ":") {
		host = host + ":22"
	}
	conn, err := ssh.Dial("tcp", host, this.ClientConfig)
	if err != nil {
		return nil, err
	}

	this.Session, err = conn.NewSession()
	if err != nil {
		return nil, err
	}
	return this.Session, err
}

// Stop SSH session
func (this *HostConfig) StopSession() {
	if this.Session != nil {
		this.Session.Close()
	}
}

// Execute command
func (this *HostConfig) ExecuteCmd(cmd string) (string, error) {
	if this.Session == nil {
		if _, err := this.StartSession(); err != nil {
			return "", err
		}
	}

	var stdoutBuf bytes.Buffer
	this.Session.Stdout = &stdoutBuf
	this.Session.Run(cmd)

	return stdoutBuf.String(), nil
}

// To string
func (this HostConfig) String() string {
	return this.User + "@" + this.Host
}
