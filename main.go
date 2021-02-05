package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func ParseHexColor(s string) (clr color.Color, err error) {
	c := color.RGBA{}
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")
	}
	clr = c
	return
}

func img2rgba(in image.Image) *image.RGBA {
	b := in.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), in, b.Min, draw.Src)
	return dst
}

func combine(bg color.Color, imgs ...image.Image) image.Image {
	if len(imgs) == 0 {
		return nil
	}
	if len(imgs) == 1 {
		return imgs[0]
	}

	bgR, bgG, bgB, _ := bg.RGBA()

	dst := img2rgba(imgs[0])

	for _, img2 := range imgs[1:] {
		b := dst.Bounds()
		for x := b.Min.X; x < b.Max.X; x++ {
			for y := b.Min.Y; y < b.Max.Y; y++ {
				color1 := dst.At(x, y)
				color2 := img2.At(x, y)
				r2, g2, b2, a2 := color2.RGBA()
				if r, g, b, _ := color1.RGBA(); r == bgR && g == bgG && b == bgB && (r2 != bgR || b2 != bgB || g2 != bgG) && a2 != 0 {
					dst.Set(x, y, color2)
				}
			}
		}
	}

	return dst
}

func main() {
	rand.Seed(time.Now().UnixNano())

	os.MkdirAll("tmp", 0777)

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.File("index.html")
	})

	r.GET("/api/combine", func(c *gin.Context) {
		clrstr := c.Query("color")
		clr, err := ParseHexColor(clrstr)
		if err != nil {
			clr = color.White
		}

		imgs := strings.Split(strings.Replace(c.Query("url"), "\r", "", -1), "\n")
		imglist := make([]image.Image, 0)
		for _, imgurl := range imgs {
			imgurl = strings.TrimSpace(imgurl)
			if imgurl == "" {
				continue
			}
			resp, err := http.Get(imgurl)
			if err != nil {
				c.String(400, "error: %s", err)
				return
			}
			img, _, err := image.Decode(resp.Body)
			if err != nil {
				c.String(400, "error: %s", err)
				return
			}
			imglist = append(imglist, img)
		}

		dest := combine(clr, imglist...)
		if dest == nil {
			c.String(400, "no images provided")
		}

		id := strconv.FormatInt(rand.Int63(), 36)
		f, err := os.Create("tmp/" + id + ".png")
		if err != nil {
			c.String(500, "internal error")
			log.Println(err)
			return
		}

		err = png.Encode(f, dest)
		if err != nil {
			c.String(500, "internal error")
			log.Println(err)
			return
		}

		c.Redirect(303, "/api/image/"+id)
	})

	r.GET("/api/image/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.Header("Content-Disposition", "attachment; filename=combined.png")
		c.File("tmp/" + id + ".png")
	})

	r.Run()
}
