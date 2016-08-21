# Symboltab

Half of a memory-mapped persistent, perhaps transactional, symbol table written
in Go.

This is a bit of an experiment, and lacks any kind of thread-safety, locking or
transactional support.

So far we have the reverse table, so we can 

- save a string and get an index for it
- look up the string by the index.

We don't yet have the bit that lets you look up the string and find a pre-existing
index. For that bit existing key-value stores probably do an excellent job, so
I might just implement that with boltdb or goleveldb