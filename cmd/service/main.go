package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	s3Bucket   = os.Getenv("S3_BUCKET")
	awsRegion  = os.Getenv("AWS_REGION")
	apiToken   = os.Getenv("GO_S3_API_TOKEN")
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/upload", uploadHandler, apiTokenMiddleware)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

func apiTokenMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		clientToken := c.Request().Header.Get("X-API-Token")
		if clientToken == "" || clientToken != apiToken {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API Token")
		}
		return next(c)
	}
}

func uploadHandler(c echo.Context) error {
	// Read form file
	file, err := c.FormFile("image")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Error retrieving the file")
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error opening the file")
	}
	defer src.Close()

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error loading AWS configuration")
	}

	s3Client := s3.NewFromConfig(cfg)

	uploadKey := filepath.Base(file.Filename)
	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(uploadKey),
		Body:   src,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error uploading to S3")
	}

	return c.String(http.StatusOK, "File uploaded successfully")
}
