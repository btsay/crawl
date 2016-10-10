package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/btsay/repository"
	"gopkg.in/olivere/elastic.v3"
)

//defile vars
var (
	Config        config
	Log           *log.Logger
	Repository    repository.Repository
	ElasticClient *elastic.Client
)

type config struct {
	Database         string `json:"database"`
	Elastic          string `json:"elastic"`
	SpiderNumber     int64  `json:"spider_number"`
	SpiderListenPort int    `json:"spider_listen_port"`
	EnableMetadata   bool   `json:"enable_metadata"`
}

//Init utilsl
func Init() {
	initLog()
	initConfig()
	initDatabase()
	initElastic()
}

func initElastic() {
	client, err := elastic.NewClient(elastic.SetURL(Config.Elastic))
	exit(err)
	ElasticClient = client
	ElasticClient.CreateIndex("torrent").Do()
}

func initConfig() {
	f, err := os.Open("config/crawl.conf")
	exit(err)
	b, err := ioutil.ReadAll(f)
	exit(err)
	err = json.Unmarshal(b, &Config)
	exit(err)
}

func initLog() {
	Log = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
}

func initDatabase() {
	repo, err := repository.NewMysqlRepository(Config.Database, 1024, 1024)
	exit(err)
	Repository = repo
}

func exit(err error) {
	if err != nil {
		Log.Fatalln(err)
	}
}
