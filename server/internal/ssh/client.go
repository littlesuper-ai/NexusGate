package ssh

import (
	"bytes"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	Host     string
	Port     int
	User     string
	Password string
}

func NewClient(host string, port int, user, password string) *Client {
	return &Client{Host: host, Port: port, User: user, Password: password}
}

func (c *Client) Run(command string) (string, error) {
	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", fmt.Errorf("ssh dial failed: %w", err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("ssh session failed: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(command); err != nil {
		return "", fmt.Errorf("command failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// RunUCI executes a UCI command on an OpenWrt device.
func (c *Client) RunUCI(uciCmd string) (string, error) {
	return c.Run(fmt.Sprintf("uci %s", uciCmd))
}

// GetSystemInfo retrieves basic system info from an OpenWrt device.
func (c *Client) GetSystemInfo() (string, error) {
	cmd := `echo "{\"hostname\":\"$(uci get system.@system[0].hostname)\",\"model\":\"$(cat /tmp/sysinfo/model 2>/dev/null)\",\"firmware\":\"$(cat /etc/openwrt_release | grep DISTRIB_REVISION | cut -d\\' -f2)\",\"uptime\":$(cat /proc/uptime | cut -d' ' -f1 | cut -d. -f1),\"kernel\":\"$(uname -r)\"}"`
	return c.Run(cmd)
}
