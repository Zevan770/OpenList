package anyshare

import (
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
)

type Addition struct {
	driver.RootID
	Address     string `json:"address" required:"true" help:"AnyShare server address, e.g. https://myserver:443"`
	AccessToken string `json:"access_token" required:"true" type:"text" help:"Bearer token for API authentication"`
}

var config = driver.Config{
	Name:             "AnyShare",
	LocalSort:        true,
	ProxyRangeOption: true,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &AnyShare{}
	})
}
