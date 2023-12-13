package generate

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"

	"github.com/AxisCommunications/go-dpop"
	"github.com/MakeNowJust/heredoc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type GenerateOpts struct {
	IO                  *iostreams.IOStreams
	PrivateKeyLocation  string
	PersonalAccessToken string
	Hostname            string
}

type PasswordReader interface {
	Read() ([]byte, error)
}

type ConsolePasswordReader struct{}

func (pr ConsolePasswordReader) Read() ([]byte, error) {
	return term.ReadPassword(int(os.Stdin.Fd()))
}

func NewCmdGenerate(f *cmdutils.Factory) *cobra.Command {
	opts := &GenerateOpts{
		IO: f.IO,
	}
	cmd := &cobra.Command{
		Use:   "dpop-gen [flags]",
		Short: "Generates a DPoP (demonstrating-proof-of-possession) proof JWT",
		Long: heredoc.Doc(`
		[Experiment] Demonstrating-proof-of-possession (DPoP, <https://gitlab.com/gitlab-com/gl-security/appsec/security-feature-blueprints/-/blob/main/sender_constraining_access_tokens/index.md>) is a technique to
		cryptographically bind personal access tokens to their owners. The tools to manage the client aspects of DPoP are
		provided by this command.

		The command generates a DPoP proof JWT that can be used alongside a Personal Access Token (PAT) to authenticate
		to the GitLab API. It is valid for 5 minutes and will need to be generated again once it expires. Your SSH
		private key will be used to sign the JWT. RSA, ed25519, and ECDSA keys are supported.
		`),
		Example: heredoc.Doc(`
			# Generate a DPoP JWT for authentication to GitLab
			$ glab dpop-gen [flags]
			$ glab dpop-gen --private-key "~/.ssh/id_rsa" --pat "glpat-xxxxxxxxxxxxxxxxxxxx"
			# No PAT required if the user has previously used the glab auth login command with a PAT
			$ glab dpop-gen --private-key "~/.ssh/id_rsa"
			# Generate a DPoP JWT for another GitLab instance
			$ glab dpop-gen --private-key "~/.ssh/id_rsa" --hostname "https://gitlab.com"
		`),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.PrivateKeyLocation == "" {
				return fmt.Errorf("private key location is required")
			}
			if opts.PersonalAccessToken == "" {
				cfg, err := f.Config()
				if err != nil {
					return fmt.Errorf("could not get config: %w", err)
				}

				token, err := cfg.Get(opts.Hostname, "token")
				if err != nil {
					return err
				}

				if token != "" {
					opts.PersonalAccessToken = token
				} else {
					return fmt.Errorf("personal access token is required")
				}
			}

			privateKey, err := loadPrivateKey(opts.PrivateKeyLocation, ConsolePasswordReader{})
			if err != nil {
				return err
			}

			proofString, err := generateDPoPProof(&privateKey, opts.PersonalAccessToken)
			if err != nil {
				return err
			}

			log.Println("DPoP Proof:", proofString)

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.PrivateKeyLocation, "private-key", "p", "", "Location of the private SSH key on the local system.")
	cmd.Flags().StringVar(&opts.PersonalAccessToken, "pat", "", "Personal Access Token (PAT) to generate a DPoP proof for. If this is not provided, the token set with `glab auth login` will be used. In the absence of both, there will be an error.")
	cmd.Flags().StringVarP(&opts.Hostname, "hostname", "h", "gitlab.com", "The hostname of the GitLab instance to authenticate with")

	return cmd
}

func generateDPoPProof(key *crypto.PrivateKey, accessToken string) (string, error) {
	signingMethod, err := getSigningMethod(key)
	if err != nil {
		return "", err
	}

	hashedToken := sha256.Sum256([]byte(accessToken))
	base64UrlEncodedHash := base64.RawURLEncoding.EncodeToString(hashedToken[:])

	uuidObj, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := &dpop.ProofTokenClaims{
		RegisteredClaims: &jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute * 5)),
			ID:        uuidObj.String(),
		},
		AccessTokenHash: base64UrlEncodedHash,
	}

	if signer, ok := (*key).(crypto.Signer); ok {
		return dpop.Create(signingMethod, claims, signer)
	} else {
		return "", fmt.Errorf("key type does not implement crypto.Signer")
	}
}

func getSigningMethod(key *crypto.PrivateKey) (jwt.SigningMethod, error) {
	var signingMethod jwt.SigningMethod
	switch key := (*key).(type) {
	case *rsa.PrivateKey:
		{
			if key.N.BitLen() < 2048 {
				// Minimum should be 2048 as per https://www.rfc-editor.org/rfc/rfc7518.html#section-3.3
				return nil, fmt.Errorf("rsa key size must be at least 2048 bits")
			} else if key.N.BitLen() > 8192 {
				// Maximum should be 8192 as per https://docs.gitlab.com/ee/user/ssh.html#rsa-ssh-keys
				return nil, fmt.Errorf("rsa key size must be at most 8192 bits")
			}
			signingMethod = jwt.SigningMethodRS512
		}
	case *ed25519.PrivateKey:
		signingMethod = jwt.SigningMethodEdDSA
	default:
		return nil, fmt.Errorf("unsupported key type")
	}
	return signingMethod, nil
}

func loadPrivateKey(path string, passwordReader PasswordReader) (crypto.PrivateKey, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	privateKey, err := ssh.ParseRawPrivateKey(keyBytes)
	if err != nil {
		var passphraseMissingErr *ssh.PassphraseMissingError
		if errors.As(err, &passphraseMissingErr) {
			fmt.Println("SSH private key is encrypted, please enter passphrase: ")
			passphrase, err := passwordReader.Read()
			if err != nil {
				return nil, err
			}

			privateKey, err = ssh.ParseRawPrivateKeyWithPassphrase(keyBytes, passphrase)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return privateKey.(crypto.PrivateKey), nil
}
