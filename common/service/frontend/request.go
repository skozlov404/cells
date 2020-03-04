package frontend

import (
	"net/http"

	"github.com/pydio/cells/common"
	config2 "github.com/pydio/go-os/config"
)

type RequestStatus struct {
	Config        config2.Config
	AclParameters common.ConfigValues
	AclActions    common.ConfigValues
	WsScopes      []string

	User     *User
	NoClaims bool
	Lang     string

	Request *http.Request
}
