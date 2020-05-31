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
	h.router.GET("/user/:email/todo", listTODO())
}

func createTODO() func(*gin.Context) {
	return func(c *gin.Context) {
		email := strings.ToLower(c.Param("email"))
		logrus.Debugf("createTODO email=%s", email)

		err := verifySelfPermit(c, email)
		if err != nil {
			c.JSON(403, gin.H{"message": err})
			return
		}

		m := make(map[string]string)
		err = c.BindJSON(&m)
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

func listTODO() func(*gin.Context) {
	return func(c *gin.Context) {
		email := strings.ToLower(c.Param("email"))
		logrus.Debugf("listTODO email=%s", email)

		err := verifySelfPermit(c, email)
		if err != nil {
			c.JSON(403, gin.H{"message": err})
			return
		}

		todos := []TODO{}

		err = db.Where("email = ?", email).Find(&todos).Error
		if err != nil {
			logrus.Warnf("Error listing TODO. err=%s", err)
			c.JSON(500, gin.H{"message": "Server error"})
			return
		}

		logrus.Infof("listTODO email=%s was successful", email)
		c.JSON(200, todos)
		return
	}
}

func verifySelfPermit(c *gin.Context, resourceOwner string) error {
	scope, _ := c.Get("scope")
	logrus.Debugf("JWT scope = %s", scope)
	sub, _ := c.Get("sub")
	logrus.Debugf("JWT sub = %s", sub)
	if scope != "basic" || sub != resourceOwner {
		return fmt.Errorf("User %s not authorized to access resource from %s", sub, resourceOwner)
	}
	return nil
}
