package scheduler

import (
	"errors"
	"fmt"
	"github.com/fmyxyz/goreptile/analyzer"
	"github.com/fmyxyz/goreptile/downloader"
	"github.com/fmyxyz/goreptile/itempipeline"
	"github.com/fmyxyz/goreptile/middleware"
	"strings"
)

func generateItemPipelLine(itemProcessors []itempipeline.ProcessItem) itempipeline.ItemPipeline {
	return itempipeline.NewItemPipeline(itemProcessors)
}

func generateAnalyzerPool(poolSize uint32) (analyzer.AnalyzerPool, error) {
	gen := func() analyzer.Analyzer {
		return analyzer.NewAnalyzer()
	}
	return analyzer.NewAnalyzerPool(poolSize, gen)
}
func generateChannelManager(channelLen uint) middleware.ChannelManager {
	return middleware.NewChannelManager(channelLen)
}

func generatePageDownloadPool(poolSize uint32, genHttpClient GenHttpClient) (downloader.PageDownloaderPool, error) {
	gen := func() downloader.PageDownloader {
		return downloader.NewPageDownloader(genHttpClient())
	}
	return downloader.NewDownloaderPool(poolSize, gen)
}

func getPrimaryDomain(host string) (string, error) {
	hp := strings.Split(host, ":")
	if len(hp) != 2 {
		return "", errors.New("The host is invalid!")
	}
	return hp[0], nil
}

func parseCode(code string) []string {
	return strings.Split(code, "-")
}

func generateCode(code string, id uint32) string {
	return fmt.Sprintf("%s-%d", code, id)
}
