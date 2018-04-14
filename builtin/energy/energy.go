package energy

import (
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vechain/thor/state"
	"github.com/vechain/thor/thor"
)

var (
	tokenSupplyKey = thor.Bytes32(crypto.Keccak256Hash([]byte("token-supply")))
	totalAddKey    = thor.Bytes32(crypto.Keccak256Hash([]byte("total-add")))
	totalSubKey    = thor.Bytes32(crypto.Keccak256Hash([]byte("total-sub")))
)

type Energy struct {
	addr  thor.Address
	state *state.State
}

func New(addr thor.Address, state *state.State) *Energy {
	return &Energy{addr, state}
}

func (e *Energy) getStorage(key thor.Bytes32, val interface{}) {
	e.state.GetStructedStorage(e.addr, key, val)
}

func (e *Energy) setStorage(key thor.Bytes32, val interface{}) {
	e.state.SetStructedStorage(e.addr, key, val)
}

// InitializeTokenSupply initialize VET token supply info.
func (e *Energy) InitializeTokenSupply(supply *big.Int) {
	e.setStorage(tokenSupplyKey, supply)
}

// GetTotalSupply returns total supply of energy.
func (e *Energy) GetTotalSupply(blockNum uint32) *big.Int {
	var tokenSupply big.Int
	e.getStorage(tokenSupplyKey, &tokenSupply)
	var tokenSupplyTime uint64
	e.getStorage(tokenSupplyKey, &tokenSupplyTime)

	// calc grown energy for total token supply
	energyState := state.EnergyState{Energy: &big.Int{}}
	grown := energyState.CalcEnergy(&tokenSupply, blockNum)

	var totalAdd, totalSub big.Int
	e.getStorage(totalAddKey, &totalAdd)
	e.getStorage(totalSubKey, &totalSub)
	grown.Add(grown, &totalAdd)
	return grown.Sub(grown, &totalSub)
}

func (e *Energy) GetTotalBurned() *big.Int {
	var totalAdd, totalSub big.Int
	e.getStorage(totalAddKey, &totalAdd)
	e.getStorage(totalSubKey, &totalSub)
	return new(big.Int).Sub(&totalSub, &totalAdd)
}

// GetBalance returns energy balance of an account at given block time.
func (e *Energy) GetBalance(addr thor.Address, blockNum uint32) *big.Int {
	return e.state.GetEnergy(addr, blockNum)
}

func (e *Energy) AddBalance(addr thor.Address, amount *big.Int, blockNum uint32) {
	bal := e.state.GetEnergy(addr, blockNum)
	if amount.Sign() != 0 {
		var totalAdd big.Int
		e.getStorage(totalAddKey, &totalAdd)
		e.setStorage(totalAddKey, totalAdd.Add(&totalAdd, amount))

		e.state.SetEnergy(addr, new(big.Int).Add(bal, amount), blockNum)
	} else {
		e.state.SetEnergy(addr, bal, blockNum)
	}
}

func (e *Energy) SubBalance(addr thor.Address, amount *big.Int, blockNum uint32) bool {
	bal := e.state.GetEnergy(addr, blockNum)
	if amount.Sign() != 0 {
		if bal.Cmp(amount) < 0 {
			return false
		}
		var totalSub big.Int
		e.getStorage(totalSubKey, &totalSub)
		e.setStorage(totalSubKey, totalSub.Add(&totalSub, amount))

		e.state.SetEnergy(addr, new(big.Int).Sub(bal, amount), blockNum)
	} else {
		e.state.SetEnergy(addr, bal, blockNum)
	}
	return true
}
