package numbers

const MaxOpcodeCount = 201

const MaxMultiSigKeyCount = 20

const BIP16SwitchTime int64 = 1333238400

const HashTypeMask = 0x1f

const MaxScriptEvalStackSize = 1000

const MaxScriptElementSize = 520

const MaxScriptSize = 10000

const MaxOpReturnRelay = 40

const PubKeyHashLen = 20

const MaxPubKeyLen = 65

const MinPubKeyLen = 33

const MaxBlockSize = 1000000

const SatoshiInCoin = 100000000

const SatoshiInTotal = SatoshiInCoin * 21000000

const MinCoinBaseSigScriptSize = 2

const MaxCoinBaseSigScriptSize = 100 

// Threshold for nLockTime: below this value it is interpreted as block number, 
// otherwise as UNIX timestamp.
const LockTimeThreshold = 500000000; // Tue Nov  5 00:53:20 1985 UTC

const MaxStandardTxSize = 100000

// Versions -----------------------------------------
const TxCurrentVersion = 1

// Rough numbers ------------------------------------
const MaxSigScriptSize = 1650