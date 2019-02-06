package sri

import (
	"crypto/sha512"
	"fmt"
	"bufio"
	"os"
	"io"
	"regexp"
	"encoding/base64"
)


func HashSHA256File(filePath string) (string, error){
	var hashValue string
	file, err := os.Open(filePath)
	if err != nil {
		return hashValue, err
	}
	defer file.Close()
	hash := sha512.New()
	if _, err := io.Copy(hash, file); err != nil {
		return hashValue,  err
	}
	signedStr := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	// sha256 string
	// hashInBytes := hash.Sum(nil)
	// hashValue = hex.EncodeToString(hashInBytes)
	return signedStr, nil

}

func Match(filePath string) error {
	fi, err := os.Open(filePath)
	if err != nil {
        return err
    }
    defer fi.Close()
	rep, err := regexp.Compile(`\s*<script.*src=[\"\'](\S*\.js)[\"\'].*></script>`)
	if err != nil{
        return err
	}


    br := bufio.NewReader(fi)
    for {
        a, _, c := br.ReadLine()
        if c == io.EOF {
            break
        }
        params := rep.FindSubmatch(a)
        if len(params) > 1 {
        	fmt.Println(string(a))
        	fmt.Println(string(params[1]))
        }

    }
    return nil
}