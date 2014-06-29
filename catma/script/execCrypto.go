package script

import (
    "crypto/sha1"
    "crypto/sha256"
    "code.google.com/p/go.crypto/ripemd160"
    "github.com/oxfeeefeee/kaiju/klib"
    )

func execRipemd160(ctx *execContext, op Opcode, _ []byte) error {
    if ctx.stack.empty() {
        return errStackItemMissing
    }
    r := ripemd160.New()
    r.Write(ctx.stack.pop())
    ctx.stack.push(r.Sum(nil))
    return nil
}

func execSha1(ctx *execContext, op Opcode, _ []byte) error {
    if ctx.stack.empty() {
        return errStackItemMissing
    }
    s := sha1.New()
    s.Write(ctx.stack.pop())
    ctx.stack.push(s.Sum(nil))
    return nil
}

func execSha256(ctx *execContext, op Opcode, _ []byte) error {
    if ctx.stack.empty() {
        return errStackItemMissing
    }
    h := sha256.Sum256(ctx.stack.pop())
    ctx.stack.push(h[:])
    return nil
}

func execHash160(ctx *execContext, op Opcode, _ []byte) error {
    if ctx.stack.empty() {
        return errStackItemMissing
    }
    h := sha256.Sum256(ctx.stack.pop())
    r := ripemd160.New()
    r.Write(h[:])
    ctx.stack.push(r.Sum(nil))
    return nil
}

func execHash256(ctx *execContext, op Opcode, _ []byte) error {
    if ctx.stack.empty() {
        return errStackItemMissing
    }
    h := *klib.Sha256Sha256(ctx.stack.pop())
    ctx.stack.push(h[:])
    return nil
}
