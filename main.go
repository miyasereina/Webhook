//
// main.go
// Copyright (C) 2020 forseason <me@forseason.vip>
//
// Distributed under terms of the MIT license.
//

package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"

	"io/ioutil"
	"os"
	"os/exec"
)

var Configs []RepositoryConfig

func main() {
	loadConfig()
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	apiv1 := r.Group("/api/v1")
	{
		apiv1.POST("webhook", webhook)

	}
	r.Run(":8000")
}

func webhook(c *gin.Context) {

	message,err := c.GetRawData()
	if err != nil{
		fmt.Println(err.Error())
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(message))


	var form WebhookForm
	if c.ShouldBind(&form) != nil {

		c.JSON(400, gin.H{"message": "invalid parameters"})
		return
	}
	form.Signature=c.GetHeader("X-Hub-Signature")
	found := false
	for _, config := range Configs {
		signature:="sha256="+HmacSha256(string(message),config.Secret)
		fmt.Print(signature)

		if signature == form.Signature && config.FullName == form.Repository.FullName && config.Ref == form.Ref {
			cmd := exec.Command("bash", "-c", "cd "+config.Dir+" && git pull")
			if stdout, err := cmd.CombinedOutput(); err != nil {
				c.JSON(500, gin.H{"message": stdout})
				return
			}
			if config.Exec != "" {
				cmd = exec.Command("bash", "-c", config.Exec)
				if stdout, err := cmd.CombinedOutput(); err != nil {
					c.JSON(500, gin.H{"message": stdout})
					return
				}
			}
			found = true
		}
	}
	if !found {
		c.JSON(404, gin.H{"message": "repository full_name not found in config.json"})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

func loadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var rawConfigs []RepositoryConfig
	if decoder.Decode(&rawConfigs) != nil {
		fmt.Println("cannot decode config.json")

		}

	for _, config := range rawConfigs {
		if config.Secret == "" || config.Dir == "" || config.FullName == "" || config.Ref == "" {
			fmt.Println("wrong format with config.json")

			}
		}

	Configs = rawConfigs

}

func HmacSha256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	//	fmt.Println(h.Sum(nil))
	//sha := hex.EncodeToString(h.Sum(nil))
	//	fmt.Println(sha)

	return hex.EncodeToString(h.Sum(nil))
	//return base64.StdEncoding.EncodeToString([]byte(nil))
}



type WebhookForm struct {
	Signature    string                `json:"X-Hub-Signature"`
	Ref        string                `json:"ref" binding:"required"`
	Repository WebhookRepositoryForm `json:"repository" binding:"required"`
}

type WebhookRepositoryForm struct {
	FullName string `json:"full_name" binding:"required"`
}

type RepositoryConfig struct {
	Secret   string `json:"secret" binding:"required"`
	Ref      string `json:"ref" binding:"required"`
	Dir      string `json:"dir" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
	Exec     string `json:"exec"`
}
