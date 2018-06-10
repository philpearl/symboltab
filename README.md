[![GoDoc](https://godoc.org/github.com/philpearl/symboltab?status.svg)](https://godoc.org/github.com/philpearl/symboltab) 
[![Build Status](https://travis-ci.org/philpearl/symboltab.svg)](https://travis-ci.org/philpearl/symboltab)


I've called this a "symbol table". It converts a string ID to an integer sequence number. The integers start at 1 and increase by 1 for each new unique string. The intention is to store a very large number of strings, so the library is light on GC. 

The idea behind the symbol table is to convert string IDs into integer IDs that can then be used for fast comparison and array/slice lookups