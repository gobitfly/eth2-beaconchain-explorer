package utils

import (
	"crypto/sha256"
	"fmt"

	"github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/go-ssz"
	"github.com/sirupsen/logrus"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

func init() {
	err := e2types.InitBLS()
	if err != nil {
		logrus.Fatalf("error in e2types.InitBLS(): %v", err)
	}
}

// VerifyBlsToExecutionChangeSignature verifies the signature of an bls_to_execution_change message
// see: https://github.com/wealdtech/ethdo/blob/master/cmd/validator/credentials/set/process.go
// see: https://github.com/prysmaticlabs/prysm/blob/76ed634f7386609f0d1ee47b703eb0143c995464/beacon-chain/core/blocks/withdrawals.go
func VerifyBlsToExecutionChangeSignature(op *capella.SignedBLSToExecutionChange) error {
	genesisForkVersion := phase0.Version{}
	genesisValidatorsRoot := phase0.Root{}
	copy(genesisForkVersion[:], MustParseHex(Config.Chain.ClConfig.GenesisForkVersion))
	copy(genesisValidatorsRoot[:], MustParseHex(Config.Chain.GenesisValidatorsRoot))

	forkDataRoot, err := (&phase0.ForkData{
		CurrentVersion:        genesisForkVersion,
		GenesisValidatorsRoot: genesisValidatorsRoot,
	}).HashTreeRoot()
	if err != nil {
		return fmt.Errorf("failed hashing hashtreeroot: %w", err)
	}

	domain := phase0.Domain{}
	domainBLSToExecutionChange := MustParseHex(Config.Chain.DomainBLSToExecutionChange)
	copy(domain[:], domainBLSToExecutionChange[:])
	copy(domain[4:], forkDataRoot[:])

	root, err := op.Message.HashTreeRoot()
	if err != nil {
		return fmt.Errorf("failed to generate message root: %w", err)
	}

	sigBytes := make([]byte, len(op.Signature))
	copy(sigBytes, op.Signature[:])

	sig, err := e2types.BLSSignatureFromBytes(sigBytes)
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}

	container := &phase0.SigningData{
		ObjectRoot: root,
		Domain:     domain,
	}
	signingRoot, err := ssz.HashTreeRoot(container)
	if err != nil {
		return fmt.Errorf("failed to generate signing root: %w", err)
	}

	pubkeyBytes := make([]byte, len(op.Message.FromBLSPubkey))
	copy(pubkeyBytes, op.Message.FromBLSPubkey[:])
	pubkey, err := e2types.BLSPublicKeyFromBytes(pubkeyBytes)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}
	if !sig.Verify(signingRoot[:], pubkey) {
		return fmt.Errorf("signature does not verify")
	}

	return nil
}

// VerifyVoluntaryExitSignature verifies the signature of an voluntary_exit message
func VerifyVoluntaryExitSignature(op *phase0.SignedVoluntaryExit, forkVersion, pubkeyBytes []byte) error {
	currentVersion := phase0.Version{}
	genesisValidatorsRoot := phase0.Root{}
	copy(currentVersion[:], forkVersion)
	copy(genesisValidatorsRoot[:], MustParseHex(Config.Chain.GenesisValidatorsRoot))

	forkDataRoot, err := (&phase0.ForkData{
		CurrentVersion:        currentVersion,
		GenesisValidatorsRoot: genesisValidatorsRoot,
	}).HashTreeRoot()
	if err != nil {
		return fmt.Errorf("failed hashing hashtreeroot: %w", err)
	}

	domain := phase0.Domain{}
	domainVoluntaryExit := MustParseHex(Config.Chain.DomainVoluntaryExit)
	copy(domain[:], domainVoluntaryExit[:])
	copy(domain[4:], forkDataRoot[:])

	root, err := op.Message.HashTreeRoot()
	if err != nil {
		return fmt.Errorf("failed to generate message root: %w", err)
	}

	sigBytes := make([]byte, len(op.Signature))
	copy(sigBytes, op.Signature[:])

	sig, err := e2types.BLSSignatureFromBytes(sigBytes)
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}

	container := &phase0.SigningData{
		ObjectRoot: root,
		Domain:     domain,
	}
	signingRoot, err := ssz.HashTreeRoot(container)
	if err != nil {
		return fmt.Errorf("failed to generate signing root: %w", err)
	}

	pubkey, err := e2types.BLSPublicKeyFromBytes(pubkeyBytes)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}
	if !sig.Verify(signingRoot[:], pubkey) {
		return fmt.Errorf("signature does not verify")
	}

	return nil
}

func FixAddressCasing(add string) string {
	return common.HexToAddress(add).Hex()
}

func VersionedBlobHash(commitment []byte) common.Hash {
	hasher := sha256.New()
	hasher.Write(commitment[:])
	var vhash common.Hash
	hasher.Sum(vhash[:0])
	vhash[0] = 0x01
	return vhash
}
