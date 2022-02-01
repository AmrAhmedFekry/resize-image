package main

import (
	"bytes"
	"context"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/chai2010/webp"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/nfnt/resize"
)

func main() {
	handleRequests()
}

/**
* Handle Requests
 */
func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/resize_image", resizeImage).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func resizeImage(w http.ResponseWriter, r *http.Request) {
	imageLink := r.FormValue("image_link")
	imageDirctory := r.FormValue("directory_name")
	extension := filepath.Ext(imageLink)
	resp, err := http.Get(imageLink)
	imageName := strings.TrimSuffix(filepath.Base(imageLink), extension)

	if err != nil {
		return
	}
	defer resp.Body.Close()
	if extension == ".jpg" {
		ResizeJpeg(resp.Body, imageDirctory, imageName+"-resized.jpg")
	} else if extension == ".jpeg" {
		ResizeJpeg(resp.Body, imageDirctory, imageName+"-resized.jpeg")
	} else if extension == ".png" {
		ResizePng(resp.Body, imageDirctory, imageName+"-resized.png")
	} else if extension == ".webp" {
		ResizeWebp(resp.Body, imageDirctory, imageName+"-resized.webp")
	}

}

func ResizeWebp(r io.Reader, imageDirctory string, fileName string) {
	img, err := webp.Decode(r)
	if err != nil {
		panic(err)
	}

	newImage := resize.Resize(500, 0, img, resize.Lanczos3)

	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	webp.Encode(&buf, newImage, nil)
	uploadToAws(buf.Bytes(), imageDirctory, fileName, "webp")
}

func ResizeJpeg(r io.Reader, imageDirctory string, fileName string) {
	img, err := jpeg.Decode(r)
	if err != nil {
		panic(err)
	}

	newImage := resize.Resize(500, 0, img, resize.Lanczos3)

	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	jpeg.Encode(&buf, newImage, nil)
	uploadToAws(buf.Bytes(), imageDirctory, fileName, "jpg")
}

func ResizePng(r io.Reader, imageDirctory string, fileName string) {
	img, err := png.Decode(r)
	if err != nil {
		panic(err)
	}

	newImage := resize.Resize(500, 0, img, resize.Lanczos3)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	png.Encode(&buf, newImage)
	uploadToAws(buf.Bytes(), imageDirctory, fileName, "png")
}

func uploadToAws(imageBytes []byte, imageDirctory string, fileName string, fileType string) {
	godotenv.Load(".env")

	s3Config := &aws.Config{
		Region:      aws.String(os.Getenv("AWS_DEFAULT_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	}
	s3Session := session.New(s3Config)

	uploader := s3manager.NewUploader(s3Session)
	input := &s3manager.UploadInput{
		Bucket:      aws.String(os.Getenv("AWS_BUCKET")),        // bucket's name
		Key:         aws.String(imageDirctory + "/" + fileName), // files destination location
		Body:        bytes.NewReader(imageBytes),                // content of the file
		ContentType: aws.String("image/" + fileType),            // content
	}
	uploader.UploadWithContext(context.Background(), input)
}
