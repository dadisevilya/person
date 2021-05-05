# Gett Storages

## Postgres
to be filled

### ReadOnly Database
```go
type DBConf struct {
	...
	ReadOnly         bool
}
```
Use can use __ReadOnly__ option to force all sql transactions to be __READ_ONLY__. It is useful while working on a replica.

You can also use config to do the same (like this)
```yaml
prod:
    driver: postgres
    open: dbname=driver_earnings  sslmode=disable
    ...
    read_only: true
```


## Redis
to be filled


> Running lua script
```go
package iNeedLua

import 	"github.com/gtforge/services_common_go/gett-storages"

//lua code here
const someLuaScript = "return 0;"

var script = gettStorages.NewLuaScript(someLuaScript, "MyNRName")

func MyLuaScript(key, old, new string) (bool, error) {
	var (
		res interface{}
		err error
	)
    
    res, err = 	gettStorages.RedisClient.RunLua(script, []string{key}, old, new)
	if err != nil {
		return false, err
	}
	return res.(int64) == 1, err
}

```
    - make sure your `LuaScriptName` is unique, otherwise it will cause redis failures