package downloader

import "github.com/fmyxyz/goreptile/base"

//网页下载器
type PageDownloader interface {
	Id() uint32
	Download(req base.Request) (*base.Response, error)
}

//网页下载器池
type PageDownloaderPool interface {
	Tack() (PageDownloader ,error)
	Return(pl PageDownloader) error
	Total()uint32
	Used() uint32
}

