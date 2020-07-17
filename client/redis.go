package client

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/wangfeiso/rwlock/lua"
	"strconv"
	"time"
)

var Redis *redis.Client
var opts *redis.Options

// error 定义
const NoScriptError = "NOSCRIPT No matching script. Please use EVAL."
const EofError = "EOF"

const LockCmd = "LOCK"
const UnlockCmd = "UNLOCK"
const RLockCmd = "RLOCK"
const RUnlockCmd = "RUNLOCK"

func Init(opt *redis.Options) {
	Redis = redis.NewClient(opt)
	ping, err := Redis.Ping().Result()
	if err != nil {
		return
	}
	if ping != "PONG" {

	}
	opts = opt
	InitLua()
}

func InitLua() string {
	hashID, err := Redis.ScriptLoad(lua.ScriptContent).Result()
	if err != nil {
		return ""
	}
	return hashID
}

type responseLock struct {
	OpRet  bool   `json:"opRet"`
	ErrMsg string `json:"errMsg"`
	Debug  string `json:"debug"`
}

func (r responseLock) IsError() bool {
	if len(r.ErrMsg) > 0 {
		return true
	}
	return false
}
func (r responseLock) Success() bool {
	return r.OpRet
}
func (r responseLock) Error() string {
	return r.ErrMsg
}

func Lock(shaHashID, key string, uniqID string, expireTime, retryTime int64) {
	res, err := send(shaHashID, key, uniqID, LockCmd, expireTime)
	if err != nil {

	}
	if res.Success() {
		return
	}
	// 如果没有拿到锁，时间驱动
	for {
		res, err := send(shaHashID, key, uniqID, LockCmd, expireTime)
		if err != nil {

		}
		if res.IsError() {
			panic(res.ErrMsg)
		}
		if res.Success() {
			return
		}
		// TODO 改成随机
		time.Sleep(10 * time.Millisecond)
		break
	}
}

func Unlock(shaHashID, key, uniqID string) {
	res, err := send(shaHashID, key, uniqID, UnlockCmd, 0)
	if res.Success() {
		return
	}
	if res.IsError() {
		panic(res.Error())
	}
	if err != nil {
		handleError(err)
	}
}

func RLock(shaHashID, key string) {
	res, err := send(shaHashID, key, "", UnlockCmd, 0)
	if res.Success() {
		return
	}
	if err != nil {
		handleError(err)
	}
}

func RUnlock(shaHashID, key string) {
	res, err := send(shaHashID, key, "", RUnlockCmd, 0)
	if res.Success() {
		return
	}
	if err != nil {
		handleError(err)
	}
}

func send(shaHashID, key string, uniqID, lockCmd string, expireTime int64) (*responseLock, error) {
	var ret interface{}
	var err error
	switch lockCmd {
	case LockCmd:
		ret, err = Redis.EvalSha(shaHashID, []string{key, lockCmd}, []string{uniqID, strconv.Itoa(int(expireTime))}).Result()
	case UnlockCmd:
		ret, err = Redis.EvalSha(shaHashID, []string{key, lockCmd}, []string{uniqID}).Result()
	case RLockCmd, RUnlockCmd:
		ret, err = Redis.EvalSha(shaHashID, []string{key, lockCmd}, []string{}).Result()
	}

	if err != nil {
		return nil, err
	}
	var retJson = ret.(string)
	var res responseLock
	if err := json.Unmarshal([]byte(retJson), &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func handleError(err error) {
	if err.Error() == EofError {
		handleEofError()
	}
	if err.Error() == NoScriptError {
		handleNoScriptError()
	}
}

// redis重启
func handleEofError() {
	Init(opts)
}

//script
func handleNoScriptError() {
	InitLua()
}