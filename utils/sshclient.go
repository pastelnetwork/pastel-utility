package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"

	"github.com/pastelnetwork/gonode/common/log"
	"github.com/pkg/errors"

	scp "github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
)

type scriptType byte

const (
	cmdLine scriptType = iota
	rawScript
	scriptFile
)

// A Client implements an SSH client that supports running commands and scripts remotely.
type Client struct {
	client *ssh.Client
}

// DialWithPasswd starts a client connection to the given SSH server with passwd authmethod.
func DialWithPasswd(addr, user, passwd string) (*Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
		Config: ssh.Config{
			KeyExchanges: []string{"diffie-hellman-group-exchange-sha256",
				"curve25519-sha256@libssh.org ecdh-sha2-nistp256",
				"ecdh-sha2-nistp384 ecdh-sha2-nistp521",
				"diffie-hellman-group14-sha1",
				"diffie-hellman-group1-sha1",
				"aes128-ctr",
				"aes192-ctr",
				"aes256-ctr",
				"arcfour256",
				"arcfour128",
				"arcfour",
			},
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	return Dial("tcp", addr, config)
}

// DialWithKey starts a client connection to the given SSH server with key authmethod.
func DialWithKey(addr, user, keyfile string) (*Client, error) {
	key, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	return Dial("tcp", addr, config)
}

// DialWithKeyWithPassphrase same as DialWithKey but with a passphrase to decrypt the private key
func DialWithKeyWithPassphrase(addr, user, keyfile string, passphrase string) (*Client, error) {
	key, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	return Dial("tcp", addr, config)
}

// Dial starts a client connection to the given SSH server.
// This wraps ssh.Dial.
func Dial(network, addr string, config *ssh.ClientConfig) (*Client, error) {
	client, err := ssh.Dial(network, addr, config)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: client,
	}, nil
}

// Close closes the underlying client network connection.
func (c *Client) Close() error {
	return c.client.Close()
}

// UnderlyingClient get the underlying client.
func (c *Client) UnderlyingClient() *ssh.Client {
	return c.client
}

// Cmd creates a RemoteScript that can run the command on the client. The cmd string is split on newlines and each line is executed separately.
func (c *Client) Cmd(cmd string) *RemoteScript {
	return &RemoteScript{
		scriptType: cmdLine,
		client:     c.client,
		script:     bytes.NewBufferString(cmd + "\n"),
	}
}

// Script creates a RemoteScript that can run the script on the client.
func (c *Client) Script(script string) *RemoteScript {
	return &RemoteScript{
		scriptType: rawScript,
		client:     c.client,
		script:     bytes.NewBufferString(script + "\n"),
	}
}

// Scp implements scp commmand to copy local file to remote host
func (c *Client) Scp(srcFile string, destFile string) error {
	// Connect to the remote server
	scpClient, err := scp.NewClientBySSH(c.client)
	if err != nil {
		return errors.Errorf("failed to create scp client: %v", err)
	}

	err = scpClient.Connect()
	if err != nil {
		return errors.Errorf("failed to connect to scp remote: %v", err)
	}

	// Close client connection after the file has been copied
	defer scpClient.Close()

	// Open a file
	f, err := os.Open(srcFile)
	if err != nil {
		return errors.Errorf("failed to read %s file: %v", srcFile, err)
	}
	defer f.Close()

	// Close the file after it has been copied

	// Finaly, copy the file over
	// Usage: CopyFile(fileReader, remotePath, permission)
	err = scpClient.CopyFile(f, destFile, "0777")

	if err != nil {
		return errors.Errorf("failed to transfer file: %v", err)
	}

	return nil
}

// ScriptFile creates a RemoteScript that can read a local script file and run it remotely on the client.
func (c *Client) ScriptFile(fname string) *RemoteScript {
	return &RemoteScript{
		scriptType: scriptFile,
		client:     c.client,
		scriptFile: fname,
	}
}

// Shell create a noninteractive shell on client.
func (c *Client) Shell() *RemoteShell {
	return &RemoteShell{
		client:     c.client,
		requestPty: false,
	}
}

// ShellCmd executes a remote command, and also print log of it
func (c *Client) ShellCmd(ctx context.Context, cmd string) error {
	log.WithContext(ctx).Infof("Remote Command: %s started", cmd)
	defer log.WithContext(ctx).Infof("Remote Command: %s finished", cmd)

	stdin := bytes.NewBufferString(cmd)
	var stdout, stderr io.Writer

	return c.Shell().SetStdio(stdin, stdout, stderr).Start()
}

// A RemoteScript represents script that can be run remotely.
type RemoteScript struct {
	client     *ssh.Client
	scriptType scriptType
	script     *bytes.Buffer
	scriptFile string
	err        error

	stdout io.Writer
	stderr io.Writer
}

// Run runs the script on the client.
//
// The returned error is nil if the command runs, has no problems
// copying stdin, stdout, and stderr, and exits with a zero exit
// status.
func (rs *RemoteScript) Run() error {
	if rs.err != nil {
		fmt.Println(rs.err)
		return rs.err
	}

	if rs.scriptType == cmdLine {
		return rs.runCmds()
	} else if rs.scriptType == rawScript {
		return rs.runScript()
	} else if rs.scriptType == scriptFile {
		return rs.runScriptFile()
	} else {
		return errors.New("not supported RemoteScript type")
	}
}

// Output runs the script on the client and returns its standard output.
func (rs *RemoteScript) Output() ([]byte, error) {
	if rs.stdout != nil {
		return nil, errors.New("stdout already set")
	}
	var out bytes.Buffer
	rs.stdout = &out
	err := rs.Run()
	return out.Bytes(), err
}

// SmartOutput runs the script on the client. On success, its standard ouput is returned. On error, its standard error is returned.
func (rs *RemoteScript) SmartOutput() ([]byte, error) {
	if rs.stdout != nil {
		return nil, errors.New("stdout already set")
	}
	if rs.stderr != nil {
		return nil, errors.New("stderr already set")
	}

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	rs.stdout = &stdout
	rs.stderr = &stderr
	err := rs.Run()
	if err != nil {
		return stderr.Bytes(), err
	}
	return stdout.Bytes(), err
}

// Cmd appends a command to the RemoteScript.
func (rs *RemoteScript) Cmd(cmd string) *RemoteScript {
	if _, err := rs.script.WriteString(cmd + "\n"); err != nil {
		rs.err = err
	}

	return rs
}

// SetStdio specifies where its standard output and error data will be written.
func (rs *RemoteScript) SetStdio(stdout, stderr io.Writer) *RemoteScript {
	rs.stdout = stdout
	rs.stderr = stderr
	return rs
}

func (rs *RemoteScript) runCmd(cmd string) error {
	session, err := rs.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = rs.stdout
	session.Stderr = rs.stderr

	return session.Run(cmd)
}

func (rs *RemoteScript) runCmds() error {
	for {
		statment, err := rs.script.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := rs.runCmd(statment); err != nil {
			return err
		}
	}

	return nil
}

func (rs *RemoteScript) runScript() error {
	session, err := rs.client.NewSession()
	if err != nil {
		return err
	}

	session.Stdin = rs.script
	session.Stdout = rs.stdout
	session.Stderr = rs.stderr

	if err := session.Shell(); err != nil {
		return err
	}
	return session.Wait()
}

func (rs *RemoteScript) runScriptFile() error {
	var buffer bytes.Buffer
	file, err := os.Open(rs.scriptFile)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(&buffer, file)
	if err != nil {
		return err
	}

	rs.script = &buffer
	return rs.runScript()
}

// A RemoteShell represents a login shell on the client.
type RemoteShell struct {
	client         *ssh.Client
	requestPty     bool
	terminalConfig *TerminalConfig

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

// A TerminalConfig represents the configuration for an interactive shell session.
type TerminalConfig struct {
	Term   string
	Height int
	Weight int
	Modes  ssh.TerminalModes
}

// Terminal create a interactive shell on client.
func (c *Client) Terminal(config *TerminalConfig) *RemoteShell {
	return &RemoteShell{
		client:         c.client,
		terminalConfig: config,
		requestPty:     true,
	}
}

// SetStdio specifies where the its standard output and error data will be written.
func (rs *RemoteShell) SetStdio(stdin io.Reader, stdout, stderr io.Writer) *RemoteShell {
	rs.stdin = stdin
	rs.stdout = stdout
	rs.stderr = stderr
	return rs
}

// Start starts a remote shell on client.
func (rs *RemoteShell) Start() error {
	session, err := rs.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	if rs.stdin == nil {
		session.Stdin = os.Stdin
	} else {
		session.Stdin = rs.stdin
	}
	if rs.stdout == nil {
		session.Stdout = os.Stdout
	} else {
		session.Stdout = rs.stdout
	}
	if rs.stderr == nil {
		session.Stderr = os.Stderr
	} else {
		session.Stderr = rs.stderr
	}

	if rs.requestPty {
		tc := rs.terminalConfig
		if tc == nil {
			tc = &TerminalConfig{
				Term:   "xterm",
				Height: 40,
				Weight: 80,
			}
		}
		if err := session.RequestPty(tc.Term, tc.Height, tc.Weight, tc.Modes); err != nil {
			return err
		}
	}

	if err := session.Shell(); err != nil {
		return err
	}

	return session.Wait()
}
