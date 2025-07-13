# gorlock

Redis Lock wrapper for running distributed tasks. You can customize settings or configure and manage own redis client.

# Usage

```go
  import (
    "fmf"
    
    "github.com/vanng822/gorlock"
  )
  // using default instance
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
  // using default instance with LockWaiting=true
  gorlock.RunWaiting("somekey", func() error {
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

  // setting up own config/settings
  redisConfig := gorlock.RedisConfig{
    Address:        "localhost:6379",
    Database:       0,
    ConnectTimeout: 3 * time.Second,
  }

  waitingSettings := gorlock.Settings{
    KeyPrefix:     "gorlock",
    LockTimeout:   10 * time.Second,
    LockWaiting:   true,
    RetryTimeout:  3 * time.Second,
    RetryInterval: 100 * time.Millisecond,
  }
  
  lock := gorlock.New(&waitingSettings, &redisConfig)

  // Wait for the lock if it is already locked due to LockWaiting=true
  lock.Run("somekey", func() error {
    fmt.Println("Doing some job")
    return nil
  })
```
