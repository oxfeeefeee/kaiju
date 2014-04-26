//
//
// Here is how stuff gets downloaded:
// 0. [getblocks / getheaders] "getblocks" or "getheaders" which will be responed with "inv"
// 1. [inv] "inv" contains the information of what remote peer has
// 2. [getdata] If we find interesting stuff in "inv", we send "getdata" to request them
// 3. [tx/block/headers] Remote peer sends "tx"/"block"/"headers" in response to "getdata"
package catchUp 

