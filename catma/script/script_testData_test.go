// The file name needs to end with "_test" so that it doesn't get compiled into binary
package script

func testScripts() map[string]PKScriptType {
    return map[string]PKScriptType{
        "0x21 0x035e7f0d4d0841bcd56c39337ed086b1a633ee770c1ffdd94ac552a95ac2ce0efc CHECKSIG": PKS_PUBKEY,
        "DUP HASH160 0x14 0xe52b482f2faa8ecbf0db344f93c84ac908557f33 EQUALVERIFY CHECKSIG": PKS_PUBKEYHASH,
        "HASH160 0x14 0x7a052c840ba73af26755de42cf01cc9e0a49fef0 EQUAL": PKS_SCRIPTHASH,
        "2 0x21 0x033bcaa0a602f0d44cc9d5637c6e515b0471db514c020883830b7cefd73af04194 0x21 0x03a88b326f8767f4f192ce252afe33c94d25ab1d24f27f159b3cb3aa691ffe1423 2 CHECKMULTISIG": PKS_MULTISIG,
    }
}