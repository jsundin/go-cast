package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/eliukblau/pixterm/pkg/ansimage"
	"github.com/go-vgo/robotgo"
	"github.com/kbinani/screenshot"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

type config_t struct {
	IPCheckHost string

	QRWidth  int
	QRHeight int
	QRMargin int

	Host string
	Port int

	SecretLen int
	ViewHits  bool
}

func findIP(conf config_t) net.IP {
	conn, err := net.Dial("udp", conf.IPCheckHost)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	la := conn.LocalAddr().(*net.UDPAddr)
	return la.IP
}

func findRect() image.Rectangle {
	fmt.Print("Click upper left position: ")
	robotgo.AddEvent("mleft")
	x0, y0 := robotgo.GetMousePos()
	fmt.Printf("%d,%d\n", x0, y0)

	fmt.Print("Click lower right position: ")
	robotgo.AddEvent("mleft")
	x1, y1 := robotgo.GetMousePos()
	fmt.Printf("%d,%d\n", x1, y1)

	if x1 <= x0 || y1 <= y0 {
		panic("bad rectangle")
	}

	return image.Rect(x0, y0, x1, y1)
}

func makeQR(conf config_t, secret string) {
	enc := qrcode.NewQRCodeWriter()
	myURL := fmt.Sprintf("http://%s:%d/%s", conf.Host, conf.Port, secret)
	img, err := enc.Encode(myURL, gozxing.BarcodeFormat_QR_CODE, conf.QRWidth, conf.QRHeight, map[gozxing.EncodeHintType]interface{}{
		gozxing.EncodeHintType_MARGIN: conf.QRMargin,
	})
	if err != nil {
		panic(err)
	}

	aimg, err := ansimage.NewFromImage(img, color.Black, ansimage.NoDithering)
	if err != nil {
		panic(err)
	}
	aimg.DrawExt(false, false)
	fmt.Println("Visit:", myURL)
}

func getRandomString(n int) string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJLMNOPQRSTUVWXYZ0123456789"
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := ""
	for i := 0; i < n; i++ {
		s += string(letters[rnd.Int()%len(letters)])
	}
	return s
}

func cap(r image.Rectangle) image.Image {
	img, err := screenshot.CaptureRect(r)
	if err != nil {
		panic(err)
	}
	return img
}

func main() {
	conf := config_t{
		QRWidth:   75,
		QRHeight:  75,
		QRMargin:  1,
		Port:      8000,
		SecretLen: 10,
	}

	if conf.IPCheckHost == "" {
		conf.IPCheckHost = "1.1.1.1:80"
	}

	if conf.Host == "" {
		conf.Host = findIP(conf).String()
	}

	secret := getRandomString(conf.SecretLen)

	rect := findRect()

	makeQR(conf, secret)

	http.HandleFunc("/"+secret, func(w http.ResponseWriter, r *http.Request) {
		if conf.ViewHits {
			fmt.Println("hit from:", r.RemoteAddr)
		}
		img := cap(rect)
		w.Header().Add("Content-Type", "image/png")
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Expires", "0")
		png.Encode(w, img)
	})

	err := http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), nil)
	if err != nil {
		panic(err)
	}
}
