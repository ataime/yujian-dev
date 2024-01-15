package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const uploadPath = "./uploads"

// User struct to hold user data
type User struct {
	Code       string
	AvatarURLs []string
	Bio        string
}

const (
	dbUser     = "root"
	dbPassword = "123456"
	dbName     = "yujian"
)

var db *sql.DB
var tmpluser *template.Template
var tmpladmin *template.Template

func init() {
	var err error

	// Initialize the database connection
	dataSourceName := dbUser + ":" + dbPassword + "@/" + dbName
	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the HTML template
	tmpluser = template.Must(template.ParseFiles("userprofile.html"))
	tmpladmin = template.Must(template.ParseFiles("uploadpage.html"))
}

func main() {

	// 设置静态文件服务
	fs := http.FileServer(http.Dir("./uploads"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	http.HandleFunc("/", homePage)
	http.HandleFunc("/user/", userPage)
	http.HandleFunc("/admin/", adminPage)
	http.HandleFunc("/upload", uploadHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to the User Profile App"))
}

func adminPage(w http.ResponseWriter, r *http.Request) {
	// Render the template with user data
	err := tmpladmin.Execute(w, nil)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func userPage(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Path[len("/user/"):]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Query the database for the user
	var user User
	var avatarURLs string
	row := db.QueryRow("SELECT code, avatar_urls, bio FROM users WHERE id = ?", userID)
	err = row.Scan(&user.Code, &avatarURLs, &user.Bio)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Split the avatar URLs
	user.AvatarURLs = strings.Split(avatarURLs, ",") // Assuming the URLs are comma-separated

	for i, url := range user.AvatarURLs {
		user.AvatarURLs[i] = fmt.Sprintf("http://localhost:8080/%s", strings.TrimSpace(url))
	}

	// Render the template with user data
	err = tmpluser.Execute(w, user)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// processFile 处理并保存上传的文件
func processFile(file multipart.File, newFileName string) (string, error) {
	// 确保上传目录存在
	err := os.MkdirAll(uploadPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	// 创建文件的路径
	filePath := filepath.Join(uploadPath, newFileName)

	// 创建新文件
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// 将上传的文件复制到新文件
	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Error parsing multipart form", http.StatusBadRequest)
		return
	}

	bio := r.FormValue("bio")
	fmt.Println(bio)

	code := r.FormValue("code")
	fmt.Println("code: ", code)

	pass := r.FormValue("pass")
	fmt.Println("pass: ", pass)

	if pass != "yujian" {
		http.Error(w, "Incorrect password", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["images[]"]
	var urls []string
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		// 使用时间戳生成新的文件名
		ext := filepath.Ext(fileHeader.Filename)
		newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

		// 保存文件
		filePath, err := processFile(file, newFileName)
		if err != nil {
			http.Error(w, "Error saving file.", http.StatusInternalServerError)
			return
		}

		// 将文件路径添加到 URLs 列表
		urls = append(urls, filePath)
		// 可以输出文件名来确认文件已收到
		fmt.Println("Received file:", fileHeader.Filename)
	}
	fmt.Println(urls)
	imgstr := strings.Join(urls, ",")

	rs, err := db.Exec("insert into users(code,avatar_urls,bio) values(?,?,?)", code, imgstr, bio)
	fmt.Println(err)

	rowCount, err := rs.RowsAffected()
	fmt.Println(err)

	fmt.Println("Affected rows: ", rowCount)

	fmt.Fprintf(w, "Upload successful")
}
