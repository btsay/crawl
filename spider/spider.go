package spider

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/btsay/crawl/utils"
	"github.com/btsay/spider"
)

//Run the spider
func Run() {
	manage.run()

	idList := spider.GenerateIDList(utils.Config.SpiderNumber)
	for k, id := range idList {
		go func(port int, id spider.ID) {
			address := fmt.Sprintf(":%v", utils.Config.SpiderListenPort+port)
			spider.RunDhtNode(&id, manage.out, address)
		}(k, id)
	}

	go store()

	for result := range manage.out {
		if len(result.Infohash) == 40 {
			hash := strings.ToUpper(result.Infohash)
			if result.IsAnnouncePeer {
				go increaseResourceHeat(hash)
				if utils.Config.EnableMetadata {
					go getMetadata(result)
				}
			}
			if manage.isHashinfoExist(hash) {
				continue
			}
			receive(hash)
		}
	}
}

func increaseResourceHeat(key string) {
	indexType := strings.ToLower(string(key[0]))
	searchResult, err := utils.ElasticClient.Get().Index("torrent").Type(indexType).Id(key).Do()
	if err == nil && searchResult != nil && searchResult.Source != nil {
		var tdata torrentSearch
		err = json.Unmarshal(*searchResult.Source, &tdata)
		if err == nil {
			tdata.Heat++
			_, err = utils.ElasticClient.Index().Index("torrent").Type(indexType).Id(key).BodyJson(tdata).Refresh(false).Do()
			if err != nil {
				utils.Log.Println(err)
			}
		}
	}
}

func getMetadata(result spider.Infohash) (err error) {
	infohash := strings.ToUpper(result.Infohash)
	trt, err := utils.Repository.GetTorrentByInfohash(infohash)
	if err == nil && len(trt.Infohash) != 0 {
		//资源已存在
		return
	}
	infohashByte, _ := hex.DecodeString(result.Infohash)
	manage.wire.fetchMetadata(Request{
		Port:     result.Port,
		IP:       result.IP.String(),
		InfoHash: infohashByte})
	return
}

//根据infohash的首字符(0~F)，将infohash写入到对应chan中
func receive(hash string) {
	if c, ok := manage.storeMap[string(hash[0])]; ok {
		c <- hash
	}
}

func store() {
	for k, v := range manage.storeMap {
		go storeSingle(k, v)
	}
	for resp := range manage.wire.Response() {
		if len(resp.MetadataInfo) > 0 {
			metadata, err := Decode(resp.MetadataInfo)
			if err != nil {
				fmt.Println(err)
				continue
			}
			storeTorrent(metadata, resp.InfoHash)
		}
	}
}

//批量处理爬取到的infohash，如果此infohash已经抓取过了，则资源热度+1，否则存入预处理表
func storeSingle(k string, v chan string) {
	var hashs []string
	for hash := range v {
		hashs = append(hashs, hash)
		if len(hashs) >= 100 {
			data, err := utils.Repository.BatchGetTorrentByInfohash(hashs)
			if err != nil {
				utils.Log.Println(err)
				continue
			}
			var (
				hashMap = make(map[string]int)
			)

			for _, item := range hashs {
				hashMap[item] = 0
			}
			for _, item := range data {
				hashMap[item.Infohash]++
			}

			for key, value := range hashMap {
				if value == 0 {
					go StoreInfohash(key)
				}
			}
			hashs = make([]string, 0)
		}
	}
}

//StoreInfohash into temp table
func StoreInfohash(infohash string) (err error) {
	if len(infohash) == 40 {
		return utils.Repository.CreateInfohash(infohash)
	}
	return
}
