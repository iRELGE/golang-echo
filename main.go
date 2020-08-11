package main

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4/middleware"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

type (
	// User : just struct
	User struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}
	// CustomValidator : our validator
	CustomValidator struct {
		validator *validator.Validate
	}
)

//Validate : use validator package to validate data that it receive from client
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// Handler
func users(c echo.Context) (err error) {
	// create a new user from struct User
	u := new(User)
	//close body after finish whit request
	defer c.Request().Body.Close()
	//decoding data based on the Content-Type header.
	if err = c.Bind(u); err != nil {
		return
	}
	//validate data if its the right expression using validator
	if err = c.Validate(u); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error)
	}
	//return a responce json
	return c.JSON(http.StatusOK, u)
}
func first(c echo.Context) error {
	//return a simple responce string
	return c.String(http.StatusOK, "Hello, World!")
}
func cat(c echo.Context) error {
	//queryparam can be retreive it by key [NAME] example: path?catname=xname
	catName := c.QueryParam("catname")
	//param can be retrieve it also by name [key] example : path/yourParamValue
	name := c.Param("name")

	return c.JSON(http.StatusOK, fmt.Sprintf(name, catName))
}
func mainAdmin(c echo.Context) error {
	//retrieve a Authorizasion value request from header
	auth := c.Request().Header.Get(echo.HeaderAuthorization)
	//just a simple decoding of your basic aut hrader value of authorize
	b, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return err
	}
	//parse value to string
	cred := string(b)
	return c.JSON(http.StatusOK, cred)
}

// ServerHeader : create my own custom  that return a hander function
func ServerHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// set some value to header
		c.Response().Header().Set(echo.HeaderServer, "rabie/1.0")
		c.Response().Header().Set("fake header", "rabie/lol")
		//return a handler function
		return next(c)
	}
}

func main() {
	e := echo.New()
	//using use to spesify our custome middleware to e
	e.Use(ServerHeader)
	//create a group that whit by part of any middlewar specified
	g := e.Group("/admin")
	//custom our default middleware
	g.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `[${time_rfc3339}] ${status} ${method} ${host}${path} ${bytes_out} ${remote_ip}` + "\n",
		//using basec auth for group admin any function in group admin need an authentification
	}), middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		// Be careful to use constant time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(username), []byte("joe")) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(":secret:")) == 1 {
			return true, nil
		}
		return false, nil
	}))
	//real root localhost:port/admin/main
	g.GET("/main", mainAdmin)

	e.Validator = &CustomValidator{validator: validator.New()}
	e.GET("/", first)
	e.GET("/cat/:name", cat)
	e.POST("/users", users)
	e.Logger.Fatal(e.Start(":1323"))

}
