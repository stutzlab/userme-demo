package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *HTTPServer) setupUserTODO() {
	h.router.POST("/user/:email/todo", createTODO())
}

func createTODO() func(*gin.Context) {
	return func(c *gin.Context) {
		email := strings.ToLower(c.Param("email"))
		logrus.Debugf("createTODO email=%s", email)

		scope, _ := c.Get("scope")
		logrus.Debugf("JWT scope = %s", scope)

		sub, _ := c.Get("sub")
		logrus.Debugf("JWT sub = %s", sub)

		m := make(map[string]string)
		err := c.BindJSON(&m)
		// data, _ := ioutil.ReadAll(c.Request.Body)
		// err := json.Unmarshal(data, &m)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("Couldn't parse body contents. err=%s", err)})
			return
		}

		title, exists := m["title"]
		if !exists {
			c.JSON(400, gin.H{"message": "'title' field is required"})
			return
		}

		todo := TODO{
			Email: email,
			Title: title,
		}

		err = db.Create(&todo).Error
		if err != nil {
			logrus.Warnf("Error creating TODO. err=%s", err)
			c.JSON(500, gin.H{"message": "Server error"})
			return
		}

		logrus.Infof("createTODO email=%s title=%s was successful", email, title)
		c.JSON(201, gin.H{"message": "TODO created successfully"})
		return
	}
}
