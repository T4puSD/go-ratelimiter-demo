## Rate Limiter
A web server with go gin module to demonstrate rate limiting implementation with redis

### Dependencies
- Redis

### How to run
To run the project a redis server must be installed either in host or in docker environment.   
After ensuring that you can use the following command to start the server:
```sh
go run ./main.go
```

> Note: If your redis server is not running in default port or host or have a password set in redis then change the following code section in `main.go` file:   

```go
redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Replace with your Redis server address
		Password: "",               // Replace with your Redis password
		DB:       0,                // Replace with the desired Redis database index
	})
```


### How to test the rate limiter
To test the rate limiter in effect his the following endpoint
```sh
curl http://localhost:8080/limited-route
```

After hitting the endpoint 10 times you will see too many request response code from the server which will get cooled down after 1 minute.


### How to change the default max limit and cooldown time window
To change the default max limit and cooldown time window change the following parameters in the `main.go` file:   
- limit
- interval

```go
limiter := &RateLimiter{
	limit:       10,          // Maximum number of requests allowed
	interval:    time.Minute, // Time interval for rate limiting
	redisClient: redisClient, // Redis client instance
	luaScriptID: luaScriptID, // Lua script ID
}
```
