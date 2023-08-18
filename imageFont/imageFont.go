package imageFont

import (
	"bytes"
	"context"
	"fmt"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"gptbot/store"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"os"
)

const (
	templatePath = "./imageFont/template.jpg"
	fontPath     = "./imageFont/simkai.ttf"
	lineDistance = 60
)

type ImageFont struct {
	newTemplateImage  *image.RGBA
	templateFileImage image.Image
	font              *truetype.Font
	content           *freetype.Context
	x                 int
	y                 int
}

var (
	fontKai *truetype.Font // 字体
)

func NewImageFont() (*ImageFont, error) {
	_ = os.MkdirAll("./imageFont/", 0644)
	_, err := os.Stat(templatePath)
	if os.IsNotExist(err) {
		_, err := store.GetStoreClient().Object.GetToFile(context.Background(), "imageFont/template.jpg", templatePath, nil)
		if err != nil {
			return nil, err
		}
	}
	_, err = os.Stat(fontPath)
	if os.IsNotExist(err) {
		_, err = store.GetStoreClient().Object.GetToFile(context.Background(), "imageFont/simkai.ttf", fontPath, nil)
		if err != nil {
			return nil, err
		}
	}

	templateFile, err := os.Open(templatePath)
	if err != nil {
		return nil, err
	}
	defer templateFile.Close()
	// 解码
	templateFileImage, err := jpeg.Decode(templateFile)
	if err != nil {
		return nil, err
	}
	// 新建一张和模板文件一样大小的画布
	newTemplateImage := image.NewRGBA(templateFileImage.Bounds())
	// 将模板图片画到新建的画布上
	draw.Draw(newTemplateImage, templateFileImage.Bounds(), templateFileImage, templateFileImage.Bounds().Min, draw.Over)

	fontKai, err = loadFont(fontPath)
	if err != nil {
		return nil, err
	}

	content := freetype.NewContext()
	content.SetClip(newTemplateImage.Bounds())
	content.SetDst(newTemplateImage)
	content.SetSrc(image.Black)
	content.SetDPI(50)

	content.SetFontSize(60)
	content.SetFont(fontKai)

	return &ImageFont{
		newTemplateImage:  newTemplateImage,
		templateFileImage: templateFileImage,
		font:              fontKai,
		content:           content,
		x:                 20,
		y:                 50,
	}, nil
}

func (f *ImageFont) Write(str string) {
	f.content.DrawString(str, freetype.Pt(f.x, f.y))
	f.y += lineDistance
}

func (f *ImageFont) GetImage() []byte {
	buf := []byte{}
	buffer := bytes.NewBuffer(buf)
	_ = jpeg.Encode(buffer, f.newTemplateImage, &jpeg.Options{Quality: 75})
	return buffer.Bytes()
}

func (f *ImageFont) Clear() {

	f.newTemplateImage = image.NewRGBA(f.templateFileImage.Bounds())
	// 将模板图片画到新建的画布上
	draw.Draw(f.newTemplateImage, f.templateFileImage.Bounds(), f.templateFileImage, f.templateFileImage.Bounds().Min, draw.Over)

	f.y = 50
	f.content = freetype.NewContext()
	f.content.SetClip(f.newTemplateImage.Bounds())
	f.content.SetDst(f.newTemplateImage)
	f.content.SetSrc(image.Black)
	f.content.SetDPI(50)

	f.content.SetFontSize(60)
	f.content.SetFont(fontKai)
}

// 根据路径加载字体文件
// path 字体的路径
func loadFont(path string) (font *truetype.Font, err error) {
	var fontBytes []byte
	fontBytes, err = ioutil.ReadFile(path) // 读取字体文件
	if err != nil {
		err = fmt.Errorf("加载字体文件出错:%s", err.Error())
		return
	}
	font, err = freetype.ParseFont(fontBytes) // 解析字体文件
	if err != nil {
		err = fmt.Errorf("解析字体文件出错,%s", err.Error())
		return
	}
	return
}
