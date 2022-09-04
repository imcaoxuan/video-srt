package main

import "net/http"
import "io/ioutil"
import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/xlzd/gotp"
	"os"
	"path"
	"path/filepath"
	"time"
	"videosrt/videosrt"
	//	"strconv"
)

type Result struct {
	Successful bool
	SubTitle   []byte
	Audio      string
}

//定义配置文件
const CONFIG = "config.ini"
const VIDEO_DIR = "./video/"

// REPLACE WITH YOUR SECRET
const SECRET = "4S62BZNFXXSZLCRO"

func process(video string) {

	//致命错误捕获
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("执行错误 : ", err)
			os.Exit(500)
		}
	}()

	appDir, err := filepath.Abs(filepath.Dir(os.Args[0])) //应用执行根目录
	if err != nil {
		panic(err)
	}

	//初始化
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "")
	}

	//var video string

	//设置命令行参数
	//flag.StringVar(&video, "f", "", "enter a video file waiting to be processed .")

	//flag.Parse()

	//if video == "" && os.Args[1] != "" && os.Args[1] != "-f" {
	//	video = os.Args[1]
	//}

	//获取应用
	app := videosrt.NewApp(CONFIG)

	appDir = videosrt.WinDir(appDir)

	//初始化应用
	app.Init(appDir)

	//调起应用
	app.Run(videosrt.WinDir(video))

	//延迟退出
	time.Sleep(time.Second * 1)
}

func do(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	// otp, _ := strconv.ParseInt(r.PostForm.Get("otp"), 10, 64)
	var otp = r.FormValue("otp")
	totp := gotp.NewDefaultTOTP(SECRET)
	if !totp.Verify(otp, time.Now().Unix()) {
		fmt.Printf("invalid otp %s\n", otp)
		json.NewEncoder(w).Encode(Result{false, []byte{}, ""})
		return
	}
	file, handler, err := r.FormFile("file")
	if err == nil {
		data, err := ioutil.ReadAll(file)
		if err == nil {
			bytes := []byte(data)
			fmt.Println(handler.Filename)
			hash := sha256.New()
			hash.Write(bytes)
			sum := hash.Sum(nil)
			fmt.Printf("%x\n", sum)
			var filename = hex.EncodeToString(sum)
			var filePath string = VIDEO_DIR + filename + path.Ext(handler.Filename)
			ioutil.WriteFile(filePath, bytes, 0666)
			process(filePath)
			srtFile, _ := os.Open(VIDEO_DIR + filename + ".srt")
			srtBytes, _ := ioutil.ReadAll(srtFile)
			json.NewEncoder(w).Encode(Result{true, srtBytes, ""})
		}
	}

}
func main() {
	const address string = ":8090"
	http.HandleFunc("/do", do)
	fmt.Printf("running on %s\n", address)
	http.ListenAndServe(address, nil)
}
