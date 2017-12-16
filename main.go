package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var (
	tmplIndex  = loadTemplate("template/index.html", "template/layout.html")
	tmplSignUp = loadTemplate("template/signup.html", "template/layout.html")
	tmplSignIn = loadTemplate("template/signin.html", "template/layout.html")
)

func loadTemplate(filename ...string) *template.Template {
	t := template.New("")
	t = template.Must(t.ParseFiles(filename...))
	t = t.Lookup("layout")
	return t
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("mysql", "root@tcp(localhost:3306)/web1")
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.GET("/", index)
	r.GET("/signup", signUp)
	r.POST("/signup", postSignUp)
	r.GET("/signin", signIn)
	r.POST("/signin", postSignIn)
	r.Run(":4000")
}

// User type
type User struct {
	ID       int    `json:"id"`
	Username string `json:"name"`
	Password string `json:"-"`
	Status   string `json:"status,omitempty"`
}

func index(c *gin.Context) {
	tmplIndex.Execute(c.Writer, nil)
}

func signUp(c *gin.Context) {
	tmplSignUp.Execute(c.Writer, nil)
}

func postSignUp(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if utf8.RuneCountInString(username) < 4 {
		c.String(http.StatusBadRequest, "username required")
		return
	}
	if password == "" {
		c.String(http.StatusBadRequest, "password required")
		return
	}
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	_, err = db.Exec(`
		insert into users (
			username, password
		) values (
			?, ?
		)
	`, username, string(hashedPass))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/")
}

func signIn(c *gin.Context) {
	tmplSignIn.Execute(c.Writer, nil)
}

func postSignIn(c *gin.Context) {

}
