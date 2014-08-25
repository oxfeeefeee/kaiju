// Adapted from Satoshi client
package script

import (
    "fmt"
    )

const (
    // 0x00 -- 0x4b are used to push data with the length of the Opcode value
    // the Satoshi client not listing them makes it a little confusing. 
    //
    // With Satoshi client 0x00 is called OP_0. 0x00 is unique because it pushes 
    // an empty item on to the stack, calling is OP_PUSHDATA00 and OP_0 are both fine.
    OP_PUSHDATA00 Opcode = iota
    OP_PUSHDATA01
    OP_PUSHDATA02
    OP_PUSHDATA03
    OP_PUSHDATA04
    OP_PUSHDATA05
    OP_PUSHDATA06
    OP_PUSHDATA07
    OP_PUSHDATA08
    OP_PUSHDATA09
    OP_PUSHDATA0a
    OP_PUSHDATA0b
    OP_PUSHDATA0c
    OP_PUSHDATA0d
    OP_PUSHDATA0e
    OP_PUSHDATA0f
    
    OP_PUSHDATA10
    OP_PUSHDATA11
    OP_PUSHDATA12
    OP_PUSHDATA13
    OP_PUSHDATA14
    OP_PUSHDATA15
    OP_PUSHDATA16
    OP_PUSHDATA17
    OP_PUSHDATA18
    OP_PUSHDATA19
    OP_PUSHDATA1a
    OP_PUSHDATA1b
    OP_PUSHDATA1c
    OP_PUSHDATA1d
    OP_PUSHDATA1e
    OP_PUSHDATA1f

    OP_PUSHDATA20
    OP_PUSHDATA21
    OP_PUSHDATA22
    OP_PUSHDATA23
    OP_PUSHDATA24
    OP_PUSHDATA25
    OP_PUSHDATA26
    OP_PUSHDATA27
    OP_PUSHDATA28
    OP_PUSHDATA29
    OP_PUSHDATA2a
    OP_PUSHDATA2b
    OP_PUSHDATA2c
    OP_PUSHDATA2d
    OP_PUSHDATA2e
    OP_PUSHDATA2f

    OP_PUSHDATA30
    OP_PUSHDATA31
    OP_PUSHDATA32
    OP_PUSHDATA33
    OP_PUSHDATA34
    OP_PUSHDATA35
    OP_PUSHDATA36
    OP_PUSHDATA37
    OP_PUSHDATA38
    OP_PUSHDATA39
    OP_PUSHDATA3a
    OP_PUSHDATA3b
    OP_PUSHDATA3c
    OP_PUSHDATA3d
    OP_PUSHDATA3e
    OP_PUSHDATA3f

    OP_PUSHDATA40
    OP_PUSHDATA41
    OP_PUSHDATA42
    OP_PUSHDATA43
    OP_PUSHDATA44
    OP_PUSHDATA45
    OP_PUSHDATA46
    OP_PUSHDATA47
    OP_PUSHDATA48
    OP_PUSHDATA49
    OP_PUSHDATA4a
    OP_PUSHDATA4b

    OP_PUSHDATA1        //0x4c
    OP_PUSHDATA2        //0x4d
    OP_PUSHDATA4        //0x4e
    OP_1NEGATE          //0x4f
    OP_RESERVED         //0x50
    OP_1                //0x51
    OP_2                //0x52
    OP_3                //0x53
    OP_4                //0x54
    OP_5                //0x55
    OP_6                //0x56
    OP_7                //0x57
    OP_8                //0x58
    OP_9                //0x59
    OP_10               //0x5a
    OP_11               //0x5b
    OP_12               //0x5c
    OP_13               //0x5d
    OP_14               //0x5e
    OP_15               //0x5f
    OP_16               //0x60

    // control
    OP_NOP              //0x61
    OP_VER              //0x62
    OP_IF               //0x63
    OP_NOTIF            //0x64
    OP_VERIF            //0x65
    OP_VERNOTIF         //0x66
    OP_ELSE             //0x67
    OP_ENDIF            //0x68
    OP_VERIFY           //0x69
    OP_RETURN           //0x6a

    // stack ops
    OP_TOALTSTACK       //0x6b
    OP_FROMALTSTACK     //0x6c
    OP_2DROP            //0x6d
    OP_2DUP             //0x6e
    OP_3DUP             //0x6f
    OP_2OVER            //0x70
    OP_2ROT             //0x71
    OP_2SWAP            //0x72
    OP_IFDUP            //0x73
    OP_DEPTH            //0x74
    OP_DROP             //0x75
    OP_DUP              //0x76
    OP_NIP              //0x77
    OP_OVER             //0x78
    OP_PICK             //0x79
    OP_ROLL             //0x7a
    OP_ROT              //0x7b
    OP_SWAP             //0x7c
    OP_TUCK             //0x7d

    // splice ops
    OP_CAT              //0x7e
    OP_SUBSTR           //0x7f
    OP_LEFT             //0x80
    OP_RIGHT            //0x81
    OP_SIZE             //0x82

    // bit logic
    OP_INVERT           //0x83
    OP_AND              //0x84
    OP_OR               //0x85
    OP_XOR              //0x86
    OP_EQUAL            //0x87
    OP_EQUALVERIFY      //0x88
    OP_RESERVED1        //0x89
    OP_RESERVED2        //0x8a

    // numeric
    OP_1ADD             //0x8b
    OP_1SUB             //0x8c
    OP_2MUL             //0x8d
    OP_2DIV             //0x8e
    OP_NEGATE           //0x8f
    OP_ABS              //0x90
    OP_NOT              //0x91
    OP_0NOTEQUAL        //0x92

    OP_ADD              //0x93
    OP_SUB              //0x94
    OP_MUL              //0x95
    OP_DIV              //0x96
    OP_MOD              //0x97
    OP_LSHIFT           //0x98
    OP_RSHIFT           //0x99

    OP_BOOLAND          //0x9a
    OP_BOOLOR           //0x9b
    OP_NUMEQUAL         //0x9c
    OP_NUMEQUALVERIFY   //0x9d
    OP_NUMNOTEQUAL      //0x9e
    OP_LESSTHAN         //0x9f
    OP_GREATERTHAN      //0xa0
    OP_LESSTHANOREQUAL  //0xa1
    OP_GREATERTHANOREQUAL //0xa2
    OP_MIN              //0xa3
    OP_MAX              //0xa4

    OP_WITHIN           //0xa5

    // crypto
    OP_RIPEMD160        //0xa6
    OP_SHA1             //0xa7
    OP_SHA256           //0xa8
    OP_HASH160          //0xa9
    OP_HASH256          //0xaa
    OP_CODESEPARATOR    //0xab
    OP_CHECKSIG         //0xac
    OP_CHECKSIGVERIFY   //0xad
    OP_CHECKMULTISIG    //0xae
    OP_CHECKMULTISIGVERIFY //0xaf

    // expansion
    OP_NOP1             //0xb0
    OP_NOP2             //0xb1
    OP_NOP3             //0xb2
    OP_NOP4             //0xb3
    OP_NOP5             //0xb4
    OP_NOP6             //0xb5
    OP_NOP7             //0xb6
    OP_NOP8             //0xb7
    OP_NOP9             //0xb8
    OP_NOP10            //0xb9

    OP_INVALIDOPCODE Opcode = 0xff
)

type Opcode byte

// The same as Satoshi client
func (c Opcode) String() string {
    str, _ := c.attr()
    return str
}

// Disabled Opcodes make the script invalid no matter what, so we set the exec func as
// nil to mark them
func (c Opcode) attr() (string, execFunc) {
    if c >= OP_PUSHDATA00 && c <= OP_PUSHDATA4b {
        return fmt.Sprintf("OP_PUSHDATA%02x", byte(c)), execPushData
    }
    switch c {
    case OP_PUSHDATA1           : return "OP_PUSHDATA1",            execPushData
    case OP_PUSHDATA2           : return "OP_PUSHDATA2",            execPushData
    case OP_PUSHDATA4           : return "OP_PUSHDATA4",            execPushData
    case OP_1NEGATE             : return "-1",                      execPushNumber
    case OP_RESERVED            : return "OP_RESERVED",             execInvalid
    case OP_1                   : return "1",                       execPushNumber
    case OP_2                   : return "2",                       execPushNumber
    case OP_3                   : return "3",                       execPushNumber
    case OP_4                   : return "4",                       execPushNumber
    case OP_5                   : return "5",                       execPushNumber
    case OP_6                   : return "6",                       execPushNumber
    case OP_7                   : return "7",                       execPushNumber
    case OP_8                   : return "8",                       execPushNumber
    case OP_9                   : return "9",                       execPushNumber
    case OP_10                  : return "10",                      execPushNumber
    case OP_11                  : return "11",                      execPushNumber
    case OP_12                  : return "12",                      execPushNumber
    case OP_13                  : return "13",                      execPushNumber
    case OP_14                  : return "14",                      execPushNumber
    case OP_15                  : return "15",                      execPushNumber
    case OP_16                  : return "16",                      execPushNumber

    // control
    case OP_NOP                 : return "OP_NOP",                  execNop
    case OP_VER                 : return "OP_VER",                  execInvalid
    case OP_IF                  : return "OP_IF",                   execBranching
    case OP_NOTIF               : return "OP_NOTIF",                execBranching
    case OP_VERIF               : return "OP_VERIF",                execInvalid
    case OP_VERNOTIF            : return "OP_VERNOTIF",             execInvalid
    case OP_ELSE                : return "OP_ELSE",                 execBranching
    case OP_ENDIF               : return "OP_ENDIF",                execBranching
    case OP_VERIFY              : return "OP_VERIFY",               execControl
    case OP_RETURN              : return "OP_RETURN",               execControl

    // stack ops
    case OP_TOALTSTACK          : return "OP_TOALTSTACK",           execStackOp
    case OP_FROMALTSTACK        : return "OP_FROMALTSTACK",         execStackOp
    case OP_2DROP               : return "OP_2DROP",                execStackOp
    case OP_2DUP                : return "OP_2DUP",                 execStackOp
    case OP_3DUP                : return "OP_3DUP",                 execStackOp
    case OP_2OVER               : return "OP_2OVER",                execStackOp
    case OP_2ROT                : return "OP_2ROT",                 execStackOp
    case OP_2SWAP               : return "OP_2SWAP",                execStackOp
    case OP_IFDUP               : return "OP_IFDUP",                execStackOp
    case OP_DEPTH               : return "OP_DEPTH",                execStackOp
    case OP_DROP                : return "OP_DROP",                 execStackOp
    case OP_DUP                 : return "OP_DUP",                  execStackOp
    case OP_NIP                 : return "OP_NIP",                  execStackOp
    case OP_OVER                : return "OP_OVER",                 execStackOp
    case OP_PICK                : return "OP_PICK",                 execStackOp
    case OP_ROLL                : return "OP_ROLL",                 execStackOp
    case OP_ROT                 : return "OP_ROT",                  execStackOp
    case OP_SWAP                : return "OP_SWAP",                 execStackOp
    case OP_TUCK                : return "OP_TUCK",                 execStackOp

    // splice ops
    case OP_CAT                 : return "OP_CAT",                  nil
    case OP_SUBSTR              : return "OP_SUBSTR",               nil
    case OP_LEFT                : return "OP_LEFT",                 nil
    case OP_RIGHT               : return "OP_RIGHT",                nil
    case OP_SIZE                : return "OP_SIZE",                 execSize

    // bit logic
    case OP_INVERT              : return "OP_INVERT",               nil
    case OP_AND                 : return "OP_AND",                  nil
    case OP_OR                  : return "OP_OR",                   nil
    case OP_XOR                 : return "OP_XOR",                  nil
    case OP_EQUAL               : return "OP_EQUAL",                execEqual
    case OP_EQUALVERIFY         : return "OP_EQUALVERIFY",          execEqual
    case OP_RESERVED1           : return "OP_RESERVED1",            execInvalid
    case OP_RESERVED2           : return "OP_RESERVED2",            execInvalid

    // numeric
    case OP_1ADD                : return "OP_1ADD",                 execNumeric1
    case OP_1SUB                : return "OP_1SUB",                 execNumeric1
    case OP_2MUL                : return "OP_2MUL",                 nil
    case OP_2DIV                : return "OP_2DIV",                 nil
    case OP_NEGATE              : return "OP_NEGATE",               execNumeric1
    case OP_ABS                 : return "OP_ABS",                  execNumeric1
    case OP_NOT                 : return "OP_NOT",                  execNumeric1
    case OP_0NOTEQUAL           : return "OP_0NOTEQUAL",            execNumeric1
    case OP_ADD                 : return "OP_ADD",                  execNumeric2
    case OP_SUB                 : return "OP_SUB",                  execNumeric2
    case OP_MUL                 : return "OP_MUL",                  nil
    case OP_DIV                 : return "OP_DIV",                  nil
    case OP_MOD                 : return "OP_MOD",                  nil
    case OP_LSHIFT              : return "OP_LSHIFT",               nil
    case OP_RSHIFT              : return "OP_RSHIFT",               nil
    case OP_BOOLAND             : return "OP_BOOLAND",              execNumeric2
    case OP_BOOLOR              : return "OP_BOOLOR",               execNumeric2
    case OP_NUMEQUAL            : return "OP_NUMEQUAL",             execNumeric2
    case OP_NUMEQUALVERIFY      : return "OP_NUMEQUALVERIFY",       execNumeric2
    case OP_NUMNOTEQUAL         : return "OP_NUMNOTEQUAL",          execNumeric2
    case OP_LESSTHAN            : return "OP_LESSTHAN",             execNumeric2
    case OP_GREATERTHAN         : return "OP_GREATERTHAN",          execNumeric2
    case OP_LESSTHANOREQUAL     : return "OP_LESSTHANOREQUAL",      execNumeric2
    case OP_GREATERTHANOREQUAL  : return "OP_GREATERTHANOREQUAL",   execNumeric2
    case OP_MIN                 : return "OP_MIN",                  execNumeric2
    case OP_MAX                 : return "OP_MAX",                  execNumeric2
    case OP_WITHIN              : return "OP_WITHIN",               execWithin

    // crypto
    case OP_RIPEMD160           : return "OP_RIPEMD160",            execRipemd160
    case OP_SHA1                : return "OP_SHA1",                 execSha1
    case OP_SHA256              : return "OP_SHA256",               execSha256
    case OP_HASH160             : return "OP_HASH160",              execHash160
    case OP_HASH256             : return "OP_HASH256",              execHash256
    case OP_CODESEPARATOR       : return "OP_CODESEPARATOR",        execSeparator
    case OP_CHECKSIG            : return "OP_CHECKSIG",             execCheckSig
    case OP_CHECKSIGVERIFY      : return "OP_CHECKSIGVERIFY",       execCheckSig
    case OP_CHECKMULTISIG       : return "OP_CHECKMULTISIG",        execCheckMultiSig
    case OP_CHECKMULTISIGVERIFY : return "OP_CHECKMULTISIGVERIFY",  execCheckMultiSig 

    // expanson
    case OP_NOP1                : return "OP_NOP1",                 execNop
    case OP_NOP2                : return "OP_NOP2",                 execNop
    case OP_NOP3                : return "OP_NOP3",                 execNop
    case OP_NOP4                : return "OP_NOP4",                 execNop
    case OP_NOP5                : return "OP_NOP5",                 execNop
    case OP_NOP6                : return "OP_NOP6",                 execNop
    case OP_NOP7                : return "OP_NOP7",                 execNop
    case OP_NOP8                : return "OP_NOP8",                 execNop
    case OP_NOP9                : return "OP_NOP9",                 execNop
    case OP_NOP10               : return "OP_NOP10",                execNop

    case OP_INVALIDOPCODE       : return "OP_INVALIDOPCODE",        nil
    }
    return "OP_UNKNOWN", nil
}

// Returns the number to be pushed for OP_PUSHDATA00, OP_1, OP_2 ...
// returns -1 othewise
func (c Opcode) number() int {
    if c == OP_PUSHDATA00 {
        return 0
    }
    if c >= OP_1 && c <= OP_16 {
        return int(c) - int(OP_1) + 1
    }
    return -1
}
