package authority

import (
	"github.com/go-errors/errors"
	"github.com/pquerna/otp/totp"
	"image"
)

var TotpValidatedKey = "totpValidated"

type tOtp struct {
}

func (t *tOtp) GenerateBindingQRCode(username string) (image.Image, string, error) {
	if key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "home-dashboard",
		AccountName: username,
	}); err != nil {
		return nil, "", errors.New(err)
	} else if qrCode, err := key.Image(128, 128); err != nil {
		return nil, "", errors.New(err)
	} else {
		return qrCode, key.Secret(), nil
	}
}

func (t *tOtp) Validate(code string, secret string) bool {
	return totp.Validate(code, secret)
}

var TOTP = &tOtp{}
