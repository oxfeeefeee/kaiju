The design and implementation of KDB (draft)
=====

A solution to mitigate the Bitcoin blockchain size problem.

Abstract
-----
Even with a hard limit of block size, the Bitcoin blockchain has grown 10GB in the past year. We argue that, a full Bitcoin node doesn't need to store the whole blockchain to verify transactions, it can choose to only store unspent outputs.

1. Only Headers Are Essential 
----
Becaus *ALL* the inputs and outputs of the proof-of-work mining are stored only in the block headers, Bitcoin nodes only have a consensus on the headers. As long as the headers are agreed on the content of blocks is also agreed on. That's because a header contains the merkle root of the corresponding block.

That means to keep the p2p net work healthy, we only need to have as much as nodes with verified headers. Different from what the Satoshi paper purposed (which is stubbing off branches of the tree but still keep the integrity of the merkel tree), once a block is buried deep enough, we can build a simple database containing only the upspent transactions, and not keep the block itself.

2. Download-Verify-Extract-Dicard
----
As long as the internet has a reliable source where people can download the full blockchain from, a Bitcoin node with KDB works like this:
- Download the header-chain
- Downlaod all the blocks from number 2, for each block:
    * Verify all the transactions.
    * Verify the merkel root with the header-chain
    * Keep the infomation of the unspent TXs in a TXDatabase
    * Remove the spent TXs in the TXDatabase
    * Discard the block
    
  up until the 100th (or any other number you see fits) newest block. After this is done, we left with a TXDatabase with all unspent TXs.
- When a new block is found, we extract information from the now 101th newest block, put it in the TXDatabase, and discard the block.

When the node verifies a new TX, it can look up the inputs in the TXDatabase, because all the unspent TXs (except the ones in the un-processed, newly found blocks) are stored there. And they are reliable because all the infomation is extract from the verified blockchain.

In summary, A KDB Bitcoin node still need to download the full blockchain, then it verifies the blocks, extracts the infomation of upspent TXs from them and build a compact database for verifing TXs to come.

3. Implementation details
----

To build a smaller and faster TXDatabase, there are some optimizations can be done:
- A TX uses a 256 bit hash to reference a privious TX as an input. We can instead use a 64 bit hash and deal with the rare but possible collisions  
- For a standard Pay-to-PubkeyHash TX, the only infomation we need to keep is the PubkeyHash, the rest are all the same among all the standard TXs.

To be continued...