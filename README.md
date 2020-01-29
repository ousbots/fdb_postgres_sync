# FDB -> Postgres synchronization
This is a prototype of synchronizing data between FDB and postgres.  The codebase includes a full test / demo setup.


The synchronization happens via the following high-level flow:

Data input:  
* (single FDB transaction) Read the old data, append the new data, then write the combined data, then mark the data as “dirty”

Synchronization:  
* (single FDB transaction) Range read the “dirty” ids, discarding any ids marked “write” and marking the remaining ids as “write”
* For each “dirty” id:
	* (single FDB transaction) Read the data
	* Write the data to postgres
	* (single FDB transaction) Read the data from FDB again (new data), then write the difference between the new data and the dirty data to FDB (“clean” data), then if the “clean” data is empty, remove the “dirty” mark from the id, and finally remove the “writing” mark

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
