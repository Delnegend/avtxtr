package social

import (
	"crypto/md5"
	"encoding/hex"
	"net/mail"
	"strings"
)

func Gravatar(email string) (string, error) {
	// validate email
	if _, err := mail.ParseAddress(email); err != nil {
		return "", err
	}
	email = strings.ToLower(strings.TrimSpace(email))

	md5Hash := md5.New()
	md5Hash.Write([]byte(email))
	md5HashInBytes := md5Hash.Sum(nil)
	md5HashString := hex.EncodeToString(md5HashInBytes)

	return "http://www.gravatar.com/avatar/" + md5HashString + "?s=400", nil
}
