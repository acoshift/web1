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
}

func getUser(id int) (*User, error) {
	var u User
	err := db.QueryRow(`
		select
			id, username
		from users
		where id = ?
	`, id).Scan(&u.ID, &u.Username)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func getUsers() ([]*User, error) {
	rows, err := db.Query(`
		select
			id, username
		from users
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	us := make([]*User, 0)
	for rows.Next() {
		var u User
		err = rows.Scan(&u.ID, &u.Username)
		if err != nil {
			return nil, err
		}
		us = append(us, &u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return us, nil
}

func index(c *gin.Context) {
	sess := sessions.Default(c)

	userID, _ := sess.Get("userId").(int)
	u, _ := getUser(userID)

	data := map[string]interface{}{
		"User": u,
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
