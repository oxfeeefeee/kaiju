The design and implementation of KDB (draft)
=====

A solution to mitigate the Bitcoin blockchain size problem and more.

Abstract
-----
A full Bitcoin node doesn't need to store the whole blockchain to verify transactions, it can choose to only store unspent outputs, plus whatever number of blocks it feels like to keep to serve the network.

1. Only Headers Are Essential 
----
Becaus *ALL* the inputs and outputs of the proof-of-work mining are stored only in the block headers, Bitcoin nodes only have a consensus on the headers. As long as the headers are agreed on the content of blocks is also agreed on. That's because a header contains the merkle root of the corresponding block.

That means to keep the p2p net work healthy, we only need to have as much as nodes with verified headers. Once a block is buried deep enough, we can build a simple database containing only the upspent transactions, and not keep the block itself.

If nobody would keep the original blocks, this scheme would be broken. But the full blockchain only needs to be available not aboundant. We can either assume we'll have enough contributor who dont mind the cost of serving the full blockchain, or we can let people randomly choose some of the blocks to keep so that the whole network would still have a full blockchain. 

2. Download-Verify-Extract-Discard
----
As long as the internet has a reliable source where people can download the full blockchain from, a Bitcoin node with KDB works like this:
- Download the header-chain
- Downlaod all the blocks from number 2, for each block:
    * Verify all the transactions.
    * Verify the merkel root with the header-chain.
    * Keep the infomation of the unspent TXs in a TXDatabase.
    * Remove the spent TXs in the TXDatabase.
    * Discard the block.
    
  up until the 100th (or any other number you see fits) newest block. After this is done, we left with a TXDatabase with all unspent TXs.
- When a new block is found, we extract information from the now 101th newest block, put it in the TXDatabase, and discard the block.

When the node verifies a new TX, it can look up the inputs in the TXDatabase, because all the unspent TXs (except the ones in the un-processed, newly found blocks) are stored there. And they are reliable because all the infomation is extract from the verified blockchain.

In summary, A KDB Bitcoin node still need to download the full blockchain, then it verifies the blocks, extracts the infomation of upspent TXs from them and build a compact database for verifing TXs to come.

3. Letting People Choose
----
With Bitcoin Core, you don't get to choose how much storage space or network bandwidth you'd like to contribute. You either run it or not. With KDB, nodes can choose from a range between a minimal requirment and a full-blockchain-serving resources. 

People can choose to only keep unspent TXs and do not store and serve older blocks. Or they can choose how much orginal blockchain data they would keep and serve, then the system will randomly pick some blocks that will be available. In other words, instead of everyone keeps everything, every node keeps a fraction of the chain, so that the whole chain gets stored in a distribute fashion. 

4. Implementation details
----

To build a smaller and faster TXDatabase, there are some optimizations can be done:
- A TX uses a 256 bit hash to reference a privious TX as an input. We can instead use a 64 bit hash and deal with the rare but possible collisions  
- For a standard TXs, the only infomation we need to keep is the Pubkey/Hash, the rest are all the same among all the standard TXs.(Already been done by Satoshi client)

The design of KDB is inspired by CDB, which is a ready-only database. To make the ready-only database fit out needs:
- Start with a relatively large key space
- Simply mark as dead when need deleting a key
- Periodically rebuild the DB when the key space is too crowded or when there is too much garbage.

To be continued