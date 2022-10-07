package google_photo

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootID
	RefreshToken string `json:"refresh_token" required:"true"`
	ClientID     string `json:"client_id" required:"true" default:"202264815644.apps.googleusercontent.com"`
	ClientSecret string `json:"client_secret" required:"true" default:"X4Z3ca8xfWDb1Voo-F9a7ZxJ"`
}

var config = driver.Config{
	Name:        "GooglePhoto",
	OnlyProxy:   true,
	DefaultRoot: "root",
	NoUpload:    true,
	LocalSort:   true,
}

func New() driver.Driver {
	return &GooglePhoto{}
}

func init() {
	op.RegisterDriver(config, New)
}
