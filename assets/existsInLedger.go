package assets

import (
	"github.com/goledgerdev/cc-tools/errors"
	"github.com/goledgerdev/cc-tools/mock"
	sw "github.com/goledgerdev/cc-tools/stubwrapper"
)

func existsInLedger(stub *sw.StubWrapper, isPrivate bool, typeTag, key string) (bool, errors.ICCError) {
	var assetBytes []byte
	var err error
	if isPrivate {
		_, isMock := stub.Stub.(*mock.MockStub)
		if isMock {
			assetBytes, err = stub.GetPrivateData(typeTag, key)
		} else {
			assetBytes, err = stub.GetPrivateDataHash(typeTag, key)
		}
	} else {
		assetBytes, err = stub.GetState(key)
	}
	if err != nil {
		return false, errors.WrapErrorWithStatus(err, "unable to check asset existence", 400)
	}
	if assetBytes != nil {
		return true, nil
	}

	return false, nil
}

// ExistsInLedger checks if asset currently has a state on the ledger.
func (a *Asset) ExistsInLedger(stub *sw.StubWrapper) (bool, errors.ICCError) {
	if a.Key() == "" {
		return false, errors.NewCCError("asset key is empty", 500)
	}
	return existsInLedger(stub, a.IsPrivate(), a.TypeTag(), a.Key())
}

// ExistsInLedger checks if asset referenced by a key object currently has a state on the ledger.
func (k *Key) ExistsInLedger(stub *sw.StubWrapper) (bool, errors.ICCError) {
	if k.Key() == "" {
		return false, errors.NewCCError("key is empty", 500)
	}
	return existsInLedger(stub, k.IsPrivate(), k.TypeTag(), k.Key())
}
