# Keystore
Design principles:
- Storage abstract. Keystore interfaces can be implemented with memory, file, database, etc. for storage to be useable in a variety of 
contexts. Use write through caching to maintain synchronization between in memory keys and stored keys.
- Only the Admin interface mutates the keystore, all other interfaces
are read only. 
- Client side key naming. Keystore itself doesn't impose certain key algorithims/curves be used for specific contexts, it just supports a the minimum viable set of algorithms/curves for chainlink wide use cases. Clients define a name for each key which represents
the context in which they wish to use it. 
- Common serialization/encryption for all storage types. Protobuf serialization (compact, versioned) for key material and then key material encrypted before persistence with a passphase.