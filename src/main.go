package main

// TODO map reduce
import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/go-redis/redis"
	"os"
	"sync"
	"strconv"
	"time"
	"fmt"
)


var adminUser = map[string]string{
	"john":"johnpwd",
	"mary":"marypwd",
}

type redisResp struct {
	Status string `json:"status"`
	Elasped float64 `json:"elasped"`
	DataLen int `json:"data_length"`
	Result []string `json:"result"`
}

const maxLen int = 10000

var redisClient *redis.Client

func timeElapsed(f func()) float64 {
	start := time.Now()
    f()
	end := time.Now()
	elapsed := end.Sub(start)
	return elapsed.Seconds()
}

func appInit() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	mlTmp := maxLen
	times := 0
	fixedLen := 10000

	x := func() {
		for mlTmp > 0 {
			go func(ctime int) {
				for i := ctime * fixedLen; i < (ctime * fixedLen)+fixedLen; i++ {
					redisClient.Set("Data"+strconv.Itoa(i), i, 0)
				}
			}(times)
			mlTmp -= fixedLen
			times++
		}
	}
	fmt.Printf("Prepare data elapsed: %f", timeElapsed(x))
}

func main() {
	var once sync.Once
	once.Do(appInit)
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hello)

	// Redis Reoutes
	redis := e.Group("/redis")
	redis.Use(middleware.BasicAuth(func(username, password string, c echo.Context)(bool, error) {
		for k, v := range adminUser{
			if k == username && v == password {
				return true, nil
			}
		}
		return false, nil
	}))
	redis.GET("/health/", redisStatus)
	redis.GET("/data/seq/", getRedisDataViaSeq)
	redis.GET("/data/routine/unbuffered/", getRedisDataViaRoutineUnbuffered)
	redis.GET("/data/routine/buffered/", getRedisDataViaRoutineBuffered)

	// Start server
	// env PORT is from external
	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT"))) // for heroku deploy: https://stackoverflow.com/questions/32463004/heroku-go-app-crashing
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func redisStatus(c echo.Context) error {
	pong, _ := redisClient.Ping().Result()
	
	return c.JSON(http.StatusOK, struct {Status string `json:"status"`} {
		pong,
	}) 
}

func getRedisDataViaSeq(c echo.Context) error {
	var result []string

	start := time.Now()
	for i:=0; i< maxLen; i++ {
		r, err := redisClient.Get("Data"+strconv.Itoa(i)).Result()
		if  err == nil {
			result = append(result, r)
		}
	}
	end := time.Now()
	elapsed := end.Sub(start)
	

	return c.JSON(http.StatusOK, redisResp {
		"done",
		elapsed.Seconds(),
		maxLen,
		result,
	}) 
}

func getRedisDataViaRoutineUnbuffered(c echo.Context) error {
	var result []string
	dataChan := make(chan string)

	start := time.Now()


	for i:=0; i< maxLen; i++ {

		go func(dataIndex int) {
			r, err := redisClient.Get("Data"+strconv.Itoa(dataIndex)).Result()
			if  err == nil {
				dataChan <- r
			}
	    }(i)
	}

	for i := 0; i < maxLen; i++ {
		result = append(result, <-dataChan)
	}

		
	end := time.Now()


	elapsed := end.Sub(start)
	

	return c.JSON(http.StatusOK, redisResp {
		"done",
		elapsed.Seconds(),
		maxLen,
		result,
	}) 
}

func getRedisDataViaRoutineBuffered(c echo.Context) error {
	var result []string
	dataChan := make(chan string, maxLen)

	start := time.Now()


	for i:=0; i< maxLen; i++ {

		go func(dataIndex int) {
			r, err := redisClient.Get("Data"+strconv.Itoa(dataIndex)).Result()
			if  err == nil {
				dataChan <- r
			}
	    }(i)
	}

	for i := 0; i < maxLen; i++ {
		result = append(result, <-dataChan)
	}

		
	end := time.Now()


	elapsed := end.Sub(start)
	

	return c.JSON(http.StatusOK, redisResp {
		"done",
		elapsed.Seconds(),
		maxLen,
		result,
	}) 
}