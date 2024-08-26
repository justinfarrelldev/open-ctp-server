package lobby

import (
	account "github.com/justinfarrelldev/open-ctp-server/internal/account"
)

type Lobby struct {
	players []account.Account
}
