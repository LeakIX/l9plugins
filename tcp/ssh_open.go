package tcp

import (
	"context"
	"github.com/LeakIX/l9format"
	"golang.org/x/crypto/ssh"
	"net"
	"time"
)

type SSHOpenPlugin struct {
	l9format.ServicePluginBase
}

func New() l9format.ServicePluginInterface {
	return SSHOpenPlugin{}
}

func (SSHOpenPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (SSHOpenPlugin) GetProtocols() []string {
	return []string{"ssh"}
}

func (SSHOpenPlugin) GetName() string {
	return "SSHOpenPlugin"
}

func (SSHOpenPlugin) GetStage() string {
	return "open"
}

func (plugin SSHOpenPlugin) Run(ctx context.Context, event *l9format.L9Event, options map[string]string)  bool {
	conn, err := plugin.GetL9NetworkConnection(event)
	if err != nil {
		return  false
	}
	_, _, _, _ = ssh.NewClientConn(conn, net.JoinHostPort(event.Ip, event.Port), &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
			event.SSH.Fingerprint = ssh.FingerprintSHA256(key)
			return nil
		},
		BannerCallback: func(message string) error {
			event.SSH.Banner = message
			return nil
		},
		Timeout:           5*time.Second,

	})
	return false
}
