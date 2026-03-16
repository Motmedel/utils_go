package token_source

import "github.com/Motmedel/utils_go/pkg/cloud/gcp/types/credentials_file"

type TokenSourceWithCredentialsFile interface {
	CredentialsFile() *credentials_file.CredentialsFile
}
