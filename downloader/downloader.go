package downloader

import (
	"errors"
	"fmt"
	"github.com/fmyxyz/goreptile/base"
	"github.com/fmyxyz/goreptile/middleware"
	"net/http"
	"reflect"
)

//网页下载器
type PageDownloader interface {
	Id() uint32
	Download(req base.Request) (*base.Response, error)
}

//网页下载器池
type PageDownloaderPool interface {
	Take() (PageDownloader, error)
	Return(pl PageDownloader) error
	Total() uint32
	Used() uint32
}

type myPageDownloader struct {
	httpClient http.Client
	id         uint32
}

func (this *myPageDownloader) Id() uint32 {
	return this.id
}

func (this *myPageDownloader) Download(req base.Request) (*base.Response, error) {
	httpResp, err := this.httpClient.Do(req.HttpReq())
	if err != nil {
		return nil, err
	}
	return base.NewResponse(httpResp, req.Depth()), nil
}

var downloaderIdGenertor middleware.IdGenertor = middleware.NewCyclicIdGenertor()

func genDownloaderId() uint32 {
	return downloaderIdGenertor.GetUint32()
}

func NewPageDownloader(client *http.Client) PageDownloader {
	var id = genDownloaderId()
	if client == nil {
		client = &http.Client{}
	}
	return &myPageDownloader{
		id:         id,
		httpClient: *client,
	}
}

type myDownloaderPool struct {
	pool  middleware.Pool
	etype reflect.Type
}

func (this *myDownloaderPool) Take() (PageDownloader, error) {
	entity, err := this.pool.Take()
	if err != nil {
		return nil, err
	}
	dl, ok := entity.(PageDownloader)
	if !ok {
		errMsg := fmt.Sprintf("The type of entity is NOT %s!\n", this.etype)
		panic(errors.New(errMsg))
	}
	return dl, nil

}

func (this *myDownloaderPool) Return(pl PageDownloader) error {
	return this.pool.Return(pl)
}

func (this *myDownloaderPool) Total() uint32 {
	return this.pool.Total()
}

func (this *myDownloaderPool) Used() uint32 {
	return this.pool.Used()
}

func NewDownloaderPool(total uint32, gen GenPageDownloader) (PageDownloaderPool, error) {
	etype := reflect.TypeOf(gen())
	genEntity := func() middleware.Entity {
		return gen()
	}
	pool, err := middleware.NewPool(total, etype, genEntity)
	if err != nil {
		return nil, err
	}
	thisPool := &myDownloaderPool{
		pool:  pool,
		etype: etype,
	}
	return thisPool, nil
}

type GenPageDownloader func() PageDownloader
