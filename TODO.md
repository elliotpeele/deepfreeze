# TODO items for backup tool
* publish to amazon glacier
* publish to local filesystem
* incremental backups
* store files in a nGB file, probably 1GB
* hash content to check for differences
* store data in a database, probably sqlite
* database should be encrypted
* all backup files should be individually encrypted
* backend store files must be anonimized such that if someone gets access to the backend store, they can not figure out what the files are
* need to figure out if go has a standard python struct like library
* need to be able to delete old backups from backend store
* keys, database, and backing store should all be in different locations
* database should be signed when stored so that it can be verfied
* database should be encrypted when stored
