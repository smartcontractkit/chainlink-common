# myapp Configuration

## Example

```toml
# ----- Global Configuration -----

[chain]
id = 1
rpc = 'https://mainnet.infura.io/v3/API_KEY'
contract = '0xYourContractAddress'

[system]
env = 'dev'
log_level = ''
metrics = false

# ----- Command: myapp bar -----
[bar]
retries = 3
backoff = '2s'

# ----- Command: myapp foo -----
[foo]
timeout = '5s'


```

# Global Configuration

## chain
```toml
[chain]
id = 1 # Default
rpc = 'https://mainnet.infura.io/v3/API_KEY' # Example
contract = '0xYourContractAddress' # Example
```


### id
```toml
id = 1 # Default
```
id Target chain ID selector

### rpc
```toml
rpc = 'https://mainnet.infura.io/v3/API_KEY' # Example
```
rpc RPC endpoint URL

### contract
```toml
contract = '0xYourContractAddress' # Example
```
contract Core contract address

## system
```toml
[system]
env = 'dev' # Default
log_level = '' # Default
metrics = false # Default
```


### env
```toml
env = 'dev' # Default
```
env Environment profile selector (dev, prod)

### log_level
```toml
log_level = '' # Default
```
log_level Logging severity

### metrics
```toml
metrics = false # Default
```
metrics Enable metrics collection

# Command: myapp bar

## bar
```toml
[bar]
retries = 3 # Default
backoff = '2s' # Default
```
Bar-specific settings

### retries
```toml
retries = 3 # Default
```
retries Bar max retries

### backoff
```toml
backoff = '2s' # Default
```
backoff Bar retry backoff

# Command: myapp foo

## foo
```toml
[foo]
timeout = '5s' # Default
```
Foo-specific settings

### timeout
```toml
timeout = '5s' # Default
```
timeout Foo operation timeout

