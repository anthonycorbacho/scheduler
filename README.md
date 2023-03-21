## Scheduler

**Scheduler** is a simple implementation of a FiFo Scheduler.

Get it:
```
go get -u github.com/anthonycorbacho/scheduler
```

### Quick start
It is very easy to get started, simply create an instance of the FiFoScheduler and start adding tasks.

```go
package main

import (
	"fmt"

	"github.com/anthonycorbacho/scheduler"
)

func main() {
	s := NewFIFOScheduler()
	defer s.Stop()
	
	jobCreator := func(i int) Job {
		return func(ctx context.Context) {
			fmt.Printf("got a task %d\n", i)
		}
	}

	var jobs []Job
	for i := 0; i < 100; i++ {
		jobs = append(jobs, jobCreator(i))
	}

	for _, j := range jobs {
		_ = s.Schedule(j)
	}
}
```

