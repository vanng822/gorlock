# gorlock

Redis Lock wrapper for running distributed tasks

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

