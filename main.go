package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const uploadPath = "./uploads"

type User struct {
	gorm.Model
	Deleted    int8   `gorm:"default:0"`
	Code       int64  `gorm:"unique;type:bigint"`
	AvatarURLs string `gorm:"type:text"`
	Bio        string `gorm:"type:text"`
}

var (
	db *gorm.DB
)

func init() {
	var err error
	// dsn := fmt.Sprintf("root:123456@tcp(db:3306)/meetyou?charset=utf8mb4&parseTime=True&loc=Local")
	dsn := fmt.Sprintf("root:uy&tr$e189@tcp(127.0.0.1:3306)/meetyou?charset=utf8mb4&parseTime=True&loc=Local")
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&User{})
	// db.Save(&User{Code: "yujian", AvatarURLs: "https://i.pravatar.cc/300", Bio: "I am an engineer"})
}

type Profiles struct {
	AvatarURLs []string
	Bio        template.HTML
	Code       int64
}

type UserList struct {
	Users []UserOne
}

type UserOne struct {
	Code int64  `gorm:"uniqueIndex NOT NULL"`
	URL  string `gorm:"type:text"`
	Bio  string `gorm:"type:text"`
}

func main() {
	router := gin.Default()
	router.Static("/uploads", "./uploads")
	router.Static("/statics", "./statics")
	router.GET("/", homePage)
	router.POST("/upload", uploadHandler)

	// 定义路由和处理函数
	router.GET("/tmp", func(c *gin.Context) {
		t, _ := template.ParseFiles("header.html", "footer.html")
		err := t.Execute(c.Writer, map[string]string{"Title": "My titleXXXXXXXX", "Body": "Hi this is my body"})
		if err != nil {
			panic(err)
		}
	})

	router.GET("/user/list/:page", func(c *gin.Context) {
		var users []User
		page := cast.ToInt(c.Param("page")) - 1
		if page < 0 {
			page = 0
		}
		limit := 20
		err := db.Where("deleted = ?", 0).Order("id").Limit(limit).Offset(page * limit).Find(&users).Error
		if err != nil {
			c.String(http.StatusNotFound, "User List Empty")
			return
		}

		t, _ := template.ParseFiles("userlist.html")

		var userones []UserOne
		for _, one := range users {
			urls := strings.Split(one.AvatarURLs, ",")
			url := ""
			if len(urls) > 0 {
				url = urls[0]
			}
			userones = append(userones, UserOne{Code: one.Code, URL: url})
		}
		ones := UserList{Users: userones}
		// fmt.Printf("ones : %+v", ones)
		err = t.Execute(c.Writer, ones)
		if err != nil {
			panic(err)
		}
	})

	router.GET("/add", func(c *gin.Context) {
		t, _ := template.ParseFiles("uploadpage.html")
		err := t.Execute(c.Writer, nil)
		if err != nil {
			panic(err)
		}
	})

	router.GET("/user/:code", func(c *gin.Context) {
		var user User
		code := c.Param("code")
		result := db.Where("code = ?", code).Where("deleted = ?", 0).First(&user)
		if result.Error == gorm.ErrRecordNotFound {
			c.String(http.StatusNotFound, "User not found")
			return
		}
		t, _ := template.ParseFiles("userprofile.html")
		profiles := Profiles{
			AvatarURLs: strings.Split(user.AvatarURLs, ","),
			Bio:        template.HTML(user.Bio),
			Code:       user.Code,
		}
		err := t.Execute(c.Writer, profiles)
		if err != nil {
			panic(err)
		}
	})

	log.Fatal(router.Run(":80"))
}

func homePage(c *gin.Context) {
	c.String(http.StatusOK, "欢迎使用我的应用")
}

func uploadHandler(c *gin.Context) {
	// 解析表单
	bio := c.PostForm("bio")
	code := cast.ToInt64(c.PostForm("code"))
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, "Error parsing multipart form")
		return
	}
	files := form.File["images[]"]
	for _, file := range files {
		// 处理文件，例如保存到服务器
		filename := file.Filename
		// 可以添加更多的文件处理逻辑，如保存文件等
		fmt.Println("Received file:", filename)
	}
	var urls []string
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			c.String(http.StatusInternalServerError, "Error opening file")
			return
		}
		defer file.Close()
		// 生成新的文件名
		ext := filepath.Ext(fileHeader.Filename)
		newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

		// 保存文件
		filePath, err := processFile(file, newFileName)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error saving file")
			return
		}

		// 添加文件路径到 URLs 列表
		urls = append(urls, filePath)
		fmt.Println("Received file:", fileHeader.Filename)
	}

	imgStr := strings.Join(urls, ",")
	user := User{Code: code, AvatarURLs: imgStr, Bio: bio, Deleted: 0}

	// 将数据保存到数据库
	result := db.Create(&user)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Error saving to database")
		return
	}

	fmt.Fprintf(c.Writer, "Upload successful")
}

// processFile 处理并保存上传的文件
func processFile(file multipart.File, newFileName string) (string, error) {
	err := os.MkdirAll(uploadPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(uploadPath, newFileName)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
