# Peer Node CLI

To run the peer node, you can create an exectuable and then use the exectuable's CLI

For example:

```bash
$ make build

$ bin/peer [arguments]
```

You can also combine this into one step with:

```bash
$ make all [arguments]
```

## CLI functions

Get a file from the DHT. You should pass a specific hash.

```bash
$ get [fileHash] 
```

Storing a file in the DHT for a given price. You should pass ONLY the file name, given the file is in the files folder (inside peers).

```bash
$ store [filename] [amount]
```

Import a file into the files directory. You can pass it any filepath, but if the path is relative. It will be rooted in the ./peer folder. It is best to just use an absolute path.

```bash
$ import [filepath]
```

Send a certain amount of coin to an address

```bash
$ send [amount] [ip] 
```

Hash a file. Only files inside the files folder can be found. Only pass relative paths. You should not need to hash any files: this should be handled internally.

```bash
$ hash [filename]
```

Listing all files stored for IPFS

```bash
$ list
```

Getting current peer node location

```bash
$ location
```

Testing network speeds

```bash
$ network
```

Runs the peer node as a backend application for a front end GUI

```bash
$ run
```

#### File System:

* There is a folder called <i>files</i>. This is where all the files that are available to the user is stored

* Any file directly stored inside <i>files</i> folder is considered <i>uploaded</i> to the client.

* Any file that has been requested by the user is stored in the <i>files/requested</i> folder.

* Any file that is available to be requested for by anyone on the network is in <i>files/stored</i>.

* Technically, you can import the files manually if you drag them inside the desired folder. There is currently no protection against this.

* The <i>transactions</i> folder stores all of the transactions that have been processed and stored.

#### Notes:

* Files that are on the network should be in the files folder. This can be done manually or by using the CLI

* Inside the config file, set your public key and private key location. If you don't want to, the CLI will generate a key-pair for you.

* Only .txt, .json and .mp4 file formats are currently supported.
