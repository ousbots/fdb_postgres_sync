# FDB -> Postgres synchronization
This is a prototype of synchronizing data between FDB and postgres.  The codebase includes a full test / demo setup.

The synchronization happens with the following flow:  
Data input:  
* When new data is input, old data is read, the new data is appended, the new data is written, then the data is marked as "dirty" (single FDB transaction)

Synchronization:  
* Range read the "dirty" ids, skipping any that are marked "writing" and mark the unmarked ids as "writing" (single FDB transaction)
* For each "dirty" id:
	* Read the data (single FDB transaction)
	* Write the data to postgres
	* Read the data from FDB again (new data), then write the difference between the new data and the dirty data to FDB ("clean" data), then if the "clean" data is empty, remove the "dirty" mark from the id, and finally remove the "writing" mark (single FDB transaction)

Marking an id means writing the id as an FDB key to an FDB directory (e.g. "dirty", "writing") and clearing means removing that id key from the FDB directory.  This enables range-reading for id marks.


## Code Layout
`data.go`: The functions used to add, modify, and get the data into FDB.  
`db.go`: Database setup.  
`main.go`: Starts the "data hammer" and synchronization goroutines and checks if the results were synchronized correctly.  
`types.go`: Data type definitions and helper functions.  
`writer.go`: The main synchronization code that reads from FDB and writes to Postgres.  


# Running

```
git clone git@github.com:ousbots/postgres_fdb_sync
cd postgres_fdb_sync
go build
pushd postgres && ./setup.sh && popd && ./postgres_fdb_sync
```
