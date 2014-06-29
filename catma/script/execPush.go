package script

// Push explicite or implicite data on to the stack

// OP_PUSHDATAXX
// OP_PUSHDATAX
func execPushData(ctx *execContext, _ Opcode, operand []byte) error {
    ctx.stack.push(operand)
    return nil
}

// OP_1NEGATE
// OP_1
// OP_2
// OP_3
// ...
// OP_16
func execPushNumber(ctx *execContext, op Opcode, _ []byte) error {
    ctx.stack.push(intToStackItem(int(op) - int(OP_1) + 1))
    return nil
}
