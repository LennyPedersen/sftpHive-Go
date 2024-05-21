package sftp

import (
	"strconv"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func NewSFTPClient(username, password, server string, port int) (*sftp.Client, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", server+":"+strconv.Itoa(port), config)
	if err != nil {
		return nil, err
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		return nil, err
	}

	return client, nil
}
