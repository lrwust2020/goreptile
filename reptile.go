package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fmyxyz/goreptile/analyzer"
	"github.com/fmyxyz/goreptile/base"
	"github.com/fmyxyz/goreptile/itempipeline"
	sched "github.com/fmyxyz/goreptile/scheduler"
	"github.com/fmyxyz/goreptile/tool"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	var channelArgs = base.NewChannelArgs(10, 10, 10, 10)

	var poolBaseArgs = base.NewPoolBaseArgs(3, 3)
	var crawlDepth uint32 = 1
	var httpClientGennerator = genHttpClien
	var respParsers = getResponseParsers()
	var itemProcessors = gerItemProcessors()

	var startUrl = "https://www.csdn.net/"
	firstHttpReq, err := http.NewRequest("GET", startUrl, nil)
	if err != nil {
		logger.Println(err)
		return
	}

	scheduler := sched.NewScheduler()

	interverNs := 10 * time.Millisecond
	var maxIdleCount uint = 1000
	checkChan := tool.Monitoring(scheduler, interverNs, maxIdleCount, true, false, record)

	//go func() {
	scheduler.Start(
		channelArgs,
		poolBaseArgs,
		crawlDepth,
		httpClientGennerator,
		respParsers,
		itemProcessors,
		firstHttpReq,
	)
	//}()

	//scheduler.Stop()
	<-checkChan
}

func record(level byte, content string) {
	if content == "" {
		return
	}
	switch level {
	case 0:
		logger.SetPrefix("info")
		logger.Println(content)
	case 1:
		logger.SetPrefix("warn")
		logger.Println(content)
	case 2:
		logger.SetPrefix("info")
		logger.Println(content)
	}
}

func gerItemProcessors() []itempipeline.ProcessItem {
	processItems := []itempipeline.ProcessItem{
		processItem,
	}
	return processItems
}

func processItem(item base.Item) (result base.Item, err error) {
	if item == nil {
		return nil, errors.New("Invalid item!")
	}
	result = make(map[string]interface{})
	for k, v := range item {
		result[k] = v
	}
	if _, ok := result["number"]; !ok {
		result["number"] = len(result)
	}
	time.Sleep(10 * time.Millisecond)
	return result, nil
}

func getResponseParsers() []analyzer.ParseResponse {
	parses := []analyzer.ParseResponse{
		parserForATag,
	}
	return parses
}
func genHttpClien() *http.Client {
	return &http.Client{}
}

var logger = log.New(os.Stdout, "scheduler", log.LstdFlags)

func parserForATag(httpResp *http.Response, respDepth uint32) ([]base.Data, []error) {
	if httpResp.StatusCode != 200 {
		err := errors.New(fmt.Sprintf("Unsupported status code %d. (httpResponse=%v)", httpResp.StatusCode))
		return nil, []error{err}
	}
	var reqUrl *url.URL = httpResp.Request.URL
	var httpRespBody io.ReadCloser = httpResp.Body
	defer func() {
		if httpRespBody != nil {
			httpRespBody.Close()
		}

	}()
	var dataList = make([]base.Data, 0)
	var errs = make([]error, 0)
	doc, err := goquery.NewDocumentFromReader(httpRespBody)
	if err != nil {
		errs = append(errs, err)
		return dataList, errs
	}

	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		var href, exist = sel.Attr("href")
		if !exist || href == "" || href == "#" || href == "/" {
			return
		}
		href = strings.TrimSpace(href)
		var lowerHref = strings.ToLower(href)

		if !strings.HasPrefix(lowerHref, "javascript") {
			aUrl, err := url.Parse(href)
			if err != nil {
				errs = append(errs, err)
				return
			}
			if !aUrl.IsAbs() {
				aUrl = reqUrl.ResolveReference(aUrl)
			}
			httpReq, err := http.NewRequest("GET", aUrl.String(), nil)
			if err != nil {
				errs = append(errs, err)
			} else {
				req := base.NewRequest(httpReq, respDepth)
				dataList = append(dataList, req)
			}
		}
		text := strings.TrimSpace(sel.Text())

		if text != "" {
			imap := make(map[string]interface{})
			imap["a.text"] = text
			imap["parent_url"] = reqUrl
			item := base.Item(imap)
			dataList = append(dataList, item)
		}

	})
	return dataList, errs
}
