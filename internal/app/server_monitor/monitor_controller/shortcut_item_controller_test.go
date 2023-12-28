package monitor_controller

import (
	"net/http"
	url2 "net/url"
	"testing"
)

func TestExtractWebsiteInfoFromUrl(t *testing.T) {
	urls := []string{
		"http://www.baidu.com/?tn=sitehao123_15",
		"http://www.sina.com.cn/",
		"https://weibo.com/",
		"http://www.sohu.com/",
		"http://tuijian.hao123.com/",
		"http://www.qq.com/",
		"http://www.163.com/",
		"http://map.baidu.com/",
		"https://game.hao123.com/?idfrom=4086",
		"http://wan.baidu.com/home?idfrom=4087",
		"http://tuijian.hao123.com/?type=rec",
		"http://v.hao123.baidu.com/",
		"https://s.click.taobao.com/8CXfaJu",
		"https://mos.m.taobao.com/union/taobaoTMPC?pid=mm_43125636_4246598_115221650492",
		"https://union-click.jd.com/jdc?d=iEZf6v",
		"https://p4psearch.1688.com/hamlet.html?scene=6&cosite=hao123daohang&location=mingzhan",
		"https://s.click.taobao.com/nRjsjOu",
		"https://www.ctrip.com/?allianceid=1630&sid=1911",
		"https://haokan.baidu.com/",
		"https://www.taobao.com/",
		"https://www.bilibili.com/",
		"http://www.iqiyi.com/",
		"http://v.hao123.baidu.com/dianshi",
		"https://mos.m.taobao.com/union/jhsjx2020?pid=mm_43125636_4246598_109944300468",
		"http://tejia.hao123.com/",
		"http://www.eastmoney.com/",
		"https://www.zhihu.com/",
		"https://yiyan.baidu.com/?from=25",
		"https://www.12306.cn/",
		"https://www.ifeng.com/",
		"https://www.chsi.com.cn/",
		"https://www.douban.com/",
	}
	for _, url := range urls {
		t.Run("extract "+url, func(t *testing.T) {
			_url, err := url2.Parse(url)
			if err != nil {
				t.Error(err.Error())
				return
			}
			item, err := extractWebsiteInfoFromUrl(_url, http.Header{
				"User-Agent":      []string{"Mozilla/5.0 (X11; Linux x86_64; rv:84.0) Gecko/20100101 Firefox/84.0"},
				"Accept-Encoding": []string{"gzip, deflate"},
				"Accept":          []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
			})
			if err != nil {
				t.Error(err.Error())
				return
			}

			t.Logf("%+v", item)
		})
	}
}
