package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var router *gin.Engine

func TestSetupMongoDB(t *testing.T) {
	client, err := setupMongoDB()
	defer client.Disconnect(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestSetupRedis(t *testing.T) {
	client, err := setupRedis()
	defer client.Close()

	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestMain(m *testing.M) {
	router = setupRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()

	exitCode := m.Run()

	os.Exit(exitCode)
}

func TestHTTPHandlers(t *testing.T) {

	req := httptest.NewRequest("GET", "/videos", nil)
	req.Header.Add("X-Test-Request", "true")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Add more assertions as needed
}
