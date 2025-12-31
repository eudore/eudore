package middleware

import (
	"strconv"
	"strings"
	"unsafe"

	"github.com/eudore/eudore"
)

// NewUserAgentFunc creates middleware that parses the [eudore.HeaderUserAgent]
// to extract and save device information to the [eudore.ParamBrowser].
//
// It uses the provided or default parsing rules [DefaultUserAgentRules].
func NewUserAgentFunc(rules []string) Middleware {
	if rules == nil {
		rules = DefaultUserAgentRules
	}
	node := &agentNode{}
	for i := 0; i < len(rules); i += 2 {
		node.insert(rules[i+1], rules[i])
	}

	return func(ctx eudore.Context) {
		val := node.match(ctx.GetHeader(eudore.HeaderUserAgent))
		if val != "" {
			// format
			ptr := *(*[]byte)(unsafe.Pointer(&val))
			for i, v := range ptr {
				if v == '_' {
					ptr[i] = '.'
				}
			}

			// replace driver
			strs := strings.SplitN(val, " ", 3)
			if len(strs) == 3 {
				name, ok := DefaultUserAgentMapping[strs[2]]
				if ok {
					strs[2] = name
					val = strings.Join(strs, " ")
				}
			}
			ctx.SetParam(eudore.ParamBrowser, val)
		}
	}
}

type agentNode struct {
	path    string
	name    string
	pattern string
	child   []*agentNode
	next    *agentNode
}

func (node *agentNode) insert(path, pattern string) *agentNode {
	next := node
	rules := splitUserAgentRule(path)
	for i, rule := range rules {
		if strings.HasPrefix(rule, "$") {
			if i == len(rules)-1 && len(rule) > 1 {
				jump := node.find(rule[1:])
				if jump != nil {
					next.next = jump
					break
				}
			}
			next = next.insertParamNode("$"+strconv.Itoa((i+1)/2), rule[1:])
		} else {
			next = next.insertConstNode(rule)
		}
	}

	next.pattern = pattern
	return next
}

func (node *agentNode) insertConstNode(path string) *agentNode {
	if path == "" {
		return node
	}
	for i := range node.child {
		prefix, find := getSubsetPrefix(path, node.child[i].path)
		if find {
			if len(node.child[i].path) > len(prefix) {
				node.child[i].path = node.child[i].path[len(prefix):]
				node.child[i] = &agentNode{
					path:  prefix,
					child: []*agentNode{node.child[i]},
				}
			}
			return node.child[i].insertConstNode(path[len(prefix):])
		}
	}

	last := &agentNode{path: path}
	node.child = append(node.child, last)
	for i := len(node.child) - 1; i > 0; i-- {
		if node.child[i].path[0] != '#' && node.child[i-1].path[0] == '#' {
			node.child[i], node.child[i-1] = node.child[i-1], node.child[i]
		}
	}
	for i := len(node.child) - 1; i > 0; i-- {
		if node.child[i].name == "" && node.child[i-1].name != "" {
			node.child[i], node.child[i-1] = node.child[i-1], node.child[i]
		}
	}
	return last
}

func (node *agentNode) insertParamNode(name, set string) *agentNode {
	for _, child := range node.child {
		if child.name != "" && child.path == set {
			return child
		}
	}

	next := &agentNode{path: set, name: name}
	node.child = append(node.child, next)
	return next
}

func (node *agentNode) find(path string) *agentNode {
	if path == "" {
		return node
	}

	for _, child := range node.child {
		if strings.HasPrefix(path, child.path) {
			return child.find(path[len(child.path):])
		}
	}
	return nil
}

func (node *agentNode) match(path string) string {
	if path == "" {
		return node.pattern
	}

	for _, child := range node.child {
		if child.name != "" {
			pos := indexBytesSet(path, child.path)
			data := child.match(path[pos:])
			if data != "" {
				return strings.Replace(data, child.name, path[:pos], 1)
			}
		}
		if strings.HasPrefix(path, child.path) {
			data := child.match(path[len(child.path):])
			if data != "" {
				return data
			}
		}
	}
	if node.next != nil {
		data := node.next.match(path)
		if data != "" {
			if node.pattern == "$$" {
				return data
			}
			return strings.Replace(node.pattern, "$$", data, 1)
		}
	}

	return ""
}

func indexBytesSet(path, set string) int {
	pos := len(path)
	for i := range set {
		p := strings.IndexByte(path[:pos], set[i])
		if p != -1 {
			pos = p
		}
	}
	return pos
}

func splitUserAgentRule(str string) []string {
	if strings.HasSuffix(str, "$") {
		str += "{ }"
	}
	var data []string
	var last int
	for str != "" {
		switch {
		case strings.HasPrefix(str, "${"):
			pos := strings.IndexByte(str, '}')
			if pos == -1 {
				pos = 2
			}
			data = append(data, "$"+str[2:pos])
			str = str[pos+1:]
		case strings.HasPrefix(str, "$"):
			data = append(data, str[:2])
			str = str[1:]
		default:
			pos := strings.IndexByte(str, '$')
			if pos == -1 {
				pos = len(str)
			}
			data = append(data, str[last:pos])
			str = str[pos:]
		}
	}
	return data
}

var userAgentRules = []string{
	// pre node
	"", "#desktop ",
	"", "#apple ",
	"", "#android ",
	"", "#driver ",
	// hit hight
	"Windows/10 Chrome/$1", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/$ Safari/537.36",
	"Windows/10 Firefox/$2", "Mozilla/5.0 (Windows NT 10.0; Win64$) Gecko/20100101 Firefox/$",
	"MacOSX/$1 Chrome/$2", "Mozilla/5.0 (Macintosh; Intel Mac OS X ${ )}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/$ Safari/537.36",
	"Linux Firefox/$2", "Mozilla/5.0 (X11; Linux x86_64; $) Gecko/20100101 Firefox/$",
	// Windows and Linux
	"Chrome/$2", "#desktop AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Safari/$",
	"Edge/$4", "#desktop AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Safari/$ Edg/$",
	"Edge/$5", "#desktop AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Safari/$ Edg$/$",
	"Firefox/$1", "#desktop Gecko/20100101 Firefox/$",
	"Firefox/$2", "#desktop Gecko/$ Firefox/$",
	"Opera/$4", "#desktop AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Safari/$ OPR/$",
	"OperaGX/$4", "#desktop AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Safari/$ OPRGX/$ OPR/$",
	"Safari/$2", "#desktop AppleWebKit/$ (KHTML, like Gecko) Version/$ Safari/$",
	"Yandex/$3", "#desktop AppleWebKit/$ (KHTML, like Gecko) Chrome/$ YaBrowser/$ Safari/537.36",
	"Yandex/$3", "#desktop AppleWebKit/$ (KHTML, like Gecko) Chrome/$ YaBrowser/$ Yowser/$ Safari/537.36",
	"QQBrowser/$5", "#desktop AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Safari/$ Core/$ QQBrowser/$",
	"UCBrowser/$3", "#desktop AppleWebKit/$ (KHTML, like Gecko) Chrome/$ UBrowser/$ Safari/537.36",
	"Sogou/$4", "#desktop AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Safari/$ SLBrowser/$ SLBChan/$",
	"115Browser/$4", "#desktop AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Safari/$ 115Browser/$",
	"Maxthon/$2", "#desktop AppleWebKit/$ (KHTML, like Gecko) Maxthon/$ Chrome/$ Safari/537.36",
	"HeadlessChrome/$2", "#desktop AppleWebKit/$ (KHTML, like Gecko) HeadlessChrome/$ Safari/537.36",
	"VivoBrowser/$3", "#desktop AppleWebKit/$ (KHTML, like Gecko) Safari/$ VivoBrowser/$ Chrome$",
	"Windows/10 $$", "Mozilla/5.0 (Windows NT 10.0$) ${#desktop }",
	"Windows/8.1 $$", "Mozilla/5.0 (Windows NT 6.3$) ${#desktop }",
	"Windows/8 $$", "Mozilla/5.0 (Windows NT 6.2$) ${#desktop }",
	"Windows/7 $$", "Mozilla/5.0 (Windows NT 6.1$) ${#desktop }",
	"Windows/vista $$", "Mozilla/5.0 (Windows NT 6.0$) ${#desktop }",
	"Windows/xp $$", "Mozilla/5.0 (Windows NT 5.2$) ${#desktop }",
	"Windows/xp $$", "Mozilla/5.0 (Windows NT 5.1$) ${#desktop }",
	"Windows/2000 $$", "Mozilla/5.0 (Windows NT 5.0$) ${#desktop }",
	"Linux $$", "Mozilla/5.0 (X11; Linux x86_64$) ${#desktop }",
	"Linux/$1 $$", "Mozilla/5.0 (X11; Linux $) ${#desktop }",
	"Linux/Ubuntu $$", "Mozilla/5.0 (X11; Ubuntu; Linux$) ${#desktop }",
	"ChromeOS/$2 $$", "Mozilla/5.0 (X11; CrOS $ $) ${#desktop }",
	"Linux/Ubuntu Chromium/$2", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/$ (KHTML, like Gecko) Ubuntu Chromium/$ Chrome/$ Safari/$",
	"$$", "Mozilla/5.0 (Windows; U; ${Mozilla/5.0 (}",

	// Apple
	"Safari/$2", "#apple AppleWebKit/$ (KHTML, like Gecko) Version/$ Safari/$",
	"Safari/$2", "#apple AppleWebKit/$ (KHTML, like Gecko) Version/$ Mobile/$ Safari/$",
	"Chrome/$2", "#apple AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Safari/$",
	"Chrome/$2", "#apple AppleWebKit/$ (KHTML, like Gecko) CriOS/$ Mobile/$ Safari/$",
	"Firefox/$2", "#apple AppleWebKit/$ (KHTML, like Gecko) FxiOS/$ Safari/$",
	"Firefox/$2", "#apple AppleWebKit/$ (KHTML, like Gecko) FxiOS/$ Mobile/$ Safari/$",
	"Edge/$3", "#apple AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Safari/537.36 Edg/$",
	"115Browser/$4", "#apple AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Safari/$ 115Browser/$",
	"Maxthon/$4", "#apple AppleWebKit/$ (KHTML, like Gecko) Version/$ Safari/$ Maxthon/$",
	"Firefox/$1", "#apple Gecko/20100101 Firefox/$",
	"Firefox/$2", "#apple Gecko/$ Firefox/$",
	"Opera/$4", "#apple AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Safari/$ OPR/$",
	"OperaGX/$4", "#apple AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Safari/$ OPRGX/$ OPR/$",
	"Yandex/$3", "#apple AppleWebKit/$ (KHTML, like Gecko) Chrome/$ YaBrowser/$ Safari/$",
	"Yandex/$3", "#apple AppleWebKit/$ (KHTML, like Gecko) Version/$ YaBrowser/$ Mobile/$ Safari/$",
	"Yandex/$3", "#apple AppleWebKit/$ (KHTML, like Gecko) Version/$ YaBrowser/$ SA/3 Mobile/$ Safari/$",
	"HeadlessChrome/$2", "#apple AppleWebKit/$ (KHTML, like Gecko) HeadlessChrome/$ Safari/$",
	"MacOSX/$1 $$", "Mozilla/5.0 (Macintosh; Intel Mac OS X $; $) ${#apple }",
	"MacOSX/$1 $$", "Mozilla/5.0 (Macintosh; U; Intel Mac OS X $; $) ${#apple }",
	"MacOSX/$1 $$", "Mozilla/5.0 (Macintosh; Intel Mac OS X $) ${#apple }",
	"iPhone/$1 $$", "Mozilla/5.0 (iPhone; CPU iPhone OS $ like Mac OS X) ${#apple }",
	"iPhone/$1 $$", "Mozilla/5.0 (iPhone; U; CPU iPhone OS $ like Mac OS X$) ${#apple }",
	"iPad/$1 $$", "Mozilla/5.0 (iPad; CPU OS $ like Mac OS X) ${#apple }",
	"iPad/$1 $$", "Mozilla/5.0 (iPad; U; CPU OS $ like Mac OS X$) ${#apple }",

	// android
	"Chrome/$2", "#android AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Safari/$",
	"Chrome/$2", "#android AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Mobile Safari/$",
	"Chrome/$3", "#android AppleWebKit/$ (KHTML, like Gecko) Version/$ Chrome/$ Safari/$",
	"Chrome/$3", "#android AppleWebKit/$ (KHTML, like Gecko) Version/$ Chrome/$ Mobile Safari/$",
	"AndroidBrowser/$2", "#android AppleWebKit/$ (KHTML, like Gecko) Version/$ Mobile Safari/$",
	"Edge/$4", "#android AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Mobile Safari/$ Edge/$",
	"Edge/$4", "#android AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Mobile Safari/$ EdgA/$",
	"Yandex/$3", "#android AppleWebKit/$ (KHTML, like Gecko) Chrome/$ YaBrowser/$ Mobile Safari/$",
	"Yandex/$3", "#android AppleWebKit/$ (KHTML, like Gecko) Chrome/$ YaBrowser/$ SA/3 Mobile Safari/$",
	"Opera/$4", "#android AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Mobile Safari/$ OPR/$",
	"MiuiBrowser/$5", "#android AppleWebKit/$ (KHTML, like Gecko) Version/$ Chrome/$ Mobile Safari/$ XiaoMi/MiuiBrowser/$",
	"VivoBrowser/$4", "#android AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Mobile Safari/$ VivoBrowser/$",
	"VivoBrowser/$5", "#android AppleWebKit/$ (KHTML, like Gecko) Version/$ Chrome/$ Mobile Safari/$ VivoBrowser/$",
	"HeyTapBrowser/$4", "#android AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Safari/$ HeyTapBrowser/$", // Oppo
	"HeyTapBrowser/$4", "#android AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Mobile Safari/$ HeyTapBrowser/$", // Oppo
	"HeyTapBrowser/$6", "#android AppleWebKit/$ (KHTML, like Gecko) Version/$ Chrome/$ Android $ Safari/$ HeyTapBrowser/$", // Oppo
	"SamsungBrowser/$2", "#android AppleWebKit/$ (KHTML, like Gecko) SamsungBrowser/$ Chrome/$ Safari/$",
	"SamsungBrowser/$2", "#android AppleWebKit/$ (KHTML, like Gecko) SamsungBrowser/$ Chrome/$ Mobile Safari/$",
	"MQQBrowser/$4", "#android AppleWebKit/$ (KHTML, like Gecko) Version/$ Chrome/$ MQQBrowser/$ Mobile Safari/537.36 COVC/047601",
	"UCBrowser/$4", "#android AppleWebKit/$ (KHTML, like Gecko) Version/$ Chrome/$ UCBrowser/$ Mobile Safari/$",
	"Maxthon/$5", "#android AppleWebKit/$ (KHTML, like Gecko) Version/$ Chrome/$ Mobile Safari/$ Maxthon/$",
	"Phoenix/$4", "#android AppleWebKit/$ (KHTML, like Gecko) Chrome/$ Mobile Safari/$ PHX/$",
	"Silk/$3", "#android AppleWebKit/$ (KHTML, like Gecko) Silk/$ like Chrome/$ Safari/$", // Amazon
	// android driver
	"$$ Vivo/V$1", "#driver V${ )} $) ${#android }",
	"$$ Vivo/V$1", "#driver V$) ${#android }",
	"$$ Vivo/$1", "#driver vivo ${ )} $) ${#android }",
	"$$ Oppo/OPD$1", "#driver OPD$ Build/$) ${#android }",
	"$$ Oppo/$1", "#driver OPPO $) ${#android }",
	"$$ Redmi-Note/$1", "#driver Redmi Note $ $) ${#android }",
	"$$ Redmi/$1", "#driver Redmi $ $) ${#android }",
	"$$ Mi-Note/$1", "#driver Mi Note $ $) ${#android }",
	"$$ Mi-Note/$1", "#driver Mi Note $) ${#android }",
	"$$ Mi-MIX/$1", "#driver Mi MIX $) ${#android }",
	"$$ Mi/$1", "#driver Mi $ $) ${#android }",
	"$$ Mi/$1", "#driver MI $ $) ${#android }",
	"$$ Xiaomi/$1", "#driver Xiaomi $ $) ${#android }",
	"$$ Samsung/SM-$1", "#driver SM-$ Build/$) ${#android }",
	"$$ Samsung/SM-$1", "#driver SM-${/)}/$) ${#android }",
	"$$ Samsung/SM-$1", "#driver SM-$) ${#android }",
	"$$ Samsung/SM-$1", "#driver SAMSUNG SM-${ )} $) ${#android }",
	"$$ Samsung/SM-$1", "#driver SAMSUNG SM-$) ${#android }",
	"$$ LG/$1", "#driver LG-${/)}/$ Build/$) ${#android }",
	"$$ LG/$1", "#driver LG-$ Build/$) ${#android }",
	"$$ Pixel-Pro/$1", "#driver Pixel $ Pro) ${#android }",
	"$$ Pixel-Pro/$1", "#driver Pixel $ Pro Build/$) ${#android }",
	"$$ Pixel/$1", "#driver Pixel $ Build/$) ${#android }",
	"$$ Pixel/$1", "#driver Pixel $) ${#android }",
	"$$ Motorola/razr", "#driver motorola razr $) ${#android }",
	"$$ Motorola/moto", "#driver moto g $) ${#android }",
	"$$ Motorola/moto", "#driver moto g $)) ${#android }",
	"$$ Nokia/$1", "#driver Nokia ${ )} $) ${#android }",
	"$$ Nokia/$1", "#driver Nokia $) ${#android }",
	"$$ ONEPLUS/$1", "#driver ONEPLUS $) ${#android }",
	"$$ Lenovo/$1", "#driver Lenovo $) ${#android }",
	"$$ HTC/$1", "#driver HTC ${ )} $) ${#android }",
	"$$ HTC/$1", "#driver HTC $) ${#android }",
	"$$ Surface", "#driver Surface Duo) ${#android }",
	"$$ Google/NexusOne", "#driver Nexus One $) ${#android }",
	"$$", "#driver K) ${#android }",
	"$$", "#driver en-us; ${#driver }",
	"$$", "#driver en-US; ${#driver }",
	"$$", "#driver en-mk; ${#driver }",
	"$$", "#driver zh-cn; ${#driver }",
	"$$", "#driver fr-fr; ${#driver }", // Français-France
	"$$", "#driver vi-vn; ${#driver }", // -Vietnam
	"$$", "#driver de-; ${#driver }", // Deutsch
	"$$ $1", "#driver $ Build/$) ${#android }",
	"$$ $1", "#driver ${ )} $) ${#android }",
	"$$ $1", "#driver $) ${#android }",
	"Android/$1 $$", "Mozilla/5.0 (Linux; Android $; ${#driver }",
	"Android/$1 $$", "Mozilla/5.0 (Linux; U; Android $; ${#driver }",
	"Android/$1 $$", "Mozilla/5.0 (Linux; arm_64; Android $; ${#driver }",
	"Android/$1 $$ WindowsPhone/10", "Mozilla/5.0 (Windows Phone 10.0; Android $; $ $) ${#android }",
	"Windows $$ X11", "Mozilla/5.0 (X11; Windows) ${#android }",

	// Android Firefox
	"Android/$1 Firefox/$5 Samsung/SM-$2", "Mozilla/5.0 (Android $; Mobile; SM-$/$) Gecko/$ Firefox/$",
	"Android/$1 Firefox/$5 Samsung/SM-$2", "Mozilla/5.0 (Android $; Mobile; SM-$;$) Gecko/$ Firefox/$",
	"Android/$1 Firefox/$4", "Mozilla/5.0 (Android $; $) Gecko/$ Firefox/$",
	"Android Firefox/$3", "Mozilla/5.0 (Android; $) Gecko/$ Firefox/$",
	"none Firefox/$3", "Mozilla/5.0 ($) Gecko/$ Firefox/$",
	// Opera
	"Windows/7 Opera/$1", "Opera/$ (Windows NT 6.1${}",
	"Windows/vista Opera/$1", "Opera/$ (Windows NT 6.0${}",
	"Windows/xp Opera/$1", "Opera/$ (Windows NT 5.1${}",
	"MacOSX Opera/$1", "Opera/$ (Macintosh; Intel Mac OS X;${}",
	"Linux Opera/$1", "Opera/$ (X11; Linux${}",
	"Android/$2 Opera/$1", "Opera/$ (Android $; Linux; Opera ${}",
	"none Opera/$1", "Opera/$ ${}",
	"none Opera/$1", "Opera/$",
	// Dalvik
	"Android/$2 Dalvik/$1", "Dalvik/$ (Linux; U; Android $; ${}",

	// bot
	"Bot Googlebot/2.1", "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Googlebot/2.1; +http://www.google.com/bot.html) Chrome/$ Safari/537.36",
	"Bot Googlebot/2.1", "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/$ Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
	"Bot Bingbot/2.0", "Mozilla/5.0 (compatible; Bingbot/2.0; +http://www.bing.com/bingbot.htm)",
	"Bot Bingbot/2.0", "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm) Chrome/$ Safari/537.36", // error doc
	"Bot Bingbot/2.0", "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/$ Mobile Safari/537.36 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
	"Bot PetalBot", "Mozilla/5.0 (compatible;PetalBot;+https://webmaster.petalsearch.com/site/petalbot)",
	"Bot PetalBot", "Mozilla/5.0 (Linux; Android 7.0;) AppleWebKit/537.36 (KHTML, like Gecko) Mobile Safari/537.36 (compatible; PetalBot;+https://webmaster.petalsearch.com/site/petalbot)",
	"Bot GPTBot", "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; GPTBot/1.2; +https://openai.com/gptbot)", // outdated doc
	"Bot Baiduspider", "Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)",
	"Bot Baiduspider", "Mozilla/5.0 (compatible; Baiduspider-render/2.0; +http://www.baidu.com/search/spider.html)",
	"Bot Baiduspider", "Mozilla/5.0 (Linux;u;Android 4.2.2;zh-cn;) AppleWebKit/534.46 (KHTML,like Gecko) Version/5.1 Mobile Safari/10600.6.3 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)",
	"Bot Baiduspider", "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1 (compatible; Baiduspider-render/2.0; +http://www.baidu.com/search/spider.html)",
	"Bot Bytespider", "Mozilla/5.0 (compatible; Bytespider; https://zhanzhang.toutiao.com/) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.0.0 Safari/537.36",
	"Bot Bytespider", "Mozilla/5.0 (Linux; Android 5.0) AppleWebKit/537.36 (KHTML, like Gecko) Mobile Safari/537.36 (compatible; Bytespider; https://zhanzhang.toutiao.com/)",
	"Bot Bytespider", "Mozilla/5.0 (iPhone; CPU iPhone OS 7_1_2 like Mac OS X) AppleWebKit/537.36 (KHTML, like Gecko) Version/7.0 Mobile Safari/537.36 (compatible; Bytespider; https://zhanzhang.toutiao.com/)",
	"Bot Bytespider", "Mozilla/5.0 (compatible; Bytespider; spider-feedback@bytedance.com) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.0.0 Safari/537.36",
	"Bot 360Spider", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36; 360Spider",
	// bot less
	"Bot Googlebot/2.1", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
	"Bot Googlebot/1.0", "Mozilla/5.0 (X11; Linux x86_64; Storebot-Google/1.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/$ Safari/537.36",
	"Bot Googlebot/1.0", "Mozilla/5.0 (Linux; Android 8.0; Pixel 2 Build/OPD3.170816.012; Storebot-Google/1.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/$ Mobile Safari/537.36",
	"Bot Googlebot/1.0", "Mozilla/5.0 (compatible; Google-InspectionTool/1.0;)",
	"Bot Googlebot/1.0", "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/$ Mobile Safari/537.36 (compatible; Google-InspectionTool/1.0;)",
	"Bot Googlebot", "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/$ Mobile Safari/537.36 (compatible; GoogleOther)",
	"Bot Googlebot", "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; GoogleOther) Chrome/$ Safari/537.36",
	"Bot Bingbot/2.0", "Mozilla/5.0 (compatible; adidxbot/2.0; +http://www.bing.com/bingbot.htm)",
	"Bot Bingbot/2.0", "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; MicrosoftPreview/2.0; +https://aka.ms/MicrosoftPreview) Chrome/$ Safari/537.36",
	"Bot Bingbot/2.0", "Mozilla/5.0 (iPhone; CPU iPhone OS 7_0 like Mac OS X) AppleWebKit/537.51.1 (KHTML, like Gecko) Version/7.0 Mobile/11A465 Safari/9537.53 (compatible; adidxbot/2.0; +http://www.bing.com/bingbot.htm)",
	"Bot Bingbot/2.0", "Mozilla/5.0 (Windows Phone 8.1; ARM; Trident/7.0; Touch; rv:11.0; IEMobile/11.0; NOKIA; Lumia 530) like Gecko (compatible; adidxbot/2.0; +http://www.bing.com/bingbot.htm)",
	"Bot Bingbot/2.0", "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/$ Mobile Safari/537.36 (compatible; MicrosoftPreview/2.0; +https://aka.ms/MicrosoftPreview)",
	"Bot Applebot/0.1", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/$ Safari/$ (Applebot/0.1; +http://www.apple.com/go/applebot)",
	"Bot Applebot/0.1", "Mozilla/5.0 (iPhone; CPU iPhone OS $ like Mac OS X) AppleWebKit/$ (KHTML, like Gecko) Version/$ Mobile/$ Safari/$ (Applebot/0.1; +http://www.apple.com/go/applebot)",
	// bot other
	"Bot $1", "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; $; +${}",
	"Bot $1", "Mozilla/5.0 (compatible; $; +http$)",
	"Bot $1", "Mozilla/5.0 (compatible; $; http$)",
	"Bot $1", "Mozilla/5.0 (compatible; $; $; +http$)",
	"Bot Googlebot/2.1", "Googlebot/2.1 (+http://www.google.com/bot.html)",
	"Bot GoogleBot/1.0", "Googlebot-Image/1.0",
	"Bot GoogleBot/1.0", "Googlebot-Video/1.0",
	"Bot GoogleBot/1.0", "GoogleOther-Image/1.0",
	"Bot GoogleBot/1.0", "GoogleOther-Video/1.0",
	"Bot GoogleBot", "Google-CloudVertexBot",
	"Bot facebook/1.1", "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
	"Bot facebook/1.1", "facebookexternalhit/1.1",
	"Bot facebook/1.0", "facebookcatalog/1.0",
	"Bot Sogou/4.0", "Sogou $ spider/4.0(+http://www.sogou.com/docs/help/webmasters.htm#07)",
	"Bot Sogou/4.0", "Sogou $ Spider/4.0(+http://www.sogou.com/docs/help/webmasters.htm#07)",
	"Bot Sogou/3.0", "Sogou $ Spider/3.0(+http://www.sogou.com/docs/help/webmasters.htm#07)",
	// none
	"none curl/$1", "curl/$",
	"none golang/$1", "Go-http-client/$",
	"none java/$1", "Java/$",
	"none python-requests/$1", "python-requests/$",
}

var userAgentMapping = map[string]string{
	"KFRAPWI":           "AmazonFireHD/8",
	"I2212":             "iQOO/11", // Vivo
	"A37f":              "Oppo/A37F",
	"CPH2631":           "Oppo/A60",
	"CPH1613":           "Oppo/F3P",
	"CPH1607":           "Oppo/R9s",
	"CPH2519":           "Oppo/FindN3",
	"RMX3840":           "Realme/12P",
	"24129RT7CC":        "Redmi/4",
	"M2101K6G":          "Redmi-Note/10P",
	"2201116SG":         "Redmi-Note/11P",
	"23129RAA4G":        "Redmi-Note/13",
	"Samsung/SM-A5560":  "Samsung/A55",
	"Samsung/SM-F9560":  "SamsungFlip",
	"Samsung/SM-F956U":  "SamsungFlip",
	"Samsung/SM-S901B":  "Samsung/S22",
	"Samsung/SM-S911B":  "Samsung/S23",
	"Samsung/SM-S911U":  "Samsung/S23",
	"Samsung/SM-S928B":  "Samsung/S24",
	"Samsung/SM-S928W":  "Samsung/S24",
	"Samsung/SM-S931B":  "Samsung/S25",
	"Samsung/SM-S931U":  "Samsung/S25",
	"Samsung/SM-G900P":  "Samsung/S5",
	"Samsung/SM-G935F":  "Samsung/S7",
	"Samsung/SM-T230NU": "SamsungTab/4",
	"Samsung/SM-T550":   "SamsungTab/A",
	"Samsung/SM-T827R4": "SamsungTab/S3",
	"Samsung/SM-X906C":  "SamsungTab/S8",
	"Samsung/SM-G556B":  "SamsungXcover/7",
	"Samsung/SM-F956U1": "Samsung/ZFold6",
	"24030PN60G":        "Xiaomi/14",
	"M2102J20SG":        "Xiaomi/X3P",
}
