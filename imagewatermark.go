package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/nfnt/resize"
)

var (
	source   string
	dest     string
	water    string
	position int
	waterImg image.Image
)

func main() {
	files, err := ioutil.ReadDir(source)
	if err != nil {
		log.Fatalln(err)
	}
	for _, img := range files {
		if !strings.HasSuffix(img.Name(), "JPG") {
			continue
		}
		resizeImg(img)
	}
}

func mark(img os.FileInfo) {
	fp, err := os.Open(fmt.Sprintf("%s/%s", source, img.Name()))
	if err != nil {
		log.Println(err)
		return
	}
	defer fp.Close()
	imgDec, err := jpeg.Decode(fp)
	if err != nil {
		log.Println(err)
		return
	}
	imgb := imgDec.Bounds()
	rotate270 := image.NewRGBA(image.Rect(0, 0, imgb.Bounds().Dy(), imgb.Bounds().Dx()))
	for x := imgb.Min.Y; x < imgb.Max.Y; x++ {
		for y := imgb.Max.X - 1; y >= imgb.Min.X; y-- {
			rotate270.Set(x, imgb.Max.X-y, imgDec.At(y, x))
		}
	}
	b := rotate270.Bounds()
	newImg := image.NewRGBA(b)
	offset := getWaterPos(rotate270, position)
	draw.Draw(newImg, b, rotate270, image.ZP, draw.Src)
	draw.Draw(newImg, waterImg.Bounds().Add(offset), waterImg, image.ZP, draw.Over)
	imgOut, err := os.Create(dest + "/" + img.Name())
	if err != nil {
		log.Println(err)
		return
	}
	defer imgOut.Close()
	jpeg.Encode(imgOut, newImg, &jpeg.Options{Quality: 100})
}

func resizeImg(img os.FileInfo) {
	fp, err := os.Open(source + "/" + img.Name())
	if err != nil {
		log.Println(err)
		return
	}
	defer fp.Close()
	dec, err := jpeg.Decode(fp)
	if err != nil {
		log.Println(err)
		return
	}
	m := resize.Resize(uint(dec.Bounds().Dx()/2), 0, dec, resize.Lanczos3)
	idx := strings.LastIndex(img.Name(), ".")
	fileName := img.Name()
	out, err := os.Create(dest + "/" + fileName[:idx] + ".big.JPG")
	if err != nil {
		log.Println(err)
		return
	}
	defer out.Close()
	jpeg.Encode(out, m, &jpeg.Options{Quality: 100})
}

func getWaterPos(img image.Image, pos int) image.Point {
	var X, Y int
	X = int(float32(img.Bounds().Dx()) * 0.05)
	Y = int(float32(img.Bounds().Dy()) * 0.05)
	var offset image.Point
	switch pos {
	case 5:
		offset = image.Pt(img.Bounds().Dx()/2-waterImg.Bounds().Dx()/2, img.Bounds().Dy()/2-waterImg.Bounds().Dy()/2)
	case 4:
		offset = image.Pt(img.Bounds().Dx()-waterImg.Bounds().Dx()-X, img.Bounds().Dy()-waterImg.Bounds().Dy()-Y)
	case 3:
		offset = image.Pt(X, img.Bounds().Dy()-waterImg.Bounds().Dy()-Y)
	case 1:
		offset = image.Pt(X, Y)
	case 2:
		offset = image.Pt(img.Bounds().Dx()-waterImg.Bounds().Dx()-X, Y)
	default:
		offset = image.Pt(rand.Intn(img.Bounds().Dx()-waterImg.Bounds().Dx()-2*X)+X, rand.Intn(img.Bounds().Dy()-waterImg.Bounds().Dy()-2*Y)+Y)
	}
	return offset
}

func init() {
	flag.StringVar(&source, "s", "", "待处理源图片所在目录")
	flag.StringVar(&dest, "o", "", "处理完毕后图片输出目录")
	flag.IntVar(&position, "p", 0, "水印位置：0随机，1左上，2右上，3左下，4右下，5居中")
	flag.StringVar(&water, "w", "", "水印图片")
	flag.Parse()
	if source == "" {
		log.Fatalln("请输入源图片目录")
	}
	if dest == "" {
		log.Fatalln("请输入输出目录")
	}
	if water == "" {
		log.Fatalln("请输入水印图片")
	}

	_, err := os.Stat(dest)
	if err != nil {
		os.Mkdir(dest, os.ModePerm)
	}

	fp, err := os.Open(water)
	if err != nil {
		log.Fatalln(err)
	}
	defer fp.Close()
	waterImg, err = png.Decode(fp)
	if err != nil {
		log.Fatalln(err)
	}
}
