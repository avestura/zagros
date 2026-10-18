//go:build !windows

package ext

type unsupportedSecrets struct{}

func init() { secretBackend = unsupportedSecrets{} }

func (unsupportedSecrets) Set(string, string) error { return errSecretsUnsupported }
func (unsupportedSecrets) Get(string) (string, bool, error) {
	return "", false, errSecretsUnsupported
}
func (unsupportedSecrets) Delete(string) error { return errSecretsUnsupported }
