# scoutup
Simple Dev tool for running local Blockscout instances.

The primary goal of this tool is to provide a simple way to run a local explorer for development and testing purposes.

### Prerequisites
* docker

### Build
```
go build
```

### Anvil
1. Run local `anvil` node:
```
anvil --host 0.0.0.0
```
2. Run local explorer for anvil using `scoutup`:
```
./scoutup
```
3. The command above will run the local Blockscout instance with the default config using docker.

### Supersim
1. Run `supersim`:
```
./main --l1.host 0.0.0.0 --l2.host 0.0.0.0 --interop.autorelay true
```
2. Run local explorers for all `supersim` chains:
```
./scoutup --supersim
```
3. The command above will fetch the configuration of the supersim network via the supersim admin RPC, set up the necessary Blockscout instances accordingly, and run the local Blockscout instances using docker.

### Cleanup
`scoutup` attempts to stop and remove all running containers and delete all temporary files when stopping. However, depending on the termination process, some dangling containers and temporary files may remain. In such cases, it is recommended to run the following command to clean up:
```
./scoutup clean
```
