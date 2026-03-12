package shipinternal

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type LocalSSHKey struct {
	Name           string
	PublicKey      string
	FingerprintMD5 string
}

func DiscoverLocalSSHKey() (LocalSSHKey, error) {
	for _, path := range []string{
		"~/.ssh/id_ed25519.pub",
		"~/.ssh/id_rsa.pub",
		"~/.ssh/id_ecdsa.pub",
		"~/.ssh/id_dsa.pub",
	} {
		key, err := localSSHKeyFromPublicKeyFile(expandHome(path))
		if err == nil {
			return key, nil
		}
	}

	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			defer conn.Close()
			signers, err := agent.NewClient(conn).Signers()
			if err == nil {
				for i, signer := range signers {
					pub := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey())))
					return newLocalSSHKey(fmt.Sprintf("ship-agent-%d", i+1), pub), nil
				}
			}
		}
	}

	return LocalSSHKey{}, fmt.Errorf("no local SSH public key found; create one with ssh-keygen or load one into ssh-agent")
}

func localSSHKeyFromPublicKeyFile(path string) (LocalSSHKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return LocalSSHKey{}, err
	}
	name := strings.TrimSuffix(filepath.Base(path), ".pub")
	return newLocalSSHKey(name, strings.TrimSpace(string(data))), nil
}

func newLocalSSHKey(name, publicKey string) LocalSSHKey {
	return LocalSSHKey{
		Name:           name,
		PublicKey:      publicKey,
		FingerprintMD5: md5Fingerprint(publicKey),
	}
}

func md5Fingerprint(publicKey string) string {
	parsed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		return ""
	}
	sum := md5.Sum(parsed.Marshal())
	raw := hex.EncodeToString(sum[:])
	parts := make([]string, 0, len(raw)/2)
	for i := 0; i < len(raw); i += 2 {
		parts = append(parts, raw[i:i+2])
	}
	return strings.Join(parts, ":")
}
