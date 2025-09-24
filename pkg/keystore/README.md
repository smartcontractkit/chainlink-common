# Keystore
Design principles:
- Storage abstract. Keystore interfaces can be implemented with memory, file, database, etc. for storage to be useable in a variety of 
contexts. For peristed keystores, do write-through caching, that is writes update the in memory cache and reads only read from the in memory cache. 
- Client side key naming. Keystore itself doesn't impose certain key algorithims/curves be used for specific contexts, it just supports a the minimum viable set of algorithms/curves for chainlink wide use cases. Clients define a name for each key which represents
the context in which they wish to use it. 
- Protobuf serialization (compact, versioned) for key material and then key material encrypted before persistence with a passphase.
Same serialization + encryption method can be used for any persistence approach.

Note EVM directory just a demo of how to build atop the base layer. Would be moved to 
chainlink-evm.
