# gorlock

Redis Lock wrapper for running distributed tasks. You can customize settings or configure and manage own redis client.

# Usage

```go
  import (
    "fmf"
    
    "github.com/vanng822/gorlock"
  )
  gorlock.Run("somekey", func() error {
    fmt.Println("Doing some job")
    return nil
  })
```

Or

```go
  import (
    "fmf"
    
    "github.com/vanng822/gorlock"
  )
  gorlock.RunWaiting("somekey", func() error {
    fmt.Println("Doing some job")
    return nil
  })
```

