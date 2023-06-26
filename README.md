# go-cache

Goroutine-safe generic flexible implementation of in-memory cache. 

### Installation

go get github.com/tunein/go-cache

### Usage

```go
    // Default cache ttl
    ttl := 1 * time.Hour
    cc = New[string, float32](ttl)

    // You can pass a function that will be called to fetch value if nothing is found using provided key
    var loader = func(s string) (float32, error) {
	// put your logic of value retrieval here
        return 1, nil
	}
    value, err := cc.Get("someKey")

    // or you can set value directly with default TTL (defined when creating a Cache instance)
    cc.Set("your-key", 100)

    // or you can override default TTL, e.g. never to expire
    cc.SetWithExpire("your-new-key", 200, 0)

    // also you can use your own func to Update existing value (atomic operation)
    cc.Update("never-ending-story", func(in float32){ /* your calculations */ })
 ```
