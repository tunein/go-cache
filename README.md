# go-cache
Generic in-memory cache


### Usage

```
    // Default cache ttl
    ttl := 1 * time.Hour
    cc = New[string, float32](ttl)

    // You can pass a function that will be called to fetch value if nothing is found using provided key
    var loader = func(s string) (float32, error) {
					// put your logic of value retrieval here
                    return 1, nil
				}
    value, err := cc.Get("someKey")

    // or you can set value directly
    cc.Set("your-key", 100)

    // also you can override default cache ttl and set dedicated ttl to an item (e.g. store forever)
    cc.SetWithExpire("never-ending-story", 9999)
 ```
