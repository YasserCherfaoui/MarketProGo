package file

import (
	"io"
	"net/http"

	"github.com/YasserCherfaoui/MarketProGo/cfg"
	"github.com/gin-gonic/gin"
)

func ProxyFilePreview(c *gin.Context) {
	fileId := c.Param("fileId")
	if fileId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileId is required"})
		return
	}

	conf, err := cfg.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load config"})
		return
	}
	// Build the Appwrite /view URL for the file
	viewURL := conf.AppwriteEndpoint + "/storage/buckets/" + conf.AppwriteBucketId + "/files/" + fileId + "/view?project=" + conf.AppwriteProject

	req, err := http.NewRequest("GET", viewURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request to appwrite"})
		return
	}
	req.Header.Set("X-Appwrite-Project", conf.AppwriteProject)
	req.Header.Set("X-Appwrite-Key", conf.AppwriteKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch file from appwrite"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, "application/json", body)
		return
	}

	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	c.Status(http.StatusOK)
	io.Copy(c.Writer, resp.Body)
}
