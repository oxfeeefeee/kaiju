// Constants
package kaiju

import (
    "time"
    "math/rand"
)

const NetWorkMagicMain = 0xD9B4BEF9

const NetWorkMagicTestNet = 0xDAB5BFFA

const AlertPublicKey = "04fc9702847840aaf195de8442ebecedf5b095cdbb9bc716bda9110971b28a49e0ead8564ff0db22209e0374782c093bb899692d524e9d6a6956e7c5ecbcd68284"

// Bitcoin network protocol version
const ProtocolVersion uint32 = 70002

// What serivices does this node provides
const NodeServices uint64 = 1

const ListenPort int = 8333

const UserAgent = "/Kaiju:0.1.0/"

//--------------------------------------------------

// Cannot declared as const but works as a const
var NounceInVersionMsg uint64 = uint64(rand.New(rand.NewSource(time.Now().UnixNano())).Int63()) 

const MaxAddrListSize = 30000

const MaxInvListSize = 50000

const MaxStrSize = 100 * 1024

const MaxAlertSize = 100 * 1024

const MaxAlertSingnatureSize = 1024

const MaxDialConcurrency = 1000

const DialTimeout = 2000

const MaxMessagePayload = 4 * 1024 * 1024

const KDBCapacity = 10 * 1024 * 1024

