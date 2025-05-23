package cubing

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/guojia99/cubing-pro/src/internel/crawler/cubing/cubing_city"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type DCubingCompetition struct {
	startYear int
	endYear   int

	currYear int
	nexYear  int

	cubingCity []string
	oldKey     map[string]bool // 原本就有的比赛Key
}

func NewDCubingCompetition() *DCubingCompetition {
	c := &DCubingCompetition{
		startYear: 2015,
		endYear:   2025,
	}
	c.updateYear()
	c.updateCityList()
	return c
}

func (c *DCubingCompetition) updateYear() {
	currentYear := time.Now().Year()
	currentMonth := int(time.Now().Month())
	c.currYear = currentYear
	c.nexYear = currentYear + 1
	c.endYear = currentYear
	if currentMonth >= 9 {
		c.endYear += 1
	}
}

func (c *DCubingCompetition) updateCityList() {
	c.oldKey = make(map[string]bool)
	c.cubingCity = make([]string, 0)

	// 粗饼城市
	cubingCity, oldKeys := cubing_city.GetCubingCityListAndOldKey(c.startYear, c.endYear)
	for _, o := range oldKeys {
		c.oldKey[o] = true
	}
	c.cubingCity = append(c.cubingCity, cubingCity...)

	// one城市
	c.cubingCity = append(c.cubingCity, cubing_city.GetOneCityList()...)

	// 其他补充的city
	c.cubingCity = append(c.cubingCity, cubing_city.OtherCitys()...)

	c.cubingCity = utils.RemoveDuplicates(c.cubingCity)
}

// CompNameWithMouth 第一个为 今年， 第二个为明年
var CompNameWithMouth = map[string][2][]int{
	"Open":        {{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, {10, 11, 12}},
	"Spring":      {{1, 2, 3}, {10, 11, 12}},
	"Spring-Open": {{1, 2, 3}, {10, 11, 12}},
	"Summer":      {{2, 3, 4, 5, 6}, {}},
	"Summer-Open": {{2, 3, 4, 5, 6}, {}},
	"Autumn":      {{6, 7, 8, 9, 10}, {}},
	"Autumn-Open": {{6, 7, 8, 9, 10}, {}},
	"Winter":      {{1, 2}, {9, 10, 11, 12}},
	"Winter-Open": {{1, 2}, {9, 10, 11, 12}},
	"Newcomers":   {{1, 2}, {12}},
	"New-Year":    {{1, 2}, {11, 12}},
}

func (c *DCubingCompetition) newKeyWithCity() []string {
	currentMonth := int(time.Now().Month())
	var outKeys []string

	for key, ms := range CompNameWithMouth {
		curYearMouth := ms[0]
		nextYearMouth := ms[1]

		for _, city := range c.cubingCity {
			if slices.Contains(curYearMouth, currentMonth) {
				outKeys = append(outKeys, fmt.Sprintf("%s-%s-%d", city, key, c.currYear))
			}
			if slices.Contains(nextYearMouth, currentMonth) {
				outKeys = append(outKeys, fmt.Sprintf("%s-%s-%d", city, key, c.nexYear))
			}
		}
	}

	for _, key := range cubing_city.OtherKeys() {
		outKeys = append(outKeys, fmt.Sprintf("%s-%d", key, c.currYear))
		outKeys = append(outKeys, fmt.Sprintf("%s-%d", key, c.nexYear))
	}

	outKeys = utils.RemoveDuplicates(outKeys)
	return outKeys
}

type TCubingCompetition struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Url    string
	Date   string `json:"date"`
	Events string `json:"events"`
}

// 随机生成 UA
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.3 Safari/605.1.15",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 16_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
}

var languages = []string{
	"zh-CN,zh;q=0.9,en;q=0.8",
	//"zh-CN,zh-HK;q=0.9,en;q=0.8",
	//"zh-TW,zh;q=0.9,en;q=0.7",
}

func RandomHeaders() map[string]interface{} {
	rand.Seed(time.Now().UnixNano())
	return map[string]interface{}{
		"Cache-Control":             "no-cache",
		"Accept-Language":           languages[rand.Intn(len(languages))],
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"User-Agent":                userAgents[rand.Intn(len(userAgents))],
		"Content-Type":              "text/html; charset=UTF-8",
		"Date":                      time.Now().UTC().Format(time.RFC1123),
		"Eagleid":                   fmt.Sprintf("%x", rand.Uint64()),
		"Strict-Transport-Security": "max-age=5184000",
		"Timing-Allow-Origin":       "*",
		"Vary":                      "Accept-Encoding",
	}
}

func (c *DCubingCompetition) getPage(id, url string) (TCubingCompetition, bool, error) {
	resp, err := utils.HTTPRequestFull("GET", url, nil, RandomHeaders(), nil)
	if err != nil {
		return TCubingCompetition{}, false, err
	}

	log.Printf("[%d]check => %s | %s\n", resp.StatusCode, id, url)
	if resp.StatusCode != 200 {
		return TCubingCompetition{}, false, fmt.Errorf("[e] %d", resp.StatusCode)
	}

	//fmt.Println(string(resp))
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
	if err != nil {
		return TCubingCompetition{}, false, err
	}

	var out = TCubingCompetition{
		ID:  id,
		Url: url,
	}
	doc.Find("h1.heading-title").Each(func(i int, s *goquery.Selection) {
		out.Name = s.Text()
	})
	doc.Find("dt#events").NextFiltered("dd").Each(func(i int, s *goquery.Selection) {
		out.Events = s.Text()
	})
	doc.Find("dt").Each(func(i int, s *goquery.Selection) {
		if strings.TrimSpace(s.Text()) == "日期" {
			dd := s.NextFiltered("dd") // 获取紧随其后的 <dd>
			out.Date = strings.TrimSpace(dd.Text())
		}
	})

	return out, true, nil
}

const competitionBase = "https://cubing.com/competition/"

func (c *DCubingCompetition) GetNewCompetitions() []TCubingCompetition {
	baseKeys := c.newKeyWithCity()

	log.Printf("===> 尝试获取新比赛%d条\n", len(baseKeys))

	var (
		find []TCubingCompetition
		wg   sync.WaitGroup
		mu   sync.Mutex
	)

	// 控制最大并发数为 10
	concurrencyLimit := make(chan struct{}, 10)
	idx := 0

	for _, nKey := range baseKeys {
		idx++
		if _, ok := c.oldKey[nKey]; ok {
			continue
		}

		// 获取一个“并发许可”
		concurrencyLimit <- struct{}{}
		wg.Add(1)
		go func(nKey string, idx int) {
			defer wg.Done()
			defer func() { <-concurrencyLimit }() // 释放一个“并发许可”

			pUrl := fmt.Sprintf("%s%s", competitionBase, nKey)
			url, isFind, _ := c.getPage(nKey, pUrl)
			if isFind {
				mu.Lock()
				find = append(find, url)
				mu.Unlock()
				log.Printf("=========== find = %s ==> %s\n", nKey, url)
			}
			time.Sleep(time.Millisecond * 100)
		}(nKey, idx)
	}

	wg.Wait() // 等待所有 goroutine 完成
	return find
}
