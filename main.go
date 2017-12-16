package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"unicode/utf8"

	"github.com/gin-contrib/sessions"
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
	store := sessions.NewCookieStore([]byte("supersecret"))
	r.Use(sessions.Sessions("s", store))
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
}

func getUser(id int) (*User, error) {
	return nil, nil
}

func index(c *gin.Context) {
	sess := sessions.Default(c)

	userID, _ := sess.Get("userId").(int)

	data := map[string]interface{}{
		"isSignIn": userID > 0,
	}
	tmplIndex.Execute(c.Writer, data)
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

	var (
		id         int
		hashedPass string
	)
	err := db.QueryRow(`
		select
			id, password
		from users
		where username = ?
	`, username).Scan(
		&id,
		&hashedPass,
	)
	if err == sql.ErrNoRows {
		c.String(http.StatusBadRequest, "wrong username or password")
		return
	}
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(hashedPass),
		[]byte(password),
	)
	if err != nil {
		c.String(http.StatusBadRequest, "wrong username or password")
		return
	}

	sess := sessions.Default(c)
	sess.Set("userId", id)
	sess.Save()
	c.Redirect(http.StatusSeeOther, "/")
}
