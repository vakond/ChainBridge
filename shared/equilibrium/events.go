package equilibrium

import (
	"github.com/centrifuge/go-substrate-rpc-client/types"
)

type Currency byte
const (
	Unknown Currency = 0
	Usd     Currency = 1
	Eq      Currency = 2
	Eth     Currency = 3
	Btc     Currency = 4
	Eos     Currency = 5
)

// EventBalancesTransfer is emitted when a Substrate client calls Currency::transfer.
type EventBalancesTransfer struct {
	Phase    types.Phase
	From     types.AccountID
	To       types.AccountID
	Currency Currency
	Value    types.U64
	Topics   []types.Hash
}
