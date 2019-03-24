package sri

import (
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"bufio"
	"os"
	"hash"
	"io"
	"regexp"
	"strings"
	"encoding/base64"
	"sync"
)


type Srier struct {
	Webdir	string
	fnname	string
	pool	sync.Pool
	cache	map[string]string
}

var (
	repScript		*regexp.Regexp
	repCss			*regexp.Regexp
	repImg			*regexp.Regexp
	repIntegrity	*regexp.Regexp
	HashFunc		map[string]func() hash.Hash
)

func init() {
	repScript, _ = regexp.Compile(`\s*<script.*src=([\"\'])(\S*\.js)([\"\']).*></script>`)
	repCss, _ = regexp.Compile(`\s*<link.*href=([\"\'])(\S*\.css)([\"\']).*>`)
	repImg, _ = regexp.Compile(`\s*<img.*src=([\"\'])(\S*)([\"\']).*>`)
	repIntegrity, _ = regexp.Compile(`.*\s+integrity=[\"\'](\S*)[\"\'].*`)
	HashFunc = make(map[string]func() hash.Hash)
	HashFunc["sha256"] = sha256.New
	HashFunc["sha512"] = sha512.New
}

func NewSrier() *Srier{
	return &Srier{
		Webdir:	"/data/web/static",
		fnname:	"sha512",
		pool:	sync.Pool{
			New:	func() interface{} {
				return sha512.New()
			},
		},
		cache:	make(map[string]string),
	}
}

func (sri *Srier) Hash(name string) *Srier {
	fn, ok := HashFunc[name]
	if ok {
		sri.fnname = name
		sri.pool = sync.Pool {
			New: func() interface{} {
				return fn()
			},
		}
	}
	return sri
}

func (sri *Srier) Calculate(path string) error {
	// 检测文件大小
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fileInfo.Size() > 10 << 20 {
		return fmt.Errorf("%s file is to long, size: %d", path, fileInfo.Size())
	}
	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	basedir := path[:strings.LastIndex(path, "/")+1]
	br := bufio.NewReader(file)
	target := &strings.Builder{}
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// match script
		params := repScript.FindStringSubmatch(line)
		// match css
		if len(params) == 0 {
			params = repCss.FindStringSubmatch(line)
		}
		// 判断是否匹配数据
		if len(params) > 1 {
			filePath := params[2]
			// 计算SRI
			val, _ := sri.getValue(basedir ,filePath)
			// 捕获原SRI
			paramInter := repIntegrity.FindStringSubmatch(line)
			if len(paramInter) > 1 {
				oldsri := paramInter[1]
				// 对比SRI 不同修改
				if oldsri != val {
					line = strings.ReplaceAll(line, oldsri, val)
				}
			}else {
				// 添加SRI值
				line = strings.ReplaceAll(line, filePath, fmt.Sprintf(`%s%s integrity=%s%s`, filePath, params[1], params[3], val))
			}
		}
		target.WriteString(line)
	}
	fmt.Println(target.String())
	return nil
}

func (sri *Srier) getPath(basedir ,path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}else if path[0] == '/' {
		return sri.Webdir + path
	}else {
		return basedir + path
	}
}

func (sri *Srier) getValue(basedir ,path string) (val string, err error) {
	// 转换文件路径
	path = sri.getPath(basedir, path)
	// 检测缓存
	val, ok := sri.cache[path]
	if ok {
		return val, err
	}
	// 缓存SRI值
	defer func() {
		val = sri.fnname + "-" + val
		sri.cache[path] = val
	}()
	// 读取数据源
	var read io.ReadCloser 
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {

	}else {
		read, err = os.Open(path)
	}
	if err != nil {
		return
	}
	// 计算SRI
	h := sri.pool.Get().(hash.Hash)
	h.Reset()
	defer read.Close()
	defer sri.pool.Put(h)
	_, err = io.Copy(h, read)
	if err != nil {
		return
	}
	val = base64.StdEncoding.EncodeToString(h.Sum(nil))
	return
}




func GetStatic(path string) ([]string, error) {
	// 检测文件大小
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if fileInfo.Size() > 10 << 20 {
		return nil, fmt.Errorf("%s file is to long, size: %d", path, fileInfo.Size())
	}
	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var statics []string
	br := bufio.NewReader(file)
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		// match script
		params := repScript.FindStringSubmatch(line)
		// match css
		if len(params) == 0 {
			params = repCss.FindStringSubmatch(line)
		}
		if len(params) == 0 {
			params = repImg.FindStringSubmatch(line)
		}
		// 判断是否匹配数据
		if len(params) > 1 {
			statics = append(statics, params[2])
		}
	}
	return statics, nil
}