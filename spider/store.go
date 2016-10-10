package spider

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/btsay/crawl/utils"
	"github.com/btsay/repository"
)

type files []repository.File

func (a files) Len() int           { return len(a) }
func (a files) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a files) Less(i, j int) bool { return a[i].Length > a[j].Length }

type torrentSearch struct {
	Name       string
	Length     int64
	Heat       int64
	CreateTime time.Time
}

func storeTorrent(data interface{}, infohash []byte) (err error) {
	defer func() {
		e := recover()
		if e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	var t repository.Torrent
	t.CreateTime = time.Now()

	if info, ok := data.(map[string]interface{}); ok {
		//get name
		if name, ok := info["name"].(string); ok {
			t.Name = name
			if t.Name == "" {
				return fmt.Errorf("store name len is 0")
			}
		}
		//get infohash
		t.Infohash = hex.EncodeToString(infohash)
		if len(t.Infohash) != 40 {
			return fmt.Errorf("store infohash len is not 40")
		}
		//get files
		if v, ok := info["files"]; !ok {
			t.Length = int64(info["length"].(int))
			t.FileCount = 1
			t.Files = append(t.Files, repository.File{Name: t.Name, Length: t.Length})
		} else {
			var tmpFiles files
			fls := v.([]interface{})
			tmpFiles = make([]repository.File, len(fls))
			for i, item := range fls {
				fl := item.(map[string]interface{})
				flName := fl["path"].([]interface{})
				tmpFiles[i] = repository.File{
					Name:   flName[0].(string),
					Length: int64(fl["length"].(int)),
				}
			}
			sort.Sort(tmpFiles)

			for k, v := range tmpFiles {
				if len(v.Name) > 0 {
					t.Length += v.Length
					t.FileCount++
					if k < 5 {
						t.Files = append(t.Files, repository.File{
							Name:   v.Name,
							Length: v.Length,
						})
					}
				}
			}
		}
		err = utils.Repository.CreateTorrent(t)
		if err == nil {
			manage.storeCount++
			data := torrentSearch{
				Name:       t.Name,
				Length:     t.Length,
				CreateTime: time.Now(),
			}
			indexType := strings.ToLower(string(t.Infohash[0]))
			utils.ElasticClient.Index().Index("torrent").Type(indexType).Id(t.Infohash).BodyJson(data).Refresh(false).Do()
		}
	}
	return
}
