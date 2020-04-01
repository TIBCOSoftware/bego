package redisutils

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

type RedisHdl = *RedisHandle

var rd = make([]RedisHdl, 0, 8)

type RedisConfig struct {
	Network string `json:"network"`
	Address string `json:"address"`
}

type RedisHandle struct {
	config  RedisConfig
	pool    *redis.Pool
	network string
	address string
}

func NewRedisHdl(config RedisConfig) RedisHdl {
	handle := RedisHandle{
		config:  config,
		network: config.Network,
		address: config.Address,
	}
	if handle.network == "" {
		handle.network = "tcp"
	}
	if handle.address == "" {
		handle.address = ":6379"
	}
	handle.newPool(handle.network, handle.address)
	rd = append(rd, &handle)
	return &handle
}

func (rh *RedisHandle) newPool(network string, address string) {
	rh.pool = &redis.Pool{
		// Maximum number of idle connections in the pool.
		MaxIdle: 80,
		// max number of connections
		MaxActive: 12000,
		// Dial is an application supplied function for creating and
		// configuring a connection.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(network, address)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

func (rh *RedisHandle) getPool() *redis.Pool {
	return rh.pool
}

func (rh *RedisHandle) Set(key string, value string, nx bool, px int64) (bool, error) {
	c := rh.getPool().Get()
	defer c.Close()
	args := []interface{}{key, value}
	if nx {
		args = append(args, "NX")
	}
	if px > 0 {
		args = append(args, "PX", px)
	}
	ok, err := redis.String(c.Do("SET", args...))
	return ok == "OK", err
}

const delIfEqualScript = `if redis.call("get",KEYS[1]) == ARGV[1]
then
    return redis.call("del",KEYS[1])
else
    return 0
end`

func (rh *RedisHandle) DelIfEqual(key string, value string) (int, error) {
	c := rh.getPool().Get()
	defer c.Close()
	return redis.Int(c.Do("EVAL", delIfEqualScript, 1, key, value))
}

func (rh *RedisHandle) HSet(key string, field string, value interface{}) (bool, error) {
	c := rh.getPool().Get()
	defer c.Close()
	exists, err := redis.Int(c.Do("HSET", key, field, value))
	return exists == 0, err
}

func (rh *RedisHandle) HSetNX(key string, field string, value interface{}) (bool, error) {
	c := rh.getPool().Get()
	defer c.Close()
	exists, err := redis.Int(c.Do("HSETNX", key, field, value))
	return exists == 0, err
}

func (rh *RedisHandle) HSetAll(key string, kvs map[string]interface{}) error {

	var args = []interface{}{key}
	for f, v := range kvs {
		args = append(args, f, v)
	}
	c := rh.getPool().Get()
	defer c.Close()
	_, error := c.Do("HMSET", args...)

	return error
}

func (rh *RedisHandle) HGetAll(key string) map[string]interface{} {
	hgetall := make(map[string]interface{})
	c := rh.getPool().Get()
	defer c.Close()
	vals, error := c.Do("HGETALL", key)
	if error != nil {
		fmt.Printf("error [%v]\n", error)
	} else {
		vals, err2 := redis.Values(vals, error)
		if err2 != nil {
			fmt.Printf("error [%v]\n", err2)
		} else {
			i := 0
			key := ""
			value := ""
			for _, val := range vals {
				ba := val.([]byte)
				s := string(ba)
				//fmt.Printf("Value [%s]\n", s)
				if i%2 == 0 {
					key = s
				} else {
					value = s
					hgetall[key] = value
				}
				i++
			}
		}
	}
	return hgetall
}

func (rh *RedisHandle) HGet(key string, field string) interface{} {
	c := rh.getPool().Get()
	defer c.Close()
	vals, error := c.Do("HGET", key, field)
	if error != nil {
		fmt.Printf("error [%v]\n", error)
	} else {
		ba := vals.([]byte)
		s := string(ba)
		return s
	}
	return nil
}

func (rh *RedisHandle) HLen(key string) int {
	c := rh.getPool().Get()
	defer c.Close()
	vals, error := c.Do("HLEN", key)
	if error != nil {
		fmt.Printf("error [%v]\n", error)
	} else {
		len, _ := redis.Int(vals, error)
		return len
	}
	return 0
}

func (rh *RedisHandle) HIncrBy(key string, field string, by int) int {
	c := rh.getPool().Get()
	defer c.Close()
	i, err := c.Do("HINCRBY", key, field, 1)

	if err != nil {
		fmt.Printf("error: [%s]", err)
		return -1
	}
	current := int(i.(int64))
	return current
}

func (rh *RedisHandle) IncrBy(key string, by int) int {
	c := rh.getPool().Get()
	defer c.Close()
	i, err := c.Do("INCRBY", key, 1)

	if err != nil {
		fmt.Printf("error: [%s]", err)
		return -1
	}
	current := int(i.(int64))
	return current
}

func (rh *RedisHandle) HGetAsInt(key string, field string) int {
	c := rh.getPool().Get()
	defer c.Close()
	i, err := c.Do("HGET", key, field)
	j := -1
	if err == nil {
		j, _ = redis.Int(i, err)
	}
	return j
}

func (rh *RedisHandle) GetAsInt(key string) int {
	c := rh.getPool().Get()
	defer c.Close()
	i, err := c.Do("GET", key)
	j := -1
	if err == nil {
		j, _ = redis.Int(i, err)
	}
	return j
}

func (rh *RedisHandle) Del(key string) int {
	c := rh.getPool().Get()
	defer c.Close()
	i, err := c.Do("DEL", key)
	j := -1
	if err == nil {
		j, _ = redis.Int(i, err)
	}
	return j
}

func (rh *RedisHandle) HDel(key string, field string) int {
	c := rh.getPool().Get()
	defer c.Close()
	i, err := c.Do("HDEL", key, field)
	j := -1
	if err == nil {
		j, _ = redis.Int(i, err)
	}
	return j
}

type LIterator struct {
	key      string
	scanIdx  int                    //iterator index
	keyIdx   int                    //local array current index
	keys     []string               //array of keys in the current call
	valueMap map[string]interface{} //map of key/value of the current keys
	rh       *RedisHandle
}

func (iter *LIterator) HasNext() bool {

	if iter.scanIdx == 0 { // there is nothing more with redis
		if len(iter.keys) == 0 {
			return false
		} else {
			if iter.keyIdx == len(iter.keys) {
				return false
			}
		}
	} else {
		if len(iter.keys) == 0 || iter.keyIdx >= len(iter.keys) {
			c := iter.rh.getPool().Get()
			defer c.Close()
			if arr, err := redis.Values(c.Do("HSCAN", iter.key, iter.scanIdx)); err != nil {
				panic(err)
			} else {
				// now we get the iter and the keys from the multi-bulk reply
				iter.scanIdx, _ = redis.Int(arr[0], nil)
				iter.keys, _ = redis.Strings(arr[1], nil)
				iter.keyIdx = 0
				return iter.HasNext()
			}
		}
	}

	return true

}

func (iter *LIterator) Next() string {
	str := iter.keys[iter.keyIdx]
	iter.keyIdx++
	return str
}

func (rh *RedisHandle) GetListIterator(key string) *LIterator {

	iter := LIterator{}
	iter.key = key
	iter.rh = rh
	c := rh.getPool().Get()
	defer c.Close()

	if arr, err := redis.Values(c.Do("HSCAN", key, 0, "COUNT", 1)); err != nil {
		panic(err)
	} else {
		// now we get the iter and the keys from the multi-bulk reply
		iter.scanIdx, _ = redis.Int(arr[0], nil)
		iter.keys, _ = redis.Strings(arr[1], nil)
		iter.keyIdx = 0
	}

	return &iter
}

type MapIterator struct {
	LIterator
	currKey   string
	currValue interface{}
}

func (iter *MapIterator) HasNext() bool {

	if iter.scanIdx == 0 { // there is nothing more with redis
		if len(iter.keys) == 0 {
			return false
		} else {
			if iter.keyIdx > len(iter.keys)-2 {
				return false
			}
		}
	} else {
		if len(iter.keys) == 0 || iter.keyIdx >= len(iter.keys)-1 {
			c := iter.rh.getPool().Get()
			defer c.Close()
			if arr, err := redis.Values(c.Do("HSCAN", iter.key, iter.scanIdx)); err != nil {
				panic(err)
			} else {
				iter.scanIdx, _ = redis.Int(arr[0], nil)
				iter.keys, _ = redis.Strings(arr[1], nil)
				iter.keyIdx = 0
				return iter.HasNext()
			}
		}
	}

	return true

}

func (iter *MapIterator) Remove() {
	iter.rh.HDel(iter.key, iter.currKey)
}

func (iter *MapIterator) Next() (string, interface{}) {
	iter.currKey = iter.keys[iter.keyIdx]
	iter.currValue = iter.keys[iter.keyIdx+1]

	iter.keyIdx += 2
	return iter.currKey, iter.currValue
}

func (rh *RedisHandle) GetMapIterator(key string) *MapIterator {

	iter := MapIterator{}
	iter.key = key
	iter.rh = rh
	c := rh.getPool().Get()
	defer c.Close()

	if arr, err := redis.Values(c.Do("HSCAN", key, 0, "COUNT", 1)); err != nil {
		panic(err)
	} else {
		// now we get the iter and the keys from the multi-bulk reply
		iter.scanIdx, _ = redis.Int(arr[0], nil)
		iter.keys, _ = redis.Strings(arr[1], nil)
		iter.keyIdx = 0
	}

	return &iter
}

func Shutdown() {
	for _, value := range rd {
		value.pool.Close()
	}
	rd = rd[:0]
}

// GetKeys returns all keys present in redis server
func (rh *RedisHandle) GetKeys(pattern string) ([]string, error) {

	conn := rh.getPool().Get()
	defer conn.Close()

	iter := 0
	keys := []string{}
	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return keys, fmt.Errorf("error retrieving '%s' keys", pattern)
		}

		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}

	return keys, nil
}
