2 Runtime Verification, Inc.
incremental algorithm enjoys O(h) time and space complexity to reconstruct
(more precisely, compute the root of) a Merkle tree of height h, while a naive al￾gorithm would require O(2h
) time or space complexity. The efficient incremental
algorithm, however, leads to the deposit contract implementation being unintu￾itive, and makes it non-trivial to ensure its correctness. The correctness of the
deposit contract, however, is critical for the security of Ethereum 2.0, since it is
a gateway for becoming a validator. Considering the utmost importance of the
deposit contract for the Ethereum blockchain, formal verification is demanded
to ultimately guarantee its correctness.
In this paper, we present our formal verification of the deposit contract.3 The
scope of verification is to ensure the correctness of the contract bytecode within
a single transaction, without considering transaction-level or off-chain behaviors.
We take the compiled bytecode as the verification target to avoid the need to
trust the compiler.
We adopt a refinement-based verification approach. Specifically, our verifica￾tion effort consists of the following two tasks:
– Verify that the incremental Merkle tree algorithm implemented in the deposit
contract is correct w.r.t. the original full-construction algorithm.
– Verify that the compiled bytecode is correctly generated from the source code
of the deposit contract.
Intuitively, the first task amounts to ensuring the correctness of the contract
source code, while the second task amounts to ensuring the compiled bytecode
being a sound refinement of the source code (i.e., translation validation of the
compiler). This refinement-based approach allows us to avoid reasoning about
the complex algorithmic details, especially specifying and verifying loop invari￾ants, directly at the bytecode level. This separation of concerns helped us to save
a significant amount of verification effort. See Section 2 for more details.
Verification Target. The specific target of our formal verification is the lat￾est version (2ac62c4) of the deposit contract written in Solidity, provided that
the contract is compiled by the Solidity compiler v0.6.8 with the optimization
enabled (–optimize-runs 5000000).
2 Our Refinement-Based Verification Approach
We illustrate our refinement-based formal verification approach used in the de￾posit contract verification. We present our approach using the K framework and
its verification infrastructure [22,24,16], but it can be applied to other program
verification frameworks.
Let us consider a sum program that computes the summation from 1 to n:
int sum(int n) { int s = 0; int i = 1;
while(i <= n) { s = s + i; i = i + 1; } return s; }
3 This was done as part of a contract funded by the Ethereum Foundation [8].2 Runtime Verification, Inc.
incremental algorithm enjoys O(h) time and space complexity to reconstruct
(more precisely, compute the root of) a Merkle tree of height h, while a naive al￾gorithm would require O(2h
) time or space complexity. The efficient incremental
algorithm, however, leads to the deposit contract implementation being unintu￾itive, and makes it non-trivial to ensure its correctness. The correctness of the
deposit contract, however, is critical for the security of Ethereum 2.0, since it is
a gateway for becoming a validator. Considering the utmost importance of the
deposit contract for the Ethereum blockchain, formal verification is demanded
to ultimately guarantee its correctness.
In this paper, we present our formal verification of the deposit contract.3 The
scope of verification is to ensure the correctness of the contract bytecode within
a single transaction, without considering transaction-level or off-chain behaviors.
We take the compiled bytecode as the verification target to avoid the need to
trust the compiler.
We adopt a refinement-based verification approach. Specifically, our verifica￾tion effort consists of the following two tasks:
– Verify that the incremental Merkle tree algorithm implemented in the deposit
contract is correct w.r.t. the original full-construction algorithm.
– Verify that the compiled bytecode is correctly generated from the source code
of the deposit contract.
Intuitively, the first task amounts to ensuring the correctness of the contract
source code, while the second task amounts to ensuring the compiled bytecode
being a sound refinement of the source code (i.e., translation validation of the
compiler). This refinement-based approach allows us to avoid reasoning about
the complex algorithmic details, especially specifying and verifying loop invari￾ants, directly at the bytecode level. This separation of concerns helped us to save
a significant amount of verification effort. See Section 2 for more details.
Verification Target. The specific target of our formal verification is the lat￾est version (2ac62c4) of the deposit contract written in Solidity, provided that
the contract is compiled by the Solidity compiler v0.6.8 with the optimization
enabled (–optimize-runs 5000000).
2 Our Refinement-Based Verification Approach
We illustrate our refinement-based formal verification approach used in the de￾posit contract verification. We present our approach using the K framework and
its verification infrastructure [22,24,16], but it can be applied to other program
verification frameworks.
Let us consider a sum program that computes the summation from 1 to n:
int sum(int n) { int s = 0; int i = 1;
while(i <= n) { s = s + i; i = i + 1; } return s; }
3 This was done as part of a contract funded by the Ethereum Foundation [8].End-to-End Formal Verification of Ethereum 2.0 Deposit Smart Contract 3
Given this program, we first manually write an abstract model of the program
in the K framework [22]. Such a K model is essentially a state transition system
of the program, and can be written as follows:
rule: sum(n) ⇒ loop(s: 0, i: 1, n: n)
rule: loop(s: s, i: i, n: n) ⇒ loop(s: s + i, i: i + 1, n: n) when i ≤ n
rule: loop(s: s, i: i, n: n) ⇒ return(s) when i > n
These transition rules correspond to the initialization, the while loop, and the
return statement, respectively. The indexed tuple (s: s, i: i, n: n) repre￾sents the state of the program variables s, i, and n.
4
Then, given the abstract model, we specify the functional correctness prop￾erty in reachability logic [23], as follows:
claim: sum(n) ⇒ return( n(n+1)
2
) when n > 0
This reachability claim says that sum(n) will eventually return n(n+1)
2
in all
possible execution paths, if n is positive. We verify this specification using the
K reachability logic theorem prover [24], which requires us only to provide the
following loop invariant:5
invariant: loop(s: i(i−1)
2
, i: i, n: n) ⇒ return( n(n+1)
2
) when 0 < i ≤ n + 1
Once we prove the desired property of the abstract model, we manually refine
the model to a bytecode specification, by translating each transition rule of the
abstract model into a reachability claim at the bytecode level, as follows:
claim: evm(pc: pcbegin, calldata: #bytes(32, n), stack: [], · · · )
⇒ evm(pc: pcloophead, stack: [0, 1, n], · · · )
claim: evm(pc: pcloophead, stack: [s, i, n], · · · )
⇒ evm(pc: pcloophead, stack: [s + i, i + 1, n], · · · ) when i ≤ n
claim: evm(pc: pcloophead, stack: [s, i, n], · · · )
⇒ evm(pc: pcend, stack: [], output: #bytes(32, s), · · · ) when i > n
Here, the indexed tuple evm(pc:_, calldata:_, stack:_, output:_) repre￾sents (part of) the Ethereum Virtual Machine (EVM) state, and #bytes(N,V )
denotes a sequence of N bytes of the two’s complement representation of V .
We verify this bytecode specification against the compiled bytecode using
the same K reachability theorem prover [24,16]. Note that no loop invariant is
needed in this bytecode verification, since each reachability claim involves only a
bounded number of execution steps—specifically, the second claim involves only
a single iteration of the loop.
Then, we manually prove the soundness of the refinement, which can be
stated as follows: for any EVM states σ1 and σ2, if σ1 ⇒ σ2, then α(σ1) ⇒ α(σ2),
where the abstraction function α is defined as follows:
4 Note that this abstract model can be also automatically derived by instantiating
the language semantics with the particular program, if a formal semantics of the
language is available (in the K framework).
5 The loop invariants in reachability logic mentioned here look different from those
in Hoare logic. See the comparison between the two logic proof systems in [24,
Section 4]. These loop invariants can be also seen as transition invariants [17].4 Runtime Verification, Inc.
1 2 3 0
#1,2 #3,0
#1,2, 
 3,0
0 0 0 0
#0,0 #0,0
#0,0, 
 0,0
#1,2, 
 3,0,…
branch
zerohashes
updated
untouched
1 2 3 4 5 6 7 8
9 10 11 12
13 14
15
1 2 3 4
#1,2 #3,4
#1,2, 
 3,4
0 0 0 0
#0,0 #0,0
#0,0, 
 0,0
#1,2, 
 3,4,…
branch
zerohashes
updated
untouched
1 2 3 4 5 6 7 8
9 10 11 12
13 14
15
1 2 3 0
#1,2 #3,0
#1,2, 
 3,0
0 0 0 0
#0,0 #0,0
#0,0, 
 0,0
#1,2, 
 3,0,…
branch
zerohashes
updated
untouched
)
Fig. 1. Illustration of the incremental Merkle tree algorithm. The left tree is updated
to the right tree by inserting a new data hash in the fourth leaf (node 4). Only the path
from the new leaf to the root (i.e., the gray nodes) are computed by the algorithm (hence
linear-time). The bold-lined (and bold-dotted-lined) nodes denote the branch (and
zero_hashes) array, respectively, which are only nodes that the algorithm maintains
(hence linear-space). The # symbol denotes the hash value, e.g., “#1,2” (in node 9)
denotes “hash(1,2)”, and “#1,2,3,4” (in node 13) denotes “hash(hash(1,2),hash(3,4))”.
Node numbers are labeled in the upper-right corner of each node.
α(evm(pc: pcbegin, calldata: #bytes(32, n), stack: [], · · · )) = sum(n)
α(evm(pc: pcloophead, stack: [s, i, n], · · · )) = loop(s: s, i: i, n: n)
α(evm(pc: pcend, stack: [], output: #bytes(32, s), · · · )) = return(s)
Putting all the results together, we finally conclude that the compiled byte￾code will return #bytes(32, n(n+1)
2
).
Note that the abstract model and the compiler are not in the trust base,
thanks to the refinement, while the K reachability logic theorem prover [24,16]
and the formal semantics of EVM [12] are.
3 Correctness of the Incremental Merkle Tree Algorithm
In this section, we briefly describe the incremental Merkle tree algorithm of the
deposit contract, and formulate its correctness. Both the formalization of the
algorithm and the formal proof of the correctness are presented in Appendix A.
A Merkle tree [13] is a perfect binary tree [14] where leaf nodes store the hash
of data, and non-leaf nodes store the hash of their children. A partial Merkle tree
up-to m is a Merkle tree whose first (leftmost) m leaves are filled with data hashes
and the other leaves are empty and filled with zeros. The incremental Merkle
tree algorithm takes as input a partial Merkle tree up-to m and a new data
hash, and inserts the new data hash into the (m+ 1)th leaf, resulting in a partial
Merkle tree up-to m + 1.
Figure 1 illustrates the algorithm, showing how the given partial Merkle tree
up-to 3 (shown in the left) is updated to the resulting partial Merkle tree up-to
4 (in the right) when a new data hash is inserted into the 4th leaf node. Here
are a few key observations of the algorithm.
1. The only difference between the two Merkle trees is the path from the new
leaf node (i.e., node 4) to the root. All the other nodes are identical between
the two trees. End-to-End Formal Verification of Ethereum 2.0 Deposit Smart Contract 5
2. The path can be computed by using only the left (i.e., node 3 and node 9)
or right (i.e., node 14) sibling of each node in the path. All the other nodes
are not needed for the path computation.
3. All the left siblings (i.e., node 3 and node 9) of the path are “finalized” in that
they will never be updated in any subsequent execution of the algorithm. All
the leaves that are a descendant of the finalized node are non-empty.
4. All the right siblings (i.e., node 14) are zero-hashes, that is, 0 for leaf nodes
(at level 0), “hash(0,0)” for nodes at level 1, “hash(hash(0,0),hash(0,0))” for
nodes at level 2, and so on. These zero-hashes are constant.
Now we describe the algorithm. To represent a Merkle tree of height h, the
algorithm maintains only two arrays of length h, called branch and zero_hashes
respectively, that store the left and right siblings of a path from a new leaf node
to the root. When inserting a new data hash, the algorithm computes the path
from the new leaf node to the root. Each node of the path can be computed
in constant time, by retrieving only its left or right sibling from the branch or
zero_hashes array. After the path computation, the branch array is updated to
contain all the left siblings of a next new path that will be computed in the next
run of the algorithm. Here the branch array update is done in constant time,
since only a single element of the array needs to be updated, and the element
has already been computed as part of the path computation.6 Note that the
zero_hashes array is computed once at the very beginning when all the leaves
are empty, and never be updated during the lifetime of the Merkle tree.
Complexity. Both the time and space complexity of the algorithm is linear in
the tree height h. The space complexity is linear, because the size of the branch
and zero_hashes arrays is h, and no other nodes are stored by the algorithm.
The time complexity is also linear. For the path computation, the length of the
path is h, and each node can be computed in constant time by using the two
arrays. The branch array update can be also done in constant time as explained
earlier.
Implementation and optimization. Figure 2 shows the pseudocode implementa￾tion of the incremental Merkle tree algorithm [3] that is employed in the deposit
contract [6]. It consists of two main functions: deposit and get_deposit_root.
The deposit function takes as input a new deposit hash, and inserts it into the
Merkle tree. The get_deposit_root function computes and returns the root of
the current partial Merkle tree whose leaves are filled with the deposit hashes
received up to that point.
Specifically, the deposit function fills the first (leftmost) empty leaf node
with a given deposit hash, and updates a single element of the branch array.
The get_deposit_root function computes the tree root by traversing a path
from the last (rightmost) non-empty leaf to the root.
As an optimization, the deposit function does not fully compute the path
from the leaf to the root, but computes only a smaller partial path from the
6 See Appendix A for more details about updating the branch array.6 Runtime Verification, Inc.
1 # globals
2 zero_hashes: int[TREE_HEIGHT] = {0} # zero array
3 branch: int[TREE_HEIGHT] = {0} # zero array
4 deposit_count: int = 0 # max: 2^TREE_HEIGHT - 1
5
6 fun constructor() -> unit:
7 i: int = 0
8 while i < TREE_HEIGHT - 1:
9 zero_hashes[i+1] = hash(zero_hashes[i], zero_hashes[i])
10 i += 1
11
12 fun deposit(value: int) -> unit:
13 assert deposit_count < 2^TREE_HEIGHT - 1
14 deposit_count += 1
15 size: int = deposit_count
16 i: int = 0
17 while i < TREE_HEIGHT:
18 if size % 2 == 1:
19 break
20 value = hash(branch[i], value)
21 size /= 2
22 i += 1
23 branch[i] = value
24
25 fun get_deposit_root() -> int:
26 root: int = 0
27 size: int = deposit_count
28 h: int = 0
29 while h < TREE_HEIGHT:
30 if size % 2 == 1: # size is odd
31 root = hash(branch[h], root)
32 else: # size is even
33 root = hash(root, zero_hashes[h])
34 size /= 2
35 h += 1
36 return root
Fig. 2. Pseudocode implementation of the incremental Merkle tree algorithm employed
in the deposit contract [6].End-to-End Formal Verification of Ethereum 2.0 Deposit Smart Contract 7
leaf to the node that is needed to update the branch array. Indeed, for all odd￾numbered deposits (i.e., 1st deposit, 3rd deposit, · · ·), such a partial path is
empty, because the leaf node is the one needed for the branch array update.
In that case, the deposit function returns immediately in constant time. For
even-numbered deposits, the partial path is not empty but still much smaller
than the full path in most cases. This optimization is useful when the tree root
computation is not needed for every single partial Merkle tree. Indeed, in many
cases, multiple deposit hashes are inserted at once, for which only the root of
the last partial Merkle tree is needed.
Correctness. Consider a Merkle tree of height h employed in the deposit contract.
Suppose that a sequence of deposit function calls are made, say deposit(v1),
deposit(v2), · · · , and deposit(vm), where m < 2
h
. Then, the function call
get_deposit_root() will return the root of the Merkle tree whose leaves are
filled with the deposit data hashes v1, v2, · · · , vm, respectively, in order from
left to right, starting from the leftmost one.
Note that the correctness statement requires the condition m < 2
h
, that is,
the rightmost leaf must be kept empty, which means that the maximum number
of deposits that can be stored in the tree using this incremental algorithm is
2
h − 1 instead of 2
h
.
The proof of the correctness is presented in Appendix A.
Remark 1. Since the deposit function reverts when deposit_count ≥ 2
TREE_HEIGHT−
1, the loop in the deposit function cannot reach the last iteration, thus the loop
bound (in line 17 of Figure 2) can be safely decreased to TREE_HEIGHT − 1.
4 Bytecode Verification of the Deposit Contract
Now we present the formal verification of the compiled bytecode of the deposit
contract. The bytecode verification ensures that the compiled bytecode is a sound
refinement of the source code. This rules out the need to trust the compiler.
As illustrated in Section 2, we first manually refined the abstract model
(in which we proved the algorithm correctness) to the bytecode specification
(Section 4.1). For the refinement, we consulted the ABI interface standard [5]
(to identify, e.g., calldata and output in the illustrating example of Section 2),
as well as the bytecode (to identify, e.g., the pc and stack information).7 Then,
we used the KEVM verifier [16] to verify the compiled bytecode against the
refined specification. We adopted the KEVM verifier to reason about all possible
corner-case behaviors of the compiled bytecode, especially those introduced by
certain unintuitive and questionable aspects of the underlying Ethereum Virtual
Machine (EVM) [26]. This was possible because the KEVM verifier is derived
7 However, we want to note that the compiler can be augmented to extract such
information, which can automate the refinement process to a certain extent. We
leave that as future work.8 Runtime Verification, Inc.
from a complete formal semantics of the EVM, called KEVM [12]. Our formal
specification and verification artifacts are publicly available at [20].
Let us elaborate on specific low-level behaviors verified against the bytecode.
In addition to executing the incremental Merkle tree algorithm, most of the
functions perform certain additional low-level tasks, and we verified that such
tasks are correctly performed. Specifically, for example, given deposit data,8
the
deposit function computes its 32-byte hash (called Merkleization) according to
the SimpleSerialize (SSZ) specification [9]. The leaves of the Merkle tree store
only the computed hashes instead of the original deposit data. The deposit
function also emits a DepositEvent log that contains the original deposit data,
where the log message needs to be encoded as a byte sequence following the
contract event ABI specification [5]. Other low-level operations performed by
those functions that we verified include: correct zero-padding for the 32-byte
alignment, correct conversions from big-endian to little-endian, input bytes of
the SHA2-256 hash function being correctly constructed, and return values being
correctly serialized to byte sequences according to the ABI specification [5].
Our formal specification includes both positive and negative behaviors. The
positive behaviors describe the desired behaviors of the contracts in a legitimate
input state. The negative behaviors, on the other hand, describe how the con￾tracts handle exceptional cases (e.g., when benign users feed invalid inputs by
mistake, or malicious users feed crafted inputs to take advantage of the con￾tracts). The negative behaviors are mostly related to security properties.
4.1 Summary of Bytecode Specification
We summarize the formal specification of the deposit contract bytecode that we
verified. The full specification can be found at [19].
Constructor function constructor updates the storage as follows:
zero_hashes[i] ← ZH(i) for all 1 ≤ i < 32
where ZH(i) is a 32-byte word that is recursively defined as follows:
ZH(i + 1) = hash(ZH(i) ++ ZH(i)) for 0 ≤ i < 31
ZH(0) = 0
where hash denotes the SHA2-256 hash function, and ++ denotes the byte con￾catenation.
Function get_deposit_count returns LE64(deposit_count), where LE64(x)
denotes the 64-bit little-endian representation of x (for 0 ≤ x < 2
64). That
is, for a given x = Σ0≤i<8(ai
· 256i
), LE64(x) = Σ0≤i<8(a7−i
· 256i
), where
0 ≤ ai < 256. Note that LE64(deposit_count) is always defined because of the
contract invariant of deposit_count < 2
32. This function does not alter the
storage state.
8 Each deposit data consists of the public key, the withdrawal credentials, the deposit
amount, and the signature of the deposit owner.8 Runtime Verification, Inc.
from a complete formal semantics of the EVM, called KEVM [12]. Our formal
specification and verification artifacts are publicly available at [20].
Let us elaborate on specific low-level behaviors verified against the bytecode.
In addition to executing the incremental Merkle tree algorithm, most of the
functions perform certain additional low-level tasks, and we verified that such
tasks are correctly performed. Specifically, for example, given deposit data,8
the
deposit function computes its 32-byte hash (called Merkleization) according to
the SimpleSerialize (SSZ) specification [9]. The leaves of the Merkle tree store
only the computed hashes instead of the original deposit data. The deposit
function also emits a DepositEvent log that contains the original deposit data,
where the log message needs to be encoded as a byte sequence following the
contract event ABI specification [5]. Other low-level operations performed by
those functions that we verified include: correct zero-padding for the 32-byte
alignment, correct conversions from big-endian to little-endian, input bytes of
the SHA2-256 hash function being correctly constructed, and return values being
correctly serialized to byte sequences according to the ABI specification [5].
Our formal specification includes both positive and negative behaviors. The
positive behaviors describe the desired behaviors of the contracts in a legitimate
input state. The negative behaviors, on the other hand, describe how the con￾tracts handle exceptional cases (e.g., when benign users feed invalid inputs by
mistake, or malicious users feed crafted inputs to take advantage of the con￾tracts). The negative behaviors are mostly related to security properties.
4.1 Summary of Bytecode Specification
We summarize the formal specification of the deposit contract bytecode that we
verified. The full specification can be found at [19].
Constructor function constructor updates the storage as follows:
zero_hashes[i] ← ZH(i) for all 1 ≤ i < 32
where ZH(i) is a 32-byte word that is recursively defined as follows:
ZH(i + 1) = hash(ZH(i) ++ ZH(i)) for 0 ≤ i < 31
ZH(0) = 0
where hash denotes the SHA2-256 hash function, and ++ denotes the byte con￾catenation.
Function get_deposit_count returns LE64(deposit_count), where LE64(x)
denotes the 64-bit little-endian representation of x (for 0 ≤ x < 2
64). That
is, for a given x = Σ0≤i<8(ai
· 256i
), LE64(x) = Σ0≤i<8(a7−i
· 256i
), where
0 ≤ ai < 256. Note that LE64(deposit_count) is always defined because of the
contract invariant of deposit_count < 2
32. This function does not alter the
storage state.
8 Each deposit data consists of the public key, the withdrawal credentials, the deposit
amount, and the signature of the deposit owner.10 Runtime Verification, Inc.
Table 1. Gas analysis of the deposit contract. The total gas cost is the summation of
the tx cost, the opcode cost, and the memory cost. The multiple rows denote the lower
and upper bounds of the gas cost.
Total Gas Cost Tx Cost Opcode Cost Memory Cost
constructor 770,144 53,000 716,470 674
supportsInterface 21,461 21,192 254 15
21,526 21,240 271 15
get_deposit_count 23,278 21,064 2,187 27
get_deposit_root 98,259 21,064 76,849 346
98,579 21,064 77,169 346
deposit 46,162 22,812 23,252 98
157,642 25,308 131,929 405
4.2 Gas Analysis
Table 1 shows the theoretical gas analysis of the deposit contract, essentially
summarizing the gas cost description of the bytecode specification [19]. It pro￾vides the lower and upper bounds of the gas cost for each external function,
meaning that the amount of gas consumed by any possible execution falls within
the bounds.11 Each gas cost is broken down into three categories: the tx cost
(the base fee of sending a transaction, plus the calldata fee), the opcode cost
(the gas fee of executing instructions), and the memory cost (the gas fee of using
memory). The calldata fee is calculated based on the number of zero and non￾zero bytes in the given calldata, where the gas fee per zero byte is smaller than
that of non-zero byte. Thus the calldata fee varies for the same function unless
the function takes no argument.
– The gas cost of the constructor function is constant because it is executed
only once when the contract is deployed, taking no argument.
– The gas cost of the get_deposit_count functions is also constant, because
it has only a single execution path and it takes no argument.12
– The supportsInterface function has two execution paths due to the short￾circuit evaluation of the logical-or (||) operation, causing two different op￾code costs. The calldata fee varies depending on the number of zero and
non-zero bytes in the given argument—the minimum is when the argument
is 0x00000000 and the maximum is when it is, e.g., 0x01010101.
– The get_deposit_root function has 2
32 execution paths, because of the
if branches inside the loop of 32 iterations. The minimum opcode cost is
when only the else branch is taken in all iterations of the loop (i.e., when
deposit_count = 0), and the maximum is when only the then branch is
taken (i.e., when deposit_count = 232 − 1).
11 For the simplicity, here we consider only the normal execution that does not revert
in the middle. The gas cost for reverting executions is smaller than that of normal
executions.
12 We assume that the calldata is well-formed, and does not include any garbage bytes.End-to-End Formal Verification of Ethereum 2.0 Deposit Smart Contract 11
– The deposit function has 32 (non-reverting) execution paths, because of
the conditional return statement inside the loop of 32 iterations. Both the
opcode and memory costs are proportional to the number of loop iterations.13
Another factor for the opcode cost is the variable gas fee of the storage
update [25]. The calldata cost varies depending on the arguments.
5 Discussion
The deposit contract had been initially written in Vyper, and the Vyper-compiled
bytecode had been formally verified [21]. The Vyper language was chosen because
of the security-first language design that prioritizes security and auditability over
efficiency and convenience for developers. However, the community concerns [4,2]
about the current Vyper compiler motivated the deposit contract developers to
reimplement the contract in Solidity with re-verifying the new Solidity-compiled
bytecode. The underlying motivation is primarily based on the maturity of the
Solidity compiler and its wider adoption in the Ethereum community than the
Vyper compiler.
To minimize the overhead of reimplementation followed by re-verification of
the new deposit contract, the Solidity implementation is designed to keep as
much as possible the overall structure of the Vyper implementation. However,
the Solidity implementation still made several substantial improvements over the
original Vyper implementation that do not change the original interface of the
contract and incur no significant overhead in the re-verification. The improve￾ments made in the Solidity implementation include the following:
– The to_little_endian_64 function has been rewritten to improve both
readability and performance. While in the Vyper implementation, a loop was
employed to reorder the (least significant) eight bytes of the argument, in
the Solidity implementation, the loop has been unrolled while the reordering
operation is succinctly and intuitively represented by converting the integer
argument to a byte array. (See Figure 3 for the comparison of the two imple￾mentations.) Moreover, with the optimization enabled, the Solidity compiler
has been able to generate quite gas-efficient bytecode from the source code
that involves multiple byte array accesses.
– The EIP-165 [18] support has been added. The EIP-165 is a protocol for
identifying the interfaces implemented in a given contract. Supporting the
EIP-165 protocol, the new deposit contract allows other Dapps to query
what interfaces it implements.
– Message strings for the require statements have been added to help users to
easily identify the reasons for failures that may occur while using the deposit
contract.
– An additional require statement has been added to ensure that the deposit
amount is a multiple of Gwei (109 wei). The unit of deposit is Gwei and
13 The memory cost can be made constant by improving the memory footprint [15].12 Runtime Verification, Inc.
function
to_little_endian_64(... val) ... {
ret = new bytes(8);
bytes8 bytesValue = bytes8(val);
ret[0] = bytesValue[7];
ret[1] = bytesValue[6];
ret[2] = bytesValue[5];
ret[3] = bytesValue[4];
ret[4] = bytesValue[3];
ret[5] = bytesValue[2];
ret[6] = bytesValue[1];
ret[7] = bytesValue[0];
}
def
to_little_endian_64(val: ...) ...:
y: uint256 = 0
x: uint256 = val
for _ in range(8):
y = shift(y, 8)
y = y + bitwise_and(x, 255)
x = shift(x, -8)
return slice(convert(y ...) ...)
Fig. 3. Comparison of the to_little_endian_64 function in the Solidity (left) and
Vyper (right) implementations.
the incoming value is divided by it, the remainder being lost. The additional
require statement avoids lost remainders due to mistakes made by users.14
– As mentioned in Remark 1, the contract contains an infeasible (dead) path.
As a defensive programming measure, an assert(false) statement has been
added at the end of the infeasible path.
Despite the additional runtime checks, the Solidity implementation is more
gas-efficient (30∼40% less gas cost) than the original Vyper implementation,
thanks to the advanced code optimization of the Solidity compiler. The Solidity￾generated bytecode, however, still has room to further improve the gas efficiency,
e.g., duplicate address sanitation [1] and non-optimal memory footprint in the
loop [15]. Nonetheless, the gas inefficiency issues are not critical, and do not
compromise the safety of the contract.
Overall, the Solidity implementation and the compiled bytecode are in a bet￾ter shape than the original Vyper implementation, thanks to the aforementioned
improvements, and the advantage of the Solidity compiler.
6 Conclusion
We reported our end-to-end formal verification of the Ethereum 2.0 deposit
contract. We adopted the refinement-based verification approach to ensure the
end-to-end correctness of the contract while minimizing the verification effort.
Specifically, we first proved that the incremental Merkle tree algorithm is cor￾rectly implemented in the contract, and then verified that the compiled bytecode
is correctly generated from the source code. No critical bugs were found in the
14 In the current gas price, the gas cost for the additional require statement is more
than the potentially saved amount of ether (≤ 1 Gwei). However, the gas price may
change in the future, and more importantly, the additional check allows users to
detect early any potential mistake, which has its own value.End-to-End Formal Verification of Ethereum 2.0 Deposit Smart Contract 13
Solidity implementation of the deposit contract.15 We conclude that the latest
deposit contract (2ac62c4) will behave as expected—as specified in the formal
specification [19], provided that the contract is compiled by the Solidity compiler
v0.6.8 with the optimization enabled (–optimize-runs 5000000).
Trust base. The validity of the bytecode verification result assumes the cor￾rectness of the bytecode specification and the KEVM verifier. The algorithm
correctness proof is partially mechanized—only the proof of major lemmas are
mechanized in the K framework. The non-mechanized proofs are included in our
trust base. The Solidity compiler is not in the trust base.
15 Several issues found in the previous Vyper implementation have already been fixed
in the Solidity reimplementation. See [21] for more details of the previous findings.14 Runtime Verification, Inc.
References
1. Beregszaszi, A.: Solidity Issue 8670. https://github.com/ethereum/solidity/
issues/8670
2. Bugrara, S.: A Review of the Deposit Contract. https://docs.google.com/
document/d/1sbsMvngo0kBfzy8fbhJ55kl6F1tUxmECMUuK75F5H9E
3. Buterin, V.: Progressive Merkle Tree. https://github.com/ethereum/research/
blob/master/beacon_chain_impl/progressive_merkle_tree.py
4. ConsenSys Diligence: Vyper Security Review. https://diligence.consensys.
net/audits/2019/10/vyper/
5. Ethereum Foundation: Contract ABI Specification. https://solidity.
readthedocs.io/en/v0.6.1/abi-spec.html
6. Ethereum Foundation: Ethereum 2.0 Deposit Contract. https://github.com/
axic/eth2-deposit-contract/blob/r1/deposit_contract.sol
7. Ethereum Foundation: Ethereum 2.0 Specifications. https://github.com/
ethereum/eth2.0-specs
8. Ethereum Foundation: Ethereum Foundation Spring 2019 Update. https://blog.
ethereum.org/2019/05/21/ethereum-foundation-spring-2019-update/
9. Ethereum Foundation: SimpleSerialize (SSZ). https://github.com/ethereum/
eth2.0-specs/tree/master/ssz
10. Ethereum Foundation: Solidity. https://solidity.readthedocs.io
11. Ethereum Foundation: Vyper. https://vyper.readthedocs.io
12. Hildenbrandt, E., Saxena, M., Zhu, X., Rodrigues, N., Daian, P., Guth, D., Moore,
B., Zhang, Y., Park, D., Ştefănescu, A., Roşu, G.: Kevm: A complete semantics of
the ethereum virtual machine. In: Proceedings of the 31st IEEE Computer Security
Foundations Symposium. CSF 2018 (2018)
13. Merkle, R.C.: A digital signature based on a conventional encryption function. In:
A Conference on the Theory and Applications of Cryptographic Techniques on
Advances in Cryptology. CRYPTO ’87 (1988)
14. NIST: Perfect Binary Tree. https://xlinux.nist.gov/dads/HTML/
perfectBinaryTree.html
15. Park, D.: Solidity Issue 9046. https://github.com/ethereum/solidity/issues/
9046
16. Park, D., Zhang, Y., Saxena, M., Daian, P., Roşu, G.: A Formal Verification Tool
for Ethereum VM Bytecode. In: Proceedings of the 26th ACM Joint European
Software Engineering Conference and Symposium on the Foundations of Software
Engineering. ESEC/FSE 2018 (2018)
17. Podelski, A., Rybalchenko, A.: Transition invariants. In: Proceedings of the 19th
Annual IEEE Symposium on Logic in Computer Science. LICS 2004 (2004)
18. Reitwießner, C., Johnson, N., Vogelsteller, F., Baylina, J., Feldmeier, K., Entriken,
W.: ERC-165 Standard Interface Detection. https://github.com/ethereum/
EIPs/blob/master/EIPS/eip-165.md
19. Runtime Verification, Inc.: Formal Verification of Deposit Contract Bytecode.
https://github.com/runtimeverification/deposit-contract-verification/
tree/master/bytecode-verification
20. Runtime Verification, Inc.: Formal Verification of Ethereum 2.0 Deposit Contract.
https://github.com/runtimeverification/deposit-contract-verification
21. Runtime Verification, Inc.: Formal Verification of Ethereum 2.0 Deposit
Contract (Vyper Implementation). https://github.com/runtimeverification/
verified-smart-contracts/tree/master/depositEnd-to-End Formal Verification of Ethereum 2.0 Deposit Smart Contract 15
22. Serbanuta, T., Arusoaie, A., Lazar, D., Ellison, C., Lucanu, D., Rosu, G.: The K
primer (version 3.3). Electr. Notes Theor. Comput. Sci. 304, 57–80 (2014)
23. Stefanescu, A., Ciobaca, S., Mereuta, R., Moore, B.M., Serbanuta, T., Rosu, G.:
All-Path Reachability Logic. Logical Methods in Computer Science 15(2) (2019)
24. Stefanescu, A., Park, D., Yuwen, S., Li, Y., Rosu, G.: Semantics-Based Program
Verifiers for All Languages. In: Proceedings of the 2016 ACM SIGPLAN Inter￾national Conference on Object-Oriented Programming, Systems, Languages, and
Applications. OOPSLA 2016 (2016)
25. Tang, W.: EIP 2200: Structured Definitions for Net Gas Metering. https://eips.
ethereum.org/EIPS/eip-2200
26. Wood, G.: Ethereum: A Secure Decentralised Generalised Transaction Ledger.
https://ethereum.github.io/yellowpaper/paper.pdf16 Runtime Verification, Inc.
A Formalization and Correctness Proof of the
Incremental Merkle Tree Algorithm
We formalize the incremental Merkle tree algorithm [3], especially the one em￾ployed in the deposit contract [6], and prove its correctness w.r.t. the original
full-construction Merkle tree algorithm [13].
Notations. Let T be a perfect binary tree [14] (i.e., every node has exactly two
child nodes) of height h, and T(l, i) denote its node at level l and index i, where
the level of leaves is 0, and the index of the left-most node is 1. For example, if
h = 2, then T(2, 1) denotes the root whose children are T(1, 1) and T(1, 2), and
the leaves are denoted by T(0, 1), T(0, 2), T(0, 3), and T(0, 4), as follows:
T(0,1) T(0,2) T(0,3) T(0,4)
T(1,1) T(1,2)
T(2,1)
We write [[T(l, i)]] to denote the value of the node T(l, i), but we omit [[·]] when
the meaning is clear in the context.
Let us define two functions, ↑ and , as follows:
↑ x = dx/2e (1)
 x = bx/2c (2)
Moreover, let us define ↑
k x =↑ (↑
k−1 x) for k ≥ 2, ↑
1 x =↑ x, and ↑
0 x = x. Let
{T(k, ↑
k x)}
h
k=0 be a path {T(0, ↑
0 x), T(1, ↑
1 x), T(2, ↑
2 x), · · · , T(h, ↑
h x)}. We
write {T(k, ↑
k x)}k if h is clear in the context. Let us define 
k and {T(k, 
k x)}k
similarly. For the presentation purpose, let T(l, 0) denote a dummy node which
has the parent T(l + 1, 0) and the children T(l − 1, 0) and T(l − 1, 1). Note that,
however, these dummy nodes are only conceptual, allowing the aforementioned
paths to be well-defined, but not part of the tree at all.
In this notation, for a non-leaf, non-root node of index i, its left child index
is 2i − 1, its right child index is 2i, and its parent index is ↑ i. Also, note that
{T(k, ↑
k m)}k is the path starting from the m-th leaf going all the way up to
the root.
First, we show that two paths {T(k, ↑
k x)}k and {T(k, 
k
(x − 1))}k are
parallel with a “distance” of 1.
Lemma 1. For all x ≥ 1, and k ≥ 0, we have:
(↑
k x) − 1 =
k
(x − 1) (3)
Proof. Let us prove by induction on k. When k = 0, we have (↑
0 x) − 1 =
x − 1 =
0
(x − 1). When k = 1, we have two cases:End-to-End Formal Verification of Ethereum 2.0 Deposit Smart Contract 17
– When x is odd, that is, x = 2y + 1 for some y ≥ 0:
(↑ x) − 1 = (↑ (2y + 1)) − 1 = 
2y + 1
2

− 1 = y =

2y
2

= 2y = (x − 1)
– When x is even, that is, x = 2y for some y ≥ 1:
(↑ x)−1 = (↑ 2y)−1 = 
2y
2

−1 = y−1 = 
2y − 1
2

= (2y−1) = (x−1)
Thus, we have:
(↑ x) − 1 = (x − 1) (4)
Now, assume that (3) holds for some k = l ≥ 1. Then,
↑
l+1 x =↑ (↑
l x) (By the definition of ↑
k
)
=↑ ((
l
(x − 1)) + 1) (By the assumption)
= ( (
l
(x − 1))) + 1 (By Equation 4)
=
l+1 (x − 1) + 1 (By the definition of 
k
)
which concludes.
Now let us define the Merkle tree.
Definition 1. A perfect binary tree T of height h is a Merkle tree [13], if the
leaf node contains data, and the non-leaf node’s value is the hash of its children’s,
i.e.,
∀0 < l ≤ h. ∀0 < i ≤ 2
h−l
. T(l, i) = hash(T(l − 1, 2i − 1), T(l − 1, 2i)) (5)
Let Tm be a partial Merkle tree up-to m whose first m leaves contain data
and the other leaves are zero, i.e.,
Tm(0, i) = 0 for all m < i ≤ 2
h
(6)
Let Z be the zero Merkle tree whose leaves are all zero, i.e., Z(0, i) = 0 for
all 0 < i ≤ 2
h
. That is, Z = T0. Since all nodes at the same level have the same
value in Z, we write Z(l) to denote the value at the level l, i.e., Z(l) = Z(l, i)
for any 0 < i ≤ 2
h−l
.
Now we formulate the relationship between the partial Merkle trees. Given
two partial Merkle trees Tm−1 and Tm, if their leaves agree up-to m − 1, then
they only differ on the path {Tm(k, ↑
k m)}k. This is formalized in Lemma 2.
Lemma 2. Let Tm be a partial Merkle tree up-to m > 0 of height h, and let
Tm−1 be another partial Merkle tree up-to m − 1 of the same height. Suppose
their leaves agree up to m−1, that is, Tm−1(0, i) = Tm(0, i) for all 1 ≤ i ≤ m−1.
Then, for all 0 ≤ l ≤ h, and 1 ≤ i ≤ 2
h−l
,
Tm−1(l, i) = Tm(l, i) when i 6=↑
l m (7)End-to-End Formal Verification of Ethereum 2.0 Deposit Smart Contract 17
– When x is odd, that is, x = 2y + 1 for some y ≥ 0:
(↑ x) − 1 = (↑ (2y + 1)) − 1 = 
2y + 1
2

− 1 = y =

2y
2

= 2y = (x − 1)
– When x is even, that is, x = 2y for some y ≥ 1:
(↑ x)−1 = (↑ 2y)−1 = 
2y
2

−1 = y−1 = 
2y − 1
2

= (2y−1) = (x−1)
Thus, we have:
(↑ x) − 1 = (x − 1) (4)
Now, assume that (3) holds for some k = l ≥ 1. Then,
↑
l+1 x =↑ (↑
l x) (By the definition of ↑
k
)
=↑ ((
l
(x − 1)) + 1) (By the assumption)
= ( (
l
(x − 1))) + 1 (By Equation 4)
=
l+1 (x − 1) + 1 (By the definition of 
k
)
which concludes.
Now let us define the Merkle tree.
Definition 1. A perfect binary tree T of height h is a Merkle tree [13], if the
leaf node contains data, and the non-leaf node’s value is the hash of its children’s,
i.e.,
∀0 < l ≤ h. ∀0 < i ≤ 2
h−l
. T(l, i) = hash(T(l − 1, 2i − 1), T(l − 1, 2i)) (5)
Let Tm be a partial Merkle tree up-to m whose first m leaves contain data
and the other leaves are zero, i.e.,
Tm(0, i) = 0 for all m < i ≤ 2
h
(6)
Let Z be the zero Merkle tree whose leaves are all zero, i.e., Z(0, i) = 0 for
all 0 < i ≤ 2
h
. That is, Z = T0. Since all nodes at the same level have the same
value in Z, we write Z(l) to denote the value at the level l, i.e., Z(l) = Z(l, i)
for any 0 < i ≤ 2
h−l
.
Now we formulate the relationship between the partial Merkle trees. Given
two partial Merkle trees Tm−1 and Tm, if their leaves agree up-to m − 1, then
they only differ on the path {Tm(k, ↑
k m)}k. This is formalized in Lemma 2.
Lemma 2. Let Tm be a partial Merkle tree up-to m > 0 of height h, and let
Tm−1 be another partial Merkle tree up-to m − 1 of the same height. Suppose
their leaves agree up to m−1, that is, Tm−1(0, i) = Tm(0, i) for all 1 ≤ i ≤ m−1.
Then, for all 0 ≤ l ≤ h, and 1 ≤ i ≤ 2
h−l
,
Tm−1(l, i) = Tm(l, i) when i 6=↑
l m (7)18 Runtime Verification, Inc.
Proof. Let us prove by induction on l. When l = 0, we immediately have
Tm−1(0, i) = Tm(0, i) for any i 6= m by the premise and Equation 6. Now, assume
that (7) holds for some l = k. Then by Equation 5, we have Tm−1(k + 1, i) =
Tm(k + 1, i) for any i 6=↑ (↑
k m) =↑
k+1 m, which concludes.
Corollary 1 induces a linear-time incremental Merkle tree insertion algo￾rithm [3].
Corollary 1. Tm can be constructed from Tm−1 by computing only {Tm(k, ↑
k
m)}k, the path from the new leaf, Tm(0, m), to the root.
Proof. By Lemma 2.
Let us formulate more properties of partial Merkle trees.
Lemma 3. Let Tm be a partial Merkle tree up-to m of height h, and Z be the
zero Merkle tree of the same height. Then, for all 0 ≤ l ≤ h, and 1 ≤ i ≤ 2
h−l
,
Tm(l, i) = Z(l) when i >↑
l m (8)
Proof. Let us prove by induction on l. When l = 0, we immediately have
Tm(0, i) = Z(0) = 0 for any m < i ≤ 2
h by Equation 6. Now, assume that
(8) holds for some 0 ≤ l = k < h. First, for any i ≥ (↑
k+1 m) + 1, we have:
2i − 1 ≥ (2 ↑
k+1 m) + 1 = 2 
↑
k m
2

+ 1 ≥ 2
↑
k m
2
+ 1 = (↑
k m) + 1 (9)
Then, for any ↑
k+1 m < i ≤ 2
h−(k+1), we have:
Tm(k + 1, i) = hash(Tm(k, 2i − 1), Tm(k, 2i)) (By Equation 5)
= hash(Z(k), Z(k)) (By Equations 8 and 9)
= Z(k + 1) (By the definition of Z)
which concludes.
Lemma 4 induces a linear-space incremental Merkle tree insertion algorithm.
Lemma 4. A path {Tm(k, ↑
k m)}k can be computed by using only two other
paths, {Tm−1(k, 
k
(m − 1)))}k and {Z(k)}k.
Proof. We will construct the path from the leaf, Tm(0, m), which is given. Sup￾pose we have constructed the path up to Tm(q, ↑
q m) for some q > 0 by using
only two other sub-paths, {Tm−1(k, 
k
(m − 1))}
q−1
k=0 and {Z(k)}
q−1
k=0. Then, to
construct Tm(q + 1, ↑
q+1 m), we need the sibling of Tm(q, ↑
q m), where we have
two cases:
– Case (↑
q m) is odd. Then, we need the right-sibling Tm(q,(↑
q m) + 1), which
is Z(q) by Lemma 3.End-to-End Formal Verification of Ethereum 2.0 Deposit Smart Contract 19
– Case (↑
q m) is even. Then, we need the left-sibling Tm(q,(↑
q m) − 1), which
is Tm(q, 
q
(m − 1)) by Lemma 1, which is in turn Tm−1(q, 
q
(m − 1)) by
Lemma 2.
By the mathematical induction on k, we conclude.
Lemma 5. Let h = TREE_HEIGHT. For any integer 0 ≤ m < 2
h
, the two paths
{Tm(k, 
k m)}k and {Tm+1(k, 
k
(m+1))}k always converge, that is, there exists
unique 0 ≤ l ≤ h such that:
(
k m) + 1 =
k
(m + 1) is even for all 0 ≤ k < l (10)
(
k m) + 1 =
k
(m + 1) is odd for k = l (11)

k m =
k
(m + 1) for all l < k ≤ h (12)
Tm(k, 
k m) = Tm+1(k, 
k
(m + 1)) for all l < k ≤ h (13)
Proof. Equation 12 follows from Equation 11, since for an odd integer x,  (x −
1) = x. Also, Equation 13 follows from Lemma 2, since ↑
k
(m + 1) = (
k
m) + 1 6=
k m =
k
(m + 1) by Lemma 1 and Equation 12. Thus, we only need
to prove the unique existence of l satisfying (10) and (11). The existence of l is
obvious since 1 ≤ m + 1 ≤ 2
h
, and one can find the smallest l satisfying (10)
and (11). Now, suppose there exist two different l1 < l2 satisfying (10) and (11).
Then, 
l1 (m + 1) is odd since l1 satisfies (11), while 
l1 (m + 1) is even since l2
satisfies (10), which is a contradiction, thus l is unique, and we conclude.
A.1 Pseudocode
Figure 2 shows the pseudocode of the incremental Merkle tree algorithm [3]
that is employed in the deposit contract [6]. It maintains a global counter
deposit_count to keep track of the number of deposits made, and two global ar￾rays zero_hashes and branch, which corresponds to Z (Definition 1) and a cer￾tain part of {Tm(k, 
k m)}k, where m denotes the value of deposit_count. The
constructor function is called once at the beginning to initialize zero_hashes
which is never updated later. The deposit function inserts a given new leaf
value in the tree by incrementing deposit_count and updating only a single
element of branch. The get_deposit_root function computes the root of the
current partial Merkle tree Tm.
Since the loops are bounded to the tree height and the size of global arrays
is equal to the tree height, it is clear that both time and space complexities of
the algorithm are linear.
A.2 Correctness Proof
Now we prove the correctness of the incremental Merkle tree algorithm shown
in Figure 2.20 Runtime Verification, Inc.
branch
zero_hashes
updated
untouched
T4(0, "0 4)
T4(1, "1 4)
T4(2, "2 4)
T4(3, "3 4)
Z(0) Z(0) Z(0) Z(0)
Z(1)
Z(2)
Z(1)
T4(0, ￾0 3)
T4(1, ￾1 3)
Fig. 4. A partial Merkle tree T4 of height 3, illustrating the incremental Merkle tree
algorithm shown in Figure 2, where TREE_HEIGHT = 3. The bold-lined nodes correspond
to the branch array. The bold-dotted-lined nodes correspond to the zero_hashes array.
The get_deposit_root function computes the gray nodes by using only the bold￾lined nodes (i.e., branch) and the bold-dotted-lined nodes (i.e., zero_hashes), where
deposit_count = 4.
Theorem 1 (Correctness of Incremental Merkle Tree Algorithm). Sup￾pose that the constructor function is executed at the beginning, followed by
a sequence of deposit function calls, say deposit(v1), deposit(v2), · · · , and
deposit(vm), where m < 2
TREE_HEIGHT. Then, the function call get_deposit_root()
will return the root of the partial Merkle tree Tm such that Tm(0, i) = vi for all
1 ≤ i ≤ m.
Proof. By Lemmas 6, 7, 8, and 9.
Note that the correctness theorem requires the condition m < 2
h
, where h
is the tree height, that is, the rightmost leaf must be kept empty, which means
that the maximum number of deposits that can be stored in the tree using this
incremental algorithm is 2
h − 1 instead of 2
h
.
Lemma 6 (init). Once init is executed, zero_hashes denotes Z, that is,
zero_hashes[k] = Z(k) (14)
for 0 ≤ k < TREE_HEIGHT.
Proof. By the implementation of init and the definition of Z in Definition 1.
Lemma 7 (deposit). Suppose that, before executing deposit, we have:
deposit_count = m < 2
TREE_HEIGHT − 1 (15)
branch[k] = Tm(k, 
k m) if 
k m is odd (16)20 Runtime Verification, Inc.
branch
zero_hashes
updated
untouched
T4(0, "0 4)
T4(1, "1 4)
T4(2, "2 4)
T4(3, "3 4)
Z(0) Z(0) Z(0) Z(0)
Z(1)
Z(2)
Z(1)
T4(0, ￾0 3)
T4(1, ￾1 3)
Fig. 4. A partial Merkle tree T4 of height 3, illustrating the incremental Merkle tree
algorithm shown in Figure 2, where TREE_HEIGHT = 3. The bold-lined nodes correspond
to the branch array. The bold-dotted-lined nodes correspond to the zero_hashes array.
The get_deposit_root function computes the gray nodes by using only the bold￾lined nodes (i.e., branch) and the bold-dotted-lined nodes (i.e., zero_hashes), where
deposit_count = 4.
Theorem 1 (Correctness of Incremental Merkle Tree Algorithm). Sup￾pose that the constructor function is executed at the beginning, followed by
a sequence of deposit function calls, say deposit(v1), deposit(v2), · · · , and
deposit(vm), where m < 2
TREE_HEIGHT. Then, the function call get_deposit_root()
will return the root of the partial Merkle tree Tm such that Tm(0, i) = vi for all
1 ≤ i ≤ m.
Proof. By Lemmas 6, 7, 8, and 9.
Note that the correctness theorem requires the condition m < 2
h
, where h
is the tree height, that is, the rightmost leaf must be kept empty, which means
that the maximum number of deposits that can be stored in the tree using this
incremental algorithm is 2
h − 1 instead of 2
h
.
Lemma 6 (init). Once init is executed, zero_hashes denotes Z, that is,
zero_hashes[k] = Z(k) (14)
for 0 ≤ k < TREE_HEIGHT.
Proof. By the implementation of init and the definition of Z in Definition 1.
Lemma 7 (deposit). Suppose that, before executing deposit, we have:
deposit_count = m < 2
TREE_HEIGHT − 1 (15)
branch[k] = Tm(k, 
k m) if 
k m is odd (16)22 Runtime Verification, Inc.
Thus, the loop invariant holds at the beginning of the (i + 1)th iteration as well,
and we conclude.
Lemma 8 (Contract Invariant). Let m = deposit_count. Then, once init
is executed, the following contract invariant holds. For all 0 ≤ k < TREE_HEIGHT,
1. zero_hashes[k] = Z(k)
2. branch[k] = Tm(k, 
k m) if 
k m is odd
3. deposit_count ≤ 2
TREE_HEIGHT − 1
Proof. Let us prove each invariant item.
1. By Lemma 6, and the fact that zero_hashes is updated by only init.
2. By Lemma 7, and the fact that branch is updated by only deposit.
3. By the assertion of deposit (at line 13 of Figure 2), and the fact that
deposit_count is updated by only deposit.
Lemma 9 (get_deposit_root). The get_deposit_root function computes the
path {Tm(k, ↑
k
(m + 1))}k and returns the root Tm(h, 1), given a Merkle tree Tm
of height h, that is, deposit_count = m < 2
h and TREE_HEIGHT = h when
get_deposit_root is invoked.
Proof. We claim the following loop invariant at line 29, which suffices to conclude
the main claim.
h = k where 0 ≤ k ≤ h
size =
k m
root = Tm(k, ↑
k
(m + 1))
Now let us prove the above loop invariant claim by the mathematical induction
on k. The base case (k = 0) is trivial, since 
0 m = m, ↑
0
(m + 1) = m + 1,
and Tm(0, m + 1) = 0 by Definition 1. Assume that the loop invariant holds
for some k = l. Let h
0
, size0
, and root0 denote the values at the next iteration
k = l + 1. Obviously, we have h
0 = l + 1 and size0 =
l+1 m. Also, we have
(
l m) + 1 =↑
l
(m + 1) by Lemma 1. Now, we have two cases:
– Case size =
l m is odd. Then, ↑
l
(m + 1) is even. Thus,
Tm(l + 1, ↑
l+1 (m + 1)) = hash(Tm(l, 
l m), Tm(l, ↑
l
(m + 1)))
= hash(branch[l], root) (by Lemma 8)
= root0
– Case size =
l m is even. Then, ↑
l
(m + 1) is odd. Thus,
Tm(l + 1, ↑
l+1 (m + 1)) = hash(Tm(l, ↑
l
(m + 1)), Tm(l,(↑
l
(m + 1)) + 1))
= hash(root, Z(l)) (by Lemma 3)
= hash(root, zero_hashes[l]) (by Lemma 8)
= root0
Thus, we have root0 = Tm(l + 1, ↑
l+1 (m + 1)), which concludes.End-to-End Formal Verification of Ethereum 2.0 Deposit Smart Contract 23
Mechanized Proofs. The loop invariant proofs of Lemma 7 and Lemma 9 are
mechanized in the K framework, which can be found at [20].
https://ethereum.org/
