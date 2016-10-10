package main

import (
	"github.com/btsay/crawl/spider"
	"github.com/btsay/crawl/utils"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	utils.Init()
	spider.Run()
}
