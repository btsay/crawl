package spider

import (
	"fmt"
	"sync"
	"time"

	"github.com/btsay/crawl/utils"
	"github.com/btsay/spider"
)

//const
const (
	HashChanSize  = 512
	UniqHashSize  = 100 * 1000
	BlackListSize = 100 * 1000
	StoreChanSize = 1000
)

//Manager spider
var manage manager

func (p *manager) run() {
	p.initChan()
	p.initUniqHash()
	p.initBlackList()
	p.wire = NewWire()
	go p.monitor()
}

func (p *manager) initChan() {
	p.out = make(chan spider.Infohash, HashChanSize)
	p.storeMap = make(map[string]chan string)
	for i := 0; i <= 15; i++ {
		p.storeMap[fmt.Sprintf("%X", i)] = make(chan string, StoreChanSize)
	}
}

func (p *manager) initBlackList() {
	p.blacklist.mutex.Lock()
	defer p.blacklist.mutex.Unlock()
	for k := range p.blacklist.blackList {
		delete(p.blacklist.blackList, k)
	}
	p.blacklist.blackList = nil
	p.blacklist.blackList = make(map[string]bool, BlackListSize)
}

func (p *blacklist) set(s string) {
	p.mutex.Lock()
	p.blackList[s] = true
	p.mutex.Unlock()
}

func (p *manager) initUniqHash() {
	p.uniqInfohash.mutex.Lock()
	defer p.uniqInfohash.mutex.Unlock()
	for k := range p.uniqInfohash.uniqInfohash {
		delete(p.uniqInfohash.uniqInfohash, k)
	}
	p.uniqInfohash.uniqInfohash = nil
	p.uniqInfohash.uniqInfohash = make(map[string]bool, UniqHashSize)
}

func (p *manager) isHashinfoExist(hash string) bool {
	p.uniqInfohash.mutex.Lock()
	defer p.uniqInfohash.mutex.Unlock()
	if _, ok := p.uniqInfohash.uniqInfohash[hash]; ok {
		return true
	}
	manage.uniqInfohash.uniqInfohash[hash] = true
	return false
}

type blacklist struct {
	blackList map[string]bool
	mutex     sync.Mutex
}

type uniqInfohash struct {
	uniqInfohash map[string]bool
	mutex        sync.Mutex
}

type manager struct {
	wire       *Wire
	storeCount int64
	storeMap   map[string]chan string
	out        chan spider.Infohash

	uniqInfohash uniqInfohash
	blacklist    blacklist
}

func (p *manager) monitor() {
	go spider.Monitor()

	for {
		utils.Log.Printf("从DHT网络获取资源数量(BEP9): %v\n", p.storeCount)
		if len(manage.uniqInfohash.uniqInfohash) >= UniqHashSize {
			p.initUniqHash()
		}
		if len(manage.blacklist.blackList) >= BlackListSize {
			p.initBlackList()
		}
		time.Sleep(time.Minute)
	}
}
